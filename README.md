# FastLauncher

TUI Application Launcher. Alternative to rofi/wofi

![main windows](https://github.com/probeldev/fastlauncher/blob/main/guides/screenshots/main.png?raw=true)

## Support OS

Linux - Done

Windows - Work in progress

Mac Os - Done

## Examples

### Logout Manager

[Example config](https://github.com/probeldev/fastlauncher/blob/main/guides/examples/logout-manager/cfg.json) 

![Logout manager](https://github.com/probeldev/fastlauncher/blob/main/guides/screenshots/logout-manager.png?raw=true)

### Emoji Select 

[Example config](https://github.com/probeldev/fastlauncher/blob/main/guides/examples/emoji/emoji.json) 

![Emoji select](https://github.com/probeldev/fastlauncher/blob/main/guides/screenshots/emoji-select.png?raw=true)


## Installation

[Full guide for Arch Linux with KDE](https://github.com/probeldev/fastlauncher/tree/main/guides/arch_kde/readme.md)

### Go
Installation

```bash
go install github.com/probeldev/fastlauncher@latest     
```


If you get an error claiming that fastlauncher cannot be found or is not defined, you
may need to add `~/go/bin` to your $PATH (MacOS/Linux), or `%HOME%\go\bin`
(Windows)

Zsh

```bash
echo "export PATH=\$PATH:~/go/bin" >> ~/.zshrc
```

Bash

```bash
echo "export PATH=\$PATH:~/go/bin" >> ~/.bashrc
```

### Nix

```bash
nix profile install github:probeldev/fastlauncher 
```


## Usage 

### All apps from OS

```bash
fastlauncher
```

### Apps from config

```bash
fastlauncher --config ~/script/fast-launcher/cfg.json
```

Example file [cfg.json](https://github.com/probeldev/fastlauncher/blob/main/cfg.json) 

It's launched with the help of window manager. Example hyprland.conf:

```
$terminal = foot
$menu = $terminal -T fast-launcher fastlauncher --config ~/script/fast-launcher/cfg.json
bind = $mainMod, D, exec, $menu


windowrulev2 = float,title:(fast-launcher)
windowrulev2 = pin,title:(fast-launcher)
windowrulev2 = size 1000 600,title:(fast-launcher)
windowrulev2 = center(1), title:(fast-launcher)
```

### Nested menus

A config item can open a submenu inside the same launcher window instead of
starting a new process. Two ways are supported and can be combined:

- `items` — inline nested list of apps;
- `config` — path to another json config file.

```json
[
	{
		"title": "Dev",
		"description": "inline submenu",
		"items": [
			{ "title": "Project", "description": "my project", "command": "code ." }
		]
	},
	{
		"title": "More Dev",
		"description": "submenu loaded from a file",
		"config": "dev.json"
	}
]
```

Rules:

- Enter on an item with nested `items` opens the submenu, Esc goes back.
  Enter on a leaf item runs its `command` and closes the launcher.
- `command` is ignored on items that have nested `items`.
- Relative `config` paths are resolved against the directory of the config
  file that declares them (not the current working directory); absolute paths
  are used as is and `~` is expanded to the home directory.
- A broken `config` reference (missing file, invalid json, a cycle, more than
  16 nested `config` files in one branch or more than 1000 expanded config
  files) is skipped with a warning, the rest of the config still loads. Inline
  `items` are not limited by the file depth: they only count against normal
  recursion.
- An item that ends up with no children and no non-empty `command` is not
  shown. A broken reference keeps the item when it has inline `items` (as a
  group with those items) or a non-empty `command` (as a leaf).

Full example: [nested/cfg.json](https://github.com/probeldev/fastlauncher/blob/main/guides/examples/nested/cfg.json)
with its [dev.json](https://github.com/probeldev/fastlauncher/blob/main/guides/examples/nested/dev.json).
