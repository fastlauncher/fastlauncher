# FastLauncher

TUI Application Launcher. Alternative to rofi/wofi

![alt text](https://github.com/probeldev/fastlauncher/blob/main/guides/screenshots/main.png?raw=true)

## Suport OS

Linux - Done

Windows - Work in progress

Mac Os - Work in progress

## Examples

### Logout Manager

[Example config](https://github.com/probeldev/fastlauncher/blob/main/examples/logout-manager/cfg.json) 

![alt text](https://github.com/probeldev/fastlauncher/blob/main/guides/screenshots/logout-manager.png?raw=true)

## Installation

[Full guide for Arch Linux with KDE](https://github.com/probeldev/fastlauncher/tree/main/guides/arch_kde/readme.md)

### Go
Installation

    go install github.com/probeldev/fastlauncher@latest     


If you get an error claiming that lazygit cannot be found or is not defined, you
may need to add `~/go/bin` to your $PATH (MacOS/Linux), or `%HOME%\go\bin`
(Windows)

Zsh

    echo "export PATH=\$PATH:~/go/bin" >> ~/.zshrc

Bash

    echo "export PATH=\$PATH:~/go/bin" >> ~/.bashrc

### Nix

    nix profile install github:probeldev/fastlauncher 


## Usage 

### All apps from OS

    fastlauncher

### Apps from config

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
