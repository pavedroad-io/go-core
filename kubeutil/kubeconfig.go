package kubeutil

import (
	"errors"
	"strconv"
	"strings"
)

var _kinds = []string{"KubeConfig"}

var _versions = []string{"eventorchestrator/v1alpha1"}

var _resources = []string{"deployment", "service",
	"ingress", "secret", "configmap", "persistentvolume", "persistentvolumeclaim", "namespace", "Secret", "EventSource", "Sensor", "Workflow"}

type KubeConfig struct {
	ApiVersion string `json:"apiVersion"`

	Kind string `json:"kind"`

	Kubectx string `json:"kubectx"`

	Namespace string `json:"namespace"`

	Environment string `json:"environment"`

	Name string `json:"name"`

	ManifestDirectory string `json:"manifestDirectory"`
}

func (k *KubeConfig) New(conf KubeConfig) error {
	if !k.SupportedVersion(conf.ApiVersion) {
		return errors.New("Unsupported api version: " + conf.ApiVersion)
	}
	k.ApiVersion = conf.ApiVersion

	if !k.SupportedKind(conf.Kind) {
		return errors.New("Unsupported kind: " + conf.Kind)
	}
	k.Kind = conf.Kind

	if !k.ValidContext(conf.Kubectx) {
		return errors.New("Unsupported kubectx: " + conf.Kubectx)
	}

	mderror := k.ValidManifestDirectory(conf.ManifestDirectory)
	if mderror != nil {
		return mderror
	}
	k.ManifestDirectory = conf.ManifestDirectory

	if conf.Name == "" {
		return errors.New("conf.Name cannot be empty")
	}
	k.Name = conf.Name
	k.Namespace = conf.Namespace

	return nil
}

func (k *KubeConfig) ValidManifestDirectory(dir string) error {
	parts := strings.Split(dir, "/")
	if len(parts) != 2 {
		return errors.New(("Invalid manifest directory need customerNumber/resourceType: " + dir))
	}

	if _, err := strconv.Atoi(parts[0]); err != nil {
		return errors.New("customerNumber must be a int: " + parts[0])
	}

	if !k.SupportedResource(parts[1]) {
		return errors.New(("Unsupported resource type: " + parts[1]))
	}

	return nil
}
func (k *KubeConfig) ValidContext(ctx string) bool {
	//TODO implement

	return true
}

func (k *KubeConfig) GetNamespace() string {
	return k.Namespace
}

func (k *KubeConfig) SupportedVersion(version string) bool {
	for _, v := range _versions {
		if v == version {
			return true
		}
	}
	return false
}

func (k *KubeConfig) SupportedResource(resource string) bool {
	for _, r := range _resources {
		if r == resource {
			return true
		}
	}
	return false
}

func (k *KubeConfig) SupportedCommand(cmd string) bool {
	switch cmd {
	case kuAttach, kuApply, kuAutoscale, kuCreate, kuDelete, kuDescribe, kuExplain, kuExspose, kuGet, kuList, kuLogs, kuRollout, kuSet, kuScale, kuWatch:
		return true

	default:
		return false
	}
}

func (k *KubeConfig) SupportedKind(kind string) bool {
	for _, k := range _kinds {
		if k == kind {
			return true
		}
	}
	return false
}

func (k *KubeConfig) GetKubectx() string {
	return k.Kubectx
}

func (k *KubeConfig) GetEnvironment() string {
	return k.Environment
}

func (k *KubeConfig) GetName() string {
	return k.Name
} // GetName

func (k *KubeConfig) GetManifestDirectory() string {
	return k.ManifestDirectory
}
