package kubeutil

import (
	"testing"
)

func TestGoodConfig(t *testing.T) {
	goodConfig := KubeConfig{
		ApiVersion:        "eventorchestrator/v1alpha1",
		Kind:              "KubeConfig",
		Kubectx:           "dev-kubectx",
		Environment:       "dev",
		Name:              "Good configuration",
		Namespace:         "argo-events",
		ManifestDirectory: "0001/EventSource",
	}

	if err := goodConfig.New(goodConfig); err != nil {
		t.Errorf("Error creating good config: %v", err)
	}
}

func TestBadConfig(t *testing.T) {
	badConfig := KubeConfig{
		ApiVersion:        "v2",
		Kind:              "KubeBar",
		Kubectx:           "",
		Environment:       "",
		Name:              "Bad configuration",
		Namespace:         "argo",
		ManifestDirectory: "/xxxxx/none",
	}

	if err := badConfig.New(badConfig); err == nil {
		t.Errorf("Bad config sould error: %v", err)
	}
}
