## 准备工作

1.  时间同步
2.  主机名修改
3.  主机名与ip在/etc/hosts绑定
4.  ssh双机互信
5.  关闭防火墙
6.  关闭swap
7.  关闭selinux

## etcd集群安装

etcd服务器IP:
- 192.168.14.203
- 192.168.14.9
- 192.168.14.34

### cfssl工具准备

```
wget https://pkg.cfssl.org/R1.2/cfssl_linux-amd64 -O /usr/local/bin/cfssl
wget https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64 -O /usr/local/bin/cfssljson
wget https://pkg.cfssl.org/R1.2/cfssl-certinfo_linux-amd64 -O /usr/local/bin/cfssl-certinfo
chmod +x /usr/local/cfssl*
```

### etcd证书创建

```
~]# cat ca-config.json
{
  "signing": {
    "default": {
      "expiry": "876000h"
    },
    "profiles": {
      "etcd": {
        "usages": [
            "signing",
            "key encipherment",
            "server auth",
            "client auth"
        ],
        "expiry": "876000h"
      }
    }
  }
}
```
```
~]# cat ca-csr.json
{
  "CN": "etcd",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "ST": "zhejiang",
      "L": "hangzhou",
      "O": "etcd",
      "OU": "System"
    }
  ]
}
```
- 生成CA证书
```
~]# cfssl gencert -initca ca-csr.json | cfssljson -bare ca
2019/10/08 14:10:38 [INFO] generating a new CA key and certificate from CSR
2019/10/08 14:10:38 [INFO] generate received request
2019/10/08 14:10:38 [INFO] received CSR
2019/10/08 14:10:38 [INFO] generating key: rsa-2048
2019/10/08 14:10:38 [INFO] encoded CSR
2019/10/08 14:10:38 [INFO] signed certificate with serial number 729202919502466530277041968346292222430741265108
```
```
~]# cat etcd-csr.json
{
  "CN": "etcd",
  "hosts": [
    "127.0.0.1",
    "192.168.14.203",
    "192.168.14.9",
    "192.168.14.34"
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "ST": "zhejiang",
      "L": "hangzhou",
      "O": "etcd",
      "OU": "System"
    }
  ]
}
```
- 生成etcd证书
```
~]# cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=etcd etcd-csr.json | cfssljson -bare etcd
2019/10/08 14:13:54 [INFO] generate received request
2019/10/08 14:13:54 [INFO] received CSR
2019/10/08 14:13:54 [INFO] generating key: rsa-2048
2019/10/08 14:13:54 [INFO] encoded CSR
2019/10/08 14:13:54 [INFO] signed certificate with serial number 725478311285159828781967151519197153159047282910
2019/10/08 14:13:54 [WARNING] This certificate lacks a "hosts" field. This makes it unsuitable for
websites. For more information see the Baseline Requirements for the Issuance and Management
of Publicly-Trusted Certificates, v.1.1.6, from the CA/Browser Forum (https://cabforum.org);
specifically, section 10.2.3 ("Information Requirements").
```

### etcd安装

-  准备etcd TLS证书文件

```
~]# mkdir -p /etc/etcd/etcdSSL
~]# cp ca.pem etcd.pem etcd-key.pem /etc/etcd/etcdSSL/
```
etcd集群节点都需要准备一份

