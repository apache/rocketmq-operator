package controller

import (
	"github.com/operator-sdk-samples/rocketmq-operator/pkg/controller/metaservice"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, metaservice.Add)
}
