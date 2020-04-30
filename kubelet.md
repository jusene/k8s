## Kubernetes kubelet证书过期问题

kubelet证书默认有效期为1年

### 解决方案

- 添加参数

kubelet组件配置
```
--feature-gates=RotateKubeletServerCertificate=true
--feature-gates=RotateKubeletClientCertificate=true
--rotate-certificates
```

controller-manager组件配置
```
--experimental-cluster-signing-duration=87600h0m0s
--feature-gates=RotateKubeletServerCertificate=true
```

- 创建自动批准CSR请求的ClusterRole

```
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: system:certificates.k8s.io:certificatesigningrequests:selfnodeserver
rules:
- apiGroups: ["certificates.k8s.io"]
  resources: ["certificatesigningrequests/selfnodeserver"]
  verbs: ["create"]
```

kubelet apply -f tls-instructs-csr.yml

- 自动批准 kubelet-bootstrap 用户 TLS bootstrapping 首次申请证书的 CSR 请求
```
kubectl create clusterrolebinding node-client-auto-approve-csr --clusterrole=system:certificates.k8s.io:certificatesigningrequests:nodeclient --user=kubelet-bootstrap
```

- 自动批准 system:nodes 组用户更新 kubelet 自身与 apiserver 通讯证书的 CSR 请求
```
kubectl create clusterrolebinding node-client-auto-renew-crt --clusterrole=system:certificates.k8s.io:certificatesigningrequests:selfnodeclient --group=system:nodes
```

- 自动批准 system:nodes 组用户更新 kubelet 10250 api 端口证书的 CSR 请求
```
kubectl create clusterrolebinding node-server-auto-renew-crt --clusterrole=system:certificates.k8s.io:certificatesigningrequests:selfnodeserver --group=system:nodes
```

### 重启kube-controller-manager和kubelet

```
systemctl restart kube-controller-manager

# 首先需要自己删除kubelet证书
rm -rf kubelet*
systemctl restart kubelet

等会会自动重新签署kubelet证书，并且是10年期的

openssl x509 -in kubelet-client-current.pem -noout -dates
```
