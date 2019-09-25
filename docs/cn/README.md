# RocketMQ Operator 用户手册

## 目录
- [概览](#overview)
- [Quick Start](#quick-start)
    - [Prepare Volume Persistence](#prepare-volume-persistence)
        - [Prepare HostPath](#prepare-hostpath)
        - [Prepare Storage Class of NFS](#prepare-storage-class-of-nfs)
    - [Define Your RocketMQ Cluster](#define-your-rocketmq-cluster)
    - [Deploy RocketMQ Operator & Create RocketMQ Cluster](#deploy-rocketmq-operator--create-rocketmq-cluster)
    - [Verify the Data Storage](#verify-the-data-storage)
        - [Verify HostPath Storage](#verify-hostpath-storage)
        - [Verify NFS storage](#verify-nfs-storage)
- [Horizontal Scale](#horizontal-scale)
    - [Name Server Cluster Scale](#name-server-cluster-scale)
    - [Broker Cluster Scale](#broker-cluster-scale)
        - [Up-scale Broker in Out-of-order Message Scenario](#up-scale-broker-in-out-of-order-message-scenario)
- [Topic Transfer](#topic-transfer)
- [Clean the Environment](#clean-the-environment)

## 概览

RocketMQ Operator是一款管理Kubernetes集群上部署的RocketMQ服务集群的工具。
它是使用[Operator SDK](https://github.com/operator-framework/operator-sdk)构建的，基于[Operator Framework](https://github.com/operator-framework/)提供的Operator标准框架。

![RocketMQ-Operator architecture](../img/operator-arch.png)

## 快速开始

### 持久化准备工作

我们支持多种持久化您的RocketMQ数据的方式，包括：```EmptyDir```, ```HostPath``` 以及 ```NFS```。 您可以通过在示例资源文件例如```rocketmq_v1alpha1_nameservice_cr.yaml```中配置来选择您想要的存储方式：

```
...
 # storageMode can be EmptyDir, HostPath, NFS
  storageMode: HostPath
...
```

如果您选择的是```EmptyDir```，那么您不需要做额外的准备工作来持久化数据。但是该方式下数据的持久化生命周期与对应的pod的生命周期相同，这意味着如果pod被删除您将有可能会丢失对应pod上的数据。

如果您选择的是其他持久化存储方式，请根据下面的指导完成相应的准备工作。

#### HostPath （宿主机本地路径）准备工作

该存储方式下，RocketMQ的数据（包括所有的日志和存储文件）会存储在每个pod对应所在的宿主机的本地路径上。 因此您需要事先准备好您希望挂载的本地路径及其统一名称。

为了方便您的准备工作，我们提供一个脚本```deploy/storage/hostpath/prepare-host-path.sh```用来帮助您创建```HostPath```本地路径：

```
$ cd deploy/storage/hostpath

$ sudo su

$ ./prepare-hostpath.sh 
Changed hostPath /data/rocketmq/nameserver uid to 3000, gid to 3000
Changed hostPath /data/rocketmq/broker uid to 3000, gid to 3000
```

> 注意： 您需要在每台可能被部署RocketMQ的Kubernetes节点上运行此脚本，以便在每台节点上都准备好```HostPath```路径。
> 
> 默认创建的路径名与```example```下的示例自定义资源配置文件中的一致。您可以查看```prepare-host-path.sh```脚本的内容以获得更多相关信息。

#### 基于NFS的StorageClass准备工作

如果您选择NFS作为您的RocketMQ数据存储方式，您将需要先在您的Kubernetes集群上部署好NFS服务端并在其他节点上部署好NFS客户端，然后准备一个基于```NFS provider```的StorageClass资源用来自动创建PV和PVC资源。如果您对以上概念不了解也没有关系，仅需按照以下步骤操作即可：

1. 在您的Kubernetes集群上部署好NFS服务端并在其他节点上部署好NFS客户端，部署方法可以参考[NFS部署文档](nfs_install_cn.md)。 请在下一步准备工作之前确保NFS服务工作正常，以下是一个简单的确认NFS工作正常的方法：

    1) 在您的NFS客户端节点上检查NFS共享路径是否存在：
    ```
   $ showmount -e 192.168.130.32
   Export list for 192.168.130.32:
   /data/k8s * 
    ```
    2) 在您的NFS客户端节点上创建一个测试路径，并将该路径挂载到NFS共享路径上(这一步您可能需要管理员权限)：
    ```
   $ mkdir -p   ~/test-nfc
   $ mount -t nfs 192.168.130.32:/data/k8s ~/test-nfc
    ```
    3) 在您的NFS客户端节点上，在上面的测试路径下创建一个测试文件：
    ```
   $ touch ~/test-nfc/test.txt
    ```
   4) 在您的NFS服务端节点上, 检查共享路径。 如果共享路径下存在您刚刚在NFS客户端节点上创建的测试文件，证明NFS是在正常工作的：
   ```
   $ ls -ls /data/k8s/
   total 4
   4 -rw-r--r--. 1 root root 4 Jul 10 21:50 test.txt
   ```

2. Modify the following configurations of the ```deploy/storage/nfs-client.yaml``` file:
``` 
...
            - name: NFS_SERVER
              value: 192.168.130.32
            - name: NFS_PATH
              value: /data/k8s
      volumes:
        - name: nfs-client-root
          nfs:
            server: 192.168.130.32
            path: /data/k8s
...
```
Replace ```192.168.130.32``` and ```/data/k8s``` with your true NFS server IP address and NFS server data volume path.
 
3. Create a NFS storage class for RocketMQ, run

```
$ cd deploy/storage
$ ./deploy-storage-class.sh
```
If the storage class is successfully deployed, you can get the pod status like:
```
$ kubectl get pods
NAME                                      READY   STATUS    RESTARTS   AGE
nfs-client-provisioner-7cf858f754-7vxmm   1/1     Running   0          136m
```

### Define Your RocketMQ Cluster

RocketMQ Operator provides several CRDs to allow users define their RocketMQ service component cluster, which includes the Namesrv cluster and the Broker cluster.

1. Check the file ```rocketmq_v1alpha1_nameservice_cr.yaml``` in the ```example``` directory, for example:
```
apiVersion: rocketmq.apache.org/v1alpha1
kind: NameService
metadata:
  name: name-service
spec:
  # size is the the name service instance number of the name service cluster
  size: 1
  # nameServiceImage is the customized docker image repo of the RocketMQ name service
  nameServiceImage: docker.io/library/rocketmq-namesrv:4.5.0-alpine
  # imagePullPolicy is the image pull policy
  imagePullPolicy: Always
  # storageMode can be EmptyDir, HostPath, NFS
  storageMode: HostPath
  # hostPath is the local path to store data
  hostPath: /data/rocketmq/nameserver
  # volumeClaimTemplates defines the storageClass
  volumeClaimTemplates:
    - metadata:
        name: namesrv-storage
        annotations:
          volume.beta.kubernetes.io/storage-class: rocketmq-storage
      spec:
        accessModes: [ "ReadWriteOnce" ]
        resources:
          requests:
            storage: 1Gi
```

which defines the RocketMQ name service (namesrv) cluster scale.

2. Check the file ```rocketmq_v1alpha1_broker_cr.yaml``` in the ```example``` directory, for example:
```
apiVersion: rocketmq.apache.org/v1alpha1
kind: Broker
metadata:
  # name of broker cluster
  name: broker
spec:
  # size is the number of the broker cluster, each broker cluster contains a master broker and [replicaPerGroup] replica brokers.
  size: 2
  # nameServers is the [ip:port] list of name service
  nameServers: 192.168.130.33:9876
  # replicationMode is the broker replica sync mode, can be ASYNC or SYNC
  replicationMode: ASYNC
  # replicaPerGroup is the number of each broker cluster
  replicaPerGroup: 1
  # brokerImage is the customized docker image repo of the RocketMQ broker
  brokerImage: docker.io/library/rocketmq-broker:4.5.0-alpine
  # imagePullPolicy is the image pull policy
  imagePullPolicy: Always
  # allowRestart defines whether allow pod restart
  allowRestart: true
  # storageMode can be EmptyDir, HostPath, NFS
  storageMode: HostPath
  # hostPath is the local path to store data
  hostPath: /data/rocketmq/broker
  # scalePodName is broker-[broker group number]-master-0
  scalePodName: broker-0-master-0
  # volumeClaimTemplates defines the storageClass
  volumeClaimTemplates:
    - metadata:
        name: broker-storage
        annotations:
          volume.beta.kubernetes.io/storage-class: rocketmq-storage
      spec:
        accessModes: [ "ReadWriteOnce" ]
        resources:
          requests:
            storage: 8Gi
``` 
which defines the RocketMQ broker cluster scale, the [ip:port] list of name service and so on.

### Deploy RocketMQ Operator & Create RocketMQ Cluster

1. To deploy the RocketMQ Operator on your Kubernetes cluster, please run the following script:

```
$ ./install-operator.sh
```

Use command ```kubectl get pods``` to check the RocketMQ Operator deploy status like:

```
$ kubectl get pods
NAME                                      READY   STATUS    RESTARTS   AGE
nfs-client-provisioner-7cf858f754-7vxmm   1/1     Running   0          146m
rocketmq-operator-564b5d75d-jllzk         1/1     Running   0          108s
```
Now you can use the CRDs provide by RocketMQ Operator to deploy your RocketMQ cluster.
 
2. Deploy the RocketMQ name service cluster by running:

``` 
$ kubectl apply -f example/rocketmq_v1alpha1_nameservice_cr.yaml 
nameservice.rocketmq.apache.org/name-service created
```

Check the status:

```
$ kubectl get pods -owide
NAME                                      READY   STATUS    RESTARTS   AGE     IP               NODE        NOMINATED NODE   READINESS GATES
name-service-0                            1/1     Running   0          3m18s   192.168.130.33   k2data-13   <none>           <none>
nfs-client-provisioner-7cf858f754-7vxmm   1/1     Running   0          150m    10.244.2.114     k2data-14   <none>           <none>
rocketmq-operator-564b5d75d-jllzk         1/1     Running   0          5m53s   10.244.2.116     k2data-14   <none>           <none>
```

We can see that there are 1 name service Pods running on 1 nodes and their IP addresses. Modify the ```nameServers``` field in the ```rocketmq_v1alpha1_broker_cr.yaml``` file using the IP addresses.

3. Deploy the RocketMQ broker clusters by running:
```
$ kubectl apply -f example/rocketmq_v1alpha1_broker_cr.yaml
broker.rocketmq.apache.org/broker created 
```

After a while after the Containers are created, the Kubernetes clusters status should be like:

``` 
$ kubectl get pods -owide
NAME                                      READY   STATUS    RESTARTS   AGE     IP               NODE        NOMINATED NODE   READINESS GATES
broker-0-master-0                         1/1     Running   0          38s     10.244.4.18      k2data-11   <none>           <none>
broker-0-replica-1-0                      1/1     Running   0          38s     10.244.1.128     k2data-13   <none>           <none>
broker-1-master-0                         1/1     Running   0          38s     10.244.2.117     k2data-14   <none>           <none>
broker-1-replica-1-0                      1/1     Running   0          38s     10.244.3.17      k2data-15   <none>           <none>
name-service-0                            1/1     Running   0          6m7s    192.168.130.33   k2data-13   <none>           <none>
nfs-client-provisioner-7cf858f754-7vxmm   1/1     Running   0          153m    10.244.2.114     k2data-14   <none>           <none>
rocketmq-operator-564b5d75d-jllzk         1/1     Running   0          8m42s   10.244.2.116     k2data-14   <none>           <none>
```

Check the PV and PVC status:
```
$ kubectl get pvc
NAME                                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS       AGE
broker-storage-broker-0-master-0        Bound    pvc-7a74871b-c005-441a-bb15-8106566c9d19   8Gi        RWO            rocketmq-storage   78s
broker-storage-broker-0-replica-1-0     Bound    pvc-521e7e9a-3795-487a-9f76-22da74db74dd   8Gi        RWO            rocketmq-storage   78s
broker-storage-broker-1-master-0        Bound    pvc-d7b76efe-384c-4f8d-9e8a-ebe209ba826c   8Gi        RWO            rocketmq-storage   78s
broker-storage-broker-1-replica-1-0     Bound    pvc-af266db9-83a9-4929-a2fe-e40fb5fdbfa4   8Gi        RWO            rocketmq-storage   78s
namesrv-storage-name-service-0          Bound    pvc-c708cb49-aa52-4992-8cac-f46a48e2cc2e   1Gi        RWO            rocketmq-storage   79s

$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                       STORAGECLASS       REASON   AGE
pvc-521e7e9a-3795-487a-9f76-22da74db74dd   8Gi        RWO            Delete           Bound    default/broker-storage-broker-0-replica-1-0 rocketmq-storage            79s
pvc-7a74871b-c005-441a-bb15-8106566c9d19   8Gi        RWO            Delete           Bound    default/broker-storage-broker-0-master-0    rocketmq-storage            79s
pvc-af266db9-83a9-4929-a2fe-e40fb5fdbfa4   8Gi        RWO            Delete           Bound    default/broker-storage-broker-1-replica-1-0 rocketmq-storage            78s
pvc-c708cb49-aa52-4992-8cac-f46a48e2cc2e   1Gi        RWO            Delete           Bound    default/namesrv-storage-name-service-0      rocketmq-storage            79s
pvc-d7b76efe-384c-4f8d-9e8a-ebe209ba826c   8Gi        RWO            Delete           Bound    default/broker-storage-broker-1-master-0    rocketmq-storage            78s
```

> Notice: if you don't choose the NFS storage mode, then the above PV and PVC won't be created.

Congratulations! You have successfully deployed your RocketMQ cluster by RocketMQ Operator.

### Verify the Data Storage

#### Verify HostPath Storage
Access on any node which contains the RocketMQ service pod, check the ```hostPath``` you configured, for example:
```
$ ls /data/rocketmq/broker
logs  store

$ cat /data/rocketmq/broker/logs/broker-1-replica-1/rocketmqlogs/broker.log
...
2019-09-12 13:12:24 INFO main - The broker[broker-1, 10.244.3.35:10911] boot success. serializeType=JSON and name server is 192.168.130.35:9876
...
```

#### Verify NFS storage
Access the NFS server node of your cluster and verify whether the RocketMQ data is stored in your NFS data volume path:

```
$ cd /data/k8s/

$ ls
default-broker-storage-broker-0-master-0-pvc-7a74871b-c005-441a-bb15-8106566c9d19   default-broker-storage-broker-1-replica-1-0-pvc-af266db9-83a9-4929-a2fe-e40fb5fdbfa4
default-broker-storage-broker-0-replica-1-0-pvc-521e7e9a-3795-487a-9f76-22da74db74dd  default-namesrv-storage-name-service-0-pvc-c708cb49-aa52-4992-8cac-f46a48e2cc2e
default-broker-storage-broker-1-master-0-pvc-d7b76efe-384c-4f8d-9e8a-ebe209ba826c

$ ls default-broker-storage-broker-1-master-0-pvc-d7b76efe-384c-4f8d-9e8a-ebe209ba826c/logs/rocketmqlogs/
broker_default.log  broker.log  commercial.log  filter.log  lock.log  protection.log  remoting.log  stats.log  storeerror.log  store.log  transaction.log  watermark.log

$ cat default-broker-storage-broker-1-master-0-pvc-d7b76efe-384c-4f8d-9e8a-ebe209ba826c/logs/rocketmqlogs/broker.log 
...
2019-09-10 14:12:22 INFO main - The broker[broker-1-master-0, 10.244.2.117:10911] boot success. serializeType=JSON and name server is 192.168.130.33:9876
...
```

## Horizontal Scale

### Name Server Cluster Scale
If the current name service cluster scale does not fit your requirements, you can simply use RocketMQ-Operator to up-scale or down-scale your name service cluster.

If you want to enlarge your name service cluster. Modify your name service CR file ```rocketmq_v1alpha1_nameservice_cr.yaml```, increase the field ```size``` to the number you want, for example, from ```size: 1``` to ```size: 2```.

> Notice: if your broker image version is 4.5.0 or earlier, you need to make sure that ```allowRestart: true``` is set in the broker CR file to enable rolling restart policy. If ```allowRestart: false```, configure it to ```allowRestart: true``` and run ```kubectl apply -f example/rocketmq_v1alpha1_broker_cr.yaml``` to apply the new config.

After configuring the ```size``` fields, simply run 
```
kubectl apply -f example/rocketmq_v1alpha1_nameservice_cr.yaml 
```
Then a new name service pod will be deployed and meanwhile the operator will inform all the brokers to update their name service list parameters, so they can register to the new name service.

> Notice: under the policy ```allowRestart: true```, the broker will gradually be updated so the update process is also not perceptible to the producer and consumer clients.

### Broker Cluster Scale

#### Up-scale Broker in Out-of-order Message Scenario
It is often the case that with the development of your business, the old broker cluster scale no longer meets your needs. You can simply use RocketMQ-Operator to up-scale your broker cluster:

1. Modify the ```size``` in the broker CR file to the number that you want the broker cluster scale will be, for example, from ```size: 1``` to ```size: 2```.

2. Choose the source broker pod, from which the old metadata like topic and subscription information data will be transferred to the newly created brokers. The source broker pod field is 
```
...
# scalePodName is broker-[broker group number]-master-0
  scalePodName: broker-0-master-0
...
```

3. Apply the new configurations:
```
kubectl apply -f example/rocketmq_v1alpha1_broker_cr.yaml
```
Then a new broker group of pods will be deployed and meanwhile the operator will copy the metadata from the source broker pod to the newly created broker pods before the new brokers are stared, so the new brokers will reload previous topic and subscription information.

## Topic Transfer

```Topic Transfer``` means that the user wants to migrate the work of providing service for a specific topic from a source(original) cluster to a target cluster without affecting the business. This may happen when the source cluster is about to shutdown, or the user wants to reduce the workload on the source cluster.

Usually the ```Topic Transfer``` process consists of 7 steps:

* Add all consumer groups of the topic to the target cluster.

* Add the topic to be transferred to the target cluster. 

* Forbid new message writing into the source cluster.

* Check the consumer group consumption progress to make sure all messages in the source cluster have been consumed.

* Delete the topic in the source cluster when all messages in the source cluster have been consumed.

* Delete the consumer groups in the source cluster.

* Add the retry-topic to the target cluster.

The TopicTransfer CRD can help you do that. Simply configure the CR file ```example/rocketmq_v1alpha1_topictransfer_cr.yaml```:

```
apiVersion: rocketmq.apache.org/v1alpha1
kind: TopicTransfer
metadata:
  name: topictransfer
spec:
  # topic defines which topic to be transferred
  topic: TopicTest
  # sourceCluster define the source cluster
  sourceCluster: broker-0
  # targetCluster defines the target cluster
  targetCluster: broker-1
```

Then apply the ```TopicTransfer``` resource:

```
$ kubectl apply -f example/rocketmq_v1alpha1_topictransfer_cr.yaml
```

The operator will automatically do the topic transfer job. 

If the transfer process is failed, the operator will roll-back the transfer operations for the atomicity of the ```TopicTransfer``` operation.

You can check the operator logs or consume progress status to monitor and verify the topic transfer process:

```
$ kubectl logs -f [operator-pod-name] 
```

```
$ sh bin/mqadmin consumerprogress -g [consumer-group] -n [name-server-ip]:9876
```

## Clean the Environment
If you want to tear down the RocketMQ cluster, to remove the broker clusters run

```
$ kubectl delete -f example/rocketmq_v1alpha1_broker_cr.yaml
```

to remove the name service clusters:

```
$ kubectl delete -f example/rocketmq_v1alpha1_nameservice_cr.yaml
```

to remove the RocketMQ Operator:

```
$ ./purge-operator.sh
```

to remove the storage class for RocketMQ:

```
$ cd deploy/storage
$ ./remove-storage-class.sh
```

> Note: the NFS persistence data will not be deleted by default.
