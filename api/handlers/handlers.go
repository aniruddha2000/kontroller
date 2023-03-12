package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	admv1beta1 "k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type Handler struct {
	Codec serializer.CodecFactory
}

func NewHandler() *Handler {
	return &Handler{
		Codec: serializer.NewCodecFactory(runtime.NewScheme()),
	}
}

func (h *Handler) KlusterValidationHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("KlusterValidationHandler is called")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Reading Requests Body: %v", err)
	}

	gvk := admv1beta1.SchemeGroupVersion.WithKind("AdmissionReview")
	var admissionReview admv1beta1.AdmissionReview
	_, _, err = h.Codec.UniversalDeserializer().Decode(body, &gvk, &admissionReview)
	if err != nil {
		log.Errorf("Error decode: %v, converting requests body to admission review type", err)
	}

	var pod v1.Pod
	gvk = v1.SchemeGroupVersion.WithKind("Pod")
	_, _, err = h.Codec.UniversalDeserializer().Decode(admissionReview.Request.Object.Raw, &gvk, &pod)
	if err != nil {
		log.Errorf("Error decode: %v, converting requests body to pod type", err)
	}

	response := admv1beta1.AdmissionResponse{}
	allow := validatePod(pod.Spec)
	if !allow {
		log.Debug("Inside failure")
		response = admv1beta1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: allow,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Pod image name: %s has latest in it", pod.Spec.Containers[0].Image),
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
		log.Errorf("Converting response to byte: %v", err)
	}

	_, err = w.Write(res)
	if err != nil {
		log.Errorf("Writing response to ResponseWritter: %v", err)
	}
}

func validatePod(podSpec v1.PodSpec) bool {
	for _, container := range podSpec.Containers {
		image := strings.Split(container.Image, ":")
		if image[1] == "latest" {
			log.Debug(image)
			return false
		}
	}
	return true
}
