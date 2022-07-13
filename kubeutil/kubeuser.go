package kubeutil

import (
	"fmt"
	"reflect"
)

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// KubeUser
//  Is used to label manifests and authenticate requests
type KubeUser struct {
	// CustomerID - required
	CustomerID int

	// UserID requesting sync
	UserID string

	// Kind being synced
	Kind string

	// AuthorizationToken - Optional
	AuthorizationToken string

	// ReferenceID - user assigned ID for debugging
	ReferenceID string
}

func (ku *KubeUser) New(customer int, user, id string) bool {
	ku.CustomerID = customer
	ku.UserID = user
	ku.Kind = "KubeUser"
	ku.AuthorizationToken = ""
	ku.ReferenceID = id
	return ku.IsValid()
}

func (ku *KubeUser) IsValid() bool {
	return ku.CustomerID != 0 && ku.Kind != ""
}

func (ku *KubeUser) GenerateLables() []Label {
	var labels []Label

	v := reflect.ValueOf(*ku)

	for i := 0; i < v.NumField(); i++ {
		l := Label{}
		fieldValue := v.Field(i)
		fieldType := v.Type().Field(i)
		fieldName := fieldType.Name
		l.Value = fmt.Sprintf("%v", fieldValue.Interface())
		l.Key = string(fieldName)
		labels = append(labels, l)
	}
	return labels
}
