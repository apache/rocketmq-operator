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

// Package topictransfer contains the implementation of the TopicTransfer CRD reconcile function
package topictransfer

import (
	"context"
	"os/exec"
	"strconv"
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
var undo = false
var status = 0

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

	topic := topicTransfer.Spec.Topic
	targetCluster := topicTransfer.Spec.TargetCluster
	sourceCluster := topicTransfer.Spec.SourceCluster

	nameServer := strings.Split(share.NameServersStr, ";")[0]
	if len(nameServer) < cons.MinIpListLength {
		reqLogger.Info("There is no available name server now thus the topic transfer process is terminated.")
		// terminate the transfer process
		return reconcile.Result{}, nil
	}

	// ConsumerGroup could be decided by listing the topics
	consumerGroups := getConsumerGroupByTopic(topic, nameServer)
	if consumerGroups == nil {
		reqLogger.Info("There is no consumer group of topic " + topic)
	}

	if undo {
		// undo the operations for atomicity
		reqLogger.Info("Transfer topic " + topic + "  from " + sourceCluster + " to " + targetCluster + " failed, rolling back...")
		switch status {
		case 7:
			fallthrough
		case 6:
			undoDeleteConsumeGroup(consumerGroups, sourceCluster, nameServer)
			fallthrough
		case 5:
			undoDeleteTopic(topic, sourceCluster, nameServer)
			fallthrough
		case 4:
			fallthrough
		case 3:
			undoStopWrite(topic, sourceCluster, nameServer)
		default:
			// for user data safety, no special operations for other status
		}
	} else {
		// step1: add all consumer groups to target cluster
		status = 1
		for i, consumerGroup := range consumerGroups {
			log.Info("Processing consumer group" + consumerGroup + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(consumerGroups)))
			addConsumerGroupToTargetClusterCommand := buildAddConsumerGroupToClusterCommand(consumerGroup, targetCluster, nameServer)
			reqLogger.Info("AddConsumerGroupToTargetClusterCommand: " + addConsumerGroupToTargetClusterCommand)
			cmd := exec.Command(cons.BasicCommand, cons.AdminToolDir, addConsumerGroupToTargetClusterCommand)
			output, err := cmd.Output()
			// validate command output
			if err != nil || !isUpdateConsumerGroupSuccess(string(output)) {
				reqLogger.Error(err, "Failed to add ConsumerGroup "+consumerGroup+" to TargetCluster "+targetCluster+" with output: "+string(output))
				// terminate the transfer process
				undo = true
				return reconcile.Result{Requeue: true}, err
			}
			reqLogger.Info("Successfully add ConsumerGroup " + consumerGroup + " to TargetCluster " + targetCluster + " with output: " + string(output))
		}

		// step2: add consumer group to target cluster
		status = 2
		addTopicToTargetClusterCommand := buildAddTopicToClusterCommand(topic, targetCluster, nameServer)
		reqLogger.Info("addTopicToTargetClusterCommand: " + addTopicToTargetClusterCommand)
		cmd := exec.Command(cons.BasicCommand, cons.AdminToolDir, addTopicToTargetClusterCommand)
		output, err := cmd.Output()
		// validate command output
		if err != nil || !isUpdateTopicCommandSuccess(string(output)) {
			reqLogger.Error(err, "Failed to add Topic "+topic+" to TargetCluster "+targetCluster+" with output: "+string(output))
			// terminate the transfer process
			undo = true
			return reconcile.Result{Requeue: true}, err
		}
		reqLogger.Info("Successfully add Topic " + topic + " to TargetCluster " + targetCluster + " with output: " + string(output))

		// step3: stop write in source cluster topic
		status = 3
		stopSourceClusterTopicWriteCommand := buildStopClusterTopicWriteCommand(topic, sourceCluster, nameServer)
		reqLogger.Info("stopSourceClusterTopicWriteCommand: " + stopSourceClusterTopicWriteCommand)
		cmd = exec.Command(cons.BasicCommand, cons.AdminToolDir, stopSourceClusterTopicWriteCommand)
		output, err = cmd.Output()
		// validate command output
		if err != nil || !isUpdateTopicCommandSuccess(string(output)) {
			reqLogger.Error(err, "Failed to stop Topic "+topic+" write in SourceCluster "+sourceCluster+" with output: "+string(output))
			// terminate the transfer process
			undo = true
			return reconcile.Result{Requeue: true}, err
		}
		reqLogger.Info("Successfully stop Topic " + topic + " write in SourceCluster " + sourceCluster + " with output: " + string(output))

		// step4: check source cluster unconsumed message
		status = 4
		for i, consumerGroup := range consumerGroups {
			log.Info("Processing consumer group" + consumerGroup + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(consumerGroups)))
			for {
				checkConsumeProgressCommand := buildCheckConsumeProgressCommand(consumerGroup, nameServer)
				reqLogger.Info("checkConsumeProgressCommand: " + checkConsumeProgressCommand)
				cmd = exec.Command(cons.BasicCommand, cons.AdminToolDir, checkConsumeProgressCommand)
				output, err = cmd.Output()
				if err != nil || !isCheckConsumeProcessCommandSuccess(string(output)) {
					reqLogger.Error(err, "Failed to check consumerGroup "+consumerGroup+" with output: "+string(output))
					// terminate the transfer process
					undo = true
					return reconcile.Result{Requeue: true}, err
				}
				reqLogger.Info(" output: " + string(output))
				if isConsumeFinished(string(output), topic, sourceCluster) {
					reqLogger.Info("Message consumption of " + consumerGroup + " in source cluster " + sourceCluster + " finished!")
					break
				}
				reqLogger.Info("Wait a moment for message consumption of " + consumerGroup + " in source cluster " + sourceCluster + " finish ...")
				time.Sleep(time.Duration(cons.CheckConsumeFinishIntervalInSecond) * time.Second)
			}
		}

		// step5: delete topic in source cluster
		status = 5
		deleteSourceClusterTopicCommand := buildDeleteSourceClusterTopicCommand(topic, sourceCluster, nameServer)
		reqLogger.Info("deleteSourceClusterTopicCommand: " + deleteSourceClusterTopicCommand)
		cmd = exec.Command(cons.BasicCommand, cons.AdminToolDir, deleteSourceClusterTopicCommand)
		output, err = cmd.Output()
		if err != nil || !isDeleteTopicCommandSuccess(string(output)) {
			reqLogger.Error(err, "Failed to delete Topic "+topic+" in SourceCluster "+sourceCluster+" with output: "+string(output))
			// terminate the transfer process
			undo = true
			return reconcile.Result{Requeue: true}, err
		}
		reqLogger.Info("Successfully delete Topic " + topic + " in SourceCluster " + sourceCluster + " with output: " + string(output))

		// step6: delete consumer group in source cluster
		status = 6
		for i, consumerGroup := range consumerGroups {
			log.Info("Processing consumer group" + consumerGroup + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(consumerGroups)))
			deleteConsumerGroupCommand := buildDeleteConsumeGroupCommand(consumerGroup, sourceCluster, nameServer)
			reqLogger.Info("deleteConsumerGroupCommand: " + deleteConsumerGroupCommand)
			cmd = exec.Command(cons.BasicCommand, cons.AdminToolDir, deleteConsumerGroupCommand)
			output, err = cmd.Output()
			if err != nil || !isDeleteConsumerGroupSuccess(string(output)) {
				reqLogger.Error(err, "Failed to delete consumer group "+consumerGroup+" in SourceCluster "+sourceCluster+" with output: "+string(output))
				// terminate the transfer process
				undo = true
				return reconcile.Result{Requeue: true}, err
			}
			reqLogger.Info("Successfully delete consumer group " + consumerGroup + " in SourceCluster " + sourceCluster + " with output: " + string(output))
		}

		// step7: create retry topic
		status = 7
		for i, consumerGroup := range consumerGroups {
			log.Info("Processing consumer group" + consumerGroup + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(consumerGroups)))
			createRetryTopicCommand := buildAddRetryTopicToClusterCommand(consumerGroup, targetCluster, nameServer)
			reqLogger.Info("createRetryTopicCommand: " + createRetryTopicCommand)
			cmd = exec.Command(cons.BasicCommand, cons.AdminToolDir, createRetryTopicCommand)
			output, err = cmd.Output()
			if err != nil || !isUpdateTopicCommandSuccess(string(output)) {
				reqLogger.Error(err, "Failed to create retry topic of consumer group "+consumerGroup+" in TargetCluster "+targetCluster+" with output: "+string(output))
				// terminate the transfer process
				undo = true
				return reconcile.Result{Requeue: true}, err
			}
			reqLogger.Info("Successfully create retry topic of consumer group " + consumerGroup + " in TargetCluster " + targetCluster + " with output: " + string(output))
		}

		reqLogger.Info("Topic " + topic + " has been successfully transferred from " + sourceCluster + " to " + targetCluster)
	}

	return reconcile.Result{}, nil
}

