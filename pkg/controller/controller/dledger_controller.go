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

// Package controller contains the implementation of the Controller CRD reconcile function
package controller

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

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
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("dledger_controller")

// SetupWithManager creates a new Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func SetupWithManager(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileController{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("dledger-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Controller
	err = c.Watch(&source.Kind{Type: &rocketmqv1alpha1.Controller{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Controller
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rocketmqv1alpha1.Controller{},
	})
	if err != nil {
		return err
	}

	return nil
}

//+kubebuilder:rbac:groups=rocketmq.apache.org,resources=controllers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rocketmq.apache.org,resources=controllers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rocketmq.apache.org,resources=controllers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods/exec,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// ReconcileController reconciles a Controller object
type ReconcileController struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Controller object and makes changes based on the state read
// and what is in the Controller.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Controller.")

	// Fetch the Controller instance
	controller := &rocketmqv1alpha1.Controller{}
	err := r.client.Get(context.TODO(), request.NamespacedName, controller)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Controller resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Controller.")
		return reconcile.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
	}

	//create headless svc
	headlessSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: tool.BuildHeadlessSvcResourceName(request.Name), Namespace: request.Namespace}, headlessSvc)
	if err != nil {
		if errors.IsNotFound(err) {
			// create;
			consoleSvc := r.generateHeadlessSvc(controller)
			err = r.client.Create(context.TODO(), consoleSvc)
			if err != nil {
				reqLogger.Error(err, "Failed to create controller headless svc")
				return reconcile.Result{}, err
			} else {
				reqLogger.Info("Successfully create controller headless svc")
			}
		} else {
			return reconcile.Result{}, err
		}
	}

	sts := r.getControllerStatefulSet(controller)
	// Check if the statefulSet already exists, if not create a new one
	found := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Controller StatefulSet.", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
		err = r.client.Create(context.TODO(), sts)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Controller StatefulSet", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to list Controller StatefulSet.")
	}

	// List the pods for this controller's statefulSet
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForController(controller.Name))
	listOps := &client.ListOptions{
		Namespace:     controller.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.client.List(context.TODO(), podList, listOps)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods.", "Controller.Namespace", controller.Namespace, "Controller.Name", controller.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)
	log.Info("controller.Status.Nodes length = " + strconv.Itoa(len(controller.Status.Nodes)))
	log.Info("podNames length = " + strconv.Itoa(len(podNames)))
	// Ensure every pod is in running phase
	for _, pod := range podList.Items {
		if !reflect.DeepEqual(pod.Status.Phase, corev1.PodRunning) {
			log.Info("pod " + pod.Name + " phase is " + string(pod.Status.Phase) + ", wait for a moment...")
		}
	}

	// Update status.Size if needed
	if controller.Spec.Size != controller.Status.Size {
		log.Info("controller.Status.Size = " + strconv.Itoa(controller.Status.Size))
		log.Info("controller.Spec.Size = " + strconv.Itoa(controller.Spec.Size))
		controller.Status.Size = controller.Spec.Size
		err = r.client.Status().Update(context.TODO(), controller)
		if err != nil {
			reqLogger.Error(err, "Failed to update Controller Size status.")
		}
	}

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, controller.Status.Nodes) {
		controller.Status.Nodes = podNames
		err = r.client.Status().Update(context.TODO(), controller)
		if err != nil {
			reqLogger.Error(err, "Failed to update Controller Nodes status.")
		}
	}

	//create svc
	controllerSvc := &corev1.Service{}
	controllerSvcName := tool.BuildSvcResourceName(request.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: controllerSvcName, Namespace: request.Namespace}, controllerSvc)
	if err != nil {
		if errors.IsNotFound(err) {
			// create;
			svcToCreate := r.generateSvc(controller)
			err = r.client.Create(context.TODO(), svcToCreate)
			if err != nil {
				reqLogger.Error(err, "Failed to create controller svc")
				return reconcile.Result{}, err
			} else {
				reqLogger.Info("Successfully create controller svc")
			}
		} else {
			return reconcile.Result{}, err
		}
	}
	share.ControllerAccessPoint = controllerSvcName + ":9878"

	return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, nil
}

