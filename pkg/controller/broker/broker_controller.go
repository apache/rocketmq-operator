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

// Package broker contains the implementation of the Broker CRD reconcile function
package broker

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"time"

	rocketmqv1alpha1 "github.com/apache/rocketmq-operator/pkg/apis/rocketmq/v1alpha1"
	cons "github.com/apache/rocketmq-operator/pkg/constants"
	"github.com/apache/rocketmq-operator/pkg/share"
	"github.com/apache/rocketmq-operator/pkg/tool"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_broker")
var isInitial = true
var cmd = []string{"/bin/bash", "-c", "echo Initial broker"}

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Broker Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBroker{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("broker-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Broker
	err = c.Watch(&source.Kind{Type: &rocketmqv1alpha1.Broker{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Broker
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rocketmqv1alpha1.Broker{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileBroker implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBroker{}

// ReconcileBroker reconciles a Broker object
type ReconcileBroker struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Broker object and makes changes based on the state read
// and what is in the Broker.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBroker) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Broker.")

	// Fetch the Broker instance
	broker := &rocketmqv1alpha1.Broker{}
	err := r.client.Get(context.TODO(), request.NamespacedName, broker)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Broker resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Broker.")
		return reconcile.Result{}, err
	}

	actualKey := broker.Namespace + "-" + broker.Spec.RocketMQName
	actual, _ := share.GetInstance().LoadOrStore(actualKey, share.ShareItem{})
	defer func() {
		reqLogger.Info("Broker actualKey:" + actualKey + " actual.NameServerStr:" + actual.NameServersStr +
			" actual.BrokerClusterName:" + actual.BrokerClusterName + " IsNameServersStrInitialized:" + strconv.FormatBool(actual.IsNameServersStrInitialized))
		share.GetInstance().Store(actualKey, actual)
	}()

	var groupNum int
	if broker.Status.Size == 0 {
		groupNum = broker.Spec.Size
	} else {
		groupNum = broker.Status.Size
	}
	if broker.Spec.NameServers == "" {
		// wait for name server ready when create broker cluster if nameServers is omitted
		for {
			actual, _ = share.GetInstance().Load(actualKey)
			if actual.IsNameServersStrInitialized {
				break
			} else {
				log.Info("Broker Waiting for name server ready..., actualKey:" +
					actualKey + " actual.NameServersStr:" + actual.NameServersStr + " IsNameServersStrInitialized:" +
					strconv.FormatBool(actual.IsNameServersStrInitialized))
				time.Sleep(time.Duration(cons.WaitForNameServerReadyInSecond) * time.Second)
			}
		}
	} else {
		actual.NameServersStr = broker.Spec.NameServers
	}
	actual.GroupNum = groupNum

	actual.BrokerClusterName = broker.Name
	replicaPerGroup := broker.Spec.ReplicaPerGroup
	reqLogger.Info("brokerGroupNum=" + strconv.Itoa(actual.GroupNum) + ", replicaPerGroup=" + strconv.Itoa(replicaPerGroup))
	for brokerGroupIndex := 0; brokerGroupIndex < actual.GroupNum; brokerGroupIndex++ {
		reqLogger.Info("Check Broker cluster " + strconv.Itoa(brokerGroupIndex+1) + "/" + strconv.Itoa(actual.GroupNum))
		dep := r.getBrokerStatefulSet(broker, brokerGroupIndex, 0)
		// Check if the statefulSet already exists, if not create a new one
		found := &appsv1.StatefulSet{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Master Broker StatefulSet.", "StatefulSet.Namespace", dep.Namespace, "StatefulSet.Name", dep.Name)
			err = r.client.Create(context.TODO(), dep)
			if err != nil {
				reqLogger.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", dep.Namespace, "StatefulSet.Name", dep.Name)
			}
		} else if err != nil {
			reqLogger.Error(err, "Failed to get broker master StatefulSet.")
		}

		for replicaIndex := 1; replicaIndex <= replicaPerGroup; replicaIndex++ {
			reqLogger.Info("Check Replica Broker of cluster-" + strconv.Itoa(brokerGroupIndex) + " " + strconv.Itoa(replicaIndex) + "/" + strconv.Itoa(replicaPerGroup))
			replicaDep := r.getBrokerStatefulSet(broker, brokerGroupIndex, replicaIndex)
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: replicaDep.Name, Namespace: replicaDep.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {
				reqLogger.Info("Creating a new Replica Broker StatefulSet.", "StatefulSet.Namespace", replicaDep.Namespace, "StatefulSet.Name", replicaDep.Name)
				err = r.client.Create(context.TODO(), replicaDep)
				if err != nil {
					reqLogger.Error(err, "Failed to create new StatefulSet of broker-"+strconv.Itoa(brokerGroupIndex)+"-replica-"+strconv.Itoa(replicaIndex), "StatefulSet.Namespace", replicaDep.Namespace, "StatefulSet.Name", replicaDep.Name)
				}
			} else if err != nil {
				reqLogger.Error(err, "Failed to get broker replica StatefulSet.")
			}
		}
	}

	// Check for name server scaling
	if broker.Spec.AllowRestart {
		// The following code will restart all brokers to update NAMESRV_ADDR env
		if actual.IsNameServersStrUpdated {
				for brokerGroupIndex := 0; brokerGroupIndex < broker.Spec.Size; brokerGroupIndex++ {
				brokerName := getBrokerName(broker, brokerGroupIndex)
				// Update master broker
				reqLogger.Info("Update Master Broker NAMESRV_ADDR of " + brokerName)
				dep := r.getBrokerStatefulSet(broker, brokerGroupIndex, 0)
				found := &appsv1.StatefulSet{}
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)
				if err != nil {
					reqLogger.Error(err, "Failed to get broker master StatefulSet of "+brokerName)
				} else {
					found.Spec.Template.Spec.Containers[0].Env[0].Value = actual.NameServersStr
					err = r.client.Update(context.TODO(), found)
					if err != nil {
						reqLogger.Error(err, "Failed to update NAMESRV_ADDR of master broker "+brokerName, "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
					} else {
						reqLogger.Info("Successfully updated NAMESRV_ADDR of master broker "+brokerName, "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
					}
					time.Sleep(time.Duration(cons.RestartBrokerPodIntervalInSecond) * time.Second)
				}
				// Update replicas brokers
				for replicaIndex := 1; replicaIndex <= replicaPerGroup; replicaIndex++ {
					reqLogger.Info("Update Replica Broker NAMESRV_ADDR of " + brokerName + " " + strconv.Itoa(replicaIndex) + "/" + strconv.Itoa(replicaPerGroup))
					replicaDep := r.getBrokerStatefulSet(broker, brokerGroupIndex, replicaIndex)
					replicaFound := &appsv1.StatefulSet{}
					err = r.client.Get(context.TODO(), types.NamespacedName{Name: replicaDep.Name, Namespace: replicaDep.Namespace}, replicaFound)
					if err != nil {
						reqLogger.Error(err, "Failed to get broker replica StatefulSet of "+brokerName)
					} else {
						for index := range replicaFound.Spec.Template.Spec.Containers[0].Env {
							if cons.EnvNameServiceAddress == replicaFound.Spec.Template.Spec.Containers[0].Env[index].Name {
								replicaFound.Spec.Template.Spec.Containers[0].Env[index].Value = actual.NameServersStr
								break
							}
						}
						err = r.client.Update(context.TODO(), replicaFound)
						if err != nil {
							reqLogger.Error(err, "Failed to update NAMESRV_ADDR of "+strconv.Itoa(brokerGroupIndex)+"-replica-"+strconv.Itoa(replicaIndex), "StatefulSet.Namespace", replicaFound.Namespace, "StatefulSet.Name", replicaFound.Name)
						} else {
							reqLogger.Info("Successfully updated NAMESRV_ADDR of "+strconv.Itoa(brokerGroupIndex)+"-replica-"+strconv.Itoa(replicaIndex), "StatefulSet.Namespace", replicaFound.Namespace, "StatefulSet.Name", replicaFound.Name)
						}
						time.Sleep(time.Duration(cons.RestartBrokerPodIntervalInSecond) * time.Second)
					}
				}
			}
		}
		actual.IsNameServersStrUpdated = false
	}

	// List the pods for this broker's statefulSet
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForBroker(broker.Name))
	listOps := &client.ListOptions{
		Namespace:     broker.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods.", "Broker.Namespace", broker.Namespace, "Broker.Name", broker.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)
	log.Info("broker.Status.Nodes length = " + strconv.Itoa(len(broker.Status.Nodes)))
	log.Info("podNames length = " + strconv.Itoa(len(podNames)))

	if len(podNames) == 0 {
		reqLogger.Info("Broker podName size is 0, wait for a moment...")
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, nil
	}

	// Ensure every pod is in running phase
	for _, pod := range podList.Items {
		if !reflect.DeepEqual(pod.Status.Phase, corev1.PodRunning) {
			reqLogger.Info("pod " + pod.Name + " phase is " + string(pod.Status.Phase) + ", wait for a moment...")
			return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, nil
		}
	}

	if broker.Status.Size != 0 && broker.Spec.Size > broker.Status.Size {
		// Get the metadata including subscriptionGroup.json and topics.json from scale source pod
		k8s, err := tool.NewK8sClient()
		if err != nil {
			log.Error(err, "Error occurred while getting K8s Client")
		}
		sourcePodName := broker.Spec.ScalePodName
		topicsCommand := getCopyMetadataJsonCommand(cons.TopicJsonDir, sourcePodName, broker.Namespace, k8s)
		log.Info("topicsCommand: " + topicsCommand)
		subscriptionGroupCommand := getCopyMetadataJsonCommand(cons.SubscriptionGroupJsonDir, sourcePodName, broker.Namespace, k8s)
		log.Info("subscriptionGroupCommand: " + subscriptionGroupCommand)
		MakeConfigDirCommand := "mkdir -p " + cons.StoreConfigDir
		ChmodDirCommand := "chmod a+rw " + cons.StoreConfigDir
		cmd = []string{"/bin/bash", "-c", MakeConfigDirCommand + " && " + ChmodDirCommand + " && " + topicsCommand + " && " + subscriptionGroupCommand}
	}

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, broker.Status.Nodes) {
		broker.Status.Nodes = podNames
		if len(broker.Status.Nodes) != 0 {
			err = r.client.Status().Update(context.TODO(), broker)
			if err != nil {
				reqLogger.Error(err, "Failed to update Broker Nodes status.")
			}
		}
	}

	// Update status.Size if needed
	if broker.Spec.Size != broker.Status.Size {
		log.Info("broker.Status.Size = " + strconv.Itoa(broker.Status.Size))
		log.Info("broker.Spec.Size = " + strconv.Itoa(broker.Spec.Size))
		broker.Status.Size = broker.Spec.Size
		err = r.client.Status().Update(context.TODO(), broker)
		if err != nil {
			reqLogger.Error(err, "Failed to update Broker Size status.")
		}
	}

	//podList := &corev1.PodList{}
	//labelSelector := labels.SelectorFromSet(labelsForBroker(broker.Name))
	//listOps := &client.ListOptions{
	//	Namespace:     broker.Namespace,
	//	LabelSelector: labelSelector,
	//}
	//err = r.client.List(context.TODO(), listOps, podList)
	//if err != nil {
	//	reqLogger.Error(err, "Failed to list pods.", "Broker.Namespace", broker.Namespace, "Broker.Name", broker.Name)
	//	return reconcile.Result{}, err
	//}
	//podNames := getPodNames(podList.Items)
	//
	//// Update status.Nodes if needed
	//if !reflect.DeepEqual(podNames, broker.Status.Nodes) {
	//	broker.Status.Nodes = podNames
	//	err := r.client.Status().Update(context.TODO(), broker)
	//	if err != nil {
	//		reqLogger.Error(err, "Failed to update Broker status.")
	//		return reconcile.Result{}, err
	//	}
	//}

	return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, nil
}

func getCopyMetadataJsonCommand(dir string, sourcePodName string, namespace string, k8s *tool.K8sClient) string {
	cmdOpts := buildInputCommand(dir)
	topicsJsonStr := exec(cmdOpts, sourcePodName, k8s, namespace)
	topicsCommand := buildOutputCommand(topicsJsonStr, dir)
	return strings.Join(topicsCommand, " ")
}

func buildInputCommand(source string) []string {
	cmdOpts := []string{
		"cat",
		source,
	}
	return cmdOpts
}

func buildOutputCommand(content string, dest string) []string {
	replaced := strings.Replace(content, "\"", "\\\"", -1)
	cmdOpts := []string{
		"echo",
		"-e",
		"\"" + replaced + "\"",
		">",
		dest,
	}
	return cmdOpts
}

func exec(cmdOpts []string, podName string, k8s *tool.K8sClient, namespace string) string {
	log.Info("On pod " + podName + ", command being run: " + strings.Join(cmdOpts, " "))
	container := cons.BrokerContainerName
	outputBytes, stderrBytes, err := k8s.Exec(namespace, podName, container, cmdOpts, nil)
	stderr := stderrBytes.String()
	output := outputBytes.String()

	if stderrBytes != nil {
		log.Info("STDERR: " + stderr)
	}

	if err != nil {
		log.Error(err, "Error occurred while running command: "+strings.Join(cmdOpts, " "))
		log.Info("stdout: " + output)
	} else {
		log.Info("output: " + output)
	}
	return output
}

func getBrokerName(broker *rocketmqv1alpha1.Broker, brokerGroupIndex int) string {
	return broker.Name + "-" + strconv.Itoa(brokerGroupIndex)
}

// getBrokerStatefulSet returns a broker StatefulSet object
func (r *ReconcileBroker) getBrokerStatefulSet(broker *rocketmqv1alpha1.Broker, brokerGroupIndex int, replicaIndex int) *appsv1.StatefulSet {
	ls := labelsForBroker(broker.Name)
	var a int32 = 1
	var c = &a
	var statefulSetName string
	if replicaIndex == 0 {
		statefulSetName = broker.Name + "-" + strconv.Itoa(brokerGroupIndex) + "-master"
	} else {
		statefulSetName = broker.Name + "-" + strconv.Itoa(brokerGroupIndex) + "-replica-" + strconv.Itoa(replicaIndex)
	}

	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      statefulSetName,
			Namespace: broker.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: c,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Resources: broker.Spec.Resources,
						Image: broker.Spec.BrokerImage,
						Name:  cons.BrokerContainerName,
						Lifecycle: &corev1.Lifecycle{
							PostStart: &corev1.Handler{
								Exec: &corev1.ExecAction{
									Command: cmd,
								},
							},
						},
						ImagePullPolicy: broker.Spec.ImagePullPolicy,
						Env: getENV(broker, replicaIndex, brokerGroupIndex),
						Ports: []corev1.ContainerPort{{
							ContainerPort: cons.BrokerVipContainerPort,
							Name:          cons.BrokerVipContainerPortName,
						}, {
							ContainerPort: cons.BrokerMainContainerPort,
							Name:          cons.BrokerMainContainerPortName,
						}, {
							ContainerPort: cons.BrokerHighAvailabilityContainerPort,
							Name:          cons.BrokerHighAvailabilityContainerPortName,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: cons.LogMountPath,
							Name:      broker.Spec.VolumeClaimTemplates[0].Name,
							SubPath:   cons.LogSubPathName + getPathSuffix(broker, brokerGroupIndex, replicaIndex),
						}, {
							MountPath: cons.StoreMountPath,
							Name:      broker.Spec.VolumeClaimTemplates[0].Name,
							SubPath:   cons.StoreSubPathName + getPathSuffix(broker, brokerGroupIndex, replicaIndex),
						}, {
							MountPath: cons.BrokerConfigPath + "/" + cons.BrokerConfigName,
							Name:      broker.Spec.Volumes[0].Name,
							SubPath:   cons.BrokerConfigName,
						}},
					}},
					Volumes: getVolumes(broker),
				},
			},
			VolumeClaimTemplates: getVolumeClaimTemplates(broker),
		},
	}
	// Set Broker instance as the owner and controller
	controllerutil.SetControllerReference(broker, dep, r.scheme)

	return dep

}

