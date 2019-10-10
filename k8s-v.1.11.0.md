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

- 创建openssl.cnf

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

- 启动docker

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

```
cat /usr/lib/systemd/system/kube-proxy.service
[Unit]
Description=Kube Proxy Service
After=network.target

[Service]
Type=simple
EnvironmentFile=-/etc/kubernetes/config
EnvironmentFile=-/etc/kubernetes/proxy
ExecStart=/usr/local/bin/kube-proxy \
            $KUBE_LOGTOSTDERR \
            $KUBE_LOG_LEVEL \
            $KUBE_MASTER \
            $KUBE_PROXY_ARGS

Restart=always
LimitNOFILE=65536

[Install]
WantedBy=default.target
```

```
cat /etc/kubernetes/proxy
###
# kubernetes proxy config

# defaults from config and proxy should be adequate

# Add your own!
KUBE_PROXY_ARGS="--kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig  --cluster-cidr=172.16.0.0/16"
```

- 启动kube-proxy

```
systemctl daemon-reload
systemctl enable kube-proxy
systemctl start kube-proxy
```

- 开启iptables上FORWARD表

```
iptables -P FORWARD ACCEPT
```

## 安装网络插件

### flannel网络插件安装

无论master节点还是node节点都需要安装

```
~]# yum install flannel

~]# cat > flannel-config.json <<  EOF
{
    "Network": "172.16.0.0/16",
    "SubnetLen": 24,
    "Backend": {
        "Type": "vxlan"
    }
}
EOF

~]# etcdctl --ca-file=/etc/etcd/etcdSSL/ca.pem  --cert-file=/etc/etcd/etcdSSL/etcd.pem   --key-file=/etc/etcd/etcdSSL/etcd-key.pem set /k8s/network/config < flannel-config.json

~]#  cat /etc/sysconfig/flanneld
# Flanneld configuration options  

# etcd url location.  Point this to the server where etcd runs
FLANNEL_ETCD_ENDPOINTS="https://192.168.14.9:2379,https://192.168.14.34:2379,https://192.168.14.203:2379"

# etcd config key.  This is the configuration key that flannel queries
# For address range assignment
FLANNEL_ETCD_PREFIX="/k8s/network"

# Any additional options that you want to pass
FLANNEL_OPTIONS="--etcd-cafile=/etc/etcd/etcdSSL/ca.pem  --etcd-certfile=/etc/etcd/etcdSSL/etcd.pem  --etcd-keyfile=/etc/etcd/etcdSSL/etcd-key.pem --log_dir=/var/log/k8s/flannel/"
```

- 启动flannel网络

```
systemctl  enable flanneld
systemctl start flanneld
```

###  docker使用flannel网络

```
~]# etcdctl   --ca-file=/etc/etcd/etcdSSL/ca.pem   --cert-file=/etc/etcd/etcdSSL/etcd.pem   --key-file=/etc/etcd/etcdSSL/etcd-key.pem ls /k8s/network/subnets
/k8s/network/subnets/172.16.54.0-24
/k8s/network/subnets/172.16.91.0-24
/k8s/network/subnets/172.16.81.0-24
```

- 每个节点下会生成子网段信息

```
~]#  cat /run/flannel/subnet.env
FLANNEL_NETWORK=172.16.0.0/16
FLANNEL_SUBNET=172.16.54.1/24
FLANNEL_MTU=1450
FLANNEL_IPMASQ=false
```

- 在node节点上的docker应用flannel网络

```
~]# cat /usr/lib/systemd/system/docker.service
[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
BindsTo=containerd.service
After=network-online.target firewalld.service containerd.service flannled.service
Wants=network-online.target
Requires=docker.socket

[Service]
Type=notify
# the default is not to use systemd for cgroups because the delegate issues still
# exists and systemd currently does not support the cgroup feature set required
# for containers run by docker
EnvironmentFile=-/var/run/flannel/subnet.env
ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock  --bip=${FLANNEL_SUBNET} --mtu=${FLANNEL_MTU}
ExecReload=/bin/kill -s HUP $MAINPID
TimeoutSec=0
RestartSec=2
Restart=always

# Note that StartLimit* options were moved from "Service" to "Unit" in systemd 229.
# Both the old, and new location are accepted by systemd 229 and up, so using the old location
# to make them work for either version of systemd.
StartLimitBurst=3

# Note that StartLimitInterval was renamed to StartLimitIntervalSec in systemd 230.
# Both the old, and new name are accepted by systemd 230 and up, so using the old name to make
# this option work for either version of systemd.
StartLimitInterval=60s

# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity

# Comment TasksMax if your systemd version does not support it.
# Only systemd 226 and above support this option.
TasksMax=infinity

# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes

# kill only the docker process, not all processes in the cgroup
KillMode=process

[Install]
WantedBy=multi-user.target

```

