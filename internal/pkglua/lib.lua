-- Contents:
-- 1. Variables
-- 2. Functions
-- 2.1 functions (local)
-- 2.2 functions (global)

-- Variables

local available_systems = {
  "all",
  "linux",
  "windows",
}

local pkgman_base_cmd = {
  arch = {
    i = { "pacman", "-S", "--needed" },
    u = { "pacman", "-R", "--needed" },
  },

  fedora = {
    i = { "dnf", "install" },
    u = { "dnf", "uninstall" },
  },

  ubuntu = {
    i = { "apt", "install" },
    u = { "apt", "purge" },
  },

  void = {
    i = { "xbps-install" },
    u = { "xbps-uninstall" },
  },
}

pkgman_base_cmd.manjaro = pkgman_base_cmd.arch
pkgman_base_cmd.mint = pkgman_base_cmd.ubuntu

local root_permission = "sudo"

-- functions (global)
function Get_Pkgman_Cmd(luaPkg, distro)
  local cmd = {}
  dofile(luaPkg)

  if Data.packages == nil then
    return nil
  end

  local pkg_list = Data.packages[distro]
  local cmd_base = pkgman_base_cmd[distro]

  if pkg_list == nil then
    return nil, "LinuxDistroIsNotSupportByPackage", distro
  end

  if cmd_base == nil then
    return nil, "LinuxDistroIsNotSupportByRaypm", distro
  end

  cmd = { root_permission }

  local base = {}
  if install then
    base = cmd_base.i
  else
    base = cmd_base.u
  end

  for _, v in pairs(base) do
    table.insert(cmd, v)
  end

  for _, v in pairs(pkg_list) do
    table.insert(cmd, v)
  end

  return cmd
end

-- Returns 'pkg' object
-- If fails, returns nil, error type and errors object
function Get_Metadata(pkg_lua_file)
  dofile(pkg_lua_file)
  local metadata = {}

  metadata.name = Data.name
  metadata.version = Data.version
  metadata.description = Data.description
  metadata.src_path = Data.src_path
  metadata.build_path = Data.build_path
  return metadata
end

function Get_Phases(pkg_lua_file, host, target)
  dofile(pkg_lua_file)
  local phases = {}

  if target ~= host then
    phases = Data.targets[target]["cross_" .. host]
  else
    phases = Data.targets[target]
  end

  if phases == nil then
    return nil, "UnknownSystem", host, target
  end

  return phases
end
