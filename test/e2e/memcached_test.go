// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	apis "github.com/apache/rocketmq-operator/pkg/apis"
	operator "github.com/apache/rocketmq-operator/pkg/apis/rocketmq/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestBroker(t *testing.T) {
	brokerList := &operator.BrokerList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, brokerList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("broker-group", func(t *testing.T) {
		t.Run("Cluster", BrokerCluster)
		t.Run("Cluster2", BrokerCluster)
	})
}

func brokerScaleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	// create broker custom resource
	exampleBroker := &operator.Broker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-broker",
			Namespace: namespace,
		},
		Spec: operator.BrokerSpec{
			Size: 3,
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), exampleBroker, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}
	// wait for example-broker to reach 3 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-broker", 3, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-broker", Namespace: namespace}, exampleBroker)
	if err != nil {
		return err
	}
	exampleBroker.Spec.Size = 4
	err = f.Client.Update(goctx.TODO(), exampleBroker)
	if err != nil {
		return err
	}

	// wait for example-broker to reach 4 replicas
	return e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-broker", 4, retryInterval, timeout)
}

func BrokerCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for rocketmq-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "rocketmq-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = brokerScaleTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}
