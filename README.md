## RocketMQ Operator
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

## Overview

RocketMQ Operator is  to manage RocketMQ service instances deployed on the Kubernetes cluster.
It is built using the [Operator SDK](https://github.com/operator-framework/operator-sdk), which is part of the [Operator Framework](https://github.com/operator-framework/).

## Quick Start

### Define Your RocketMQ Cluster

RocketMQ Operator provides several CRDs to allow users define their RocketMQ service component cluster, which includes the Namesrv cluster and the Broker cluster.

1. Check the file ```rocketmq_v1alpha1_metaservice_cr.yaml``` in the ```deploy/crds``` directory, for example:
```
apiVersion: cache.example.com/v1alpha1
apiVersion: rocketmq.operator.com/v1alpha1
kind: MetaService
metadata:
  name: meta-service
spec:
  # size is the the name service instance number of the name service cluster
  size: 2
  # metaServiceImage is the customized docker image repo of the RocketMQ name service
  metaServiceImage: rocketmqinc/rocketmq-namesrv:4.5.0-alpine
  # imagePullPolicy is the image pull policy
  imagePullPolicy: Always
```

which defines the RocketMQ name service (namesrv) cluster scale.

2. Check the file ```cache_v1alpha1_broker_cr.yaml``` in the ```deploy/crds``` directory, for example:
```
apiVersion: cache.example.com/v1alpha1
kind: Broker
metadata:
  name: broker
spec:
  # size is the number of the broker cluster, each broker cluster contains a master broker and [slavePerGroup] slave brokers.
  size: 2
  # nameServers is the [ip:port] list of name service
  nameServers: 192.168.130.33:9876;192.168.130.34:9876
  # replicationMode is the broker slave sync mode, can be ASYNC or SYNC
  replicationMode: ASYNC
  # slavePerGroup is the number of each broker cluster
  slavePerGroup: 2
  # brokerImage is the customized docker image repo of the RocketMQ broker
  brokerImage: rocketmqinc/rocketmq-broker:4.5.0-alpine
  # imagePullPolicy is the image pull policy
  imagePullPolicy: Always
``` 
which defines the RocketMQ broker cluster scale, the [ip:port] list of name service and so on.

### Deploy RocketMQ Operator & Create RocketMQ Cluster

To deploy the RocketMQ Operator on your Kubernetes cluster, please run the following script:

```
$ ./install-operator.sh
```

Use command ```kubectl get pods``` to check the RocketMQ Operator deploy status like:

```
NAME                                READY   STATUS    RESTARTS   AGE
rocketmq-operator-564b5d75d-rls5n   1/1     Running   0          13s
```
Now you can use the CRDs provide by RocketMQ Operator to deploy your RocketMQ cluster.
 
Deploy the RocketMQ name service cluster by running:

``` 
$ kubectl apply -f deploy/crds/rocketmq_v1alpha1_metaservice_cr.yaml 
metaservice.rocketmq.operator.com/meta-service created
```

Check the status:

```
$ kubectl get pods -owide
NAME                                READY   STATUS    RESTARTS   AGE     IP               NODE        NOMINATED NODE   READINESS GATES
meta-service-d4f5796d-7r9f9         1/1     Running   0          8s      192.168.130.34   k2data-14   <none>           <none>
meta-service-d4f5796d-kr2qg         1/1     Running   0          8s      192.168.130.33   k2data-13   <none>           <none>
rocketmq-operator-564b5d75d-qnpts   1/1     Running   1          3m33s   10.244.1.68      k2data-13   <none>           <none>
```

We can see that there are 2 name service Pods running on 2 nodes and their IP addresses. Modify the ```nameServers``` field in the ```cache_v1alpha1_broker_cr.yaml``` file using the IP addresses.

Deploy the RocketMQ broker clusters by running:
```
$ kubectl apply -f deploy/crds/cache_v1alpha1_broker_cr.yaml 
broker.cache.example.com/broker created 
```

After a while after the Containers are created, the Kubernetes clusters status should be like:

``` 
$ kubectl get pods -owide
NAME                                READY   STATUS    RESTARTS   AGE     IP               NODE        NOMINATED NODE   READINESS GATES
broker-0-master-56b9c84748-4b6fv    1/1     Running   0          20s     10.244.2.64      k2data-14   <none>           <none>
broker-0-slave-1-6c4c44d7d8-62pv8   1/1     Running   0          20s     10.244.2.65      k2data-14   <none>           <none>
broker-0-slave-2-56484bd645-s75dh   1/1     Running   0          20s     10.244.1.69      k2data-13   <none>           <none>
broker-1-master-7b54bd95cb-s2275    1/1     Running   0          20s     10.244.2.67      k2data-14   <none>           <none>
broker-1-slave-1-56d65c89cb-rbrcv   1/1     Running   0          20s     10.244.1.70      k2data-13   <none>           <none>
broker-1-slave-2-dc4d8d88f-sjkk7    1/1     Running   0          20s     10.244.2.66      k2data-14   <none>           <none>
meta-service-d4f5796d-7r9f9         1/1     Running   0          6m46s   192.168.130.34   k2data-14   <none>           <none>
meta-service-d4f5796d-kr2qg         1/1     Running   0          6m46s   192.168.130.33   k2data-13   <none>           <none>
rocketmq-operator-564b5d75d-qnpts   1/1     Running   1          10m     10.244.1.68      k2data-13   <none>           <none>
```

Congratulations! You have successfully deployed your RocketMQ cluster by RocketMQ Operator.

### Clean the Environment
If you want to tear down the RocketMQ cluster, run
```
$ kubectl delete -f deploy/crds/cache_v1alpha1_broker_cr.yaml 
```

to remove the broker clusters

```
$ kubectl delete -f deploy/crds/rocketmq_v1alpha1_metaservice_cr.yaml
```

to remove the name service clusters

```
$ ./purge-operator.sh
```

to remove the RocketMQ Operator.