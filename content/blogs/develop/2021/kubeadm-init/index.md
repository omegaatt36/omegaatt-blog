---
title: Kubernetes 學習筆記 - Ubuntu kubeamd 建置 cluster
date: 2021-09-19
categories:
  - develop
tags:
  - kubernetes
---

## 環境

物理機: 2278G/16G DDR4 ECC\*4/1T MX500
OS: - master: Ubuntu server 20.04 - node: Ubuntu server 20.04

## 安裝 docker、kubeadm、kubelet、kubectl

### 安裝 docker

```shell
sudo apt install apt-transport-https ca-certificates curl gnupg-agent software-properties-common -y

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io -y

sudo usermod -aG docker $USER
```

登出再登入

### 安裝 kubeadm kubelet kubectl

```shell
sudo apt-get install -y apt-transport-https curl
sudo su
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

sudo cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
exit

sudo apt-get update

# apt-cache madison kubeadm
# K_VER="1.20.5-00"
# apt install -y kubelet=${K_VER} kubectl=${K_VER} kubeadm=${K_VER}

# 不指定版本的話直接安裝即可
sudo apt install -y kubelet kubeadm kubectl
# 若需要鎖定版本可以使用 apt-mark hold
sudo apt-mark hold kubelet kubeadm kubectl
```

### 設定 kubeadm

#### master

```shell
sudo kubeadm init \
    --pod-network-cidr 網路區段 \
    --apiserver-advertise-address 本機IP \
    --apiserver-cert-extra-sans gcp IP

# kubeadm init --pod-network-cidr 172.100.0.0/16 --apiserver-advertise-address 10.140.0.2 --apiserver-cert-extra-sans 130.211.253.131
```

結束後會看到

```shell
Your Kubernetes control-plane has initialized successfully!
To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

Alternatively, if you are the root user, you can run:

  export KUBECONFIG=/etc/kubernetes/admin.conf

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
    https://kubernetes.io/docs/concepts/cluster-administration/addons/
Then you can join any number of worker nodes by running the following on each as root:
    kubeadm join 10.140.0.2:6443 --token xgfeim.4mum4i5fh1uvchv2 \
    --discovery-token-ca-cert-hash sha256:9fd129841267a930532f46ccf868f3229ec13e0b2d09c589421402aef13fa2f8
```

照著他給的指令去建立 config 後可以去尋找喜歡的 [CNI](https://kubernetes.io/docs/concepts/cluster-administration/addons/#networking-and-network-policy) 根據他的安裝指令設定網路

例如若是想使用 [Weave Net](https://www.weave.works/docs/net/latest/kubernetes/kube-addon/) 作為 CNI 則可以

```shell
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')"
```

並可在 node 透過最下面 `kubeadm join ...` 來加入這個 cluster

待 node 加入後可以透過 `kubectl get nodes` 來確認 node 狀態

#### node

執行 `kubeadm join ...` 來加入 cluster
