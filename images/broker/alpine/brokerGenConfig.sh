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

BROKER_CONFIG_FILE="$ROCKETMQ_HOME/conf/broker.conf"
BROKER_CONFIG_MOUNT_FILE="$ROCKETMQ_HOME/conf/broker-common.conf"

function create_config() {
    rm -f $BROKER_CONFIG_FILE
    echo "Creating broker configuration."
    # Remove brokerClusterName, brokerName, brokerId if configured
    sed -e '/brokerClusterName/d;/brokerName/d;/brokerId/d' $BROKER_CONFIG_MOUNT_FILE > $BROKER_CONFIG_FILE
    echo -e >> $BROKER_CONFIG_FILE
    echo "brokerClusterName=$BROKER_CLUSTER_NAME" >> $BROKER_CONFIG_FILE
    echo "brokerName=$BROKER_NAME" >> $BROKER_CONFIG_FILE
    echo "brokerId=$BROKER_ID" >> $BROKER_CONFIG_FILE
    echo "brokerIP1=`hostname -i`" >> $BROKER_CONFIG_FILE
    if [ $BROKER_ID != 0 ]; then
        sed -i 's/brokerRole=.*/brokerRole=SLAVE/g' $BROKER_CONFIG_FILE
    fi

    if [ "${enableControllerMode}" = "true" ]; then
        echo "enableControllerMode=true" >> $BROKER_CONFIG_FILE
        echo "controllerAddr=${controllerAddr}" >> $BROKER_CONFIG_FILE
    fi
}

create_config
cat $BROKER_CONFIG_FILE