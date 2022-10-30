---
title: zsh & oh-my-zsh install cheat sheet on ubuntu 22.04
date: 2022-10-30
categories:
 - linux
---

0. install deps
    ```shell
    yes | sudo apt install git curl
    ```
1. install zsh
    ```shell
    yes | sudo apt install zsh
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
5. change login shell(must logout and login again)
    ```shell
    chsh -s $(which zsh)
    logout

    ...

    login
    ```
6. install plugins(recommended)
    - [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions/blob/master/INSTALL.md#oh-my-zsh)
        ```shell
        git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
        ```
    - [zsh-docker-aliases](https://github.com/akarzim/zsh-docker-aliases#with-oh-my-zsh)
        ```shell
        git clone https://github.com/akarzim/zsh-docker-aliases.git  ~/.oh-my-zsh/custom/plugins/zsh-docker-aliases
        ```
7. configure zsh  
    - [option 1] first use
        ```shell
        p10k configure
        ```
    - [option 2] have configured .zshrc  
        [my rc file](https://github.com/omegaatt36/lab/blob/main/rc/.zshrc)
        ```shell
        curl https://raw.githubusercontent.com/omegaatt36/lab/main/rc/.zshrc -o .zshrc
        ```

        must modify .zshrc ```export ZSH="/home/raiven/.oh-my-zsh"``` to your home path

        then source rc file
        ```shell
        source .zshrc
        ```
