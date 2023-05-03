---
title: zsh & oh-my-zsh & tmux install cheat sheet on ubuntu 22.04
date: 2022-10-30
categories:
 - develop
tags:
 - linux
 - cheat_sheet
---

## oh my zsh & power lever 10k

0. install deps
    ```shell
    sudo apt install git curl -y
    ```
1. install zsh
    ```shell
    sudo apt install zsh -y
    ```
2. check is installed
    ```shell
    cat /etc/shells | grep zsh

    # raiven@k3s:~$ cat /etc/shells | grep zsh
    # /bin/zsh
    # /usr/bin/zsh
    ```
3. install [oh-my-zsh](https://github.com/ohmyzsh/ohmyzsh#basic-installation)
    ```shell
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
    ```
4. install [powerlevel10k](https://github.com/romkatv/powerlevel10k#oh-my-zsh)
    ```shell
    git clone --depth=1 https://github.com/romkatv/powerlevel10k.git ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes/powerlevel10k
    ```
5. install fonts:
    - wsl:
        ```sh
        bash -c "$(curl -fsSL https://raw.githubusercontent.com/omegaatt36/lab/main/wsl/install_fonts.sh)" --windows
        ```
    - linux:
        ```sh
        bash -c "$(curl -fsSL https://raw.githubusercontent.com/omegaatt36/lab/main/wsl/install_fonts.sh)"
        ```
6. change login shell(must logout and login again)
    ```shell
    chsh -s $(which zsh)
    logout

    ...

    login
    ```
7. install plugins(recommended)
    - [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions/blob/master/INSTALL.md#oh-my-zsh)
        ```shell
        git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
        ```
    - [zsh-docker-aliases](https://github.com/akarzim/zsh-docker-aliases#with-oh-my-zsh)
        ```shell
        git clone https://github.com/akarzim/zsh-docker-aliases.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-docker-aliases
        ```
8. configure zsh  
    - [option 1] first use
        ```shell
        p10k configure
        ```
    - [option 2] have configured .zshrc  
        [my rc file](https://github.com/omegaatt36/lab/blob/main/rc/.zshrc)
        ```shell
        curl https://raw.githubusercontent.com/omegaatt36/lab/main/rc/.zshrc -o $HOME/.zshrc
        curl https://raw.githubusercontent.com/omegaatt36/lab/main/rc/.p10k.zsh -o $HOME/.p10k.zsh
        ```

        must modify .zshrc ```export ZSH="/home/raiven/.oh-my-zsh"``` to your home path

        then source rc file
        ```shell
        source $HOME/.zshrc
        ```

## tmux
1. install tmux
    ```shell
    sudo apt install tmux -y
    ```
2. install [tpm](https://github.com/tmux-plugins/tpm) (Tmux Plugin Manager)
    ```shell
    git clone https://github.com/tmux-plugins/tpm $HOME/.tmux/plugins/tpm
    ```
3. configure tmux
    ```shell
    curl https://raw.githubusercontent.com/omegaatt36/lab/main/rc/.tmux.conf -o $HOME/.tmux.conf
    ```
4. enjoy tmux
    ```shell
    tmux
    ```

#### ref
- [install_meslo_wsl](https://gist.githubusercontent.com/romkatv/aa7a70fe656d8b655e3c324eb10f6a8b/raw/install_meslo_wsl.sh)
- [install nerd fonts](https://gist.github.com/matthewjberger/7dd7e079f282f8138a9dc3b045ebefa0)