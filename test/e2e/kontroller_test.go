package e2e

import (
	"context"
	"os"
	"testing"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestKontroller(t *testing.T) {
	f1 := features.New("create kontroller").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			r, err := resources.New(config.Client().RESTConfig())
			if err != nil {
				t.Errorf("Failed to create resource: %v", err)
				t.Fail()
			}
			r.WithNamespace(namespace)

			err = decoder.DecodeEachFile(ctx, os.DirFS("./testdata"), "*.yaml",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace))
			if err != nil {
				t.Errorf("Failed to decode: %v", err)
				t.Fail()
			}

			svc := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "validation-kontroller",
					Namespace: "default",
					Labels: map[string]string{
						"app": "validation-kontroller",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port:       443,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromInt(8443),
						},
					},
					Selector: map[string]string{
						"app": "validation-kontroller",
					},
				},
			}
			err = r.Create(ctx, &svc)
			if err != nil {
				t.Errorf("Failed to create Service: %v", err)
				t.Fail()
			}

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			r, err := resources.New(config.Client().RESTConfig())
			if err != nil {
				t.Errorf("Failed to create resource: %v", err)
				t.Fail()
			}
			r.WithNamespace(namespace)

			err = decoder.DecodeEachFile(ctx, os.DirFS("./testdata/admission"), "*.yaml",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace))
			if err != nil {
				t.Errorf("Failed to decode: %v", err)
				t.Fail()
			}
			return ctx
		}).
		Assess("Check for kontroller", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			r, err := resources.New(config.Client().RESTConfig())
			if err != nil {
				t.Errorf("Failed to create resource: %v", err)
				t.Fail()
			}
			r.WithNamespace(namespace)

			deploy := v1.Deployment{}
			err = r.Get(ctx, "validation-kontroller", namespace, &deploy)
			if err != nil {
				t.Errorf("Failed to get Deployment: %v", err)
				t.Fail()
			}

			role := rbacv1.Role{}
			err = r.Get(ctx, "kontroller-role", namespace, &role)
			if err != nil {
				t.Errorf("Failed to get Role: %v", err)
				t.Fail()
			}

			rb := rbacv1.RoleBinding{}
			err = r.Get(ctx, "kontroller-rb", namespace, &rb)
			if err != nil {
				t.Errorf("Failed to get RoleBinding: %v", err)
				t.Fail()
			}

			sa := corev1.ServiceAccount{}
			err = r.Get(ctx, "kontroller-sa", namespace, &sa)
			if err != nil {
				t.Errorf("Failed to get ServiceAccount: %v", err)
				t.Fail()
			}

			secret := corev1.Secret{}
			err = r.Get(ctx, "certs", namespace, &secret)
			if err != nil {
				t.Errorf("Failed to get Secret: %v", err)
				t.Fail()
			}

			service := corev1.Service{}
			err = r.Get(ctx, "validation-kontroller", "default", &service)
			if err != nil {
				t.Errorf("Failed to get Service: %v", err)
				t.Fail()
			}

			return ctx
		}).
		Assess("call webhook by creating pod", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			r, err := resources.New(config.Client().RESTConfig())
			if err != nil {
				t.Errorf("Failed to create resource: %v", err)
				t.Fail()
			}
			r.WithNamespace(namespace)

			err = decoder.DecodeEachFile(ctx, os.DirFS("./testdata/pod"), "*.yaml",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace))
			if err != nil {
				t.Errorf("Failed to decode: %v", err)
				t.Fail()
			}

			pod := corev1.Pod{}
			err = r.Get(ctx, "webserver", namespace, &pod)
			if err != nil {
				t.Errorf("Failed to get Pod: %v", err)
				t.Fail()
			}

			return ctx
		}).Feature()

	testenv.Test(t, f1)
}