- 安装etcd服务
```
~]# yum install -y etcd
~]# cat /usr/lib/systemd/system/etcd.service
[Unit]
Description=Etcd Server
After=network.target
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
WorkingDirectory=/var/lib/etcd/
EnvironmentFile=-/etc/etcd/etcd.conf
User=etcd
# set GOMAXPROCS to number of processors
ExecStart=/usr/bin/etcd \
  --name ${ETCD_NAME} \
  --cert-file=/etc/etcd/etcdSSL/etcd.pem \
  --key-file=/etc/etcd/etcdSSL/etcd-key.pem \
  --peer-cert-file=/etc/etcd/etcdSSL/etcd.pem \
  --peer-key-file=/etc/etcd/etcdSSL/etcd-key.pem \
  --trusted-ca-file=/etc/etcd/etcdSSL/ca.pem \
  --peer-trusted-ca-file=/etc/etcd/etcdSSL/ca.pem \
  --initial-advertise-peer-urls ${ETCD_INITIAL_ADVERTISE_PEER_URLS} \
  --listen-peer-urls=${ETCD_LISTEN_PEER_URLS} \
  --listen-client-urls=${ETCD_LISTEN_CLIENT_URLS},http://127.0.0.1:2379 \
  --advertise-client-urls=${ETCD_ADVERTISE_CLIENT_URLS} \
  --initial-cluster-token=${ETCD_INITIAL_CLUSTER_TOKEN} \
  --initial-cluster=${ETCD_CLUSTER}\
  --initial-cluster-state new \
  --data-dir=${ETCD_DATA_DIR}
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```
- 配置文件
```
~]# cat /etc/etcd/etcd.conf
# 这里需要修改
ETCD_NAME=etcd1
ETCD_DATA_DIR="/var/lib/etcd"
ETCD_LISTEN_PEER_URLS="https://192.168.14.9:2380"
ETCD_LISTEN_CLIENT_URLS="https://192.168.14.9:2379"

#[cluster]
ETCD_INITIAL_ADVERTISE_PEER_URLS="https://192.168.14.9:2380"
ETCD_INITIAL_CLUSTER_TOKEN="etcd-cluster"
ETCD_ADVERTISE_CLIENT_URLS="https://192.168.14.9:2379"
ETCD_CLUSTER="etcd1=https://192.168.14.9:2380,etcd2=https://192.168.14.34:2380,etcd3=https://192.168.14.203:2380"
```
- 启动节点
```
~]# chown -R etcd /etc/etcd/etcdSSL
~]# systemctl daemon-reload
~]# systemctl enable etcd
~]# systemctl start etcd
```

重复三个节点

### 验证etcd集群

```
~]# etcdctl --ca-file=/etc/etcd/etcdSSL/ca.pem --cert-file=/etc/etcd/etcdSSL/etcd.pem --key-file=/etc/etcd/etcdSSL/etcd-key.pem cluster-health
member 2f37f9544bd71c49 is healthy: got healthy result from https://192.168.14.203:2379
member 78ff39a4f11f0d65 is healthy: got healthy result from https://192.168.14.9:2379
member 945d4e6858099e77 is healthy: got healthy result from https://192.168.14.34:2379
cluster is healthy
```

## Kubernetes Master节点安装

### 准备kubernetes集群所需要的TLS证书文件

部署kubernetes服务所需要使用的证书如下：
1.  根证书公钥与私钥 ca.pem与ca.key
2.  API Server公钥与私钥 apiserver.pem与apiserver.key
3.  集群管理员公钥与私钥 admin.pem与admin.key
4.  节点proxy公钥与私钥
5.  节点kubelet的公钥与私钥是通过bootstrap响应的方式，在启动kubelet自动会产生，然后在master通过csr请求，就会产生

### 创建CA证书

```
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -days 100000 -out ca.pem -subj "/CN=kubernetes/O=k8s"
```

### 创建apiserver证书

创建openssl.cnf
```
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = kubernetes
DNS.2 = kubernetes.default
DNS.3 = kubernetes.default.svc
DNS.4 = kubernetes.default.svc.cluster
DNS.5 = kubernetes.default.svc.cluster.local
DNS.6 = k8s_master
IP.1 = 10.0.6.1              # ClusterServiceIP 地址
IP.2 = 192.168.14.9           # master IP地址
IP.3 = 10.0.6.200            # kubernetes DNS IP地址
```
- 生成apiserver证书
```
openssl genrsa -out apiserver.key 2048
openssl req -new -key apiserver.key -out apiserver.csr -subj "/CN=kubernetes/O=k8s" -config openssl.cnf
openssl x509 -req -in apiserver.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out apiserver.pem -days 3650 -extensions v3_req -extfile openssl.cnf
```

