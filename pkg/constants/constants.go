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

// BrokerContainerName is the name of broker container
const BrokerContainerName = "broker"

// BasicCommand is basic command of exec function
const BasicCommand = "sh"

// AdminToolDir is the RocketMQ Admin directory in operator image
const AdminToolDir = "/home/rocketmq/rocketmq-4.5.0/bin/mqadmin"

// StoreConfigDir is the directory of config file
const StoreConfigDir = "/home/rocketmq/store/config"

// TopicJsonDir is the directory of topics.json
const TopicJsonDir = "/home/rocketmq/store/config/topics.json"

// SubscriptionGroupJsonDir is the directory of subscriptionGroup.json
const SubscriptionGroupJsonDir = "/home/rocketmq/store/config/subscriptionGroup.json"

// UpdateBrokerConfig is update broker config command
const UpdateBrokerConfig = "updateBrokerConfig"

// ParamNameServiceAddress is the name of name server list parameter
const ParamNameServiceAddress = "namesrvAddr"

// EnvNameServiceAddress is the container environment variable name of name server list
const EnvNameServiceAddress = "NAMESRV_ADDR"

// EnvReplicationMode is the container environment variable name of replication mode
const EnvReplicationMode = "REPLICATION_MODE"

// EnvBrokerId is the container environment variable name of broker id
const EnvBrokerId = "BROKER_ID"

// EnvBrokerClusterName is the container environment variable name of broker cluster name
const EnvBrokerClusterName = "BROKER_CLUSTER_NAME"

// EnvBrokerName is the container environment variable name of broker name
const EnvBrokerName = "BROKER_NAME"

// LogMountPath is the directory of RocketMQ log files
const LogMountPath = "/home/rocketmq/logs"

// StoreMountPath is the directory of RocketMQ store files
const StoreMountPath = "/home/rocketmq/store"

// LogSubPathName is the sub-path name of log dir under mounted host dir
const LogSubPathName = "logs"

// StoreSubPathName is the sub-path name of store dir under mounted host dir
const StoreSubPathName = "store"

// NameServiceMainContainerPort is the main port number of name server container
const NameServiceMainContainerPort = 9876

// NameServiceMainContainerPortName is the main port name of name server container
const NameServiceMainContainerPortName = "main"

// BrokerVipContainerPort is the VIP port number of broker container
const BrokerVipContainerPort = 10909

// BrokerVipContainerPortName is the VIP port name of broker container
const BrokerVipContainerPortName = "vip"

// BrokerMainContainerPort is the main port number of broker container
const BrokerMainContainerPort = 10911

// BrokerMainContainerPortName is the main port name of broker container
const BrokerMainContainerPortName = "main"

// BrokerHighAvailabilityContainerPort is the high availability port number of broker container
const BrokerHighAvailabilityContainerPort = 10912

// BrokerHighAvailabilityContainerPortName is the high availability port name of broker container
const BrokerHighAvailabilityContainerPortName = "ha"

// StorageModeNFS is the name of NFS storage mode
const StorageModeNFS = "NFS"

// StorageModeEmptyDir is the name of EmptyDir storage mode
const StorageModeEmptyDir = "EmptyDir"

// StorageModeHostPath is the name pf HostPath storage mode
const StorageModeHostPath = "HostPath"

// RestartBrokerPodIntervalInSecond is restart broker pod interval in second
const RestartBrokerPodIntervalInSecond = 30

// MinMetadataJsonFileSize is the threshold value if file length is lower than this will be considered as invalid
const MinMetadataJsonFileSize = 5

// MinIpListLength is the threshold value if the name server list parameter length is shorter than this will be considered as invalid
const MinIpListLength = 8

// CheckConsumeFinishIntervalInSecond is the interval of checking whether the consumption process is finished in second
const CheckConsumeFinishIntervalInSecond = 5

// RequeueIntervalInSecond is an universal interval of the reconcile function
const RequeueIntervalInSecond = 6

// Topic is the topic field index of the output when using command check consume progress
const Topic = 0

// BrokerName is the broker name field index of the output when using command check consume progress
const BrokerName = 1

// Diff is the diff field index of the output when using command check consume progress
const Diff = 6

// TopicListTopic is the topic field index of the output when using command check topic list
const TopicListTopic = 1

// TopicListConsumerGroup is the consumer group field index of the output when using command check topic list
const TopicListConsumerGroup = 2
