# RAYPM - package manager for projects on Raylib
Simple package manager, allows you to install needed dependencies, like raylib.dll for Windows
Простой пакетный менеджер, позволяющий установить необходимые зависимости для создания игр на движке Raylib и портирования этих игр на другие платформы.

## Installation
```console
$ go install github.com/mxk-9/raypm@latest
```

## ToDo:
- [ ] **Use raypm as a build system** [can\_i\_use\_raypm\_as\_a\_build\_system.txt](third_party/can_i_use_raypm_as_a_build_system.txt)
- [ ] **Crossplatform** — ability to build package for another system (host ≠ target)
  [cross.txt](third\_party/cross.txt)
- [doc.txt](third\_party/doc.txt)
- [ ] Use Lua to describe package instead of json-hell
- [ ] Use MySQL to mantain package dependencies
- [ ] Temporary use json files as database
***

### [ ] raypm -help
Write custom `help` function
### [ ] raypm -init <package\_name>
Creates <package\_name> in current directory and adds to `lists` in `$HOME/.raypm/`
### [X] raypm -sync
Get a fresh package database. It's will download db and unpack to ./raypm/pkgs
+ [X] We need get access to raypm-pkgs github page and download the latest archive.
+ [X] Download and unpack pkgs into ./raypm/pkgs

### [X] raypm -clean [option]
Available options:
1. `cache` — deleting .raypm/cache/*
2. `all` — deleting entire .raypm directory

### [X] raypm -install [package]
Installes a package
+ [X] Searching package in .raypm/pkgs
+ [X] Copying uninstall phase instructions in ./raypm/store/\<package\_name\>/uninstall.json

### [ ] raypm -uninstall [package]
+ [ ] Ensures that package is installed by search in .raypm/store/\<package\_name\>
+ [ ] Call ./raypm/store/\<package_name\>/uninstall.json

### [ ] raypm -upgrade [package]
If version of installed package mismatch with it's package.json, reinstalls

### [ ] raypm -upgrade-all
Similar for `-upgrade`, but for all installed packages

### [X] raypm -search [keyword]
Search packages by one keyword(in future remade this functionality and allow to search by separate keywords, that using package's description)

### [X] raypm -info [package]
Parse information about chosen package. Output must be something like that:
```ini
Name: pkg_name [Installed]
Description: pkg for something
Version: 0.1
Depends on: (target: windows)
    + mingw
    + raylib-src

Supported systems:
    + x86_64-linux
    + x86_64-windows

Package stores: .raypm/pkgs/pkgs_name
Files:
    .raypm/store/pkg_name/file1
    .raypm/store/pkg_name/file2.txt
    .raypm/store/pkg_name/dir/file3.conf
```

### [ ] raypm -run [package]
If a package contains executable, launch it

> package.json
```json
{
    "executable" : "bin/rltest"
}
```

It will automaticly expand `bin/rltest` to `$out/bin/rltest`