### admin集群管理员证书生成

```
openssl genrsa -out admin.key 2048
openssl req -new -key admin.key -out admin.csr -subj "/CN=admin/O=system:masters/OU=System"
openssl x509 -req -in admin.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out admin.pem -days 3650
```

### 节点Proxy证书生成

```
openssl genrsa -out proxy.key 2048
openssl req -new -key proxy.key -out proxy.csr -subj "/CN=system:kube-proxy"
openssl x509 -req -in proxy.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out proxy.pem -days 3650
```

### 准备Master节点证书组件

```
mkdir /etc/kubernetes/kubernetesTLS
cp ca.pem ca.key apiserver.key apiserver.pem admin.key admin.pem proxy.key proxy.pem /etc/kubernetes/kubernetesTLS
cp kube-apiserver  /usr/local/bin
cp kube-scheduler /usr/local/bin
cp kube-controller-manager /usr/local/bin
```

### 安装kube-apiserver

- 创建TLS Bootstrapping Token

```
export BOOTSTRAP_TOKEN=$(head -c 16 /dev/urandom | od -An -t x | tr -d ' ')
cat > /etc/kubernetes/BOOTSTRAP_TOKEN << EOF
$BOOTSTRAP_TOKEN
EOF
cat > /etc/kubernetes/token.csv << EOF
$BOOTSTRAP_TOKEN,kubelet-bootstrap,10001,"system:kubelet-bootstrap"
EOF
```

- 创建admin用户的集群参数

```
# 设置集群参数
kubectl config set-cluster kubernetes --certificate-authority=/etc/kubernetes/kubernetesTLS/ca.pem --embed-certs=true --server=https://192.168.14.9:6443

# 设置管理员参数
kubectl config set-credentials admin --client-certificate=/etc/kubernetes/kubernetesTLS/admin.pem --client-key=/etc/kubernetes/kubernetesTLS/admin.key --embed-certs=true

# 设置管理员上下文参数
kubectl config set-context kubernetes --cluster=kubernetes --user=admin

# 设置集群默认上下文参数
kubectl config use-context kubernetes
```
- 配置kube-apiserver

