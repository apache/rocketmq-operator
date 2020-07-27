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

set -eux;

# You can change the DOCKERHUB_REPO to your docker repo for development purpose
DOCKERHUB_REPO="apacherocketmq/rocketmq-operator:0.2.1"
# The version of RocketMQ including the Admin Tool
ROCKETMQ_VERSION="4.5.0"

export GO111MODULE=on

# Download rocketmq.zip for the Admin Tool if it does not exist
if [ ! -e build/rocketmq.zip ]; then
  curl https://archive.apache.org/dist/rocketmq/${ROCKETMQ_VERSION}/rocketmq-all-${ROCKETMQ_VERSION}-bin-release.zip -o rocketmq.zip
  mv rocketmq.zip build
fi

# use the following 2 commands if you have updated the [kind]_type.go file or don't have zz_generated.deepcopy.go and zz_generated.openapi.go files
operator-sdk generate k8s
operator-sdk generate openapi

go mod vendor

echo "Start building RocketMQ-Operator..."
operator-sdk build $DOCKERHUB_REPO

docker push $DOCKERHUB_REPO
