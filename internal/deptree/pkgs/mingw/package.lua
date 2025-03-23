local name = "mingw"
local version = "1"
local description = "Minimalist GNU for Windows"

local link_to_devkit =
  "https://github.com/skeeto/w64devkit/releases/download/v2.0.0/w64devkit-x64-2.0.0.exe"
local devkit_file = "w64devkit-x64-2.0.0.exe"
local devkit_dir = "w64devkit"

local targets = {
  linux = {
    install_phase = "${pkgman}",
    uninstall_phase = "${pkgman}",
  },

  windows = {
    fetch_phase = string.format("${get %s %s}", link_to_devkit, devkit_file),
    unpack_phase = string.format("${unpack 7z %s %s}", devkit_file, devkit_dir),
    install_phase = string.format("${copy %s $out}", devkit_dir),
  },
}

targets.windows.cross_linux = targets.windows

local packages = {
  arch = { "mingw-w64", "make" },
  fedora = { "mingw64-gcc", "make" },
  ubuntu = { "gcc-mingw-w64", "build-essential" },
  void = { "cross-i686-w64-mingw32", "make" },
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