```
cat /usr/lib/systemd/system/kube-apiserver.service
[Unit]
Description=Kube-apiserver Service
Documentation=https://github.com/GoogleCloudPlatform/kubernetes

After=network.target
[Service]
Type=notify
EnvironmentFile=-/etc/kubernetes/config
EnvironmentFile=-/etc/kubernetes/apiserver
ExecStart=/usr/local/bin/kube-apiserver \
        $KUBE_LOGTOSTDERR \
        $KUBE_LOG_LEVEL \
        $KUBE_ETCD_SERVERS \
        $KUBE_API_ADDRESS \
        $KUBE_API_PORT \
        $KUBELET_PORT \
        $KUBE_ALLOW_PRIV \
        $KUBE_SERVICE_ADDRESSES \
        $KUBE_ADMISSION_CONTROL \
        $KUBE_API_ARGS
Restart=always
LimitNOFILE=65536

[Install]
WantedBy=default.target
```
```
cat /etc/kubernetes/config
###
# kubernetes system config
#
# The following values are used to configure various aspects of all
# kubernetes services, including
#
#   kube-apiserver.service
#   kube-controller-manager.service
#   kube-scheduler.service
#   kubelet.service
#   kube-proxy.service
# logging to stderr means we get it in the systemd journal
# 表示错误日志记录到文件还是输出到stderr。
KUBE_LOGTOSTDERR="--logtostderr=true"

# journal message level, 0 is debug
# 日志等级。设置0则是debug等级
KUBE_LOG_LEVEL="--v=0"

# Should this cluster be allowed to run privileged docker containers
# 允许运行特权容器。
KUBE_ALLOW_PRIV="--allow-privileged=true"

# How the controller-manager, scheduler, and proxy find the apiserver
# 设置master服务器的访问
KUBE_MASTER="--master=http://192.168.14.9:8080"
```
```
cat /etc/kubernetes/apiserver
###
## kubernetes system config
##
## The following values are used to configure the kube-apiserver
##
#
## The address on the local server to listen to.
KUBE_API_ADDRESS="--advertise-address=192.168.14.9 --bind-address=192.168.14.9 --insecure-bind-address=192.168.14.9"
#
## The port on the local server to listen on.
#KUBE_API_PORT="--port=8080"
#
## Port minions listen on
#KUBELET_PORT="--kubelet-port=10250"
#
## Comma separated list of nodes in the etcd cluster
KUBE_ETCD_SERVERS="--etcd-servers=https://192.168.14.9:2379,https://192.168.14.34:2379,https://192.168.14.203:2379"
#
## Address range to use for services
KUBE_SERVICE_ADDRESSES="--service-cluster-ip-range=10.0.6.0/24"
#
## default admission control policies
KUBE_ADMISSION_CONTROL="--admission-control=ServiceAccount,NamespaceLifecycle,NamespaceExists,LimitRanger,ResourceQuota,NodeRestriction"

## Add your own!
KUBE_API_ARGS="--authorization-mode=Node,RBAC  --runtime-config=rbac.authorization.k8s.io/v1beta1  --kubelet-https=true  --token-auth-file=/etc/kubernetes/token.csv  --service-node-port-range=30000-32767  --tls-cert-file=/etc/kubernetes/kubernetesTLS/apiserver.pem  --tls-private-key-file=/etc/kubernetes/kubernetesTLS/apiserver.key  --client-ca-file=/etc/kubernetes/kubernetesTLS/ca.pem  --service-account-key-file=/etc/kubernetes/kubernetesTLS/ca.key  --storage-backend=etcd3  --etcd-cafile=/etc/etcd/etcdSSL/ca.pem  --etcd-certfile=/etc/etcd/etcdSSL/etcd.pem  --etcd-keyfile=/etc/etcd/etcdSSL/etcd-key.pem  --enable-swagger-ui=true  --apiserver-count=3  --audit-log-maxage=30  --audit-log-maxbackup=3  --audit-log-maxsize=100  --audit-log-path=/var/lib/audit.log  --event-ttl=1h"
```

- 启动kube-apiserver

```
systemctl daemon-reload
systemctl enable kube-apiserver
systemctl  start kube-apiserver
```

###  安装kube-scheduler

```
cat  /usr/lib/systemd/system/kube-scheduler.service
[Unit]
Description=Kube-scheduler Service
After=network.target

[Service]
Type=simple
EnvironmentFile=-/etc/kubernetes/config
EnvironmentFile=-/etc/kubernetes/scheduler
ExecStart=/usr/local/bin/kube-scheduler \
            $KUBE_LOGTOSTDERR \
            $KUBE_LOG_LEVEL \
            $KUBE_MASTER \
            $KUBE_SCHEDULER_ARGS

Restart=always
LimitNOFILE=65536

[Install]
WantedBy=default.target
```

```
cat /etc/kubernetes/scheduler
#wing values are used to configure the kubernetes scheduler

# defaults from config and scheduler should be adequate

# Add your own!
KUBE_SCHEDULER_ARGS="--leader-elect=true --address=127.0.0.1"
```

- 启动kube-scheduler
```
systemctl daemon-reload
systemctl enable kube-scheduler
systemctl start kube-scheduler
```

### 安装kube-controller-manager

