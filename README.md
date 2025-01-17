# FastLauncher

![alt text](https://github.com/probeldev/fastlauncher/blob/main/screenshots/main.png?raw=true)

## Установка

    go install github.com/probeldev/fastlauncher@latest     


## Использование 

    fastlauncher --config ~/script/fast-launcher/cfg.json

Пример файла [cfg.json](https://github.com/probeldev/fastlauncher/blob/main/cfg.json) 

Запуск производится средствами оконного менеджера. Пример hyprland.cofg:
    
    $terminal = foot
    $menu = $terminal -T fast-launcher fastlauncher --config ~/script/fast-launcher/cfg.json
    bind = $mainMod, D, exec, $menu

## Горячие клавиши

h,j,k,l - навигация

? - Поиск
