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

// Package rocketmq contains the implementation of the Rocketmq CRD reconcile function
package rocketmq

import (
	"context"
	rocketmqv1alpha1 "github.com/apache/rocketmq-operator/pkg/apis/rocketmq/v1alpha1"
	cons "github.com/apache/rocketmq-operator/pkg/constants"
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

var log = logf.Log.WithName("controller_rocketmq")

// Add creates a new Rocketmq Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRocketmq{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("rocketmq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// Watch for changes to primary resource Rocketmq
	err = c.Watch(&source.Kind{Type: &rocketmqv1alpha1.Rocketmq{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	// Watch for changes to secondary resource Pods and requeue the owner Rocketmq
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType: &rocketmqv1alpha1.Rocketmq{},
	})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileRocketmq implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRocketmq{}

// ReconcileRocketmq reconciles a Rocketmq object
type ReconcileRocketmq struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NameService object and makes changes based on the state read
// and what is in the NameService.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRocketmq) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rocketmq")
	// Fetch the Rocketmq instance
	instance := &rocketmqv1alpha1.Rocketmq{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object, requeue the request
		return reconcile.Result{}, err
	}

	// Check if nameserver already exist, if not create a new one
	nameServiceFound := &rocketmqv1alpha1.NameService{}
	nameServiceSts := r.nameServiceForRocketmq(instance)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: nameServiceSts.Name, Namespace: nameServiceSts.Namespace}, nameServiceFound)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), nameServiceSts)
		if err != nil {
			reqLogger.Error(err, "Failed to create new nameService of rocketmq", "nameservice.namespace",
				nameServiceSts.Namespace, "nameservice.Name", nameServiceSts.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get rocketmq nameservice.")
		return reconcile.Result{}, err
	} else {
		// Resource NameService will change; Only size nameServiceImage imagePullPolicy can update
		if !reflect.DeepEqual(nameServiceSts.Spec.Size, nameServiceFound.Spec.Size) ||
			!reflect.DeepEqual(nameServiceSts.Spec.NameServiceImage, nameServiceFound.Spec.NameServiceImage) ||
			!reflect.DeepEqual(nameServiceSts.Spec.ImagePullPolicy, nameServiceFound.Spec.ImagePullPolicy) {
			nameServiceFound.Spec.ImagePullPolicy = nameServiceSts.Spec.ImagePullPolicy
			nameServiceFound.Spec.NameServiceImage = nameServiceSts.Spec.NameServiceImage
			nameServiceFound.Spec.Size = nameServiceSts.Spec.Size
			err = r.client.Update(context.TODO(), nameServiceFound)
			if err != nil {
				reqLogger.Error(err, "should update nameservice resource", "spec.nameService",
					nameServiceSts.Spec, "nameServiceFound.spec", nameServiceFound.Spec)
				return reconcile.Result{}, err
			}
		}
	}

	// Check if broker already exists, if not create a new one
	brokerFound := &rocketmqv1alpha1.Broker{}
	brokerSts := r.brokerForRocketmq(instance)

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: brokerSts.Name, Namespace: brokerSts.Namespace}, brokerFound)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), brokerSts)
		if err != nil {
			reqLogger.Error(err, "Failed to create new broker of rocketmq", "broker.namespace",
				brokerSts.Namespace, "broker.Name", brokerSts.Name)
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get rocketmq broker.")
	} else {
		// Resource broker will change; Only ReplicaPerGroup Size ImagePullPolicy BrokerImage can update
		if !reflect.DeepEqual(brokerSts.Spec.ReplicaPerGroup, brokerFound.Spec.ReplicaPerGroup) ||
			!reflect.DeepEqual(brokerSts.Spec.Size, brokerFound.Spec.Size) ||
			!reflect.DeepEqual(brokerSts.Spec.ImagePullPolicy, brokerFound.Spec.ImagePullPolicy) ||
			!reflect.DeepEqual(brokerSts.Spec.BrokerImage, brokerFound.Spec.BrokerImage) {
			brokerFound.Spec.ReplicaPerGroup = brokerSts.Spec.ReplicaPerGroup
			brokerFound.Spec.Size = brokerSts.Spec.Size
			brokerFound.Spec.ImagePullPolicy = brokerSts.Spec.ImagePullPolicy
			brokerFound.Spec.BrokerImage = brokerSts.Spec.BrokerImage
			err = r.client.Update(context.TODO(), brokerFound)
			if err != nil {
				reqLogger.Error(err, "should update broker resource", "spec.Broker",
					brokerSts.Spec, "brokerFound.spec", brokerFound.Spec)
				return reconcile.Result{}, err
			}
		}
	}
	// update instance status
	if !reflect.DeepEqual(instance.Status.Broker, brokerFound.ObjectMeta.Name) ||
		!reflect.DeepEqual(instance.Status.NameService, nameServiceFound.ObjectMeta.Name) {
		instance.Status.Broker = brokerFound.ObjectMeta.Name
		instance.Status.NameService = nameServiceFound.ObjectMeta.Name
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "update rocketmq status failed")
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, nil
}

func (r *ReconcileRocketmq) nameServiceForRocketmq(rocketmq *rocketmqv1alpha1.Rocketmq) *rocketmqv1alpha1.NameService {
	nameService := &rocketmqv1alpha1.NameService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NameService",
			APIVersion: rocketmq.TypeMeta.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   rocketmq.Name + "-name-service",
			Namespace: rocketmq.Namespace,
		},
		Spec: rocketmq.Spec.NameService,
	}
	controllerutil.SetControllerReference(rocketmq, nameService, r.scheme)
	return nameService
}

func (r *ReconcileRocketmq) brokerForRocketmq(rocketmq *rocketmqv1alpha1.Rocketmq) *rocketmqv1alpha1.Broker {
	broker := &rocketmqv1alpha1.Broker{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Broker",
			APIVersion: rocketmq.TypeMeta.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta {
			Name: rocketmq.Name + "-broker",
			Namespace: rocketmq.Namespace,
		},
		Spec: rocketmq.Spec.Broker,
	}
	controllerutil.SetControllerReference(rocketmq, broker, r.scheme)
	return broker
}