func undoStopWrite(topic string, cluster string, nameServer string) {
	addTopicToClusterCommand := buildUndoStopWriteCommand(topic, cluster, nameServer)
	log.Info("undoStopWrite: " + addTopicToClusterCommand)
	cmd := exec.Command(cons.BasicCommand, cons.AdminToolDir, addTopicToClusterCommand)
	output, err := cmd.Output()
	if err != nil || !isUpdateTopicCommandSuccess(string(output)) {
		log.Error(err, "Failed to undo stop write topic with output: "+string(output))
	}
	log.Info("Successfully undo stop write topic with output: " + string(output))
}

func undoDeleteTopic(topic string, cluster string, nameServer string) {
	addTopicToClusterCommand := buildAddTopicToClusterCommand(topic, cluster, nameServer)
	log.Info("undoDeleteTopic: " + addTopicToClusterCommand)
	cmd := exec.Command(cons.BasicCommand, cons.AdminToolDir, addTopicToClusterCommand)
	output, err := cmd.Output()
	if err != nil || !isUpdateTopicCommandSuccess(string(output)) {
		log.Error(err, "Failed to undo delete topic with output: "+string(output))
	}
	log.Info("Successfully undo delete topic with output: " + string(output))
}

func undoDeleteConsumeGroup(consumerGroups []string, cluster string, nameServer string) {
	for _, consumerGroup := range consumerGroups {
		addConsumerGroupToTargetClusterCommand := buildAddConsumerGroupToClusterCommand(consumerGroup, cluster, nameServer)
		log.Info("undoDeleteConsumeGroup: " + addConsumerGroupToTargetClusterCommand)
		cmd := exec.Command(cons.BasicCommand, cons.AdminToolDir, addConsumerGroupToTargetClusterCommand)
		output, err := cmd.Output()
		if err != nil || !isUpdateConsumerGroupSuccess(string(output)) {
			log.Error(err, "Failed to undo delete consume group with output: "+string(output))
		}
		log.Info("Successfully undo delete consume group with output: " + string(output))
	}
}

