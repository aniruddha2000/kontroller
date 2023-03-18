package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aniruddha2000/kontroller/api/handlers"
	"github.com/mattbaird/jsonpatch"
	"github.com/stretchr/testify/assert"
	admv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

			// Create request body
			jsonBytes, admissionReview, err := getAdmissionReviewObject(t, test.podRequestsObject, uid)
			if err != nil {
				t.Error(err)
			}

			// Make requests
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(jsonBytes))
			w := httptest.NewRecorder()
			h.PodValidationHandler(w, r)

			// Read response
			got, err := io.ReadAll(w.Body)
			if err != nil {
				t.Error(err)
			}

			err = json.Unmarshal(got, admissionReview)
			if err != nil {
				t.Error(err)
			}

			assert.Equalf(t, test.response, admissionReview.Response, "Want %v, got %v", test.response, admissionReview.Response)
		})
	}
}

func TestPodMutationHandler(t *testing.T) {
	uid := uuid.NewUUID()
	jsonPatchType := admv1beta1.PatchTypeJSONPatch

	tests := []struct {
		name              string
		podRequestsObject corev1.Pod
		response          *admv1beta1.AdmissionResponse
	}{
		{
			name: "succesfull pod mutation requests",
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
				UID:       uid,
				Allowed:   true,
				PatchType: &jsonPatchType,
				Patch:     nil,
			},
		},
		{
			name: "unsuccesfull pod mutation requests",
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
				UID:       uid,
				Allowed:   true,
				PatchType: &jsonPatchType,
				Patch:     nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := handlers.NewHandler()

			// Create request body
			jsonBytes, admissionReview, err := getAdmissionReviewObject(t, test.podRequestsObject, uid)
			if err != nil {
				t.Error(err)
			}

			// Create patch only if annotations not found
			if test.podRequestsObject.Annotations == nil || test.podRequestsObject.Annotations["validated-by"] != "custom webhook" {
				test.response.Patch, err = createJSONPatch(t, admissionReview.Request, test.podRequestsObject)
				if err != nil {
					t.Error(err)
				}
			}

			// Make request
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(jsonBytes))
			w := httptest.NewRecorder()
			h.PodMutationHandler(w, r)

			// Read response
			got, err := io.ReadAll(w.Body)
			if err != nil {
				t.Error(err)
			}

			err = json.Unmarshal(got, admissionReview)
			if err != nil {
				t.Error(err)
			}

			if !assert.Equal(t, test.response, admissionReview.Response) {
				t.Fatalf("Want %v, got %v", test.response.Patch, admissionReview.Response.Patch)
			}
		})
	}
}

func createJSONPatch(t *testing.T, adm *admv1beta1.AdmissionRequest, pod corev1.Pod) ([]byte, error) {
	t.Helper()

	newPod := pod.DeepCopy()
	if newPod.Annotations == nil {
		newPod.Annotations = make(map[string]string)
	}
	newPod.Annotations["validated-by"] = "custom webhook"

	newPodRaw, err := json.Marshal(newPod)
	if err != nil {
		return []byte{}, err
	}

	jsonPatch, err := jsonpatch.CreatePatch(adm.Object.Raw, newPodRaw)
	if err != nil {
		return []byte{}, err
	}

	patch, err := json.Marshal(jsonPatch)
	if err != nil {
		return []byte{}, err
	}

	return patch, nil
}

func getAdmissionReviewObject(t *testing.T, pod corev1.Pod, uid types.UID) ([]byte, *admv1beta1.AdmissionReview, error) {
	t.Helper()

	jsonBytes, err := json.Marshal(pod)
	if err != nil {
		return []byte{}, &admv1beta1.AdmissionReview{}, err
	}

	admissionReview := &admv1beta1.AdmissionReview{
		Request: &admv1beta1.AdmissionRequest{
			UID: uid,
			Object: runtime.RawExtension{
				Raw: jsonBytes,
			},
		},
	}

	jsonBytes, err = json.Marshal(admissionReview)
	if err != nil {
		return []byte{}, &admv1beta1.AdmissionReview{}, err
	}

	return jsonBytes, admissionReview, nil
}
