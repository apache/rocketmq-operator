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

#===========================================================================================
# Java Environment Setting
#===========================================================================================
error_exit ()
{
    echo "ERROR: $1 !!"
    exit 1
}

find_java_home()
{
    case "`uname`" in
        Darwin)
            JAVA_HOME=$(/usr/libexec/java_home)
        ;;
        *)
            JAVA_HOME=$(dirname $(dirname $(readlink -f $(which javac))))
        ;;
    esac
}

find_java_home

[ ! -e "$JAVA_HOME/bin/java" ] && JAVA_HOME=$HOME/jdk/java
[ ! -e "$JAVA_HOME/bin/java" ] && JAVA_HOME=/usr/java
[ ! -e "$JAVA_HOME/bin/java" ] && error_exit "Please set the JAVA_HOME variable in your environment, We need java(x64)!"

export JAVA_HOME
export JAVA="$JAVA_HOME/bin/java"
export BASE_DIR=$(dirname $0)/..
export CLASSPATH=.:${BASE_DIR}/conf:${CLASSPATH}

#===========================================================================================
# JVM Configuration
#===========================================================================================
calculate_heap_sizes()
{
    case "`uname`" in
        Linux)
            system_memory_in_mb=`free -m| sed -n '2p' | awk '{print $2}'`
            system_memory_in_mb_in_docker=$(($(cat /sys/fs/cgroup/memory/memory.limit_in_bytes)/1024/1024))
            if [ $system_memory_in_mb_in_docker -lt $system_memory_in_mb ];then
              system_memory_in_mb=$system_memory_in_mb_in_docker
            fi
            system_cpu_cores=`egrep -c 'processor([[:space:]]+):.*' /proc/cpuinfo`
            system_cpu_cores_in_docker=$(($(cat /sys/fs/cgroup/cpu/cpu.cfs_quota_us)/$(cat /sys/fs/cgroup/cpu/cpu.cfs_period_us)))
            if [ $system_cpu_cores_in_docker -lt $system_cpu_cores -a $system_cpu_cores_in_docker -ne 0 ];then
              system_cpu_cores=$system_cpu_cores_in_docker
            fi
        ;;
        FreeBSD)
            system_memory_in_bytes=`sysctl hw.physmem | awk '{print $2}'`
            system_memory_in_mb=`expr $system_memory_in_bytes / 1024 / 1024`
            system_cpu_cores=`sysctl hw.ncpu | awk '{print $2}'`
        ;;
        SunOS)
            system_memory_in_mb=`prtconf | awk '/Memory size:/ {print $3}'`
            system_cpu_cores=`psrinfo | wc -l`
        ;;
        Darwin)
            system_memory_in_bytes=`sysctl hw.memsize | awk '{print $2}'`
            system_memory_in_mb=`expr $system_memory_in_bytes / 1024 / 1024`
            system_cpu_cores=`sysctl hw.ncpu | awk '{print $2}'`
        ;;
        *)
            # assume reasonable defaults for e.g. a modern desktop or
            # cheap server
            system_memory_in_mb="2048"
            system_cpu_cores="2"
        ;;
    esac

    # some systems like the raspberry pi don't report cores, use at least 1
    if [ "$system_cpu_cores" -lt "1" ]
    then
        system_cpu_cores="1"
    fi

    # set max heap size based on the following
    # max(min(1/2 ram, 1024MB), min(1/4 ram, 8GB))
    # calculate 1/2 ram and cap to 1024MB
    # calculate 1/4 ram and cap to 8192MB
    # pick the max
    half_system_memory_in_mb=`expr $system_memory_in_mb / 2`
    quarter_system_memory_in_mb=`expr $half_system_memory_in_mb / 2`
    if [ "$half_system_memory_in_mb" -gt "1024" ]
    then
        half_system_memory_in_mb="1024"
    fi
    if [ "$quarter_system_memory_in_mb" -gt "8192" ]
    then
        quarter_system_memory_in_mb="8192"
    fi
    if [ "$half_system_memory_in_mb" -gt "$quarter_system_memory_in_mb" ]
    then
        max_heap_size_in_mb="$half_system_memory_in_mb"
    else
        max_heap_size_in_mb="$quarter_system_memory_in_mb"
    fi
    MAX_HEAP_SIZE="${max_heap_size_in_mb}M"

    # Young gen: min(max_sensible_per_modern_cpu_core * num_cores, 1/4 * heap size)
    max_sensible_yg_per_core_in_mb="100"
    max_sensible_yg_in_mb=`expr $max_sensible_yg_per_core_in_mb "*" $system_cpu_cores`

    desired_yg_in_mb=`expr $max_heap_size_in_mb / 4`

    if [ "$desired_yg_in_mb" -gt "$max_sensible_yg_in_mb" ]
    then
        HEAP_NEWSIZE="${max_sensible_yg_in_mb}M"
    else
        HEAP_NEWSIZE="${desired_yg_in_mb}M"
    fi
}

