local name = "base"
local version = "1"
local description = "Base linux packages for developing with Raylib"
local supported_systems = { "linux" }

local targets = {
  linux = {
    install_phase = "${pkgman}",
  },
}

local packages = {
  arch = {
    "alsa-lib",
    "mesa",
    "libx11",
    "libxrandr",
    "libxi",
    "libxcursor",
    "libxinerama",
  },

  fedora = {
    "alsa-lib-devel",
    "mesa-libGL-devel",
    "libX11-devel",
    "libXrandr-devel",
    "libXi-devel",
    "libXcursor-devel",
    "libXinerama-devel",
    "libatomic",
  },

  ubuntu = {
    "libasound2-dev",
    "libx11-dev",
    "libxrandr-dev",
    "libxi-dev",
    "libgl1-mesa-dev",
    "libglu1-mesa-dev",
    "libxcursor-dev",
    "libxinerama-dev",
    "libwayland-dev",
    "libxkbcommon-dev",
  },

  void = {
    "alsa-lib-devel",
    "libglvnd-devel",
    "libX11-devel",
    "libXrandr-devel",
    "libXi-devel",
    "libXcursor-devel",
    "libXinerama-devel",
    "mesa",
    "MesaLib-devel",
    "mesa-dri",
    "mesa-intel-dri",
  },
}

packages.mint = packages.ubuntu
packages.manjaro = packages.arch

Data = {
  name = name,
  description = description,
  version = version,
  supported_systems = supported_systems,
  targets = targets,
  packages = packages,
}