- 启动docker

```
systemctl daemon-reload
systemctl enable docker
systemctl start docker
```

## CoreDNS安装

```
~]# cat coredns.yml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: coredns
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:coredns
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - services
  - pods
  - namespaces
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:coredns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:coredns
subjects:
- kind: ServiceAccount
  name: coredns
  namespace: kube-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health
        kubernetes cluster.local 10.0.6.0/24 {
          pods insecure
          upstream
          fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        proxy . /etc/resolv.conf
        cache 30
        reload
    }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: coredns
  namespace: kube-system
  labels:
    k8s-app: kube-dns
    kubernetes.io/name: "CoreDNS"
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  selector:
    matchLabels:
      k8s-app: kube-dns
  template:
    metadata:
      labels:
        k8s-app: kube-dns
    spec:
      priorityClassName: system-cluster-critical
      serviceAccountName: coredns
      tolerations:
        - key: "CriticalAddonsOnly"
          operator: "Exists"
      nodeSelector:
        beta.kubernetes.io/os: linux
      containers:
      - name: coredns
        image: coredns/coredns:1.3.1
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            memory: 170Mi
          requests:
            cpu: 100m
            memory: 70Mi
        args: [ "-conf", "/etc/coredns/Corefile" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
          readOnly: true
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        - containerPort: 9153
          name: metrics
          protocol: TCP
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - all
          readOnlyRootFilesystem: true
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
      dnsPolicy: Default
      volumes:
        - name: config-volume
          configMap:
            name: coredns
            items:
            - key: Corefile
              path: Corefile
---
apiVersion: v1
kind: Service
metadata:
  name: kube-dns
  namespace: kube-system
  annotations:
    prometheus.io/port: "9153"
    prometheus.io/scrape: "true"
  labels:
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "CoreDNS"
spec:
  selector:
    k8s-app: kube-dns
  clusterIP:  10.0.6.200
  ports:
  - name: dns
    port: 53
    protocol: UDP
  - name: dns-tcp
    port: 53
    protocol: TCP
  - name: metrics
    port: 9153
    protocol: TCP
```

- 启动coredns

```
kubectl apply -f coredns.yml
```

## Traefik ingress安装

- traefik rbac

```
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: traefik-ingress-controller
rules:
  - apiGroups:
      - ""
    resources:
      - services
      - endpoints
      - secrets
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - extensions
    resources:
      - ingresses
    verbs:
      - get
      - list
      - watch
  - apiGroups:
    - extensions
    resources:
    - ingresses/status
    verbs:
    - update
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: traefik-ingress-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: traefik-ingress-controller
subjects:
- kind: ServiceAccount
  name: traefik-ingress-controller
  namespace: kube-system
```

- traefik daemonset

```
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
---
kind: DaemonSet
apiVersion: extensions/v1beta1
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
  labels:
    k8s-app: traefik-ingress-lb
spec:
  template:
    metadata:
      labels:
        k8s-app: traefik-ingress-lb
        name: traefik-ingress-lb
    spec:
      serviceAccountName: traefik-ingress-controller
      terminationGracePeriodSeconds: 60
      containers:
      - image: traefik
        name: traefik-ingress-lb
        ports:
        - name: http
          containerPort: 80
          hostPort: 80
        - name: admin
          containerPort: 8080
          hostPort: 8080
        securityContext:
          capabilities:
            drop:
            - ALL
            add:
            - NET_BIND_SERVICE
        args:
        - --api
        - --kubernetes
        - --logLevel=INFO
---
kind: Service
apiVersion: v1
metadata:
  name: traefik-ingress-service
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
    - protocol: TCP
      port: 80
      name: web
    - protocol: TCP
      port: 8080
      name: admin
```

- traefik ui ingress

