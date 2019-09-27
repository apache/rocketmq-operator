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

package constants

// BrokerContainerName: the name of broker container
const BrokerContainerName = "broker"

// BasicCommand: basic command of exec function
const BasicCommand = "sh"

// AdminToolDir: the RocketMQ Admin directory in operator image
const AdminToolDir = "/home/rocketmq/rocketmq-4.5.0/bin/mqadmin"

// StoreConfigDir: the directory of config file
const StoreConfigDir = "/home/rocketmq/store/config"

// TopicJsonDir: the directory of topics.json
const TopicJsonDir = "/home/rocketmq/store/config/topics.json"

// SubscriptionGroupJsonDir: the directory of subscriptionGroup.json
const SubscriptionGroupJsonDir = "/home/rocketmq/store/config/subscriptionGroup.json"

// UpdateBrokerConfig: update broker config command
const UpdateBrokerConfig = "updateBrokerConfig"

// ParamNameServiceAddress: the name of name server list parameter
const ParamNameServiceAddress = "namesrvAddr"

// EnvNameServiceAddress: the container environment variable name of name server list
const EnvNameServiceAddress = "NAMESRV_ADDR"

// EnvReplicationMode: the container environment variable name of replication mode
const EnvReplicationMode = "REPLICATION_MODE"

// EnvBrokerId: the container environment variable name of broker id
const EnvBrokerId = "BROKER_ID"

// EnvBrokerClusterName: the container environment variable name of broker cluster name
const EnvBrokerClusterName = "BROKER_CLUSTER_NAME"

// EnvBrokerName: the container environment variable name of broker name
const EnvBrokerName = "BROKER_NAME"

// LogMountPath: the directory of RocketMQ log files
const LogMountPath = "/home/rocketmq/logs"

// StoreMountPath: the directory of RocketMQ store files
const StoreMountPath = "/home/rocketmq/store"

// LogSubPathName: the sub-path name of log dir under mounted host dir
const LogSubPathName = "logs"

// StoreSubPathName: the sub-path name of store dir under mounted host dir
const StoreSubPathName = "store"

// NameServiceMainContainerPort: the main port number of name server container
const NameServiceMainContainerPort = 9876

// NameServiceMainContainerPort: the main port name of name server container
const NameServiceMainContainerPortName = "main"

// BrokerVipContainerPort: the VIP port number of broker container
const BrokerVipContainerPort = 10909

// BrokerVipContainerPortName: the VIP port name of broker container
const BrokerVipContainerPortName = "vip"

// BrokerMainContainerPort: the main port number of broker container
const BrokerMainContainerPort = 10911

// BrokerMainContainerPortName: the main port name of broker container
const BrokerMainContainerPortName = "main"

// BrokerHighAvailabilityContainerPort: the high availability port number of broker container
const BrokerHighAvailabilityContainerPort = 10912

// BrokerHighAvailabilityContainerPortName: the high availability port name of broker container
const BrokerHighAvailabilityContainerPortName = "ha"

// StorageModeNFS: the name of NFS storage mode
const StorageModeNFS = "NFS"

// StorageModeEmptyDir: the name of EmptyDir storage mode
const StorageModeEmptyDir = "EmptyDir"

// StorageModeHostPath: the name pf HostPath storage mode
const StorageModeHostPath = "HostPath"

// RestartBrokerPodIntervalInSecond: restart broker pod interval in second
const RestartBrokerPodIntervalInSecond = 30

// MinMetadataJsonFileSize: if file length is lower than this will be considered as invalid
const MinMetadataJsonFileSize = 5

// MinIpListLength: if the name server list parameter length is shorter than this will be considered as invalid
const MinIpListLength = 8

// CheckConsumeFinishIntervalInSecond: the interval of checking whether the consumption process is finished in second
const CheckConsumeFinishIntervalInSecond = 5

// RequeueIntervalInSecond: an universal interval of the reconcile function
const RequeueIntervalInSecond = 6

// Topic: the topic field index of the output when using command check consume progress
const Topic = 0

// BrokerName: the broker name field index of the output when using command check consume progress
const BrokerName = 1

// Diff: the diff field index of the output when using command check consume progress
const Diff = 6

// TopicListTopic: the topic field index of the output when using command check topic list
const TopicListTopic = 1

// TopicListConsumerGroup: the consumer group field index of the output when using command check topic list
const TopicListConsumerGroup = 2
