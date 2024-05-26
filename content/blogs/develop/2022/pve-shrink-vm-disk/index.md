---
title: proxmox ve shrink vm disk size
date: 2022-09-04
categories:
 - develop
tags:
 - linux
 - kubernetes
 - proxmox
---

由於 zpool 吃超過 80%，故將其中一個 VM(k8s-worker) 的硬碟縮小(200GB=>100GB)

- pve: 7.2.7
- zfs:
    - zfs-2.1.5-pve1
    - zfs-kmod-2.1.5-pve1

1. 確認 vm disk 剩餘空間
    ```bash
    $ df -h
    Filesystem      Size  Used Avail Use% Mounted on
    /dev/vda1        197G   14G   178G  7% /
    ```
2. vm 關機
3. 下載 [gparded iso](https://sourceforge.net/projects/gparted/) 到 pve
4. vm 從 gparded iso 開機並[縮減 disk 大小](https://www.howtogeek.com/114503/how-to-resize-your-ubuntu-partitions/)
5. pve server 針對 vm 的 hard disk 調整占用大小
    switch(allocate storage type)
    - case LV:
        `lvreduce -L 5G /dev/pve/disk-name (縮小到只剩 5G)`
        or
        `lvreduce -L -5G /dev/pve/disk-name (縮小 5G)`
    - case qcow2:
        `qemu-img resize --shrink <vmfile.qcow2> [+-] or size`
    - case ZFS:
        `zfs set volsize=<new size>G rpool/vm-<vm id>-disk-<disk number>`
    舉例:
        vm(`205`) 的虛擬硬碟(`vm-205-disk-0`)放在 zpool `land` 上面，已經將 vm 縮小至 100GB
        ```
        ❯ zfs list
        NAME                     USED  AVAIL     REFER  MOUNTPOINT
        land/vm-205-disk-0       236G   354G     119.3G  -
        
        ❯ zfs set volsize=100G land/vm-205-disk-0

        ❯ zfs list
        NAME                     USED  AVAIL     REFER  MOUNTPOINT
        land/vm-205-disk-0       136G   354G     19.3G  -
        ```
6. 在網頁 GUI 中修改強制刷新 disk。修改 disk 的設定，再調整回來
    ![](https://md.stranity.org/uploads/upload_d7fbc8d92382438b1b485a23f3e008e9.png)
7. 開機