func getENV(broker *rocketmqv1alpha1.Broker, replicaIndex int, brokerGroupIndex int)  []corev1.EnvVar {

	actualKey := broker.Namespace + "-" + broker.Spec.RocketMQName
	actual, _ := share.GetInstance().LoadOrStore(actualKey, share.ShareItem{})

	envs := []corev1.EnvVar{{
		Name:  cons.EnvNameServiceAddress,
		Value: actual.NameServersStr,
	}, {
		Name:  cons.EnvBrokerId,
		Value: strconv.Itoa(replicaIndex),
	}, {
		Name:  cons.EnvBrokerClusterName,
		Value: broker.Name,
	}, {
		Name:  cons.EnvBrokerName,
		Value: broker.Name + "-" + strconv.Itoa(brokerGroupIndex),
	}}
	envs = append(envs, broker.Spec.Env...)
	return envs
}

func getVolumeClaimTemplates(broker *rocketmqv1alpha1.Broker) []corev1.PersistentVolumeClaim {
	switch broker.Spec.StorageMode {
	case cons.StorageModeStorageClass:
		return broker.Spec.VolumeClaimTemplates
	case cons.StorageModeEmptyDir, cons.StorageModeHostPath:
		fallthrough
	default:
		return nil
	}
}

func getVolumes(broker *rocketmqv1alpha1.Broker) []corev1.Volume {
	switch broker.Spec.StorageMode {
	case cons.StorageModeStorageClass:
		return broker.Spec.Volumes
	case cons.StorageModeEmptyDir:
		volumes := broker.Spec.Volumes
		volumes = append(volumes, corev1.Volume{
			Name: broker.Spec.VolumeClaimTemplates[0].Name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
		return volumes
	case cons.StorageModeHostPath:
		fallthrough
	default:
		volumes := broker.Spec.Volumes
		volumes = append(volumes, corev1.Volume{
			Name: broker.Spec.VolumeClaimTemplates[0].Name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: broker.Spec.HostPath,
				}},
		})
		return volumes
	}
}

