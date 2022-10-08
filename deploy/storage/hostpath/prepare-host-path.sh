#!/bin/bash

# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

### Instruction
### This script helps you create the hostPath dir on your localhost. You may need sudo privilege to run this script!
### You may need to run this script on every node of your Kubernetes cluster to prepare hostPath.

### You can modify the following parameters according to your real deployment requirements ###
# NAME_SERVER_DATA_PATH should be the same with the hostPath field in rocketmq_v1alpha1_nameservice_cr.yaml
NAME_SERVER_DATA_PATH=/data/rocketmq/nameserver
# BROKER_DATA_PATH should be the same with the hostPath field in rocketmq_v1alpha1_broker_cr.yaml
BROKER_DATA_PATH=/data/rocketmq/broker
# CONTROLLER_DATA_PATH should be the same with the hostPath field in rocketmq_v1alpha1_controller_cr.yaml
CONTROLLER_DATA_PATH=/data/rocketmq/controller
# ROCKETMQ_UID and ROCKETMQ_GID should be the same with docker image settings
ROCKETMQ_UID=3000
ROCKETMQ_GID=3000


### prepare hostPath dir for RocketMQ-Operator
prepare_dir()
{
  mkdir -p $1
  chown -R $2  $1
  chgrp -R $3  $1
  chmod -R a+rw $1
}

prepare_dir $BROKER_DATA_PATH $ROCKETMQ_UID $ROCKETMQ_GID
prepare_dir $NAME_SERVER_DATA_PATH $ROCKETMQ_UID $ROCKETMQ_GID
prepare_dir $CONTROLLER_DATA_PATH $ROCKETMQ_UID $ROCKETMQ_GID

echo "Changed hostPath $NAME_SERVER_DATA_PATH uid to $ROCKETMQ_UID, gid to $ROCKETMQ_GID"
echo "Changed hostPath $BROKER_DATA_PATH uid to $ROCKETMQ_UID, gid to $ROCKETMQ_GID"
echo "Changed hostPath $CONTROLLER_DATA_PATH uid to $ROCKETMQ_UID, gid to $ROCKETMQ_GID"
