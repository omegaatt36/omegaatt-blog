---
title: CentOS 7 kubernetes + containerd + calico basic installation tutorial
date: 2022-02-20
categories:
 - develop
---

# 前言

鑒於最近接到 1 CentOS master + 2 Gentoo node k8s cluster 建置雜事，被自己不熟系統雷到，做個筆記紀錄一下，未來敲敲指令就可以了
```bash
                 ..                    root@master
               .PLTJ.                  -----------
              <><><><>                 OS: CentOS Linux 7 (Core) x86_64
     KKSSV' 4KKK LJ KKKL.'VSSKK        Host: KVM/QEMU (Standard PC (i440FX + PIIX, 1996) pc-i440fx-6.1)
     KKV' 4KKKKK LJ KKKKAL 'VKK        Kernel: 5.4.180-1.el7.elrepo.x86_64
     V' ' 'VKKKK LJ KKKKV' ' 'V        Uptime: 2 mins
     .4MA.' 'VKK LJ KKV' '.4Mb.        Packages: 359 (rpm)
   . KKKKKA.' 'V LJ V' '.4KKKKK .      Shell: bash 4.2.46
 .4D KKKKKKKA.'' LJ ''.4KKKKKKK FA.    Terminal: /dev/pts/0
<QDD ++++++++++++  ++++++++++++ GFD>   CPU: Common KVM processor (4) @ 3.493GHz
 'VD KKKKKKKK'.. LJ ..'KKKKKKKK FV     Memory: 97MiB / 16015MiB
   ' VKKKKK'. .4 LJ K. .'KKKKKV '
      'VK'. .4KK LJ KKA. .'KV'
     A. . .4KKKK LJ KKKKA. . .4
     KKA. 'KKKKK LJ KKKKK' .4KK
     KKSSA. VKKK LJ KKKV .4SSKK
              <><><><>
               'MKKM'
                 ''
```
此篇 CentOS 寄宿於 PVE 下，kernel 已升級為 `5.4.180-1.el7.elrepo.x86_64`
使用 containerd 作為 cri
使用 calico 作為 cni，並使用 host 唯一的網卡，不做其他進階設定。

# 安裝流程

1. 安裝依賴
  ```bash
  yum install -y conntrack iptables wget vim
  ```
2. 防火牆設成 `iptables`
  ```bash
  systemctl stop firewalld && systemctl disable firewalld && yum -y remove firewalld
  yum -y install iptables-services && systemctl start iptables && systemctl enable iptables
  ```
3. 關閉 swap
  ```bash
  swapoff -a
  sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
  ```
4. 關閉 SELINUX
  ```bash
  setenforce 0
  sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config
  ```
5. 調整 kernel 參數
  ```bash
  touch /etc/sysctl.d/kubernetes.conf
  vim /etc/sysctl.d/kubernetes.conf
  ```
  輸入下面內容
  ```bash
  net.bridge.bridge-nf-call-iptables=1
  net.bridge.bridge-nf-call-ip6tables=1
  net.ipv4.ip_forward=1
  net.ipv4.tcp_tw_recycle=0
  vm.swappiness=0
  vm.overcommit_memory=1
  vm.panic_on_oom=0
  fs.inotify.max_user_instances=8192
  fs.inotify.max_user_watches=1048576
  fs.file-max=52706963
  fs.nr_open=52706963
  net.ipv6.conf.all.disable_ipv6=1
  net.netfilter.nf_conntrack_max=2310720
  ```
  然後
  ```bash
  sysctl -p /etc/sysctl.d/kubernetes.conf
  ```
6. 調整系統時區
  ```bash
  timedatectl set-timezone Asia/Taipei
  timedatectl set-local-rtc 0
  ```
7. ipvs 設定
  ```
  modprobe br_netfilter
  vim /etc/sysconfig/modules/ipvs.modules
  ```
  主要是把 `nf_conntrack_ipv4` 改為 `nf_conntrack`
  ```bash
  #!/bin/bash
  modprobe -- ip_vs
  modprobe -- ip_vs_rr
  modprobe -- ip_vs_wrr
  modprobe -- ip_vs_sh
  modprobe -- nf_conntrack 
  ```
  更改完後執行
  ```bash
  chmod 755 /etc/sysconfig/modules/ipvs.modules
  bash /etc/sysconfig/modules/ipvs.modules
  lsmod | grep -e ip_vs -e nf_conntrack_ipv4
  ```
