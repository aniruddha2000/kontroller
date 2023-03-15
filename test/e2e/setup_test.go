//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var (
	testenv         env.Environment
	kindClusterName string
	namespace       = "e2e-ns"
)

func TestMain(m *testing.M) {
	testenv = env.New()
	kindClusterName = envconf.RandomName("e2e-cluster", 16)

	testenv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
		envfuncs.CreateNamespace(namespace),
	)

	testenv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyKindCluster(kindClusterName),
	)

	os.Exit(testenv.Run(m))
}