func getConsumerGroupByTopic(topic string, nameServer string) []string {
	var consumerGroups []string
	topicListCmd := buildTopicListCommand(nameServer)
	cmd := exec.Command(cons.BasicCommand, cons.AdminToolDir, topicListCmd)
	output, err := cmd.Output()
	if err != nil || !isTopicListSuccess(string(output)) {
		log.Error(err, "Failed to list topic with output: "+string(output))
		return nil
	}
	log.Info("topicListCmd output: " + string(output))

	lines := strings.Split(string(output), "\n")
	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(strings.TrimSpace(lines[i]))
		if len(fields) >= cons.TopicListConsumerGroup {
			if fields[cons.TopicListTopic] == topic {
				consumerGroups = append(consumerGroups, fields[cons.TopicListConsumerGroup])
			}
		}
	}
	return consumerGroups
}

func isTopicListSuccess(s string) bool {
	return strings.Contains(s, "#Topic") && strings.Contains(s, "#Consumer Group")
}

func isCheckConsumeProcessCommandSuccess(s string) bool {
	return strings.Contains(s, "#Topic")
}

func isDeleteTopicCommandSuccess(s string) bool {
	return strings.Contains(s, "delete topic") && strings.Contains(s, "success")
}

func isUpdateTopicCommandSuccess(s string) bool {
	return strings.Contains(s, "create topic") && strings.Contains(s, "success")
}