```
cat /usr/lib/systemd/system/kube-controller-manager.service
[Unit]
Description=Kube-controller-manager Service
Documentation=https://github.com/GoogleCloudPlatform/kubernetes
After=network.target
After=kube-apiserver.service
Requires=kube-apiserver.service
[Service]
Type=simple
EnvironmentFile=-/etc/kubernetes/config
EnvironmentFile=-/etc/kubernetes/controller-manager
ExecStart=/usr/local/bin/kube-controller-manager \
        $KUBE_LOGTOSTDERR \
        $KUBE_LOG_LEVEL \
        $KUBE_MASTER \
        $KUBE_CONTROLLER_MANAGER_ARGS
Restart=always
LimitNOFILE=65536

[Install]
WantedBy=default.target
```
```
cat /etc/kubernetes/controller-manager
###
# The following values are used to configure the kubernetes controller-manager

# defaults from config and apiserver should be adequate

# Add your own!
KUBE_CONTROLLER_MANAGER_ARGS=" --address=127.0.0.1  --service-cluster-ip-range=10.0.6.0/24  --cluster-name=kubernetes  --cluster-signing-cert-file=/etc/kubernetes/kubernetesTLS/ca.pem  --cluster-signing-key-file=/etc/kubernetes/kubernetesTLS/ca.key  --service-account-private-key-file=/etc/kubernetes/kubernetesTLS/ca.key  --root-ca-file=/etc/kubernetes/kubernetesTLS/ca.pem  --leader-elect=true  --cluster-cidr=172.16.0.0/16"
```

-  启动kube-controller-manager
```
systemctl daemon-reload
systemctl enable kube-controller-manager
systemctl start kube-controller-manager
```

### 创建kubeconfig文件及相关的集群参数

```
export kubernetesTLSDir=/etc/kubernetes/kubernetesTLS
export kubernetesDir=/etc/kubernetes
## 设置proxy的集群参数
kubectl config set-cluster kubernetes \
--certificate-authority=$kubernetesTLSDir/ca.pem \
--embed-certs=true \
--server=https://192.168.14.9:6443 \
--kubeconfig=$kubernetesDir/kube-proxy.kubeconfig

## 设置kube-proxy用户的参数
kubectl config set-credentials kube-proxy \
--client-certificate=$kubernetesTLSDir/proxy.pem \
--client-key=$kubernetesTLSDir/proxy.key \
--embed-certs=true \
--kubeconfig=$kubernetesDir/kube-proxy.kubeconfig

## 设置kubernetes集群中kube-proxy用户的上下文参数
kubectl config set-context default \
--cluster=kubernetes \
--user=kube-proxy \
--kubeconfig=$kubernetesDir/kube-proxy.kubeconfig

## 设置kube-proxy用户的默认上下文参数
kubectl config use-context default --kubeconfig=$kubernetesDir/kube-proxy.kubeconfig
```

### 创建kube bootstapping kubeconfig文件即集群参数

```
export BOOTSTRAP_TOKEN=`cat /etc/kubernetes/BOOTSTRAP_TOKEN`
## 设置kubelet的集群参数
kubectl config set-cluster kubernetes \
--certificate-authority=$kubernetesTLSDir/ca.pem \
--embed-certs=true \
--server=https://192.168.14.9:6443 \
--kubeconfig=$kubernetesDir/bootstrap.kubeconfig

## 设置kubelet用户的参数
kubectl config set-credentials kubelet-bootstrap \
--token=$BOOTSTRAP_TOKEN \
--kubeconfig=$kubernetesDir/bootstrap.kubeconfig

## 设置kubernetes集群中kubelet用户的默认上下文参数
kubectl config set-context default \
--cluster=kubernetes \
--user=kubelet-bootstrap \
--kubeconfig=$kubernetesDir/bootstrap.kubeconfig

## 设置kubelet用户的默认上下文参数
kubectl config use-context default \
--kubeconfig=$kubernetesDir/bootstrap.kubeconfig

## 创建kubelet的RABC角色
kubectl create --insecure-skip-tls-verify clusterrolebinding kubelet-bootstrap \
--clusterrole=system:node-bootstrapper \
--user=kubelet-bootstrap
```

