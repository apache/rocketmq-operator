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

package console

import (
	"context"
	"fmt"
	rocketmqv1alpha1 "github.com/apache/rocketmq-operator/pkg/apis/rocketmq/v1alpha1"
	cons "github.com/apache/rocketmq-operator/pkg/constants"
	"github.com/apache/rocketmq-operator/pkg/share"
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
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var log = logf.Log.WithName("controller_console")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Console Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileConsole{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("console-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Console
	err = c.Watch(&source.Kind{Type: &rocketmqv1alpha1.Console{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Console
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rocketmqv1alpha1.Console{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileConsole implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileConsole{}

// ReconcileConsole reconciles a Console object
type ReconcileConsole struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Console object and makes changes based on the state read
// and what is in the Console.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileConsole) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Console")

	// Fetch the Console instance
	instance := &rocketmqv1alpha1.Console{}
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

	if instance.Spec.NameServers == "" {
		// wait for name server ready if nameServers is omitted
		for {
			if share.IsNameServersStrInitialized {
				break
			} else {
				log.Info("Waiting for name server ready...")
				time.Sleep(time.Duration(cons.WaitForNameServerReadyInSecond) * time.Second)
			}
		}
	} else {
		share.NameServersStr = instance.Spec.NameServers
	}

	consoleDeployment := r.newDeploymentForCR(instance)

	// Set Console instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, consoleDeployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: consoleDeployment.Name, Namespace: consoleDeployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating RocketMQ Console Deployment", "Namespace", consoleDeployment, "Name", consoleDeployment.Name)
		err = r.client.Create(context.TODO(), consoleDeployment)
		if err != nil {
			return reconcile.Result{}, err
		}

		// created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Support console deployment scaling
	if !reflect.DeepEqual(instance.Spec.ConsoleDeployment.Spec.Replicas, found.Spec.Replicas) {
		found.Spec.Replicas = instance.Spec.ConsoleDeployment.Spec.Replicas
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update console CR ", "Namespace", found.Namespace, "Name", found.Name)
		} else {
			reqLogger.Info("Successfully updated console CR ", "Namespace", found.Namespace, "Name", found.Name)
		}
	}

	// TODO: update console if name server address changes

	// CR already exists - don't requeue
	reqLogger.Info("Skip reconcile: RocketMQ Console Deployment already exists", "Namespace", found.Namespace, "Name", found.Name)
	return reconcile.Result{}, nil
}

// newDeploymentForCR returns a deployment pod with modifying the ENV
func (r *ReconcileConsole) newDeploymentForCR(cr *rocketmqv1alpha1.Console) *appsv1.Deployment {
	env := corev1.EnvVar{
		Name:  "JAVA_OPTS",
		Value: fmt.Sprintf("-Drocketmq.namesrv.addr=%s -Dcom.rocketmq.sendMessageWithVIPChannel=false", share.NameServersStr),
	}
	selectLabels, matchLabels := func() (map[string]string, map[string]string) {
		selectorLabels := cr.Spec.ConsoleDeployment.Spec.Selector.MatchLabels
		labels := cr.Spec.ConsoleDeployment.Spec.Template.Labels
		if selectorLabels != nil && labels != nil {
			selectorLabels["console-cr"] = cr.Name
			labels["console-cr"] = cr.Name
		}
		return selectorLabels, labels
	}()

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: cr.Spec.ConsoleDeployment.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Resources: cr.Spec.ConsoleDeployment.Spec.Template.Spec.Containers[0].Resources,
						Image: cr.Spec.ConsoleDeployment.Spec.Template.Spec.Containers[0].Image,
						Name:  cr.Spec.ConsoleDeployment.Spec.Template.Spec.Containers[0].Name,
						ImagePullPolicy: cr.Spec.ConsoleDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy,
						Env: append(cr.Spec.ConsoleDeployment.Spec.Template.Spec.Containers[0].Env, env),
						Ports: cr.Spec.ConsoleDeployment.Spec.Template.Spec.Containers[0].Ports,
					}},
				},
			},
		},
	}
	controllerutil.SetControllerReference(cr, dep, r.scheme)
	return dep
}
