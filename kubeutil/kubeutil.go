package kubeutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	kuApply     = "apply"
	kuAttach    = "attach"
	kuAutoscale = "autoscale"
	kuCreate    = "create"
	kuDelete    = "delete"
	kuDescribe  = "describe"
	kuExplain   = "explain"
	kuExspose   = "expose"
	kuGet       = "get"
	kuList      = "list"
	kuLogs      = "logs"
	kuRollout   = "rollout"
	kuSet       = "set"
	kuScale     = "scale"
	kuWatch     = "watch"
)

const (
	manifestLocation = "manifests/"
)

type KubeUtil struct {
	_startTime        time.Time
	_endTime          time.Time
	_command          string
	_manifestRaw      []byte
	_manifest         map[string]interface{}
	_fileName         string
	_result           string
	_error            string
	_user             KubeUser
	_config           *KubeConfig
	_ctx              context.Context
	_location         string
	_additionalLabels []Label
}

func (k *KubeUtil) ExecWithContext(
	ctx context.Context,
	conf *KubeConfig,
	user KubeUser,
	cmd string,
	manifest []byte,
	filename string) error {

	fmt.Println("CTX: ", ctx)
	fmt.Println("CONF: ", conf)
	k._ctx = ctx
	if validConf := k._config.New(*conf); validConf != nil {
		return k.respondWithError("Bad config", validConf)
	}

	if err := k.init(user, conf, cmd, manifest, filename); err != nil {
		return k.respondWithError("Failed to initialize", err)
	}

	if err := k.checkAndSave(); err != nil {
		return k.respondWithError("checkAndSave", err)
	}

	if err := k.execute(); err != nil {
		return k.respondWithError("execute", err)
	}

	return (nil)
}

func (k *KubeUtil) buildCommandOptions(cmd []string) []string {

	// Always add the context
	cmd = append(cmd, "--context")
	cmd = append(cmd, k._config.GetKubectx())

	// And the namespace
	cmd = append(cmd, "--namespace")
	cmd = append(cmd, k._config.GetNamespace())

	switch k._command {

	// Commands that use a manifest
	case kuApply, kuCreate, kuDelete:
		// Add the command
		cmd = append(cmd, k._command)

		cmd = append(cmd, "-f")
		cmd = append(cmd, k._location)

	// Commands that create a list of resource types
	case kuList:
		// Add the command
		cmd = append(cmd, kuGet)

		cmd = append(cmd, k._manifest["kind"].(string))
		list := fmt.Sprintf("-l CustomerID=%v", k._user.CustomerID)
		cmd = append(cmd, list)

	// Commands that use a name and resource
	case kuGet, kuDescribe, kuExplain, kuExspose, kuLogs, kuRollout, kuScale, kuWatch:
		// Add the command
		cmd = append(cmd, k._command)

		cmd = append(cmd, k._manifest["kind"].(string))
		cmd = append(cmd, k._manifest["metadata"].(map[interface{}]interface{})["name"].(string))
	}

	// Command that support a JSON response body
	switch k._command {
	case kuApply, kuCreate, kuDescribe, kuExplain, kuExspose, kuGet, kuList, kuLogs, kuRollout, kuScale, kuWatch:
		cmd = append(cmd, "-o")
		cmd = append(cmd, "yaml")

	}
	return cmd
}

func (k *KubeUtil) execute() error {
	var kubecmd = []string{}
	kubecmd = k.buildCommandOptions(kubecmd)

	debug := "kubectl " + strings.Join(kubecmd, " ")
	fmt.Println(debug)
	data, err := exec.Command("kubectl", kubecmd...).CombinedOutput()
	if err != nil {
		k._error = string(data)
		return err
	}
	k._result = string(data)
	return nil
}

func (k *KubeUtil) respondWithError(where string, err error) error {
	k._endTime = time.Now()
	fmt.Println(where, " : ", k._command, "failed in", k._endTime.Sub(k._startTime).String())
	return err
}

func (k *KubeUtil) checkAndSave() error {
	saveLocation, _ := filepath.Abs(filepath.Join(manifestLocation, k._config.GetManifestDirectory()))

	if _, err := os.Stat(saveLocation); os.IsNotExist(err) {
		os.MkdirAll(saveLocation, 0755)
	} else if err != nil {
		return err
	}

	// Save the manifest
	k._location = filepath.Join(saveLocation, k._fileName+".yaml")

	if _, err := os.Stat(k._location); os.IsNotExist(err) {
		f, err := os.Create(k._location)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.Write(k._manifestRaw); err != nil {
			return err
		}
	} else {
		f, err := os.OpenFile(k._location, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.Write(k._manifestRaw); err != nil {
			return err
		}
	}

	return nil
}

func (k *KubeUtil) init(user KubeUser, conf *KubeConfig, cmd string, manifest []byte, filename string) error {
	k._startTime = time.Now()
	k._command = cmd
	k._manifestRaw = manifest
	k._fileName = filename
	k._manifest = make(map[string]interface{})
	k._user = user
	k._config = conf
	k._additionalLabels = user.GenerateLables()

	// Parse the manifest
	err := yaml.Unmarshal([]byte(k._manifestRaw), &k._manifest)
	if err != nil {
		k._error = err.Error()
		return err
	}

	k.LabelManifest()

	data, err := yaml.Marshal(&k._manifest)

	if err != nil {
		k._error = err.Error()
		return err
	} else {
		k._manifestRaw = data
	}

	k._result = ""
	k._error = ""
	return nil
}

func (k *KubeUtil) LabelManifest() {
	// Add lbels to the manifest if missing
	_, ok := k._manifest["metadata"].(map[interface{}]interface{})["labels"]
	if !ok {
		k._manifest["metadata"].(map[interface{}]interface{})["labels"] = make(map[string]string)

	}

	for _, v := range k._additionalLabels {
		k._manifest["metadata"].(map[interface{}]interface{})["labels"].(map[string]string)[v.Key] = interface{}(v.Value).(string)
	}
}
