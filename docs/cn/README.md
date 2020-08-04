# RocketMQ Operator

## 目录

- [概览](#概览)
- [快速开始](#快速开始)
    - [部署 RocketMQ Operator](#部署-rocketmq-operator)
    - [持久化准备工作](#持久化准备工作)
        - [HostPath （宿主机本地路径）准备工作](#hostpath-宿主机本地路径准备工作)
        - [基于NFS的StorageClass准备工作](#基于nfs的storageclass准备工作)
    - [定义您的RocketMQ集群](#定义您的rocketmq集群)
    - [创建 RocketMQ 集群](#创建-rocketmq-集群)
    - [验证数据存储](#验证数据存储)
        - [验证 HostPath 存储](#验证-hostpath-存储)
        - [验证 NFS 存储](#验证-nfs-存储)
- [水平扩缩容](#水平扩缩容)
    - [Name Server 集群扩缩容](#name-server-集群扩缩容)
    - [Broker集群扩缩容](#broker集群扩缩容)
        - [无顺序消息的Broker集群扩容](#无顺序消息的broker集群扩容)
- [Topic 迁移](#topic-迁移)
- [环境清理](#环境清理)

Switch to [English Document](../../README.md)

## 概览

RocketMQ Operator的作用是管理基于Kubernetes平台的RocketMQ服务集群。
它是使用[Operator SDK](https://github.com/operator-framework/operator-sdk)构建的，基于[Operator Framework](https://github.com/operator-framework/)提供的Operator标准框架。

![RocketMQ-Operator architecture](../img/operator-arch.png)

## 快速开始

### 部署 RocketMQ Operator

1. 克隆RocketMQ Operator项目到您的Kubernetes集群的master节点上：
```
$ git clone https://github.com/apache/rocketmq-operator.git
$ cd rocketmq-operator
```

2. 运行以下安装脚本以在您的Kubernetes集群上部署RocketMQ Operator:

```
$ ./install-operator.sh
```

3. 使用命令 ```kubectl get pods``` 来检查RocketMQ Operator的部署状态，若部署成功将产生类似下面的输出:

```
$ kubectl get pods
NAME                                      READY   STATUS    RESTARTS   AGE
rocketmq-operator-564b5d75d-jllzk         1/1     Running   0          108s
```

现在您可以使用RocketMQ Operator提供的自定义资源（CRD）来部署您的RocketMQ集群。

### 持久化准备工作

在部署RocketMQ集群之前，您需要做一些数据持久化的准备工作。目前RocketMQ Operator支持多种持久化RocketMQ数据的方式，包括：```EmptyDir```, ```HostPath``` 以及 ```NFS```。 

您可以通过在示例资源文件例如```rocketmq_v1alpha1_nameservice_cr.yaml```中配置来选择您想要的存储方式：

```
...
 # storageMode can be EmptyDir, HostPath, NFS
  storageMode: HostPath
...
```

如果您选择的是```EmptyDir```，那么您不需要做额外的准备工作来持久化数据。但是该方式下数据的持久化生命周期与对应的pod的生命周期相同，这意味着如果pod被删除您将有可能会丢失对应pod上的数据。

如果您选择的是其他持久化存储方式，请根据下面的指导完成相应的准备工作。

#### HostPath （宿主机本地路径）准备工作

如果您使用的配置是```storageMode: HostPath```，RocketMQ的数据（包括所有的日志和存储文件）会存储在每个pod对应所在的宿主机的本地路径上。 因此您需要事先准备好您希望挂载的本地路径及其统一名称。

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

如果您选择NFS作为RocketMQ数据存储方式，您将需要先在Kubernetes集群上部署好NFS服务端，并在其他节点上部署好NFS客户端，然后准备一个基于```NFS```的```StorageClass```资源用来自动创建PV和PVC资源。如果您对以上概念不了解也没有关系，仅需按照以下步骤操作即可：

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

2. 修改 ```deploy/storage/nfs-client.yaml``` 文件中的以下配置:
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
将 ```192.168.130.32``` 和 ```/data/k8s``` 替换成您真实的NFS服务端的IP和共享目录路径。
 
3. 创建一个NFS StorageClass，运行：

```
$ cd deploy/storage
$ ./deploy-storage-class.sh
```

4. 如果StorageClass创建并部署成功，您将可以通过以下命令获得类似如下集群pod状态：

```
$ kubectl get pods
NAME                                      READY   STATUS    RESTARTS   AGE
nfs-client-provisioner-7cf858f754-7vxmm   1/1     Running   0          136m
```

到这里您的NFS存储准备工作就完成了。

### 定义您的RocketMQ集群

RocketMQ Operator 提供多种自定义资源（CRD）用于让用户定义需要部署的RocketMQ服务组件集群的规模、从节点数等相关配置，服务组件集群包括Name Server集群以及Broker集群。

1. 查看```example```路径下的 ```rocketmq_v1alpha1_nameservice_cr.yaml```NameService自定义资源示例配置文件, 例如:
```
apiVersion: rocketmq.apache.org/v1alpha1
kind: NameService
metadata:
  name: name-service
spec:
  # size is the the name service instance number of the name service cluster
  size: 1
  # nameServiceImage is the customized docker image repo of the RocketMQ name service
  nameServiceImage: apacherocketmq/rocketmq-namesrv:4.5.0-alpine
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
它定义了RocketMQ Name Server 集群的规模（```size```）等。

2. 检查```example```下的```rocketmq_v1alpha1_broker_cr.yaml``` Broker自定义资源示例文件, 例如:
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
  # replicaPerGroup is the number of replica broker in each group
  replicaPerGroup: 1
  # brokerImage is the customized docker image repo of the RocketMQ broker
  brokerImage: apacherocketmq/rocketmq-broker:4.5.0-alpine
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
它定义了RocketMQ Broker集群的规模，初始化Broker集群时Name Server集群的[IP:端口]列表等参数。

> 注：```size```指的是Broker组数，一个Broker组包含一个Broker主节点以及若干个（可以是0个）Broker从节点。

### 创建 RocketMQ 集群
 
1. 部署Name Server集群，运行命令:

``` 
$ kubectl apply -f example/rocketmq_v1alpha1_nameservice_cr.yaml 
nameservice.rocketmq.apache.org/name-service created
```

检查当前Kubernetes集群pod状态:

```
$ kubectl get pods -owide
NAME                                      READY   STATUS    RESTARTS   AGE     IP               NODE        NOMINATED NODE   READINESS GATES
name-service-0                            1/1     Running   0          3m18s   192.168.130.33   k2data-13   <none>           <none>
nfs-client-provisioner-7cf858f754-7vxmm   1/1     Running   0          150m    10.244.2.114     k2data-14   <none>           <none>
rocketmq-operator-564b5d75d-jllzk         1/1     Running   0          5m53s   10.244.2.116     k2data-14   <none>           <none>
```

可以看到有1个Name Server的pod(```name-service-0```)启动了，以及它对应的IP地址。修改 ```rocketmq_v1alpha1_broker_cr.yaml``` 文件中的```nameServers```配置，将默认配置替换为此时您看到的Name Server所在pod的真实IP地址:9876。

> 如果您修改了```NameService```的```size```默认配置，启动了多个Name Server，则```nameServers```配置需修改为类似```IP1:9876;IP2:9876```的这种列表形式。

2. 部署Broker集群，运行命令:

```
$ kubectl apply -f example/rocketmq_v1alpha1_broker_cr.yaml
broker.rocketmq.apache.org/broker created 
```

之后Broker集群的容器就会被自动创建好，最终查看集群pod状态会得到类似如下的输出：

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

3. 查看PV和PVC资源状态:
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

> 注意：如果您不是选择的NFS存储方式，那么以上PV和PVC就不会被创建

恭喜您成功通过RocketMQ Operator创建了您的RocketMQ服务集群！

### 验证数据存储

#### 验证 HostPath 存储

如果您选择的是```HostPath```的存储方式，您可以通过ssh访问任何一台部署了RocketMQ服务pod的节点，检查您在自定义资源声明文件中```hostPath```对应路径下的文件内容来确认数据存储，例如默认路径下类似如下输出：

```
$ ls /data/rocketmq/broker
logs  store

$ cat /data/rocketmq/broker/logs/broker-1-replica-1/rocketmqlogs/broker.log
...
2019-09-12 13:12:24 INFO main - The broker[broker-1, 10.244.3.35:10911] boot success. serializeType=JSON and name server is 192.168.130.35:9876
...
```

#### 验证 NFS 存储

如果您选择的是```NFS```的存储方式，您可以通过ssh访问部署了NFS服务端的节点，并检查您配置的共享路径下的文件内容来确认数据存储，例如：

```
$ cd /data/k8s/

$ ls
default-broker-storage-broker-0-master-0-pvc-7a74871b-c005-441a-bb15-8106566c9d19   
default-broker-storage-broker-0-replica-1-0-pvc-521e7e9a-3795-487a-9f76-22da74db74dd  
default-broker-storage-broker-1-master-0-pvc-d7b76efe-384c-4f8d-9e8a-ebe209ba826c
default-broker-storage-broker-1-replica-1-0-pvc-af266db9-83a9-4929-a2fe-e40fb5fdbfa4
default-namesrv-storage-name-service-0-pvc-c708cb49-aa52-4992-8cac-f46a48e2cc2e

$ ls default-broker-storage-broker-1-master-0-pvc-d7b76efe-384c-4f8d-9e8a-ebe209ba826c/logs/rocketmqlogs/
broker_default.log  broker.log  commercial.log  filter.log  lock.log  protection.log  remoting.log  stats.log  storeerror.log  store.log  transaction.log  watermark.log

$ cat default-broker-storage-broker-1-master-0-pvc-d7b76efe-384c-4f8d-9e8a-ebe209ba826c/logs/rocketmqlogs/broker.log 
...
2019-09-10 14:12:22 INFO main - The broker[broker-1-master-0, 10.244.2.117:10911] boot success. serializeType=JSON and name server is 192.168.130.33:9876
...
```

## 水平扩缩容

### Name Server 集群扩缩容
如果当前的Name Server集群规模不能满足您的需求，您可以通过RocketMQ Operator轻松地实现Name Server集群的扩缩容。

例如您希望扩大您的Name Server集群规模，可以通过修改NameService自定义资源声明文件例如```rocketmq_v1alpha1_nameservice_cr.yaml```中的```size```配置项来增大为您希望的服务实例数，例如从```size: 1``` 修改为 ```size: 2```

> 注意：如果您的Broker自定义资源声明文件中使用的镜像为4.5.0或更早的版本，您需要确保您的Broker自定义资源声明文件中设置了```allowRestart: true```，使得Broker可以通过滚动重启的方式来注册新扩容出来的Name Server，该过程不会影响集群对外的持续服务。如果```allowRestart: false```，请改为```allowRestart: true```并运行```kubectl apply -f example/rocketmq_v1alpha1_broker_cr.yaml```以应用修改后的配置。

在修改了```size```之后，只需简单地运行：
```
kubectl apply -f example/rocketmq_v1alpha1_nameservice_cr.yaml 
```
之后新的Name Server就会被自动部署出来，与此同时Operator会自动通知所有的Broker去更新他们的Name Server列表参数，使得他们可以注册新的Name Server服务。

> 注意：在```allowRestart: true```配置下，Broker集群会逐步地滚动式重启服务以更新参数，这个过程对集群外部的生产者和消费者是不会感知到的。

### Broker集群扩缩容

#### 无顺序消息的Broker集群扩容
随着您业务的发展，原来的Broker集群规模可能无法满足您的生产需求，您可以通过RocketMQ Operator轻松地实现Broker集群的扩容：

1. 修改Broker自定义资源声明文件中的```size```配置为您希望的Broker集群规模，例如从```size: 1``` 修改为 ```size: 2```

2. 配置源Broker pod，所谓源Broker是指把哪个pod上的Broker元数据（包括Topic信息和订阅信息）到新扩容出来的Broker。源Broker pod对应的配置项默认为：
```
...
# scalePodName is broker-[broker group number]-master-0
  scalePodName: broker-0-master-0
...
```
表示将```broker-0-master```的元数据同步给所有新扩容出来的Broker。

3. 应用修改后的Broker自定义资源声明文件:
```
kubectl apply -f example/rocketmq_v1alpha1_broker_cr.yaml
```
之后Operator就会帮您自动创建出新扩容出的Broker组，并且在新扩容出来的Broker启动之前将元数据文件同步到每个新Broker上，因而新Broker启动之后会加载源Broker的元数据信息（包括Topic和订阅信息）。

## Topic 迁移

```Topic 迁移```是指用户希望将一个Topic的服务工作从一个源集群转移到另一个目标集群，并且在这个过程中不影响业务。这可能发生在用户想要停用源集群或减轻源集群的工作负载压力。
通常```Topic 迁移```的过程分为以下7步：

* 添加要转移的Topic的所有消费者组到目标集群。

* 添加要转移的Topic到目标集群。

* 源集群对应Topic禁止写入新的消息

* 检查所有对应消费者组的消费进度状态，直到确认源集群中的所有消息都被消费了，没有堆积。

* 当确认源集群中的所有对应消息都被消费之后，删除源集群中的对应Topic

* 删除源集群中的所有对应消费者组

* 在目标集群中创建RETRY Topic

通过Operator提供的```TopicTransfer```自定义资源可以帮助您自动完成Topic迁移的工作。只需简单地配置自定义资源声明文件```example/rocketmq_v1alpha1_topictransfer_cr.yaml```：

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

然后应用 ```TopicTransfer``` 自定义资源:

```
$ kubectl apply -f example/rocketmq_v1alpha1_topictransfer_cr.yaml
```

之后Operator就会自动地帮您完成Topic迁移的工作。

如果在Topic迁移过程中遇到了错误，Operator将自动回滚所有的Topic迁移中间过程的操作，使得集群恢复到应用```TopicTransfer```自定义资源之前的状态，以保证Topic迁移操作的原子性。

您可以通过查看Operator日志或通过RocketMQ的Admin工具来检查和验证当前Topic迁移过程的状态：

```
$ kubectl logs -f [operator-pod-name] 
```

```
$ sh bin/mqadmin consumerprogress -g [consumer-group] -n [name-server-ip]:9876
```

## 环境清理

如果您想要下线RocketMQ的Broker集群，运行：

```
$ kubectl delete -f example/rocketmq_v1alpha1_broker_cr.yaml
```

如果您想要下线RocketMQ的Name Server集群，运行：

```
$ kubectl delete -f example/rocketmq_v1alpha1_nameservice_cr.yaml
```

如果您想要清理整个RocketMQ集群以及Operator，运行：

```
$ ./purge-operator.sh
```

如果您想要清理RocketMQ的NFS StorageClass，运行：

```
$ cd deploy/storage
$ ./remove-storage-class.sh
```

> 注意：通过NFS和HostPath方式存储的数据，默认情况下即便您通过以上命令下线了整个RocketMQ集群，上面的数据也不会被删除。
