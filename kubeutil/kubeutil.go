package kubeutil

import (
	"context"
	"errors"
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
	_startTime   time.Time
	_endTime     time.Time
	_command     string
	_manifestRaw []byte
	_manifest    map[string]interface{}
	_fileName    string
	_result      string
	_error       string
	_user        KubeUser
	_config      *KubeConfig
	_ctx         context.Context
	_location    string
}

func (k *KubeUtil) ExecWithContext(
	ctx context.Context,
	conf *KubeConfig,
	user KubeUser,
	cmd string,
	manifest []byte,
	filename string) string {

	k._ctx = ctx
	if validConf := k._config.New(*conf); validConf != nil {
		k.respondWithError(errors.New("invalid configuration"))
	}

	if err := k.init(user, conf, cmd, manifest, filename); err != nil {
		k.respondWithError(err)
	}

	if err := k.checkAndSave(); err != nil {
		k.respondWithError(err)
	}

	if err := k.execute(); err != nil {
		k.respondWithError(err)
	}

	return (k._result)
}

func (k *KubeUtil) execute() error {
	var kubecmd = []string{}

	kubecmd = append(kubecmd, "--context")
	kubecmd = append(kubecmd, k._config.GetKubectx())

	kubecmd = append(kubecmd, "--namespace")
	kubecmd = append(kubecmd, k._config.GetNamespace())

	kubecmd = append(kubecmd, k._command)
	kubecmd = append(kubecmd, "-f")
	kubecmd = append(kubecmd, k._location)

	kubecmd = append(kubecmd, "-o")
	kubecmd = append(kubecmd, "json")

	debug := "kubectl " + strings.Join(kubecmd, " ")
	fmt.Println(debug)
	data, err := exec.Command("kubectl", kubecmd...).Output()
	if err != nil {
		fmt.Println("Error executing kubectl: ", err)
		return err
	}
	k._result = string(data)
	return nil
}

func (k *KubeUtil) respondWithError(err error) string {
	k._endTime = time.Now()
	fmt.Println(k._command, "failed in", k._endTime.Sub(k._startTime).String())
	return err.Error()
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
	f, err := os.Create(k._location)
	if err != nil {
		return err
	}

	if _, err := f.Write(k._manifestRaw); err != nil {
		return err
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

	// Parse the manifest
	err := yaml.Unmarshal([]byte(k._manifestRaw), &k._manifest)
	if err != nil {
		k._error = err.Error()
		return err
	}

	k._result = ""
	k._error = ""
	return nil
}
