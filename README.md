# FastLauncher

TUI Application Launcher

![alt text](https://github.com/probeldev/fastlauncher/blob/main/screenshots/main.png?raw=true)

## Suport OS

Linux - Done

Windows - Work in progress

Mac Os - Work in progress

## Installation


Installation

    go install github.com/probeldev/fastlauncher@latest     


If you get an error claiming that lazygit cannot be found or is not defined, you
may need to add `~/go/bin` to your $PATH (MacOS/Linux), or `%HOME%\go\bin`
(Windows)

Zsh

    echo "export PATH=\$PATH:~/go/bin" >> ~/.zshrc

Bash

    echo "export PATH=\$PATH:~/go/bin" >> ~/.bashrc


## Usage 

    fastlauncher --config ~/script/fast-launcher/cfg.json

Example file [cfg.json](https://github.com/probeldev/fastlauncher/blob/main/cfg.json) 

It's launched with the help of window manager. Example hyprland.conf:
    
    $terminal = foot
    $menu = $terminal -T fast-launcher fastlauncher --config ~/script/fast-launcher/cfg.json
    bind = $mainMod, D, exec, $menu


    windowrulev2 = float,title:(fast-launcher)
    windowrulev2 = pin,title:(fast-launcher)
    windowrulev2 = size 1000 600,title:(fast-launcher)
    windowrulev2 = center(1), title:(fast-launcher)


## Hotkeys

h,j,k,l - Navigation

? - Search
