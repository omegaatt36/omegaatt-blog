---
title: 解決 VM 關機時等待容器關閉
date: 2023-01-11
categories:
  - develop
tags:
  - linux
---

```
[     *] A stop job is running for libcontainer container xxxxxxxx (10s / 1m30s)_
```

過去 kubernetes 還跟 docker-shim 手牽手時，在關閉 k8s node vm 時常常出現等待 `docker-shim` 關閉，直到一分三十秒被 time out 才正確關機。若是開發機就當作去尿尿的時間就好，但若是重要的環境停機備份，或是斷電時靠 UPS 提供電源等待系統關機，那浪費時間就不好了。

這次在 home lab 中採取省電措施，不用 [1 master + 1 node](https://www.omegaatt.com/blogs/develop/2022/centos-7-kubernetes-install.html) 的組合，直接使用 [k3s](https://k3s.io/) 作為家用 kubernetes 的實驗環境。但依然在關機時會需要等待 containerd-shim 等等被 time out kill 才會完成關機。

若是沒有做特別的設定，或是沒有過分的 stateful pod，找到了 [k3s issue #2400](https://github.com/k3s-io/k3s/issues/2400#issuecomment-1312621468) 中的解決方法

[Stopping K3s](https://docs.k3s.io/upgrades/killall)

> To allow high availability during upgrades, the K3s containers continue running when the K3s service is stopped.
>
> To stop all of the K3s containers and reset the containerd state, the k3s-killall.sh script can be used.

在關閉 k3s service 後，containers 仍會繼續執行，所以可以透過 `/usr/local/bin/k3s-killall.sh` 來強制清除 containers。

於是可以將該 shell script 設定為 systemd service 關閉後執行:

```
sudo cat <<EOF >/etc/systemd/system/shutdown-k3s.service
[Unit]
Description=Kill containerd-shims on shutdown
DefaultDependencies=false
Before=shutdown.target umount.target

[Service]
ExecStart=/usr/local/bin/k3s-killall.sh
Type=oneshot

[Install]
WantedBy=shutdown.target
EOF

sudo systemctl enable shutdown-k3s.service
```

如此一來關機時就會直接清掉 containers 了。
