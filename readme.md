# RAYPM - package manager for projects on Raylib
Simple package manager, allows you to install needed dependencies, like raylib.dll for Windows
Простой пакетный менеджер, позволяющий установить необходимые зависимости для создания игр на движке Raylib и портирования этих игр на другие платформы.

## Installation

Using `go`
```
$ go install github.com/mxk-9/raypm@latest
```
***

## TODO:
- [X] **Crossplatform** — ability to build package for another system (host ≠ target)
- [ ] **Use raypm as a build system** [can\_i\_use\_raypm\_as\_a\_build\_system.txt](third_party/can_i_use_raypm_as_a_build_system.txt)
  [cross.txt](third\_party/cross.txt)
- [doc.txt](third\_party/doc.txt)
- [X] Use Lua to describe package instead of json-hell
- [ ] Use MySQL to mantain package dependencies
- [X] Temporary use json files as database
- [ ] Weird bug, my pm trying to download/unpack one thing twice and install the package twice
***

### [ ] raypm -help
Write custom `help` function
### [ ] raypm -init <package\_name>
Creates <package\_name> in current directory and adds to `lists` in `$HOME/.raypm/`
### [ ] raypm -build
### [ ] `-o <path>` key
This key will override $out