func isDeleteConsumerGroupSuccess(s string) bool {
	return strings.Contains(s, "delete subscription group") && strings.Contains(s, "success")
}

func isUpdateConsumerGroupSuccess(s string) bool {
	// return strings.Contains(s, "create subscription group") && strings.Contains(s, "success")
	return strings.Contains(s, "groupName")
}

func buildUndoStopWriteCommand(topic string, cluster string, nameServer string) string {
	cmdOpts := []string{
		"updatetopic",
		"-t",
		topic,
		"-c",
		cluster,
		"-r",
		"8",
		"-w",
		"8",
		"-p",
		"6",
		"-n",
		nameServer,
	}
	return strings.Join(cmdOpts, " ")
}

func buildTopicListCommand(nameServer string) string {
	cmdOpts := []string{
		"topiclist",
		"-c",
		"-n",
		nameServer,
	}
	return strings.Join(cmdOpts, " ")
}

func buildAddRetryTopicToClusterCommand(consumerGroup string, cluster string, nameServer string) string {
	cmdOpts := []string{
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
	return strings.Join(cmdOpts, " ")
}

func getClusterBrokerNames(cluster string) []string {
	// TODO: consider more scenarios
	return []string{cluster}
}

func isConsumeFinished(output string, topic string, cluster string) bool {
	lines := strings.Split(output, "\n")
	brokers := getClusterBrokerNames(cluster)
	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(strings.TrimSpace(lines[i]))
		if len(fields) > cons.Diff {
			for _, broker := range brokers {
				log.Info("broker = " + broker)
				log.Info("fields[cons.Topic] = " + fields[cons.Topic] + " , in line " + strconv.Itoa(i))
				log.Info("fields[cons.BrokerName] = " + fields[cons.BrokerName] + " , in line " + strconv.Itoa(i))
				log.Info("fields[cons.Diff] = " + fields[cons.Diff] + " , in line " + strconv.Itoa(i))
				if fields[cons.Topic] == topic && fields[cons.BrokerName] == broker {
					if fields[cons.Diff] != "0" {
						return false
					}
				}
			}
		}
	}
	return true
}

func buildDeleteConsumeGroupCommand(consumerGroup string, cluster string, nameServer string) string {
	cmdOpts := []string{
		"deleteSubGroup",
		"-g",
		consumerGroup,
		"-c",
		cluster,
		"-n",
		nameServer,
	}
	return strings.Join(cmdOpts, " ")
}

func buildDeleteSourceClusterTopicCommand(topic string, sourceCluster string, nameServer string) string {
	cmdOpts := []string{
		"deletetopic",
		"-t",
		topic,
		"-c",
		sourceCluster,
		"-n",
		nameServer,
	}
	return strings.Join(cmdOpts, " ")
}

func buildCheckConsumeProgressCommand(consumerGroup string, nameServer string) string {
	cmdOpts := []string{
		"consumerprogress",
		"-g",
		consumerGroup,
		"-n",
		nameServer,
	}
	return strings.Join(cmdOpts, " ")
}

func buildStopClusterTopicWriteCommand(topic string, cluster string, nameServer string) string {
	cmdOpts := []string{
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
	return strings.Join(cmdOpts, " ")
}

func buildAddConsumerGroupToClusterCommand(consumerGroup string, cluster string, nameServer string) string {
	cmdOpts := []string{
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
	return strings.Join(cmdOpts, " ")
}

func buildAddTopicToClusterCommand(topic string, cluster string, nameServer string) string {
	cmdOpts := []string{
		"updatetopic",
		"-t",
		topic,
		"-c",
		cluster,
		"-n",
		nameServer,
	}
	return strings.Join(cmdOpts, " ")
}
