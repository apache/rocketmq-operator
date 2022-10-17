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

package proxy

import (
	"context"
	rocketmqv1alpha1 "github.com/apache/rocketmq-operator/pkg/apis/rocketmq/v1alpha1"
	cons "github.com/apache/rocketmq-operator/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_proxy")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// SetupWithManager creates a new Proxy Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func SetupWithManager(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileProxy{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("proxy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Proxy
	err = c.Watch(&source.Kind{Type: &rocketmqv1alpha1.Proxy{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Proxy
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rocketmqv1alpha1.Proxy{},
	})
	if err != nil {
		return err
	}

	return nil
}

//+kubebuilder:rbac:groups=rocketmq.apache.org,resources=proxys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rocketmq.apache.org,resources=proxys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rocketmq.apache.org,resources=proxys/finalizers,verbs=update
//+kubebuilder:rbac:groups="apps",resources=statefulSets,verbs=get;list;watch;create;update;patch;delete

// ReconcileProxy reconciles a Proxy object
type ReconcileProxy struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Proxy object and makes changes based on the state read
// and what is in the Proxy.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileProxy) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Proxy")

	// Fetch the Proxy instance
	instance := &rocketmqv1alpha1.Proxy{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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
	if instance.Spec.ProxyConfigPath == "" || instance.Spec.ProxyMode == "" {
		reqLogger.Error(err, "The value of proxyConfigPath and proxyMode must be not empty.")
	}
	if instance.Spec.BrokerConfigPath == "" && instance.Spec.ProxyMode == "LOCAL" {
		reqLogger.Error(err, "ProxyMode is LOCAL, the value of brokerConfigPath must be not empty.")
	}
	proxyStatefulSet := newStatefulSetForCR(instance)

	// Set Proxy instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, proxyStatefulSet, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: proxyStatefulSet.Name, Namespace: proxyStatefulSet.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating RocketMQ Proxy StatefulSet", "Namespace", proxyStatefulSet, "Name", proxyStatefulSet.Name)
		err = r.client.Create(context.TODO(), proxyStatefulSet)
		if err != nil {
			return reconcile.Result{}, err
		}

		// created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Support proxy statefulSet scaling
	if !reflect.DeepEqual(instance.Spec.ProxyStatefulSet.Spec.Replicas, found.Spec.Replicas) {
		found.Spec.Replicas = instance.Spec.ProxyStatefulSet.Spec.Replicas
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update proxy CR ", "Namespace", found.Namespace, "Name", found.Name)
		} else {
			reqLogger.Info("Successfully updated proxy CR ", "Namespace", found.Namespace, "Name", found.Name)
		}
	}
	// CR already exists - don't requeue
	reqLogger.Info("Skip reconcile: RocketMQ Proxy StatefulSet already exists", "Namespace", found.Namespace, "Name", found.Name)
	return reconcile.Result{}, nil
}

// newStatefulSetForCR returns a StatefulSet pod with modifying the ENV
func newStatefulSetForCR(cr *rocketmqv1alpha1.Proxy) *appsv1.StatefulSet {
	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: cr.Spec.ProxyStatefulSet.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: cr.Spec.ProxyStatefulSet.Spec.Selector.MatchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: cr.Spec.ProxyStatefulSet.Spec.Template.ObjectMeta.Labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: cr.Spec.ProxyStatefulSet.Spec.Template.Spec.ServiceAccountName,
					Affinity:           cr.Spec.ProxyStatefulSet.Spec.Template.Spec.Affinity,
					ImagePullSecrets:   cr.Spec.ProxyStatefulSet.Spec.Template.Spec.ImagePullSecrets,
					Containers: []corev1.Container{{
						Resources:       cr.Spec.ProxyStatefulSet.Spec.Template.Spec.Containers[0].Resources,
						Image:           cr.Spec.ProxyStatefulSet.Spec.Template.Spec.Containers[0].Image,
						Name:            cr.Spec.ProxyStatefulSet.Spec.Template.Spec.Containers[0].Name,
						ImagePullPolicy: cr.Spec.ProxyStatefulSet.Spec.Template.Spec.Containers[0].ImagePullPolicy,
						Ports:           cr.Spec.ProxyStatefulSet.Spec.Template.Spec.Containers[0].Ports,
						Env:             getENV(cr),
						VolumeMounts:    cr.Spec.ProxyStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts,
						SecurityContext: getContainerSecurityContext(cr),
					}},
					Volumes:         cr.Spec.ProxyStatefulSet.Spec.Template.Spec.Volumes,
					SecurityContext: getPodSecurityContext(cr),
				},
			},
		},
	}

	return dep
}

func getPodSecurityContext(proxy *rocketmqv1alpha1.Proxy) *corev1.PodSecurityContext {
	var securityContext = corev1.PodSecurityContext{}
	if proxy.Spec.ProxyStatefulSet.Spec.Template.Spec.SecurityContext != nil {
		securityContext = *proxy.Spec.ProxyStatefulSet.Spec.Template.Spec.SecurityContext
	}
	return &securityContext
}

func getContainerSecurityContext(proxy *rocketmqv1alpha1.Proxy) *corev1.SecurityContext {
	var securityContext = corev1.SecurityContext{}
	if proxy.Spec.ProxyStatefulSet.Spec.Template.Spec.Containers[0].SecurityContext != nil {
		securityContext = *proxy.Spec.ProxyStatefulSet.Spec.Template.Spec.Containers[0].SecurityContext
	}
	return &securityContext
}
func getENV(proxy *rocketmqv1alpha1.Proxy) []corev1.EnvVar {
	envs := []corev1.EnvVar{{
		Name:  cons.EnvBrokerConfigPath,
		Value: proxy.Spec.BrokerConfigPath,
	}, {
		Name:  cons.EnvProxyConfigPath,
		Value: proxy.Spec.ProxyConfigPath,
	}, {
		Name:  cons.EnvProxyMode,
		Value: proxy.Spec.ProxyMode,
	}}
	return envs
}