##  Kubernetes node节点安装

将master节点相关配置拷贝到node节点

```
cd /etc/kubernetes

scp -r bootstrap.kubeconfig config kube-proxy.kubeconfig kubernetesTLS 192.168.14.203:/etc/kubernetes

scp -r bootstrap.kubeconfig config kube-proxy.kubeconfig kubernetesTLS 192.168.14.34:/etc/kubernetes
```

-  etcd集群的节点和flannel网络需要，所以也必须有etcd的证书
```
ls /etc/etcd/etcdSSL
ca.pem  etcd-key.pem  etcd.pem
```
```
scp kubelet kube-proxy 192.168.14.203:/usr/local/bin
scp kubelet kube-proxy  192.168.14.34:/usr/local/bin
```

### 安装docker-ce

这里安装的是最新版的docker-ce.安装包到官网下载安装:
https://download.docker.com/linux/centos/7/x86_64/stable/Packages/

启动docker
```
systemctl enable docker
systemctl start docker
```

### 安装kubelet

```
[Unit]
Description=Kubernetes Kubelet Server
Documentation=https://github.com/GoogleCloudPlatform/kubernetes
After=docker.service
Requires=docker.service

[Service]
EnvironmentFile=-/etc/kubernetes/config
EnvironmentFile=-/etc/kubernetes/kubelet
ExecStart=/usr/local/bin/kubelet \
            $KUBE_LOGTOSTDERR \
            $KUBE_LOG_LEVEL \
            $KUBELET_CONFIG\
            $KUBELET_ADDRESS \
            $KUBELET_PORT \
            $KUBELET_HOSTNAME \
            $KUBELET_POD_INFRA_CONTAINER \
            $KUBELET_ARGS
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

```
cat /etc/kubernetes/kubelet
# kubelet (minion) config
#
## The address for the info server to serve on (set to 0.0.0.0 or "" for all interfaces)
KUBELET_ADDRESS="--address=0.0.0.0"
#
## The port for the info server to serve on
KUBELET_PORT="--port=10250"
#
## You may leave this blank to use the actual hostname
KUBELET_HOSTNAME="--hostname-override=192.168.14.34"
#
## location of the api-server
KUBELET_CONFIG="--kubeconfig=/etc/kubernetes/kubelet.kubeconfig"
#
## pod infrastructure container
KUBELET_POD_INFRA_CONTAINER="--pod-infra-container-image=reg.ops.com/google_containers/pause-amd64:3.1"
#
## Add your own!
KUBELET_ARGS="--cluster-dns=10.0.6.200  --serialize-image-pulls=false  --bootstrap-kubeconfig=/etc/kubernetes/bootstrap.kubeconfig  --kubeconfig=/etc/kubernetes/kubelet.kubeconfig  --cert-dir=/etc/kubernetes/kubernetesTLS  --cluster-domain=cluster.local.  --hairpin-mode=promiscuous-bridge "
```

- 启动kubelet
```
systemctl enable kubelet
systemctl start kubelet
```

- 在master节点认证csr
```
kubectl get csr
NAME                                                   AGE       REQUESTOR           CONDITION
node-csr-m1zSkjPvdWiDfS6Tpct_XMULRZ5uZ4UoSSH9Exx7gjk   13s       kubelet-bootstrap   Pending

kubectl certificate approve node-csr-m1zSkjPvdWiDfS6Tpct_XMULRZ5uZ4UoSSH9Exx7gjk
certificatesigningrequest.certificates.k8s.io/node-csr-m1zSkjPvdWiDfS6Tpct_XMULRZ5uZ4UoSSH9Exx7gjk approved

kubectl get node
NAME            STATUS    ROLES     AGE       VERSION
192.168.14.34   Ready     <none>    14s       v1.11.0
```

### 安装kube-proxy