/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package topictransfer

import (
	"context"
	"os/exec"
	"strings"
	"time"

	rocketmqv1alpha1 "github.com/apache/rocketmq-operator/pkg/apis/rocketmq/v1alpha1"
	cons "github.com/apache/rocketmq-operator/pkg/constants"
	"github.com/apache/rocketmq-operator/pkg/share"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_topictransfer")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new TopicTransfer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTopicTransfer{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("topictransfer-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource TopicTransfer
	err = c.Watch(&source.Kind{Type: &rocketmqv1alpha1.TopicTransfer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner TopicTransfer
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rocketmqv1alpha1.TopicTransfer{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileTopicTransfer implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileTopicTransfer{}

// ReconcileTopicTransfer reconciles a TopicTransfer object
type ReconcileTopicTransfer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a TopicTransfer object and makes changes based on the state read
// and what is in the TopicTransfer.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTopicTransfer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling TopicTransfer")

	// Fetch the TopicTransfer topicTransfer
	topicTransfer := &rocketmqv1alpha1.TopicTransfer{}
	err := r.client.Get(context.TODO(), request.NamespacedName, topicTransfer)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// TODO: if ConsumerGroup could be decided by listing the topics

	// TODO: what if current name server is compromised
	// step1: add consumer group to target cluster
	nameServer := strings.Split(share.NameServersStr, ";")[0]
	consumerGroup := topicTransfer.Spec.ConsumerGroup
	targetCluster := topicTransfer.Spec.TargetCluster
	sourceCluster := topicTransfer.Spec.SourceCluster
	topic := topicTransfer.Spec.Topic

	addConsumerGroupToTargetCluster := buildAddConsumerGroupToClusterCommand(consumerGroup, targetCluster, nameServer)
	addConsumerGroupToTargetClusterCommand := strings.Join(addConsumerGroupToTargetCluster, " ")
	reqLogger.Info("AddConsumerGroupToTargetClusterCommand: " + addConsumerGroupToTargetClusterCommand)
	cmd := exec.Command("sh", addConsumerGroupToTargetClusterCommand)
	output, err := cmd.Output()
	if err != nil {
		reqLogger.Error(err, "Failed to add ConsumerGroup " + consumerGroup + " to TargetCluster " + targetCluster)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(3) * time.Second}, err
	}
	// isAddConsumerGroupToTargetClusterOutputValid(string(output))
	reqLogger.Info("Successfully add ConsumerGroup " + consumerGroup + " to TargetCluster " + targetCluster + " with output: " + string(output))

	// step2: add consumer group to target cluster
	addTopicToTargetCluster := buildAddTopicToClusterCommand(topic, targetCluster, nameServer)
	addTopicToTargetClusterCommand := strings.Join(addTopicToTargetCluster, " ")
	reqLogger.Info("addTopicToTargetClusterCommand: " + addTopicToTargetClusterCommand)
	cmd = exec.Command("sh", addTopicToTargetClusterCommand)
	output, err = cmd.Output()
	if err != nil {
		reqLogger.Error(err, "Failed to add Topic " + topic + " to TargetCluster " + targetCluster)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(3) * time.Second}, err
	}
	// TODO: verify output of addTopicToTargetClusterCommand
	reqLogger.Info("Successfully add Topic " + topic + " to TargetCluster " + targetCluster + " with output: " + string(output))

	// step3: stop write in source cluster topic
	stopSourceClusterTopicWrite := buildStopClusterTopicWriteCommand(topic, sourceCluster, nameServer)
	stopSourceClusterTopicWriteCommand := strings.Join(stopSourceClusterTopicWrite, " ")
	reqLogger.Info("stopSourceClusterTopicWriteCommand: " + stopSourceClusterTopicWriteCommand)
	cmd = exec.Command("sh", stopSourceClusterTopicWriteCommand)
	output, err = cmd.Output()
	if err != nil {
		reqLogger.Error(err, "Failed to stop Topic " + topic + " write in SourceCluster " + sourceCluster)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(3) * time.Second}, err
	}
	// TODO: verify output of stopSourceClusterTopicWriteCommand
	reqLogger.Info("Successfully stop Topic " + topic + " write in SourceCluster " + sourceCluster + " with output: " + string(output))

	// step4: check source cluster unconsumed message
	for {
		checkConsumeProgress := buildCheckConsumeProgressCommand(consumerGroup, nameServer)
		checkConsumeProgressCommand := strings.Join(checkConsumeProgress, " ")
		reqLogger.Info("checkConsumeProgressCommand: " + checkConsumeProgressCommand)
		cmd = exec.Command("sh", checkConsumeProgressCommand)
		output, err = cmd.Output()
		if err != nil {
			reqLogger.Error(err, "Failed to check consumerGroup " + consumerGroup)
			return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(3) * time.Second}, err
		}
		var finished bool = isConsumeFinished(string(output))
		reqLogger.Info(" output: " + string(output))
		if finished {
			break
		}
		time.Sleep(time.Duration(5) * time.Second)
	}

	// step5: delete topic in source cluster
	deleteSourceClusterTopic := buildDeleteSourceClusterTopicCommand(topic, sourceCluster, nameServer)
	deleteSourceClusterTopicCommand := strings.Join(deleteSourceClusterTopic, " ")
	reqLogger.Info("deleteSourceClusterTopicCommand: " + deleteSourceClusterTopicCommand)
	cmd = exec.Command("sh", deleteSourceClusterTopicCommand)
	output, err = cmd.Output()
	if err != nil {
		reqLogger.Error(err, "Failed to delete Topic " + topic + " in SourceCluster " + sourceCluster)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(3) * time.Second}, err
	}
	// TODO: verify output
	reqLogger.Info("Successfully delete Topic " + topic + " in SourceCluster " + sourceCluster + " with output: " + string(output))

	// step6: delete consumer group in source cluster
	deleteConsumerGroup := buildDeleteConsumeGroupCommand(consumerGroup, sourceCluster, nameServer)
	deleteConsumerGroupCommand := strings.Join(deleteConsumerGroup, " ")
	reqLogger.Info("deleteConsumerGroupCommand: " + deleteConsumerGroupCommand)
	cmd = exec.Command("sh", deleteConsumerGroupCommand)
	output, err = cmd.Output()
	if err != nil {
		reqLogger.Error(err, "Failed to delete consumer group " + consumerGroup + " in SourceCluster " + sourceCluster)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(3) * time.Second}, err
	}
	// TODO: verify output
	reqLogger.Info("Successfully delete consumer group " + consumerGroup + " in SourceCluster " + sourceCluster + " with output: " + string(output))

	// step7: create retry topic
	createRetryTopic := buildAddRetryTopicToClusterCommand(consumerGroup, targetCluster, nameServer)
	createRetryTopicCommand := strings.Join(createRetryTopic, " ")
	reqLogger.Info("createRetryTopicCommand: " + createRetryTopicCommand)
	cmd = exec.Command("sh", createRetryTopicCommand)
	output, err = cmd.Output()
	if err != nil {
		reqLogger.Error(err, "Failed to create retry topic of consumer group " + consumerGroup + " in TargetCluster " + targetCluster)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(3) * time.Second}, err
	}
	// TODO: verify output
	reqLogger.Info("Successfully create retry topic of consumer group " + consumerGroup + " in TargetCluster " + targetCluster + " with output: " + string(output))

	reqLogger.Info("Topic " + topic + " has been transferred from " + sourceCluster + " to " + targetCluster)
	return reconcile.Result{}, nil
}