8. 安裝 containerd，改用 docker ce 作為 yum repo，因為踩到了 [containerd 版本過舊的坑](https://github.com/containerd/containerd/issues/4901)
  ```bash
  yum install -y yum-utils device-mapper-persistent-data lvm2
  # 改為 docker.ce 作為 containerd 的 repo
  yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
  yum install containerd -y

  containerd config default > /etc/containerd/config.toml
  systemctl restart containerd
  systemctl status containerd
  ```
  開機自動啟動
  ```bash
  systemctl enable containerd
  ```
9. 安裝 crictl
  ```bash
  wget https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.20.0/crictl-v1.20.0-linux-amd64.tar.gz
  tar zxvf crictl-v1.20.0-linux-amd64.tar.gz -C /usr/local/bin
  ```
  `vim /etc/crictl.yaml` 設定 cri 為 containerd
  ```bash
  runtime-endpoint: unix:///run/containerd/containerd.sock
  image-endpoint: unix:///run/containerd/containerd.sock
  timeout: 10
  debug: false
  ```
  看看是否設定正常
  ```bash
  crictl  pull nginx
  crictl  images
  crictl  rmi nginx
  ```
10. 安裝 kubeadm
  `vim /etc/yum.repos.d/kubernetes.repo` 編輯 repo
  ```bash
  [kubernetes]
  name=Kubernetes
  baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
  enabled=1
  gpgcheck=1
  repo_gpgcheck=1
  gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
  ```
  > centos 7 目前可能會遇到 ["repomd.xml signature could not be verified for kubernetes"](https://github.com/kubernetes/kubernetes/issues/60134) 的問題，可將 `repo_gpgcheck` 改為 `0`

  完成後 `yum install -y kubeadm` 便會將 `kubelet`、`kubectl` 等依賴一併安裝
  開機自動啟動
  ```bash
  systemctl enable kubelet.service
  ```
11. 設定 iptables rules
  `vim /etc/sysconfig/iptables` 主要加上幾個 apiserver & nodeport 會用到的 port
  ```bash
  -A INPUT -p tcp -m state --state NEW -m tcp --dport 6443 -j ACCEPT
  -A INPUT -p tcp -m state --state NEW -m tcp --dport 2379:2380 -j ACCEPT
  -A INPUT -p tcp -m state --state NEW -m tcp --dport 10250:10253 -j  ACCEPT
  -A INPUT -p tcp -m state --state NEW -m tcp --dport 10250 -j ACCEPT
  -A INPUT -p tcp -m state --state NEW -m tcp --dport 30000:32767 -j ACCEPT
  -A INPUT -p tcp -m state --state NEW -m tcp --dport 8472 -j ACCEPT
  ```
  編輯完後套用設定
  ```bash
  systemctl daemon-reload
  systemctl restart iptables
  systemctl restart kubelet
  ```
12. kubeadm 初始化 master node
  先將 config 導出加以設置
  ```bash
  kubeadm config print init-defaults > kubeadm-init.yaml
  ```
  `vim kubeadm-config.yaml` 更改以下設置
  ```
  localAPIEndpoint:
    # master node 的 ip
    advertiseAddress: 192.168.101.91 
  nodeRegistration:
    # 更改 continer runtime interface 為 containerd
    criSocket: unix:///run/containerd/containerd.sock
  networking:
    podSubnet: "10.168.0.0/16"
    # cidr 在設定 calico 時會用到
    serviceSubnet: 10.96.0.0/16
  ```
  完成後 `kubeadm init --config=kubeadm-init.yaml` 初始化 master node，訊息會提示你將 `/etc/kubernetes/admin.conf` 複製到 `$HOME/.kube/config`
  ```bash
  mkdir -p $HOME/.kube
  cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  chown $(id -u):$(id -g) $HOME/.kube/config
  ```
  並複製下面那一串加入指令以便其他 node 加入(兩個小時會過期)
  >  kubeadm join 192.168.101.107:6443 --token abcdef.0123456789abcdef --discovery-token-ca-cert-hash sha256:c33e1d061b0a029df6bfa7182345b7c01b62b9e769a44db62d33ce690f40ac1e
  或是使用 `kubeadm token create --print-join-command` 建立一個新的
13. 設定 calico 作為 cni
  ```bash
  wget https://docs.projectcalico.org/manifests/calico.yaml
  ```
  `vim calico.yaml` 編輯內容如下，主要修改 env 下的。
  ```bash
  # 修改 CALICO_IPV4POOL_CIDR 與上面 kubeadm-init.yaml 中的 cidr 相同
  - name: CALICO_IPV4POOL_CIDR
    value: "10.168.0.0/16"
  
  # (optional)
  # 新增 IP_AUTODETECTION_METHOD，讓他在 bird-check 時綁 host nic
  - name: IP_AUTODETECTION_METHOD
    value: "cidr=192.168.101.0/24"
  ```
  設定完成後 `kubectl apply -f calico.yaml` 套用設定
  可以透過 `watch -n 1 kubectl get pods -A` 查看 namespace `kube-system` 下的 pod 是否都正常 running
  ```bash
  Every 1.0s: kubectl get pods -A                                                                                                                                                                                                     Sun Feb 20 17:19:56 2022

  NAMESPACE     NAME                                       READY   STATUS    RESTARTS   AGE
  kube-system   calico-kube-controllers-566dc76669-dngc4   1/1     Running   0          125m
  kube-system   calico-node-wg6dn                          1/1     Running   0          125m
  kube-system   coredns-64897985d-h4n45                    1/1     Running   0          123m
  kube-system   coredns-64897985d-tkxsw                    1/1     Running   0          123m
  kube-system   etcd-node                                  1/1     Running   0          134m
  kube-system   kube-apiserver-node                        1/1     Running   0          134m
  kube-system   kube-controller-manager-node               1/1     Running   0          134m
  kube-system   kube-proxy-vhq2j                           1/1     Running   0          134m
  kube-system   kube-scheduler-node                        1/1     Running   0          134m
  ```

至此 master node 就順利設定完成了

# 疑難排解
1. cri 設定成 `containerd` 後 `kubeadm init` 時 `kubelet` 開不起來
  可以參考 [containerd issue 4901](https://github.com/containerd/containerd/issues/4901)
  原因可能為 containerd 版本太舊
2. `calico-node` 沒 ready
  可能要看一下 host dns 設定是否正常
  或是修改 `calico.yaml` 中的 `IP_AUTODETECTION_METHOD` 設定，可以參考 [calico issue 4197](https://github.com/projectcalico/calico/issues/4197)