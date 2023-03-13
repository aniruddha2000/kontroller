package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mattbaird/jsonpatch"
	log "github.com/sirupsen/logrus"
	admv1beta1 "k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
)

// Handler define the attribute for webhook handlers.
type Handler struct {
	Codec serializer.CodecFactory
}

// NewHandler returns a Handler.
func NewHandler() *Handler {
	return &Handler{
		Codec: serializer.NewCodecFactory(runtime.NewScheme()),
	}
}

func (h *Handler) PodValidationHandler(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"URL": r.URL.RawPath,
	}).Info("PodValidationHandler is called")

	pod, admissionReview, err := h.getRequestsObject(w, r)
	if err != nil {
		responsewriters.InternalError(w, r, err)
		log.Errorf("Reading Requests Body: %v", err)
	}

	response := admv1beta1.AdmissionResponse{}
	allow := validatePodObjectMeta(pod.ObjectMeta)
	if !allow {
		response = admv1beta1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: allow,
			Result: &metav1.Status{
				Status:  "Failure",
				Message: fmt.Sprintf("Pod metadata: %s has no desired annotation in it %s:%s", pod.ObjectMeta.Annotations, "validated-by", "custom webhook"),
				Reason:  metav1.StatusReasonInvalid,
			},
		}
	} else {
		response = admv1beta1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: allow,
		}
	}

	admissionReview.Response = &response
	res, err := json.Marshal(admissionReview)
	if err != nil {
		responsewriters.InternalError(w, r, err)
		log.Errorf("Converting response to byte: %v", err)
	}

	_, err = w.Write(res)
	if err != nil {
		responsewriters.InternalError(w, r, err)
		log.Errorf("Writing response to ResponseWritter: %v", err)
	}
}

func validatePodObjectMeta(objectMeta metav1.ObjectMeta) bool {
	if objectMeta.Annotations == nil {
		return false
	} else {
		annotation, ok := objectMeta.Annotations["validated-by"]
		if !ok {
			return false
		} else if annotation == "custom webhook" {
			return true
		} else {
			return false
		}
	}
}

func (h *Handler) PodMutationHandler(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"URL": r.URL.RawPath,
	}).Info("PodMutationHandler is called")

	pod, admissionReview, err := h.getRequestsObject(w, r)
	if err != nil {
		responsewriters.InternalError(w, r, err)
		log.Errorf("Reading Requests Body: %v", err)
	}

	newPod := pod.DeepCopy()
	allow := validatePodObjectMeta(newPod.ObjectMeta)
	if !allow {
		if newPod.ObjectMeta.Annotations == nil {
			newPod.ObjectMeta.Annotations = make(map[string]string)
		}
		newPod.ObjectMeta.Annotations["validated-by"] = "custom webhook"
	}

	res, err := json.Marshal(newPod)
	if err != nil {
		responsewriters.InternalError(w, r, err)
		log.Errorf("Converting response to byte: %v", err)
	}

	patch, err := jsonpatch.CreatePatch(admissionReview.Request.Object.Raw, res)
	if err != nil {
		responsewriters.InternalError(w, r, err)
		log.Errorf("Creating JSONPatch: %v", err)
	}

	patchRes, err := json.Marshal(patch)
	if err != nil {
		responsewriters.InternalError(w, r, err)
		log.Errorf("Converting patch response to byte: %v", err)
	}

	jsonPatchType := admv1beta1.PatchTypeJSONPatch
	admissionReview.Response = &admv1beta1.AdmissionResponse{
		UID:       admissionReview.Request.UID,
		Allowed:   true,
		PatchType: &jsonPatchType,
		Patch:     patchRes,
	}

	admissionReviewResponse, err := json.Marshal(admissionReview)
	if err != nil {
		responsewriters.InternalError(w, r, err)
		log.Errorf("Converting admission review response to byte: %v", err)
	}

	_, err = w.Write(admissionReviewResponse)
	if err != nil {
		responsewriters.InternalError(w, r, err)
		log.Errorf("Writing response to ResponseWritter: %v", err)
	}
}

func (h *Handler) getRequestsObject(w http.ResponseWriter, r *http.Request) (v1.Pod, admv1beta1.AdmissionReview, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return v1.Pod{}, admv1beta1.AdmissionReview{}, err
	}

	gvk := admv1beta1.SchemeGroupVersion.WithKind("AdmissionReview")
	var admissionReview admv1beta1.AdmissionReview
	_, _, err = h.Codec.UniversalDeserializer().Decode(body, &gvk, &admissionReview)
	if err != nil {
		return v1.Pod{}, admv1beta1.AdmissionReview{}, err
	}

	var pod v1.Pod
	gvk = v1.SchemeGroupVersion.WithKind("Pod")
	_, _, err = h.Codec.UniversalDeserializer().Decode(admissionReview.Request.Object.Raw, &gvk, &pod)
	if err != nil {
		return v1.Pod{}, admv1beta1.AdmissionReview{}, err
	}

	return pod, admissionReview, nil
}
