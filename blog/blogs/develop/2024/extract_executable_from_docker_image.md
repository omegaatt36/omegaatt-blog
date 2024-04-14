---
title: 如何不啟動 container 從 image 中提取可執行檔
date: 2024-04-14
categories:
 - develop
tags:
 - docker
---

::: info
雖然文章內大多指令都是使用 docker，但由於是標準的 OCI image，使用 Podman 也是一樣的效果。
:::

某段時間內，我會將所有需要執行的 binary 使用 containerize 給打包起來執行，例如：

1. 需要將前端環境給跑起來，我會啟動一個 node 環境的 container:

    ```shell
    docker run --rm -it --name f2e --net=host -v $(pwd):/app \
        docker.io/library/node:20-bookworm \
        bash
    ```

1. 需要安裝某個基於 golang 的 cli 工具，會使用[自己寫的腳本](https://gist.github.com/omegaatt36/487d643aca6443f9524eed2975cbd746)，在 container 內進行建構。
1. 需要透過 Liquibase 來進行 db migration，並不會選擇在本地安裝 maven 環境，一樣是啟動一個 container 來執行。

這麼做的好處是，「目前」大多數的 server/cli tools 至少會編譯 x86 架構下的可執行檔，我只需要確保 container 環境內可以工作，就可以在不同硬體的開發環境下游走。極度偏激的來說，無法確保安裝的 binary 有沒有受過污染，無論開源或閉源與否，也可以透過 Podman 來啟動 rootless 的 container 確保本機的安全。

但有時候會遇到別人打包的 image 十分肥大，即便程式可執行檔只需要幾 MB，編譯後的 image 確有幾百 MB（大多數是腳本語言），蒙生了從 image 中提取 executable 的想法，藉此也學習 docker layer 間的關係。

## 事前準備

會使用 film36exp 這個 side project 作為文章內的範例，可以參考 [Dockerfile](https://github.com/omegaatt36/film36exp/blob/0610348fe69a78dbdefc6374e1aab2163b4e3e85/Dockerfile)。由於是使用 `gcr.io/distroless/static-debian12` 作為 prod 的 base image，層數會比使用 `debian:bookworm` 來的多很多，不坊在閱讀完部落格後，嘗試替換 build stage 的 base image 來驗證。

1. 先 `git clone` 後到當前的 commit

    ```shell
    git clone https://github.com/omegaatt36/film36exp.git
    cd film36exp
    git reset --hard 0610348
    ```

1. 打包一個 docker image

    ```shell
    docker build --build-arg CMD=api -t film36exp-api:latest .
    ```

## 如何解析 image

1. 首先透過 `docker save` 保存 image 成檔案

    ```shell
    docker save -o film36exp-api.tar film36exp-api
    ```

1. 解壓縮

    ```shell
    mkdir film36exp-api
    tar -C film36exp-api -xf film36exp-api.tar
    cd film36exp-api
    ```

1. 查看目錄結構

    ```shell
    ❯ ls --tree
     .
    ├──  blobs
    │  └──  sha256
    │     ├──  05c3ed42d4e43c35c17176230a9370c2333d51e55982577a93fdb620ea24ef72
    │     ├──  09f3168ca02760375469a5120ce700ecbde03852a0956e2fd50965c1c5123023
    │     ├──  1a73b54f556b477f0a8b939d13c504a3b4f4db71f7a09c63afbc10acb3de5849
    │     ├──  2a92d6ac9e4fcc274d5168b217ca4458a9fec6f094ead68d99c77073f08caac1
    │     ├──  3d6fa0469044370439d20eaf7e0d25450e01335a93c13ba46e368d7785914c0c
    │     ├──  4d049f83d9cf21d1f5cc0e11deaf36df02790d0e60c1a3829538fb4b61685368
    │     ├──  6cab0ce007d2d5ba6dcb59175947ced48139ea894b59b3f9a2079ee87456bc85
    │     ├──  9e1c613df631db2e76e3b37db5971420fa1a469a31cbed0a8b7c1bc3bf41b2a2
    │     ├──  11a7b0414ee466033cc36ffc991a542750af9c6d313b65222a0d2721429fca2d
    │     ├──  52e8589849b53f257b72aa6ecdd438e22c90e8c795db9177165758c0ef7bb12c
    │     ├──  68dc859147597f8d988de5b5a8d3e0041c08d3dd1fa6eefe2ecb5d23ac8eaa28
    │     ├──  73a34eb6fcb677bf81f1b8c36fe88b038105d5f30c0cef1b0a715eac7a07537a
    │     ├──  945d17be9a3e27af5ca1c671792bf1a8f2c3f4d13d3994665d95f084ed4f8a60
    │     ├──  953df972f540cc2389cdb187baae2adb879f5f5d6637cff6cb525f76219c80df
    │     ├──  2038c1e1ade120b9c9aaaa3632161e24f0d565ed87b75c97ea1712001eed4c04
    │     ├──  4581c0ea206de8590e0aee7ee54ccff7dd3c1f13409b2fe65d45fb5dddc20b5f
    │     ├──  49626df344c912cfe9f8d8fcd635d301bd41127cd326914212cf2443a96cf421
    │     ├──  a6c9a0b765bfd83b258a972bd5b8a1a48af15023883b52358a85ea1e7c632e57
    │     ├──  a548c2945819b785f42216b8138115022ef38de6759b19a30ada73c4aaa8fe62
    │     ├──  ac805962e47900b616b2f4b4584a34ac7b07d64ac1fd2c077478cf65311addcc
    │     ├──  af5aa97ebe6ce1604747ec1e21af7136ded391bcabe4acef882e718a87c86bcc
    │     ├──  b1bc7f7021c1125d3a288b89bd51d915ea6a83598b5aeac4c673e1d30e178e44
    │     ├──  b336e209998fa5cf0eec3dabf93a21194198a35f4f75612d8da03693f8c30217
    │     ├──  bbb6cacb8c82e4da4e8143e03351e939eab5e21ce0ef333c42e637af86c5217b
    │     ├──  f0cab5029bbb46537901976e0e9b05edcfbfcc33ec2552ba8909b2713ef3cd2a
    │     ├──  f3ddf7095a620f60a37d4500a0408b25ad61faa18adfdc65899ec95cf043e800
    │     ├──  f4aee9e53c42a22ed82451218c3ea03d1eea8d6ca8fbe8eb4e950304ba8a8bb3
    │     └──  fdf90b3af235abc789d4fd7c97286a4dc1732c25f18dcccf4117d68fe4b6b732
    ├──  index.json
    ├──  manifest.json
    ├──  oci-layout
    └──  repositories
    ```

    可以看到裡面的幾個資料夾/檔案：
    1. `blobs`: 儲存 image 的所有層（layer）數據。這些層是 image 的基本組成單元，每一層代表 image 在建構過程中的一次變更。
    1. `blobs/sha256`: blobs 下的一個目錄，存放具體的層數據。每個檔案都以其 SHA256 hash 值命名，這個 hash 值保證了檔案的唯一性和完整性。這些 hash 檔案包含了 image 的實際內容，從應用程式的二進制執行檔到操作系統的庫等等。
    1. `index.json`: 是 image 的索引文件，它包含了關於 image 版本、資料結構和指向 manifest 的指針。這個檔案是解析 image 時的起點，用於找到管理 image 的 manifest 文件。
    1. `manifest.json`: 存儲 image 的元數據和 config，如 image 中每一層的內容、大小和 SHA256 hash 等。它也包含了 image 建構的歷史和指示如何重新組合這些層以重建 image 的指令。
    1. `oci-layout`: 指定了 image 遵循的 Open Container Initiative（OCI）標準的版本。OCI 是一個幫助確保容器映像格式在不同容器技術間具有一致性和兼容性的開放標準。
    1. `repositories`: 記錄了 image 的倉庫資訊，包括 image 的標籤和對應的層。這個檔案對於理解 image 在倉庫中的組織結構非常有用。

    所有純文字檔案都是使用 JSON 來進行存儲，而二進制檔案則是 tar 後的壓縮檔，後續會透過 `jq` 來進行縮排後的輸出。
1. 查看 `index.json`

    ```shell
    ❯ jq < index.json
    {
    "schemaVersion": 2,
    "mediaType": "application/vnd.oci.image.index.v1+json",
    "manifests": [
        {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "digest": "sha256:a6c9a0b765bfd83b258a972bd5b8a1a48af15023883b52358a85ea1e7c632e57",
        "size": 2211,
        "annotations": {
            "io.containerd.image.name": "docker.io/library/film36exp-api:latest",
            "org.opencontainers.image.ref.name": "latest"
        }
        }
    ]
    }
    ```

    `index.json` 中包含了 manifests 的資訊，也就是 `sha256:a6c9a0b765bfd83b258a972bd5b8a1a48af15023883b52358a85ea1e7c632e57`，我們可以在 `blobs/sha256` 內找到他:

    ```shell
    ❯ jq < blobs/sha256/$(jq ".manifests[].digest" < index.json | tr -d '"'  | sed -e "s/^sha256://") | jq
    {
        "schemaVersion": 2,
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "config": {
            "mediaType": "application/vnd.oci.image.config.v1+json",
            "digest": "sha256:fdf90b3af235abc789d4fd7c97286a4dc1732c25f18dcccf4117d68fe4b6b732",
            "size": 2089
        },
        "layers": [
            {
            "mediaType": "application/vnd.oci.image.layer.v1.tar",
            "digest": "sha256:3d6fa0469044370439d20eaf7e0d25450e01335a93c13ba46e368d7785914c0c",
            "size": 327680
            },
            # 以下省略
        ]
    }
    ```

    繼續深挖 `config`

    ```shell
    ❯ jq < blobs/sha256/$(jq < blobs/sha256/$(jq ".manifests[].digest" < index.json | tr -d '"'  | sed -e "s/^sha256://") | jq ".config.digest"  | tr -d '"'  | sed -e "s/^sha256://")
    {
        "architecture": "amd64",
        "config": {
            "User": "0",
            "Env": [
            "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
            "SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt"
            ],
            "Cmd": [
            "./app"
            ],
            "WorkingDir": "/home/app/",
            "ArgsEscaped": true
        },
        "history": [
            # 省略
        ]
        "os": "linux",
        "rootfs": {
            "type": "layers",
            "diff_ids": [
                "sha256:3d6fa0469044370439d20eaf7e0d25450e01335a93c13ba46e368d7785914c0c",
                "sha256:49626df344c912cfe9f8d8fcd635d301bd41127cd326914212cf2443a96cf421",
                "sha256:945d17be9a3e27af5ca1c671792bf1a8f2c3f4d13d3994665d95f084ed4f8a60",
                "sha256:4d049f83d9cf21d1f5cc0e11deaf36df02790d0e60c1a3829538fb4b61685368",
                "sha256:af5aa97ebe6ce1604747ec1e21af7136ded391bcabe4acef882e718a87c86bcc",
                "sha256:ac805962e47900b616b2f4b4584a34ac7b07d64ac1fd2c077478cf65311addcc",
                "sha256:bbb6cacb8c82e4da4e8143e03351e939eab5e21ce0ef333c42e637af86c5217b",
                "sha256:2a92d6ac9e4fcc274d5168b217ca4458a9fec6f094ead68d99c77073f08caac1",
                "sha256:1a73b54f556b477f0a8b939d13c504a3b4f4db71f7a09c63afbc10acb3de5849",
                "sha256:f4aee9e53c42a22ed82451218c3ea03d1eea8d6ca8fbe8eb4e950304ba8a8bb3",
                "sha256:b336e209998fa5cf0eec3dabf93a21194198a35f4f75612d8da03693f8c30217",
                "sha256:4581c0ea206de8590e0aee7ee54ccff7dd3c1f13409b2fe65d45fb5dddc20b5f",
                "sha256:68dc859147597f8d988de5b5a8d3e0041c08d3dd1fa6eefe2ecb5d23ac8eaa28"
                ]
            }
        }
    }
    ```

    有沒有一個更方便直接查看這些內容的方式？那就是 `manifest.json`，已經將 config & index.json & manifest 給祖好了，我們可以在 `manifest.json` 內直接查看諸如 `Config` 與 `Layers` 等等，而這些也就是 `docker image inspect film36exp-api` 相妨。

## RootFS

到這裡我們知道 image 內的 `index.json` 內宣告了 manifest 檔案的路徑，以及裡面 `config` 與 `layers`，那麼這些 layer 到底是什麼？

查看 `config` 可以看到 `rootfs`，我們知道 RootFS 其實就是在 container 內的根目錄 `/`，雖然 container 與主機共享一個 kernel，但 container 也有自己完整的 RootFS，例如我們可以查看 `debian:bookworm` 的根目錄：

```shell
❯ docker run --rm debian:bookworm ls /
bin
boot
dev
etc
home
lib
lib64
media
mnt
opt
proc
root
run
sbin
srv
sys
tmp
usr
var
```

docker image 是透過 layer 的方式來逐步構建 RootFS，每一層都是前一層的增量更新。也就是說我們可以在透過 `layers.diff_ids` 這個陣列得知每一層的 rootfs 增量修改，透過簡易的腳本來解壓縮每一個 layer：

```bash
#!/bin/bash

image_config=$(jq '.[].Config' < manifest.json | tr -d '"')

lines=$(jq ".rootfs.diff_ids[]" < "${image_config}" | tr -d '"')

for line in ${lines}
do
    layer="${line#sha256:}"
    mkdir "${layer}"
    tar -xf "blobs/sha256/${layer}" -C "${layer}"
done
```

接著透過 config 中的 `Cmd` 顯示 `["./app"]`，我們可以知道我們要找的 executable 為 `app`：

```shell
❯ find . -type f -name "app"
./68dc859147597f8d988de5b5a8d3e0041c08d3dd1fa6eefe2ecb5d23ac8eaa28/home/app/app
```

不難看出我們已經找到可執行檔了，試著跑跑看

```shell
❯ ./68dc859147597f8d988de5b5a8d3e0041c08d3dd1fa6eefe2ecb5d23ac8eaa28/home/app/app
NAME:
   app - A new cli application

USAGE:
   app [global options] command [command options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-level value    (default: debug) [$LOG_LEVEL]
   --db-dialect value   [sqlite3|postgres] [$DB_DIALECT]
   --db-host value      postgres -> host, sqlite3 -> filepath [$DB_HOST]
   --db-port value      (default: 0) [$DB_PORT]
   --db-name value       [$DB_NAME]
   --db-user value       [$DB_USER]
   --db-password value   [$DB_PASSWORD]
   --db-silence-logger  (default: false) [$DB_SILENCE_LOGGER]
   --app-env value      (default: local) [$APP_ENV]
   --app-port value     (default: 8070) [$APP_PORT]
   --help, -h           show help
2024/04/14 10:54:03 Required flag "db-dialect" not set
```

## 還要更快

前面這個方法會將所有 layer 給解壓縮，但我們其實只需要「一個」 diff_id 就好，可以透過 `docker history` 來逐 layer 查看異動，或是透過其他更好用的 cli tool 來協助我們查看歷史與變更。

### [dive](https://github.com/wagoodman/dive)

```shell
docker run --rm -it \
    -v /var/run/docker.sock:/var/run/docker.sock \
    wagoodman/dive:latest film36exp-api
```

可以看到最底層的 Current Layer Contents 已經包含了完整的執行環境
    ![20240414_110044.png](/assets/dev/20240414/110044.png)

選最後一個 layer 會看到 diff 是 `/home/app/app`，也能在 Layer Details 內看到 hash 為 `sha256:68dc859147597f8d988de5b5a8d3e0041c08d3dd1fa6eefe2ecb5d23ac8eaa28`
    ![20240414_110053.png](/assets/dev/20240414/110053.png)

過去還有一些 [dockviz](https://github.com/justone/dockviz) 與 [sen](https://github.com/TomasTomecek/sen)，不過目前看起來最好用的 terminal ui 還是 dive。

## 寫在最後

透過拆解 image ，學習一個 OCI 的 image 內包了什麼，以及可以從更多方向對 image 進行 debug。

若是嫌棄別人打包的 image 太肥大，也可以透過這個方式提取可執行檔或是部份內容，自己打包成小一點的 image。