```
---
apiVersion: v1
kind: Service
metadata:
  name: traefik-web-ui
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
  - name: web
    port: 80
    targetPort: 8080
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: traefik-web-ui
  namespace: kube-system
spec:
  rules:
  - host: traefik-ui.ops.com
    http:
      paths:
      - path: /
        backend:
          serviceName: traefik-web-ui
          servicePort: web
```

## Nginx ingress安装

- namespace.yml

```
apiVersion: v1
kind: Namespace
metadata:
  name: ingress-nginx
```

- default-backend.yml

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: default-http-backend
  labels:
    app: default-http-backend
  namespace: ingress-nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: default-http-backend
  template:
    metadata:
      labels:
        app: default-http-backend
    spec:
      terminationGracePeriodSeconds: 60
      containers:
      - name: default-http-backend
        # Any image is permissible as long as:
        # 1. It serves a 404 page at /
        # 2. It serves 200 on a /healthz endpoint
        image: reg.ops.com/google_containers/defaultbackend:1.4
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 30
          timeoutSeconds: 5
        ports:
        - containerPort: 8080
        resources:
          limits:
            cpu: 100m
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 200Mi
---
apiVersion: v1
kind: Service
metadata:
  name: default-http-backend
  namespace: ingress-nginx
  labels:
    app: default-http-backend
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: default-http-backend
```

- configmap.yml

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: nginx-configuration
  namespace: ingress-nginx
  labels:
    app: ingress-nginx
```

- tcp-services-configmap.yml

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: tcp-services
  namespace: ingress-nginx
```

- udp-services-configmap.yml

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: udp-services
  namespace: ingress-nginx
```

- rbac.yml

```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nginx-ingress-serviceaccount
  namespace: ingress-nginx

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: nginx-ingress-clusterrole
rules:
  - apiGroups:
      - ""
    resources:
      - configmaps
      - endpoints
      - nodes
      - pods
      - secrets
    verbs:
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - get
  - apiGroups:
      - ""
    resources:
      - services
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - "extensions"
    resources:
      - ingresses
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ""
    resources:
        - events
    verbs:
        - create
        - patch
  - apiGroups:
      - "extensions"
    resources:
      - ingresses/status
    verbs:
      - update

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: nginx-ingress-role
  namespace: ingress-nginx
rules:
  - apiGroups:
      - ""
    resources:
      - configmaps
      - pods
      - secrets
      - namespaces
    verbs:
      - get
  - apiGroups:
      - ""
    resources:
      - configmaps
    resourceNames:
      # Defaults to "<election-id>-<ingress-class>"
      # Here: "<ingress-controller-leader>-<nginx>"
      # This has to be adapted if you change either parameter
      # when launching the nginx-ingress-controller.
      - "ingress-controller-leader-nginx"
    verbs:
      - get
      - update
  - apiGroups:
      - ""
    resources:
      - configmaps
    verbs:
      - create
  - apiGroups:
      - ""
    resources:
      - endpoints
    verbs:
      - get

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: nginx-ingress-role-nisa-binding
  namespace: ingress-nginx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: nginx-ingress-role
subjects:
  - kind: ServiceAccount
    name: nginx-ingress-serviceaccount
    namespace: ingress-nginx

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: nginx-ingress-clusterrole-nisa-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: nginx-ingress-clusterrole
subjects:
  - kind: ServiceAccount
    name: nginx-ingress-serviceaccount
    namespace: ingress-nginx
```

- with-rbac.yml

```
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: nginx-ingress-controller
  namespace: ingress-nginx 
spec:
  selector:
    matchLabels:
      app: ingress-nginx
  template:
    metadata:
      labels:
        app: ingress-nginx
      annotations:
        prometheus.io/port: '10254'
        prometheus.io/scrape: 'true'
    spec:
      nodeSelector:
        custom/ingress-controller-ready: "true"
      serviceAccountName: nginx-ingress-serviceaccount
      hostNetwork: true
      containers:
        - name: nginx-ingress-controller
          image: reg.ops.com/google_containers/nginx-ingress-controller:0.11.0
          args:
            - /nginx-ingress-controller
            - --default-backend-service=$(POD_NAMESPACE)/default-http-backend
            - --configmap=$(POD_NAMESPACE)/nginx-configuration
            - --tcp-services-configmap=$(POD_NAMESPACE)/tcp-services
            - --udp-services-configmap=$(POD_NAMESPACE)/udp-services
            - --annotations-prefix=nginx.ingress.kubernetes.io
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          ports:
          - name: http
            containerPort: 80
          - name: https
            containerPort: 443
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 10254
              scheme: HTTP
            initialDelaySeconds: 10
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          readinessProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 10254
              scheme: HTTP
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
```

