package kubeutil

import (
	"testing"
)

func TestLables(t *testing.T) {

	testUser := KubeUser{
		CustomerID:         "1",
		UserID:             "test",
		Kind:               "KubeUser",
		AuthorizationToken: "#########",
		ReferenceID:        "123",
	}

	lables := testUser.GenerateLables()
	if len(lables) != 4 {
		t.Errorf("Expected 4 labels, got %d", len(lables))
	}
}
