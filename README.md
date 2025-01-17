# FastLauncher

TUI Application Launcher

![alt text](https://github.com/probeldev/fastlauncher/blob/main/screenshots/main.png?raw=true)

## Installation

    go install github.com/probeldev/fastlauncher@latest     


## Usage 

    fastlauncher --config ~/script/fast-launcher/cfg.json

Example file [cfg.json](https://github.com/probeldev/fastlauncher/blob/main/cfg.json) 

It's launched with the help of window manager. Example hyprland.conf:
    
    $terminal = foot
    $menu = $terminal -T fast-launcher fastlauncher --config ~/script/fast-launcher/cfg.json
    bind = $mainMod, D, exec, $menu

## Hotkeys

h,j,k,l - Navigation

? - Search
