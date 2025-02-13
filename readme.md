# RAYPM - package manager for my projects with Go & Raylib
Это простой пакетный менеджер, позволяющий установить необходимые зависимости для создания игр на движке Raylib и портирования этих игр на другие платформы.


## Installation
```console
$ go install github.com/mxk-9/raypm@latest
```

## ToDo:
+ Docs:
  + [doc.txt](third_party/doc.txt)
  + [package\_cache.txt](third_party/doc/package_cache.txt)
- All installed packages will store in $PROJECT_ROOT/.raypm/store:
```console
raylib-src
raylib-dll-mingw (dependends on raylib-src and mingw)
raylib-dll-mvsc
raylib-android (depends on raylib-src, mingw(if target==windows), android)
mingw
base
android
```

- [ ] `package.json`'s will store in $PROJECT_ROOT/.raypm/pkgs
- [ ] I will have separate repo with pkgs, each release will contain creation date
- [ ] all downloaded content(include raypm-pkgs) in $PROJECT_ROOT/.raypm/cache
- `raylib-dll-mingw` will depends on `raylib-src` and `mingw`. It also will stores custom `Makefile`.
- [ ] For `build_phase` `raypm` will create .raypm/cache/build\_\<pkgname\>\_\<hash\>, copy all needed resources and execute all nessesary command to build package.
Then, copy final result to `$out`, full path to .raypm/store/\<package_name\>\_version.
- [ ] While running, $PROJECT\_ROOT/.raypm will RW, and in the end — RO


## Working with GitHub
### [ ] raypm -help
Write custom `help` function
### [ ] raypm -sync
Get a fresh package database. It's will download db and unpack to ./raypm/pkgs
+ [ ] We need get access to raypm-pkgs github page and download the latest archive.
+ [ ] Download and unpack pkgs into ./raypm/pkgs

### [ ] raypm -clean [option]
Available options:
1. `cache` — deleting .raypm/cache/*
2. `all` — deleting entire .raypm directory

### [ ] raypm -install [package]
Installes a package
+ [ ] Searching package in .raypm/pkgs
+ [ ] Copying uninstall phase instructions in ./raypm/store/\<package\_name\>/uninstall.json

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

### [ ] raypm -search [keyword]
Search packages by one keyword(in future remade this functionality and allow to search by separate keywords, that using package's description)

### [ ] raypm -info [package]
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
