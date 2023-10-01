---
title: KDE neon / Ubuntu 啟用 PipeWire 與 LDAC/AAC/AptX 藍芽編碼
date: 2023-10-01
categories:
 - develop
tags:
 - linux
---

[Ubuntu 22.10 將 Audio Server 從 PulseAudio 切換成 PipeWire](https://www.ghacks.net/2022/05/23/ubuntu-22-10-dropping-pulseaudio/)

## 遇到了什麼問題

明明使用著有 LDAC 或是 AptX 解碼能力的藍芽耳機，卻只能被迫接受 sbc 的低傳輸率音質嗎，身為規格黨怎麼可以忍受這件事（即便大部份時間都只使用 Youtube Music 的 128K bps opus）

## 解決方法

我們可以在 KDE neon 上啟用 PipeWire，並安裝更多藍芽轉碼器

- 安裝 WirePlumber（PipeWire Manager）
    ```bash
    sudo apt install -y pipewire-media-session- wireplumber
    ```
- 透過 systemd 管理 WirePlumber 的開機自啟 daemon
    ```bash
    systemctl --user --now enable wireplumber.service
    ```
- 安裝 ALSA
    ```bash
    sudo apt install -y pipewire-audio-client-libraries
    ```
- 安裝藍芽轉碼器
    ```bash
    sudo apt install -y \
        libfdk-aac2 \
        libldacbt-{abr,enc}2 \
        libopenaptx0
    sudo apt install -y \
        libspa-0.2-bluetooth \
        pipewire-audio-client-libraries \
        pipewire-pulse
    ```
- 解除安裝 PulseAudio
    ```bash
    sudo apt remove -y pulseaudio-module-bluetooth
    ```
- 登出或重新啟動
- 檢查成果
    ```bash
    ❯ LANG=C pactl info | grep '^Server Name'
    Server Name: PulseAudio (on PipeWire 0.3.48)
    ```
## 成效

讓耳機連線至電腦後，選擇設定檔會出現更多轉碼器，諸如 SBC 與 LDAC，但若訊源只是普通串流，那其實聽起來不太可能會有差異。

## 如何更好

LHDC 與 LC3 與 aptx adaptive