calculate_heap_sizes

# Dynamically calculate parameters, for reference.
Xms=$MAX_HEAP_SIZE
Xmx=$MAX_HEAP_SIZE
Xmn=$HEAP_NEWSIZE

# For controller deployed in K8s
if test -n "$MY_POD_NAME" ;then
    echo "MY_POD_NAME=$MY_POD_NAME"
    controllerDLegerSelfIdValue=$(echo "$MY_POD_NAME" | tr -d -)
    export controllerDLegerSelfId=$controllerDLegerSelfIdValue
fi

# Set for `JAVA_OPT`.
JAVA_OPT="${JAVA_OPT} -server -Xms${Xms} -Xmx${Xmx} -Xmn${Xmn}"
JAVA_OPT="${JAVA_OPT} -XX:+UseConcMarkSweepGC -XX:+UseCMSCompactAtFullCollection -XX:CMSInitiatingOccupancyFraction=70 -XX:+CMSParallelRemarkEnabled -XX:SoftRefLRUPolicyMSPerMB=0 -XX:+CMSClassUnloadingEnabled -XX:SurvivorRatio=8  -XX:-UseParNewGC"
JAVA_OPT="${JAVA_OPT} -verbose:gc -Xloggc:/dev/shm/rmq_srv_gc.log -XX:+PrintGCDetails"
JAVA_OPT="${JAVA_OPT} -XX:-OmitStackTraceInFastThrow"
JAVA_OPT="${JAVA_OPT}  -XX:-UseLargePages"
JAVA_OPT="${JAVA_OPT} -Djava.ext.dirs=${JAVA_HOME}/jre/lib/ext:${BASE_DIR}/lib"
JAVA_OPT="${JAVA_OPT} ${JAVA_OPT_EXT}"
JAVA_OPT="${JAVA_OPT} -cp ${CLASSPATH}"

echo "=============env beg============="
env
echo "=============env end============="

#TODO cp from config map
cp ${ROCKETMQ_HOME}/conf/controller/controller-standalone.conf ${ROCKETMQ_HOME}/conf/controller/controller.conf


sed -i "s/^controllerDLegerGroup.*$/controllerDLegerGroup=${controllerDLegerGroup}/" ${ROCKETMQ_HOME}/conf/controller/controller.conf

sed -i "s/^controllerDLegerPeers.*$/controllerDLegerPeers=${controllerDLegerPeers}/" ${ROCKETMQ_HOME}/conf/controller/controller.conf

sed -i "s/^controllerDLegerSelfId.*$/controllerDLegerSelfId=${controllerDLegerSelfId}/" ${ROCKETMQ_HOME}/conf/controller/controller.conf

#perl not exists, sed can't process 
#perl -pi -e "s|controllerStorePath.*$|controllerStorePath=${controllerStorePath}|g" ${ROCKETMQ_HOME}/conf/controller/controller.conf
echo "" >> ${ROCKETMQ_HOME}/conf/controller/controller.conf
echo "controllerStorePath=${controllerStorePath}" >> ${ROCKETMQ_HOME}/conf/controller/controller.conf

$JAVA ${JAVA_OPT} $@
