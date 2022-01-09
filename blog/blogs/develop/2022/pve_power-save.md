---
title: 改善 Proxmox VE/Debian 始終跑在最高頻率
date: 2022-01-09
categories:
 - develop
---

這幾天把便宜撿到的 Threadripper 2950X 平台也上 Proxmox VE 玩玩了，裝完系統後才發現自己太習慣於 windows 下的電源管理，一直都沒發現 linux 下 CPU 頻率都是拉滿的狀態，找了[debian 下進行電源管理的電源計畫設定教學](https://forum.proxmox.com/threads/fix-always-high-cpu-frequency-in-proxmox-host.84270/)達到降溫省電，順便做做紀錄。

此篇文章的硬體基於
```
root@raiven:~# neofetch
       _,met$$$$$gg.          root@raiven 
    ,g$$$$$$$$$$$$$$$P.       ----------- 
  ,g$$P"     """Y$$.".        OS: Debian GNU/Linux 10 (buster) x86_64 
 ,$$P'              `$$$.     Host: HP Z2 SFF G4 Workstation 
',$$P       ,ggs.     `$$b:   Kernel: 5.4.106-1-pve 
`d$$'     ,$P"'   .    $$$    Uptime: 276 days, 9 hours, 51 mins 
 $$P      d$'     ,    $$P    Packages: 719 (dpkg) 
 $$:      $$.   -    ,d$$'    Shell: bash 5.0.3 
 $$;      Y$b._   _,d$P'      Terminal: /dev/pts/1 
 Y$$.    `.`"Y$$$$P"'         CPU: Intel Xeon E-2278G (16) @ 5.000GHz 
 `$$b      "-.__              GPU: Intel Device 3e9a 
  `Y$$                        Memory: 48763MiB / 64099MiB 
   `Y$$.
     `$$b.                                            
       `Y$$b.
          `"Y$b._
              `"""

root@raiven:~# ^C
```

用 `watch -n 1 "cat /proc/cpuinfo | grep MHz"` 可以查看當下的 cpu 頻率狀態。
```
Every 1.0s: cat /proc/cpuinfo | grep MHz                                                                                                             raiven: Sun Jan  9 22:03:08 2022

cpu MHz         : 4757.399
cpu MHz         : 4686.194
cpu MHz         : 4764.715
cpu MHz         : 4755.255
cpu MHz         : 4659.314
cpu MHz         : 4796.930
cpu MHz         : 4735.900
cpu MHz         : 4743.142
cpu MHz         : 4762.312
cpu MHz         : 4760.803
cpu MHz         : 4798.249
cpu MHz         : 4727.302
cpu MHz         : 4755.562
cpu MHz         : 4653.955
cpu MHz         : 4776.860
cpu MHz         : 4734.148
```

會發現 CPU 一直都在頻率很高的狀態，可能導致無謂的能源浪費。

於是我們可以先安裝 [ACPI](https://zh.wikipedia.org/zh-tw/%E9%AB%98%E7%BA%A7%E9%85%8D%E7%BD%AE%E4%B8%8E%E7%94%B5%E6%BA%90%E6%8E%A5%E5%8F%A3)。
```
apt install acpi-support acpid acpi
```

接著可以查看有哪些選項可以使用
```
> cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_available_governors

# on Xeon E-2278G
performance powersave

# on Ryzen pro 5750G
conservative ondemand userspace powersave performance schedutil
```

在 E-2278 上只看到效能(performance)與節能(powersave)，嘗試將電源管理改為節能:
```
echo "powersave" | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
```

接著再次查看 CPU 頻率，就會發現已經成功讓電源管理變為節能了，而有需要時仍會跑到最大頻率 4.5GHz。
```
Every 1.0s: cat /proc/cpuinfo | grep MHz                                                                                                             raiven: Sun Jan  9 22:11:21 2022

cpu MHz         : 899.943
cpu MHz         : 900.073
cpu MHz         : 900.099
cpu MHz         : 900.045
cpu MHz         : 900.035
cpu MHz         : 900.037
cpu MHz         : 900.007
cpu MHz         : 900.038
cpu MHz         : 899.981
cpu MHz         : 900.023
cpu MHz         : 900.034
cpu MHz         : 900.019
cpu MHz         : 900.026
cpu MHz         : 899.897
cpu MHz         : 900.001
cpu MHz         : 900.034
```

---

在[留言](https://forum.proxmox.com/threads/fix-always-high-cpu-frequency-in-proxmox-host.84270/#post-373393)中也有看到 `cpufrequtils` 看起來不妨是更好的選項，也可以嘗試看看。