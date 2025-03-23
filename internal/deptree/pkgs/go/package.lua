local name = "go"
local version = "1"
local description = "The Go Programming Language"

local targets = {
  linux = {
    install_phase = "${pkgman}",
    uninstall_phase = "${pkgman}",
  },
}

local packages = {
  arch = { "go" },
  fedora = { "golang" },
  ubuntu = { "golang-go" },
  void = { "go" },
}

packages.mint = packages.ubuntu
packages.manjaro = packages.arch

Data = {
  name = name,
  version = version,
  description = description,
  targets = targets,
  packages = packages,
}
