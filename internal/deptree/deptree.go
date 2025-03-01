package deptree

import (
	"fmt"
	"os"
	"path"
	"raypm/internal/fetch"
	"raypm/internal/pkginfo"
	"raypm/internal/task"
	"raypm/internal/unpack"
	"raypm/internal/vars"
	log "raypm/pkg/slog"
)

type PkgData struct {
	PkgsPath string
	Target   string
}

type Node struct {
	Data    *PkgData
	Pkg     *pkginfo.Package
	Depends []*Node // If len(Depends) is 0, we reach the end
	// Predefined variables
	Vars *vars.Vars
}

type Tree struct {
	Nodes *Node
	Data  PkgData
}

// internals — is a path to packages
func NewDepTree(internals, packageName, target string) (depTree *Tree,
	err error) {

	depTree = &Tree{
		Data: PkgData{
			PkgsPath: internals,
			Target:   target,
		},
	}

	log.Debugln("Creating dependency tree")
	if depTree.Nodes, err = NewNode(&depTree.Data, packageName); err != nil {
		err = fmt.Errorf("Failed to build dependency tree:\n%s\n", err)
	}

	return
}

func NewNode(data *PkgData, internalName string) (depNode *Node, err error) {
	// Creates simple dependency node
	depNode = &Node{
		Data: data,
		Vars: &vars.Vars{},
	}

	depNode.Vars.Cache = path.Join(".raypm", "cache", internalName)
	depNode.Vars.Src = path.Join(depNode.Vars.Cache, "src")
	depNode.Vars.Fetch = path.Join(depNode.Vars.Cache, "fetch")
	depNode.Vars.Out = path.Join(".raypm", "store", internalName)
	depNode.Vars.Package = path.Join(".raypm", "pkgs", internalName)

	log.Debug("Vars:\n%v", depNode.Vars)

	internal := &pkginfo.Package{}

	log.Debug("Creating package item '%s'", internalName)
	if internal, err = pkginfo.NewPackageItem(path.Join(data.PkgsPath, internalName),
		data.Target); err != nil {
		return
	}

	log.Debugln("Looking for dependencies...")
	for _, item := range internal.Dependendencies {
		log.Debug("Found '%s', appending to list", item)
		if err = depNode.Append(depNode.Data, item); err != nil {
			return
		}
		depNode.Vars.Dep = append(depNode.Vars.Dep, item)
	}
	log.Debug("Success fetching dependencies for package '%s'", internalName)

	depNode.Pkg = internal

	return
}

// Expand dependency nodes array
func (dn *Node) Append(data *PkgData, internalName string) (err error) {

	localNode := &Node{}

	log.Debug("Creating node '%s'", internalName)
	if localNode, err = NewNode(dn.Data, internalName); err != nil {
		err = fmt.Errorf("Failed to create node: %s\n", err)
		return
	}

	log.Debugln("Appending")
	dn.Depends = append(dn.Depends, localNode)

	return
}

func (dn *Node) ShowNode() {
	if dn.Pkg == nil {
		return
	}

	for _, item := range dn.Depends {
		item.ShowNode()
	}

	log.Infoln(dn.Pkg.Info())
	fmt.Println()
}

func (dn *Node) InstallNode() (err error) {
	if dn.Pkg == nil {
		return
	}

	for _, item := range dn.Depends {
		if err = item.InstallNode(); err != nil {
			return
		}
	}

	log.Infoln("Installing", dn.Pkg.Name)
	pkgsPath := path.Join(".raypm", "store", dn.Pkg.Name)
	if _, err = os.Stat(pkgsPath); err == nil {
		log.Info("Package '%s' already installed", pkgsPath)
		return
	} else {
		err = nil
	}

	for _, item := range dn.Pkg.FetchPhase {
		log.Info("Getting '%s'", item.From)
		to := ""

		expanded := dn.Vars.ExpandVars(&item.To)

		for _, dest := range expanded {
			to = path.Join(to, dest)
		}
		if err = fetch.GetFile(item.From, to); err != nil {
			log.Error("Failed to get '%s'", item.From)
			return
		}
	}

	for _, item := range dn.Pkg.UnpackPhase {
		from := path.Join(dn.Vars.ExpandVars(&item.Src)...)
		log.Debug("Expanded from '%v'\nto '%v'", item.Src, from)

		log.Info("Unpacking '%s'", from)

		to := path.Join(dn.Vars.ExpandVars(&item.Dest)...)
		log.Debug("Expanded from '%v'\nto '%v'", item.Dest, to)

		if err = unpack.Unpack(item.Type, from, to, item.SelectedItems); err != nil {
			return
		}
	}

	for _, item := range dn.Pkg.BuildPhase {
		if err = task.Do("build", item, dn.Vars); err != nil {
			log.Errorln("Build phase failed")
			return
		}
	}

	outDir := dn.Vars.Out
	if _, err = os.Stat(outDir); err == nil {
		log.Error("Directory '%s' already exists!", outDir)
		err = fmt.Errorf("DirectoryAlreadyExists")
		return
	}

	if err = os.MkdirAll(outDir, 0754); err != nil {
		log.Error("Failed to create directory '%s': %s", outDir, err)
		return
	}

	for _, item := range dn.Pkg.InstallPhase {
		if err = task.Do("install", item, dn.Vars); err != nil {
			log.Errorln("Install phase failed")
			return
		}
	}

	log.Info("Package '%s' installed", dn.Pkg.Name)
	return
}

// TODO: UninstallNode
func (dn *Node) UninstallNode(withDeps bool) (err error) {
	return
}

func (dp *Tree) ShowTree() {
	log.Infoln("Packages path:", dp.Data.PkgsPath)
	log.Infoln("Target:", dp.Data.Target)

	dp.Nodes.ShowNode()
}

func (dp *Tree) Install() (err error) {
	if err = dp.Nodes.InstallNode(); err != nil {
		log.Error("Package installation failed")
		return
	}
	return
}
