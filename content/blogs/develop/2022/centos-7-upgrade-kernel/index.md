---
title: CentOS 7 upgrade kernel
date: 2022-02-20
categories:
 - develop
tags:
 - linux
aliases:
 - "/blogs/develop/2022/centos-7-upgrade-kernel.html"
---

- 查看目前 kernel 版本
    ```bash
    > uname -a
    Linux R350-1 3.10.0-1160.53.1.el7.x86_64 #1 SMP Fri Jan 14 13:59:45 UTC 2022 x86_64 x86_64 x86_64 GNU/Linux
    ```
- 添加 ELRepo 公鑰
    ```bash
    rpm --import https://www.elrepo.org/RPM-GPG-KEY-elrepo.org
    ```
- 安裝 ELRepo yum 來源
    ```bash
    rpm -Uvh https://www.elrepo.org/elrepo-release-7.0-4.el7.elrepo.noarch.rpm
    ```
- 查看可用的 kernel
    ```bash
    yum --disablerepo="*" --enablerepo="elrepo-kernel" list available
    ```
- 安裝最新的 kernel，挑一個安裝，這邊選擇 kernel-lt
    - kernel-lt 為長期支援版
    - kernel-mt 為 linus 個人維護版
    ```
    yum --enablerepo=elrepo-kernel install kernel-lt
    ```
- 查看目前已安裝的 kernel
    ```bash
    sudo awk -F\' '$1=="menuentry " {print i++ " : " $2}' /etc/grub2.cfg

    0 : CentOS Linux (5.4.180-1.el7.elrepo.x86_64) 7 (Core)
    1 : CentOS Linux (3.10.0-1160.53.1.el7.x86_64) 7 (Core)
    2 : CentOS Linux (3.10.0-1160.el7.x86_64) 7 (Core)
    3 : CentOS Linux (0-rescue-b804ce66fb404eb7a5dd04547e3e972e) 7 (Core)
    ```
    如果結果為空，或是沒有這個文件，可以先進行 `grub2-mkconfig -o /boot/grub2/grub.cfg`

- 設定新的 kernel
    ```bash
    grub2-set-default 0
    reboot
    ```
