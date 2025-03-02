# RAYPM - package manager for projects on Raylib
Это простой пакетный менеджер, позволяющий установить необходимые зависимости для создания игр на движке Raylib и портирования этих игр на другие платформы.
Simple package manager, allows you to install needed dependencies, like raylib.dll for Windows

## Installation
```console
$ go install github.com/mxk-9/raypm@latest
```

## ToDo:
- [ ] **Use raypm as a build system** [can\_i\_use\_raypm\_as\_a\_build\_system.txt](third_party/can_i_use_raypm_as_a_build_system.txt)
- [doc.txt](third_party/doc.txt)
- [X] All installed packages will store in $PROJECT_ROOT/.raypm/store:
```console
raylib-src
raylib-dll-mingw (dependends on raylib-src and mingw)
raylib-dll-mvsc
raylib-android (depends on raylib-src, mingw(if target==windows), android)
mingw
base
android
```

***

### [ ] raypm -help
Write custom `help` function
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

### [ ] raypm -reinstall [package]
Calls `uninstall` and then `install`

### [ ] raypm -uninstall [package]
+ [ ] Ensures that package is installed by search in .raypm/store/\<package\_name\>
+ [ ] Call ./raypm/store/\<package_name\>/uninstall.json

### [ ] raypm -downgrade [date]
If `date` is not define, it's just list available releases
When user choose a release, raypm calling -sync with specific link.

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
