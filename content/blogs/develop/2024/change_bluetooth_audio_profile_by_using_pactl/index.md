---
title: 透過 CLI 調整藍牙耳機音訊設定
date: 2024-01-28
categories:
 - develop
tags:
 - linux
aliases:
 - "/blogs/develop/2024/change_bluetooth_audio_profile_by_using_pactl.html"
---

接續著 [KDE neon / Ubuntu 啟用 PipeWire 與 LDAC/AAC/AptX 藍芽編碼](/blogs/develop/2023/kde_neon_pipewire_and_more_bluetooth_kde_neon_pipe_wire_and_more_bluetooth_codec)，每當連線到藍芽耳機時，語音設定檔在使用 LDAC 後，就無法同時使用麥克風，這樣一來在開會時，就需要手動到設定裡面去調整語音設定檔。

身為一個懶惰鬼，可以用鍵盤解決的就不會用滑鼠去點，發現可以通過 pactl（PulseAudio）在 cli 直接設定 audio profile。

首先，我們通過 `bluetoothctl` 列出藍芽設備的實體位置

```bash
$ bluetoothctl
> [EAH-A800]# devices
> Device 88:C9:E8:B1:5D:AE WF-1000XM4
> Device DC:22:D2:85:85:15 MX Master 3S
> Device B8:20:8E:35:CB:D0 EAH-A800
```

文章內舉例的目標為 EAH-A800，也就是 `B8:20:8E:35:CB:D0`
接著需要確認目前藍牙裝置支持的音訊設定檔。可以透過 `pactl list cards short` 來列出目前啟用的設備

```bash
$ pactl list cards short
47      alsa_card.usb-Shure_Inc_Shure_MV7-00    alsa
48      alsa_card.pci-0000_00_1f.3-platform-skl_hda_dsp_generic alsa
49      alsa_card.usb-NuPrime_NuPrime_DAC-9H-00 alsa
1002    bluez_card.B8_20_8E_35_CB_D0    module-bluez5-device.c
```

發現目標為 `1002    bluez_card.B8_20_8E_35_CB_D0`
再來通過 `pactl list cards` 來獲取所有音訊設定檔

```bash
Card #1002
        Name: bluez_card.B8_20_8E_35_CB_D0
        Driver: module-bluez5-device.c
        Owner Module: n/a
        Properties:
                api.bluez5.address = "B8:20:8E:35:CB:D0"
                api.bluez5.class = "0x240404"
                api.bluez5.connection = "disconnected"
                api.bluez5.device = ""
                api.bluez5.icon = "audio-headset"
                api.bluez5.path = "/org/bluez/hci0/dev_B8_20_8E_35_CB_D0"
                bluez5.auto-connect = "[ hfp_hf hsp_hs a2dp_sink ]"
                bluez5.profile = "off"
                device.alias = "EAH-A800"
                device.api = "bluez5"
                device.bus = "bluetooth"
                device.description = "EAH-A800"
                device.form_factor = "headset"
                device.name = "bluez_card.B8_20_8E_35_CB_D0"
                device.product.id = "0x0004"
                device.string = "B8:20:8E:35:CB:D0"
                device.vendor.id = "bluetooth:0094"
                media.class = "Audio/Device"
                factory.id = "14"
                client.id = "34"
                object.id = "86"
                object.serial = "1002"
        Profiles:
                off: Off (sinks: 0, sources: 0, priority: 0, available: yes)
                a2dp-sink: High Fidelity Playback (A2DP Sink) (sinks: 1, sources: 0, priority: 16, available: yes)
                headset-head-unit: Headset Head Unit (HSP/HFP) (sinks: 1, sources: 1, priority: 1, available: yes)
                a2dp-sink-sbc: High Fidelity Playback (A2DP Sink, codec SBC) (sinks: 1, sources: 0, priority: 18, available: yes)
                a2dp-sink-sbc_xq: High Fidelity Playback (A2DP Sink, codec SBC-XQ) (sinks: 1, sources: 0, priority: 17, available: yes)
                a2dp-sink-ldac: High Fidelity Playback (A2DP Sink, codec LDAC) (sinks: 1, sources: 0, priority: 19, available: yes)
                headset-head-unit-cvsd: Headset Head Unit (HSP/HFP, codec CVSD) (sinks: 1, sources: 1, priority: 2, available: yes)
                headset-head-unit-msbc: Headset Head Unit (HSP/HFP, codec mSBC) (sinks: 1, sources: 1, priority: 3, available: yes)
        Active Profile: a2dp-sink-ldac
        Ports:
                headset-input: Headset (type: Headset, priority: 0, latency offset: 0 usec, available)
                        Properties:
                                port.type = "headset"
                        Part of profile(s): headset-head-unit, headset-head-unit-cvsd, headset-head-unit-msbc
                headset-output: Headset (type: Headset, priority: 0, latency offset: 0 usec, available)
                        Properties:
                                port.type = "headset"
                        Part of profile(s): a2dp-sink, headset-head-unit, a2dp-sink-sbc, a2dp-sink-sbc_xq, a2dp-sink-ldac, headset-head-unit-cvsd, headset-head-unit-msbc
```

可以在最下面看到音訊設定檔

```bash
Part of profile(s): a2dp-sink, headset-head-unit, a2dp-sink-sbc, a2dp-sink-sbc_xq, a2dp-sink-ldac, headset-head-unit-cvsd, headset-head-unit-msbc
```

之後就可以使用以下指令，指定設定檔了。

```bash
pactl set-card-profile bluez_card.B8_20_8E_35_CB_D0 headset-head-unit-msbc
pactl set-card-profile bluez_card.B8_20_8E_35_CB_D0 a2dp-sink-ldac
```

透過 pactl 命令，我們可以輕鬆地在 Linux 系統中切換不同的藍牙耳機音訊設定，無需透過圖形化介面，如果跟我一樣懶 XD。
