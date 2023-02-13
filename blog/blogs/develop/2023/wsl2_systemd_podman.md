---
title: WSL2 中使用 systemd 管理 podman 的 container
date: 2023-02-13
categories:
 - develop
tags:
 - linux
 - wsl
---

原本使用 [wsl-distrod](https://github.com/nullpo-head/wsl-distrod) 來作為 wsl 中 systemd 的實現方式，後來 [microsoft 宣布 wsl 支持 systemd](https://devblogs.microsoft.com/commandline/systemd-support-is-now-available-in-wsl/) 後，distrod 等等 repo 就沒有在更新了呢，是巧合嗎?我不這麼認為。

這篇簡略紀錄 wsl2 中啟用 systemd 作為 process 管理工具以及配合紅帽推出的 [podman](https://podman.io/) 來取代(斷捨離)docker desktop。

## 環境

wsl2 + ubuntu 22.04

```bash
❯ neofetch
            .-/+oossssoo+/-.               raiven@raiven 
        `:+ssssssssssssssssss+:`           ------------- 
      -+ssssssssssssssssssyyssss+-         OS: Ubuntu 22.04.1 LTS on Windows 10 x86_64 
    .ossssssssssssssssssdMMMNysssso.       Kernel: 5.10.43.3-microsoft-standard-WSL2 
   /ssssssssssshdmmNNmmyNMMMMhssssss/      Uptime: 3 hours, 10 mins 
  +ssssssssshmydMMMMMMMNddddyssssssss+     Packages: 2005 (dpkg), 4 (snap) 
 /sssssssshNMMMyhhyyyyhmNMMMNhssssssss/    Shell: zsh 5.8.1 
.ssssssssdMMMNhsssssssssshNMMMdssssssss.   Theme: Adwaita [GTK3] 
+sssshhhyNMMNyssssssssssssyNMMMysssssss+   Icons: Adwaita [GTK3] 
ossyNMMMNyMMhsssssssssssssshmmmhssssssso   Terminal: Windows Terminal 
ossyNMMMNyMMhsssssssssssssshmmmhssssssso   CPU: AMD Ryzen 9 5900X (24) @ 3.700GHz 
+sssshhhyNMMNyssssssssssssyNMMMysssssss+   GPU: dbac:00:00.0 Microsoft Corporation Device 008e 
.ssssssssdMMMNhsssssssssshNMMMdssssssss.   Memory: 1427MiB / 15966MiB 
 /sssssssshNMMMyhhyyyyhdNMMMNhssssssss/
  +sssssssssdmydMMMMMMMMddddyssssssss+                             
   /ssssssssssshdmNNNNmyNMMMMhssssss/                              
    .ossssssssssssssssssdMMMNysssso.
      -+sssssssssssssssssyyyssss+-
        `:+ssssssssssssssssss+:`
            .-/+oossssoo+/-.
```

## wsl2 中啟用 systemd

- 請確保 wsl2 的版本是 version 2，可以在 cmd/powershell 中查看
    ```bash
    C:\Users\omega>wsl --list -v
      NAME            STATE           VERSION
    * Ubuntu-22.04    Running         2
    ```
- 在 wsl2 中修改 `/etc/wsl.conf`，確保 wsl2 重啟時
    ```bash
    sudo vim /etc/wsl.conf
    ```
- 在 cmd/powershell 中關閉 wsl 執行個體，會自己重啟
    ```bash
    wsl --shutdown
    ```

- 重啟完成後在 wsl 中檢查是否已經正確啟用
    ```bash
    ❯ systemctl list-unit-files --type=service
    UNIT FILE                                  STATE           VENDOR PRESET
    accounts-daemon.service                    enabled         enabled
    acpid.service                              masked          enabled
    alsa-restore.service                       static          -
    alsa-state.service                         static          -
    alsa-utils.service                         masked          enabled
    anacron.service                            enabled         enabled
    apparmor.service                           enabled         enabled
    apport-autoreport.service                  static          -

    ❯ sudo systemctl status | cat
    [sudo] password for raiven:
    ● raiven
        State: degraded
         Jobs: 0 queued
       Failed: 5 units
        Since: Mon 2023-02-13 19:21:06 CST; 3h 22min ago
       CGroup: /
               ├─user.slice
               │ ├─user-0.slice
               │ │ └─session-c2.scope
               │ │   ├─ 978 /bin/login -f
               │ │   └─1029 -bash
    ```
- 若啟用後發現無論是 `sudo apt update` 或是 `sudo systemctl status` 都跑得很慢，甚至跳出 `Transport Endpoint Is Not Connected`，可以參考[這篇 issue 的解決辦法](https://github.com/microsoft/WSL/issues/8904#issuecomment-1324249768)
    ```bash
    sudo ln -s /dev/null /etc/systemd/system/acpid.service
    sudo ln -s /dev/null /etc/systemd/system/acpid.path
    ```

## 安裝 podman

- 按照[官方文件](https://podman.io/getting-started/installation#installing-on-linux)安裝即可
    ```bash
    sudo apt update
    sudo apt -y install podman
    ```
- 可以檢查/修改 `/etc/containers/registries.conf` 來添加自己慣用的 container registry，好比自己的 [harbor](https://goharbor.io/) 服務
    ```bash
    vim /etc/containers/registries.conf
    
    unqualified-search-registries=[
      "quay.io",
      "gcr.io",
      "docker.io",
      "hao123.omegaatt.com"
    ]
    ```
- 可以設定別名，讓一些使用 docker 的腳本無痛銜接
    ```bash
    vim ~/.zshrc

    alias docker='podman'
    ```
- 或是在腳本中使用變數
    ```bash
    # xxx.sh
    #!/bin/bash
    DOCKER=podman
    $DOCKER ps
    ```
- 以及在 Makefile 中使用變數
    ```bash
    # Makefile
    DOCKER=podman
    ps:
	    $(DOCKER) ps
    ```

## 使用 systemd 管理 podman container

雖然說 podman 可以使用介於 Docker 與 Kubernetes 之間的 [Pod](https://docs.podman.io/en/latest/markdown/podman-pod.1.html)，但那並不是此篇文章的重點，詳細可以參考紅帽的[教學](https://access.redhat.com/documentation/zh-cn/red_hat_enterprise_linux/8/html/building_running_and_managing_containers/proc_auto-generating-a-systemd-unit-file-using-podman_assembly_porting-containers-to-systemd-using-podman)。

舉例來說我需要在某台機器上部署 [drone-runner](https://docs.drone.io/runner/docker/installation/linux/):

- 建立 podman container
    ```bash
    podman create \ 
        -v /run/podman/podman.sock:/var/run/docker.sock \
        -e DRONE_RPC_HOST=$DRONE_HOST \
        -e DRONE_RPC_PROTO=http \
        -e DRONE_RPC_SECRET=$DRONE_SECRET \
        -e DRONE_RUNNER_CAPACITY=3 \
        --name drone-runner \
        --restart on-failure 
        docker.io/drone/drone-runner-docker:1
    ```
- [將 container 轉成 systemd service file](https://docs.podman.io/en/latest/markdown/podman-generate-systemd.1.html)
    ```bash
    podman generate systemd --new --files --name drone-runner
    ```
- 複製 systemd service file
    ```bash
    sudo cp -Z container-drone-runner.service /etc/systemd/system
    ```
- 啟用
    ```bash
    sudo systemctl enable container-drone-runner.service
    sudo systemctl start container-drone-runner.service
    sudo systemctl status container-drone-runner.service
    sudo journalctl -f -u container-drone-runner.service
    ```

## 檢查成果

- 透過 `pstree` 來看 systemd 與 podman 的 container 的互動，可以看到 `systemd.conmon.drone-runner-do` 正是剛剛部署的 drone runner。
    ```bash
    ❯ pstree
    systemd─┬─ModemManager───2*[{ModemManager}]
            ├─agetty
            ├─bash───frpc───8*[{frpc}]
            ├─conmon─┬─drone-runner-do───12*[{drone-runner-do}]
            │        └─{conmon}
            ├─containerd-shim─┬─2*[entry]
            │                 ├─pause
            │                 └─13*[{containerd-shim}]
            ├─containerd-shim─┬─argocd-applicat───12*[{argocd-applicat}]
            │                 ├─pause
            │                 └─12*[{containerd-shim}]
            ├─containerd-shim─┬─argocd-server───12*[{argocd-server}]
            │                 ├─pause
            │                 └─12*[{containerd-shim}]
            ├─containerd-shim─┬─pause
            │                 ├─s6-svscan─┬─s6-supervise───s6-linux-init-s
            │                 │           ├─s6-supervise
            │                 │           ├─s6-supervise───s6-ipcserverd
            │                 │           ├─s6-supervise───php
            │                 │           ├─s6-supervise───php-fpm8───3*[php-fpm8]
            │                 │           ├─s6-supervise───crond
            │                 │           └─s6-supervise───nginx───4*[nginx]
            │                 └─13*[{containerd-shim}]
    ```