## kube dashboard

- kubernetes-dashboard

```
# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# ------------------- Dashboard Secret ------------------- #

apiVersion: v1
kind: Secret
metadata:
  labels:
    k8s-app: kubernetes-dashboard
  name: kubernetes-dashboard-certs
  namespace: kube-system
type: Opaque

---
# ------------------- Dashboard Service Account ------------------- #

apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-app: kubernetes-dashboard
  name: kubernetes-dashboard
  namespace: kube-system

---
# ------------------- Dashboard Role & Role Binding ------------------- #

kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kubernetes-dashboard-minimal
  namespace: kube-system
rules:
  # Allow Dashboard to create 'kubernetes-dashboard-key-holder' secret.
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["create"]
  # Allow Dashboard to create 'kubernetes-dashboard-settings' config map.
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["create"]
  # Allow Dashboard to get, update and delete Dashboard exclusive secrets.
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["kubernetes-dashboard-key-holder", "kubernetes-dashboard-certs"]
  verbs: ["get", "update", "delete"]
  # Allow Dashboard to get and update 'kubernetes-dashboard-settings' config map.
- apiGroups: [""]
  resources: ["configmaps"]
  resourceNames: ["kubernetes-dashboard-settings"]
  verbs: ["get", "update"]
  # Allow Dashboard to get metrics from heapster.
- apiGroups: [""]
  resources: ["services"]
  resourceNames: ["heapster"]
  verbs: ["proxy"]
- apiGroups: [""]
  resources: ["services/proxy"]
  resourceNames: ["heapster", "http:heapster:", "https:heapster:"]
  verbs: ["get"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: kubernetes-dashboard-minimal
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kubernetes-dashboard-minimal
subjects:
- kind: ServiceAccount
  name: kubernetes-dashboard
  namespace: kube-system

---
# ------------------- Dashboard Deployment ------------------- #

kind: Deployment
apiVersion: apps/v1beta2
metadata:
  labels:
    k8s-app: kubernetes-dashboard
  name: kubernetes-dashboard
  namespace: kube-system
spec:
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      k8s-app: kubernetes-dashboard
  template:
    metadata:
      labels:
        k8s-app: kubernetes-dashboard
    spec:
      containers:
      - name: kubernetes-dashboard
        image: gcrxio/kubernetes-dashboard-amd64:v1.11.0
        ports:
        - containerPort: 9090
          protocol: TCP
        args:
          #- --auto-generate-certificates
          # Uncomment the following line to manually specify Kubernetes API server Host
          # If not specified, Dashboard will attempt to auto discover the API server and connect
          # to it. Uncomment only if the default does not work.
          # - --apiserver-host=http://my-address:port
        volumeMounts:
        - name: kubernetes-dashboard-certs
          mountPath: /certs
          # Create on-disk volume to store exec logs
        - mountPath: /tmp
          name: tmp-volume
        livenessProbe:
          httpGet:
            scheme: HTTP
            path: /
            port: 9090
          initialDelaySeconds: 30
          timeoutSeconds: 30
      volumes:
      - name: kubernetes-dashboard-certs
        secret:
          secretName: kubernetes-dashboard-certs
      - name: tmp-volume
        emptyDir: {}
      serviceAccountName: kubernetes-dashboard
      # Comment the following tolerations if Dashboard must not be deployed on master
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule

---
# ------------------- Dashboard Service ------------------- #

kind: Service
apiVersion: v1
metadata:
  labels:
    k8s-app: kubernetes-dashboard
  name: kubernetes-dashboard
  namespace: kube-system
spec:
  ports:
    - port: 80
      targetPort: 9090
  selector:
    k8s-app: kubernetes-dashboard
```

- kubernetes-dashboard-ingress

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: kubernetes-dashboard
  namespace: kube-system
spec:
  rules:
  - host: kube-dashboard.ops.com
    http:
      paths:
      - path: /
        backend:
          serviceName: kubernetes-dashboard
          servicePort: 80
```

- kubernetes-dashboard-admin

```
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kubernetes-dashboard
  labels:
    k8s-app: kubernetes-dashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: kubernetes-dashboard
  namespace: kube-system
```