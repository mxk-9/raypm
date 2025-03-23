local name = "snake"
local version = "0.2.1"
local description = [[Simple snake on golang]]
local supported_systems = { "linux", "windows" }

local src_path = "."
local build_path = "build"

local targets = {
  linux = {
    dependencies = { "go", "base" },
    build_phase = string.format("go build -x -o %s %s", build_path, src_path),
  },

  windows = {
    dependencies = { "go", "mingw" },
    build_phase = [[
        ${setenv CGO_ENABLED 1}
        ${setenv CC x86_64-w64-mingw32-gcc}
        ${setenv GOOS windows}
        ${setenv GOARCH amd64}
    ]] .. string.format(
      "go build -x -ldflags '-s -w' -o %s %s",
      build_path,
      src_path
    ),
  },
}

-- Ability to build the package from Linux to Windows
targets.windows.cross_linux = {
  dependencies = { "go", "mingw", "base" },
  build_phase = targets.windows.build_phase,
}

-- What we will return(required table)
Data = {
  name = name,
  description = description,
  version = version,
  supported_systems = supported_systems,
  src_path = src_path,
  build_path = build_path,
  targets = targets,
}
