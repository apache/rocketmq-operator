package controller

import (
	"github.com/apache/rocketmq-operator/pkg/controller/console"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, console.Add)
}
