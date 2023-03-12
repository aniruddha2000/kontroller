package handlers

import (
	"fmt"
	"io"
	admv1beta1 "k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"net/http"
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
	fmt.Println("KlusterValidationHandler is called")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error %s, reading the body", err.Error())
	}

	gvk := admv1beta1.SchemeGroupVersion.WithKind("AdmissionReview")
	var admissionReview admv1beta1.AdmissionReview

	_, _, err = h.Codec.UniversalDeserializer().Decode(body, &gvk, &admissionReview)
	if err != nil {
		fmt.Printf("Error decode: %s, converting requests body to admission review type", err.Error())
	}

	var pod v1.Pod
	gvk = v1.SchemeGroupVersion.WithKind("Pod")
	_, _, err = h.Codec.UniversalDeserializer().Decode(admissionReview.Request.Object.Raw, &gvk, &pod)
	if err != nil {
		fmt.Printf("Error decode: %s, converting requests body to pod type", err.Error())
	}

	fmt.Printf("Pod resource: %v", pod)
}