func buildAddRetryTopicToClusterCommand(consumerGroup string, cluster string, nameServer string) []string {
	cmdOpts := []string{
		cons.AdminToolDir,
		"updatetopic",
		"-t",
		"%RETRY%" + consumerGroup,
		"-c",
		cluster,
		"-r",
		"1",
		"-w",
		"1",
		"-p",
		"6",
		"-n",
		nameServer,
	}
	return cmdOpts
}

func isConsumeFinished(output string) bool {
	return true
}

func buildDeleteConsumeGroupCommand(consumerGroup string, cluster string, nameServer string) []string {
	cmdOpts := []string{
		cons.AdminToolDir,
		"deleteSubGroup",
		"-g",
		consumerGroup,
		"-c",
		cluster,
		"-n",
		nameServer,
	}
	return cmdOpts
}

func buildDeleteSourceClusterTopicCommand(topic string, sourceCluster string, nameServer string) []string {
	cmdOpts := []string{
		cons.AdminToolDir,
		"deletetopic",
		"-t",
		topic,
		"-c",
		sourceCluster,
		"-n",
		nameServer,
	}
	return cmdOpts
}

func isAddConsumerGroupToTargetClusterOutputValid(s string)  {

}

func buildCheckConsumeProgressCommand(consumerGroup string, nameServer string) []string {
	cmdOpts := []string{
		cons.AdminToolDir,
		"consumerprogress",
		"-g",
		consumerGroup,
		"-n",
		nameServer,
	}
	return cmdOpts
}

func buildStopClusterTopicWriteCommand(topic string, cluster string, nameServer string) []string {
	cmdOpts := []string{
		cons.AdminToolDir,
		"updatetopic",
		"-t",
		topic,
		"-c",
		cluster,
		"-p",
		"4",
		"-w",
		"0",
		"-n",
		nameServer,
	}
	return cmdOpts
}

func buildAddConsumerGroupToClusterCommand(consumerGroup string, cluster string, nameServer string) []string {
	cmdOpts := []string{
		cons.AdminToolDir,
		"updatesubgroup",
		"-g",
		consumerGroup,
		"-c",
		cluster,
		"-m",
		"true",
		"-d",
		"true",
		"-n",
		nameServer,
	}
	return cmdOpts
}

func buildAddTopicToClusterCommand(topic string, cluster string, nameServer string) []string {
	cmdOpts := []string{
		cons.AdminToolDir,
		"updatetopic",
		"-t",
		topic,
		"-c",
		cluster,
		"-n",
		nameServer,
	}
	return cmdOpts
}
