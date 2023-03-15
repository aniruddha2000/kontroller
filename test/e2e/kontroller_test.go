//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
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
				t.Errorf("Failed to create testdata/: %v", err)
				t.Fail()
			}

			err = decoder.DecodeEachFile(ctx, os.DirFS("./testdata/admission"), "*.yaml",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace))
			if err != nil {
				t.Errorf("Failed to create testdata/admission/: %v", err)
				t.Fail()
			}

			return ctx
		}).
		Assess("Check for deployment availability", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			client, err := config.NewClient()
			if err != nil {
				t.Errorf("Error getting client: %v", err)
				t.Fail()
			}

			dep := v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "validation-kontroller",
					Namespace: namespace,
				},
			}
			err = wait.For(conditions.New(client.Resources()).DeploymentConditionMatch(&dep, v1.DeploymentAvailable,
				corev1.ConditionTrue), wait.WithTimeout(1*time.Minute))
			if err != nil {
				t.Errorf("Deployment Availability: %v", err)
				t.Fail()
			}

			return ctx
		}).
		Assess("create pod", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
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
				t.Errorf("Failed to create testdata/pod/: %v", err)
				t.Fail()
			}

			return ctx
		}).
		Assess("check pod availability and annotations", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			client, err := config.NewClient()
			if err != nil {
				t.Errorf("Error getting client: %v", err)
				t.Fail()
			}

			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webserver",
					Namespace: namespace,
				},
			}
			err = wait.For(conditions.New(client.Resources()).PodRunning(&pod), wait.WithTimeout(5*time.Minute))
			if err != nil {
				t.Errorf("Pod Availability: %v", err)
				t.Fail()
			}

			err = wait.For(conditions.New(client.Resources()).ResourceMatch(&pod, func(object k8s.Object) bool {
				p := object.(*corev1.Pod)
				if p.ObjectMeta.Annotations["validated-by"] != "custom webhook" {
					return false
				}
				return true
			}), wait.WithTimeout(20*time.Second))
			if err != nil {
				t.Errorf("Annotations not found: %v", err)
				t.Fail()
			}

			return ctx
		}).
		WithTeardown("tearing down resources", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Errorf("Failed to create resource: %v", err)
				t.Fail()
			}
			r.WithNamespace(namespace)

			err = decoder.DecodeEachFile(ctx, os.DirFS("./testdata"), "*.yaml",
				decoder.DeleteHandler(r),
				decoder.MutateNamespace(namespace))
			if err != nil {
				t.Errorf("Failed to delete testdata/: %v", err)
				t.Fail()
			}

			err = decoder.DecodeEachFile(ctx, os.DirFS("./testdata/admission"), "*.yaml",
				decoder.DeleteHandler(r),
				decoder.MutateNamespace(namespace))
			if err != nil {
				t.Errorf("Failed to decode testdata/admission: %v", err)
				t.Fail()
			}

			err = decoder.DecodeEachFile(ctx, os.DirFS("./testdata/pod"), "*.yaml",
				decoder.DeleteHandler(r),
				decoder.MutateNamespace(namespace))
			if err != nil {
				t.Errorf("Failed to decode testdata/pod: %v", err)
				t.Fail()
			}

			return ctx
		}).Feature()

	testenv.Test(t, f1)
}