func getPathSuffix(broker *rocketmqv1alpha1.Broker, brokerGroupIndex int, replicaIndex int) string {
	if replicaIndex == 0 {
		return "/" + broker.Name + "-" + strconv.Itoa(brokerGroupIndex) + "-master"
	}
	return "/" + broker.Name + "-" + strconv.Itoa(brokerGroupIndex) + "-replica-" + strconv.Itoa(replicaIndex)
}

// labelsForBroker returns the labels for selecting the resources
// belonging to the given broker CR name.
func labelsForBroker(name string) map[string]string {
	return map[string]string{"app": "broker", "broker_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func contains(item string, arr []string) bool {
	for _, value := range arr {
		if reflect.DeepEqual(value, item) {
			return true
		}
	}
	return false
}

func checkAndCopyMetadata(newPodNames []string, dir string, sourcePodName string, namespace string, k8s *tool.K8sClient) {
	cmdOpts := buildInputCommand(dir)
	jsonStr := exec(cmdOpts, sourcePodName, k8s, namespace)
	if len(jsonStr) < cons.MinMetadataJsonFileSize {
		log.Info("The file " + dir + " is abnormally too short to execute metadata transmission, please check whether the source broker pod " + sourcePodName + " is correct")
	} else {
		// for each new pod, copy the metadata from the scale source pod
		for _, newPodName := range newPodNames {
			cmdOpts = buildOutputCommand(jsonStr, dir)
			exec(cmdOpts, newPodName, k8s, namespace)
		}
	}
}
