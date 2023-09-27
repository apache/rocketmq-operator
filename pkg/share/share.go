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

// Package share defines some variables shared by different packages
package share

import (
	"context"
	"sort"
	"strings"

	rocketmqv1alpha1 "github.com/apache/rocketmq-operator/pkg/apis/rocketmq/v1alpha1"
	"github.com/apache/rocketmq-operator/pkg/tool"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetNameServersStr(r client.Reader, namespace, rocketMqName string) string {
	nameserviceList := &rocketmqv1alpha1.NameServiceList{}
	err := r.List(context.TODO(), nameserviceList, &client.MatchingFields{
		rocketmqv1alpha1.NameServiceRocketMqNameIndexKey: rocketMqName + "-" + namespace,
	})
	if err != nil {
		return ""
	}
	if len(nameserviceList.Items) != 1 {
		return ""
	}

	nameservice := nameserviceList.Items[0]
	labelSelector := labels.SelectorFromSet(tool.LabelsForNameService(nameservice.Name))
	listOps := &client.ListOptions{
		Namespace:     nameservice.Namespace,
		LabelSelector: labelSelector,
	}
	podList := &corev1.PodList{}
	err = r.List(context.Background(), podList, listOps)
	if err != nil {
		return ""
	}
	if len(podList.Items) == 0 {
		return ""
	}

	var hostIps []string
	for _, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodRunning && !strings.EqualFold(pod.Status.PodIP, "") {
			hostIps = append(hostIps, pod.Status.PodIP)
		}
	}
	sort.Strings(hostIps)

	nameServerListStr := ""
	for _, value := range hostIps {
		nameServerListStr = nameServerListStr + value + ":9876;"
	}

	return nameServerListStr[:len(nameServerListStr)-1]
}
