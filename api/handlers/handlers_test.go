package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/aniruddha2000/kontroller/api/handlers"
	admv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func TestPodValidationHandler(t *testing.T) {
	uid := uuid.NewUUID()

	tests := []struct {
		name              string
		podRequestsObject corev1.Pod
		response          *admv1beta1.AdmissionResponse
	}{
		{
			name: "succesfull pod validation requests",
			podRequestsObject: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "webserver",
					Annotations: map[string]string{
						"validated-by": "custom webhook",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "webserver",
							Image: "ngnix:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
			response: &admv1beta1.AdmissionResponse{
				UID:     uid,
				Allowed: true,
			},
		},
		{
			name: "succesfull pod validation requests",
			podRequestsObject: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "webserver",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "webserver",
							Image: "ngnix:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
			response: &admv1beta1.AdmissionResponse{
				UID:     uid,
				Allowed: false,
				Result: &metav1.Status{
					Status:  "Failure",
					Message: fmt.Sprintf("Pod metadata: has no desired annotation in it %s:%s", "validated-by", "custom webhook"),
					Reason:  metav1.StatusReasonInvalid,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := handlers.NewHandler()

			jsonBytes, err := json.Marshal(test.podRequestsObject)
			if err != nil {
				t.Error(err)
			}

			admissionReview := admv1beta1.AdmissionReview{
				Request: &admv1beta1.AdmissionRequest{
					UID: uid,
					Object: runtime.RawExtension{
						Raw: jsonBytes,
					},
				},
			}

			jsonBytes, err = json.Marshal(admissionReview)
			if err != nil {
				t.Error(err)
			}

			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(jsonBytes))
			w := httptest.NewRecorder()

			h.PodValidationHandler(w, r)

			got, err := io.ReadAll(w.Body)
			if err != nil {
				t.Error(err)
			}

			responseAdmissionReview := admv1beta1.AdmissionReview{}
			err = json.Unmarshal(got, &responseAdmissionReview)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(responseAdmissionReview.Response, test.response) {
				t.Fatalf("Want %v, got %v", responseAdmissionReview.Response, test.response)
			}
		})
	}
}