// returns a controller StatefulSet object
func (r *ReconcileController) getControllerStatefulSet(controller *rocketmqv1alpha1.Controller) *appsv1.StatefulSet {
	ls := labelsForController(controller.Name)

	// After CustomResourceDefinition version upgraded from v1beta1 to v1
	// `controller.spec.VolumeClaimTemplates.metadata` declared in yaml will not be stored by kubernetes.
	// Here is a temporary repair method: to generate a random name
	if strings.EqualFold(controller.Spec.VolumeClaimTemplates[0].Name, "") {
		controller.Spec.VolumeClaimTemplates[0].Name = uuid.New().String()
	}

	var replica = int32(controller.Spec.Size)
	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      controller.Name,
			Namespace: controller.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: tool.BuildHeadlessSvcResourceName(controller.Name),
			Replicas:    &replica,
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

					ServiceAccountName: controller.Spec.ServiceAccountName,
					Affinity:           controller.Spec.Affinity,
					Tolerations:        controller.Spec.Tolerations,
					NodeSelector:       controller.Spec.NodeSelector,
					PriorityClassName:  controller.Spec.PriorityClassName,
					ImagePullSecrets:   controller.Spec.ImagePullSecrets,
					Containers: []corev1.Container{{
						Resources:       controller.Spec.Resources,
						Image:           controller.Spec.ControllerImage,
						Name:            cons.ControllerContainerName,
						SecurityContext: getContainerSecurityContext(controller),
						ImagePullPolicy: controller.Spec.ImagePullPolicy,
						Env:             getENV(controller),
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: cons.LogMountPath,
							Name:      controller.Spec.VolumeClaimTemplates[0].Name,
							SubPath:   cons.LogSubPathName,
						}, {
							MountPath: cons.StoreMountPath,
							Name:      controller.Spec.VolumeClaimTemplates[0].Name,
							SubPath:   cons.StoreSubPathName,
						}},
						// Command: []string{"sh", "mqcontroller"},
					}},
					Volumes:         getVolumes(controller),
					SecurityContext: getPodSecurityContext(controller),
				},
			},
			VolumeClaimTemplates: getVolumeClaimTemplates(controller),
		},
	}
	// Set Controller instance as the owner and controller
	controllerutil.SetControllerReference(controller, dep, r.scheme)

	return dep

}

func getENV(controller *rocketmqv1alpha1.Controller) []corev1.EnvVar {
	var controllerDLegerPeersStr string
	for controllerIndex := 0; controllerIndex < int(controller.Spec.Size); controllerIndex++ {
		controllerDLegerPeersStr += controller.Name + strconv.Itoa(controllerIndex) + "-" + controller.Name + "-" + strconv.Itoa(controllerIndex) + "." + tool.BuildHeadlessSvcResourceName(controller.Name) + ":9878"
		if controllerIndex < int(controller.Spec.Size)-1 {
			controllerDLegerPeersStr += ";"
		}
	}
	log.Info("controllerDLegerPeersStr=" + controllerDLegerPeersStr)
	envs := []corev1.EnvVar{{
		Name:      "MY_POD_NAME",
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
	}, {
		Name:  cons.EnvControllerDLegerGroup,
		Value: "ControllerGroup-" + controller.Name,
	}, {
		Name:  cons.EnvControllerDLegerPeers,
		Value: controllerDLegerPeersStr,
	}, {
		Name:  cons.EnvControllerStorePath,
		Value: cons.StoreMountPath,
	}}
	envs = append(envs, controller.Spec.Env...)
	return envs
}

func getVolumeClaimTemplates(controller *rocketmqv1alpha1.Controller) []corev1.PersistentVolumeClaim {
	switch controller.Spec.StorageMode {
	case cons.StorageModeStorageClass:
		return controller.Spec.VolumeClaimTemplates
	case cons.StorageModeEmptyDir, cons.StorageModeHostPath:
		fallthrough
	default:
		return nil
	}
}

func getPodSecurityContext(controller *rocketmqv1alpha1.Controller) *corev1.PodSecurityContext {
	var securityContext = corev1.PodSecurityContext{}
	if controller.Spec.PodSecurityContext != nil {
		securityContext = *controller.Spec.PodSecurityContext
	}
	return &securityContext
}

func getContainerSecurityContext(controller *rocketmqv1alpha1.Controller) *corev1.SecurityContext {
	var securityContext = corev1.SecurityContext{}
	if controller.Spec.ContainerSecurityContext != nil {
		securityContext = *controller.Spec.ContainerSecurityContext
	}
	return &securityContext
}

func getVolumes(controller *rocketmqv1alpha1.Controller) []corev1.Volume {
	switch controller.Spec.StorageMode {
	case cons.StorageModeStorageClass:
		return nil
	case cons.StorageModeEmptyDir:
		volumes := []corev1.Volume{{
			Name: controller.Spec.VolumeClaimTemplates[0].Name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{}},
		}}
		return volumes
	case cons.StorageModeHostPath:
		fallthrough
	default:

		volumes := []corev1.Volume{{
			Name: controller.Spec.VolumeClaimTemplates[0].Name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: controller.Spec.HostPath,
				}},
		}}
		return volumes
	}
}

// labelsForController returns the labels for selecting the resources
// belonging to the given controller CR name.
func labelsForController(name string) map[string]string {
	return map[string]string{"app": "controller", "controller_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *ReconcileController) generateHeadlessSvc(cr *rocketmqv1alpha1.Controller) *corev1.Service {
	controllerSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   cr.Namespace,
			Name:        tool.BuildHeadlessSvcResourceName(cr.Name),
			Annotations: map[string]string{"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true"},
			Labels:      cr.Labels,
			//Finalizers:  []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP:                "None",
			PublishNotReadyAddresses: true,
			Selector:                 labelsForController(cr.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       "controller",
					Port:       9878,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(9878),
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, controllerSvc, r.scheme)
	return controllerSvc
}

func (r *ReconcileController) generateSvc(cr *rocketmqv1alpha1.Controller) *corev1.Service {
	controllerSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cr.Namespace,
			Name:       tool.BuildSvcResourceName(cr.Name),
			Labels:     labelsForController(cr.Name),
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			Selector: labelsForController(cr.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       "controller",
					Port:       9878,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(9878),
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, controllerSvc, r.scheme)
	return controllerSvc
}
