package kubeutil

import (
	"context"
	"testing"
)

/*
func TestKubeUtil_ExecWithContext(t *testing.T) {

	type fields struct {
		_startTime   time.Time
		_endTime     time.Time
		_command     string
		_manifestRaw []byte
		_manifest    map[string]interface{}
		_result      string
		_error       string
		_user        KubeUser
		_config      *KubeConfig
	}
	type args struct {
		ctx      context.Context
		conf     *KubeConfig
		user     KubeUser
		cmd      string
		manifest []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &KubeUtil{
				_startTime:   tt.fields._startTime,
				_endTime:     tt.fields._endTime,
				_command:     tt.fields._command,
				_manifestRaw: tt.fields._manifestRaw,
				_manifest:    tt.fields._manifest,
				_result:      tt.fields._result,
				_error:       tt.fields._error,
				_user:        tt.fields._user,
				_config:      tt.fields._config,
			}
			if got := k.ExecWithContext(tt.args.ctx, tt.args.conf, tt.args.user, tt.args.cmd, tt.args.manifest); got != tt.want {
				t.Errorf("KubeUtil.ExecWithContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubeUtil_init(t *testing.T) {
	type fields struct {
		_startTime   time.Time
		_endTime     time.Time
		_command     string
		_manifestRaw []byte
		_manifest    map[string]interface{}
		_result      string
		_error       string
		_user        KubeUser
		_config      *KubeConfig
	}
	type args struct {
		_user    KubeUser
		conf     *KubeConfig
		cmd      string
		manifest []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &KubeUtil{
				_startTime:   tt.fields._startTime,
				_endTime:     tt.fields._endTime,
				_command:     tt.fields._command,
				_manifestRaw: tt.fields._manifestRaw,
				_manifest:    tt.fields._manifest,
				_result:      tt.fields._result,
				_error:       tt.fields._error,
				_user:        tt.fields._user,
				_config:      tt.fields._config,
			}
			k.init(tt.args._user, tt.args.conf, tt.args.cmd, tt.args.manifest)
		})
	}
}

*/
func TestExecWithContext(t *testing.T) {
	var testCommand KubeUtil
	testUser := KubeUser{
		CustomerID:         1,
		UserID:             "test",
		Kind:               "KubeUser",
		AuthorizationToken: "#########",
		ReferenceID:        "123",
	}
	testManifest := []byte(`{"kind":"Deployment","apiVersion":"apps/v1","metadata":{"name":"test-deployment","namespace":"argo-events"},"spec":{"replicas":1,"selector":{"matchLabels":{"app":"test-deployment"}},"template":{"metadata":{"labels":{"app":"test-deployment"}},"spec":{"containers":[{"name":"test-container","image":"test-image","ports":[{"containerPort":8080}]}]}}}}`)

	testConf := &KubeConfig{
		ApiVersion:        "eventorchestrator/v1",
		Kind:              "KubeConfig",
		Kubectx:           "microk8s",
		Name:              "test-config",
		Namespace:         "argo-events",
		Environment:       "test-environment",
		ManifestDirectory: "1/Workflows",
	}

	ctx := context.Background()

	testCommand.init(testUser, testConf, "create", testManifest, "test-manifest")

	testCommand.ExecWithContext(ctx, testConf, testUser, "create", testManifest, "test-manifest")
}
