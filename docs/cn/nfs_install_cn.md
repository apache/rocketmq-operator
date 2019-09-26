## NFS 存储服务部署

本文档将通过一个示例介绍如何简单地在您的Kubernetes集群上部署NFS共享存储服务。

### 安装 NFS 服务端

选择您的一个节点来安装 NFS 服务，假设我们的共享数据目录为：```/data/k8s```

1. 关闭防火墙(需要sudo权限)

```
$ systemctl stop firewalld.service

$ systemctl disable firewalld.service
```

2. 安装配置 NFS

```
$ yum -y install nfs-utils rpcbind
```

3. 设置共享目录的权限：

```
$ chmod 755 /data/k8s/
```

4. 配置 NFS，NFS 的默认配置文件在 ```/etc/exports``` 文件下，在该文件中添加下面的配置信息：

```
$ vi /etc/exports
/data/k8s  *(rw,sync,no_root_squash)
```

5. 启动rpcbind服务

```
$ systemctl start rpcbind.service

$ systemctl enable rpcbind

$ systemctl status rpcbind
···
Sep 16 10:27:19 master systemd[1]: Started RPC bind service.
```

如果看到上面的 Started 说明rpcbind启动成功了。

6. 启动 NFS 服务：

```
$ systemctl start nfs.service

$ systemctl enable nfs

$ systemctl status nfs
···
Sep 16 10:28:01 master systemd[1]: Started NFS server and services.
```

如果看到上面的 Started 则说明 NFS Server 启动成功了。

也可以通过以下命令确认：

```
$ rpcinfo -p | grep nfs
    100003    3   tcp   2049  nfs
    100003    4   tcp   2049  nfs
    100227    3   tcp   2049  nfs_acl
    100003    3   udp   2049  nfs
    100003    4   udp   2049  nfs
    100227    3   udp   2049  nfs_acl
```

到这里我们就安装好了 NFS server，接下来我们在其他节点上安装 NFS 的客户端。

### 安装 NFS 客户端

在您的Kubernetes集群每个节点上安装NFS：

1. 先关闭防火墙：

```
$ systemctl stop firewalld.service

$ systemctl disable firewalld.service
```

2. 安装 NFS 

```
$ yum -y install nfs-utils rpcbind
```

3. 与上面的方法一样，先启动 rpc、然后启动 NFS：
```
$ systemctl start rpcbind.service 

$ systemctl enable rpcbind.service 

$ systemctl start nfs.service    

$ systemctl enable nfs.service
```

4. 客户端启动完成后，检查 NFS 是否有共享目录：

```
$ showmount -e 192.168.130.32
Export list for 192.168.130.32:
/data/k8s *
```

其中```192.168.130.32```是您的NFS服务端所在的节点IP

进一步的验证操作请查看[NFS存储准备工作文档](README.md)

