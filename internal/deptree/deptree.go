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
	"strings"
)

type PkgData struct {
	BasePath string
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

// internals — is a path to .raypm(it can be global or local)
func NewDepTree(internals, packageName, target string) (depTree *Tree,
	err error) {
	depTree = &Tree{
		Data: PkgData{
			BasePath: internals,
			PkgsPath: path.Join(internals, "pkgs"),
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

	depNode.Vars = vars.NewVars(data.BasePath, internalName)

	log.Debug("Vars:\n%v", depNode.Vars)

	internal := &pkginfo.Package{}

	log.Debug("Creating package item '%s'", internalName)
	internal, err = pkginfo.NewPackageItem(
		path.Join(data.PkgsPath, internalName),
		data.Target,
	)

	if err != nil {
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

	dn.Pkg.Info()
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
	// pkgsPath := path.Join(dn.Vars.Base, dn.Pkg.Name)
	if _, err = os.Stat(dn.Vars.Out); err == nil {
		log.Info("Package '%s' already installed", dn.Vars.Out)
		return
	} else {
		err = nil
	}

	if err = fetchPhase(&dn.Pkg.FetchPhase, dn.Vars); err != nil {
		return
	}

	if err = unpackPhase(&dn.Pkg.UnpackPhase, dn.Vars); err != nil {
		return
	}

	if err = taskPhase("build", &dn.Pkg.BuildPhase, dn.Vars); err != nil {
		return
	}

	outDir := dn.Vars.Out
	if err = os.MkdirAll(outDir, 0754); err != nil {
		log.Error("Failed to create directory '%s': %s", outDir, err)
		return
	}

	if err = taskPhase("install", &dn.Pkg.InstallPhase, dn.Vars); err != nil {
		return
	}

	log.Info("Package '%s' installed", dn.Pkg.Name)
	return
}

// TODO
func (dn *Node) BuildNode() (err error) {
	return
}

// TODO
func (dn *Node) UninstallNode(withDeps bool) (err error) {
	if dn.Pkg == nil {
		return
	}

	packageToUninstall := path.Join(".raypm", "store", dn.Pkg.Name)
	errMsg := fmt.Sprintf("Cannot delete a package '%s':", dn.Pkg.Name)

	if _, err = os.Stat(packageToUninstall); err != nil {
		log.Error(errMsg)
		log.Errorln(err)
		return
	}

	if !withDeps && len(dn.Depends) > 0 {
		log.Error(errMsg)
		log.Errorln("Package depends on:")
		for _, item := range dn.Depends {
			fmt.Printf("\t+ %s\n", item.Pkg.Name)
		}
	} else if withDeps && len(dn.Depends) > 0 {
		for _, item := range dn.Depends {
			if err = item.UninstallNode(true); err != nil {
				return
			}
		}
	}

	log.Infoln("Uninstalling", dn.Pkg.Name)
	if err = taskPhase("uninstall", &dn.Pkg.UninstallPhase, dn.Vars); err != nil {
		return
	}

	if err = os.RemoveAll(packageToUninstall); err != nil {
		log.Errorln("Failed to remove package's directory")
		log.Errorln(err)
		return
	}
	return
}

func fetchPhase(data *[]pkginfo.FetchPath, vv *vars.Vars) (err error) {
	for _, item := range *data {
		log.Info("Getting '%s'", item.From)
		to := ""

		expanded := vv.ExpandVars(&item.To)

		for _, dest := range expanded {
			to = path.Join(to, dest)
		}

		if err = fetch.GetFile(item.From, to); err != nil {
			log.Errorln("Fetch phase failed:")
			log.Error("Failed to get '%s': %s", item.From, err)
			break
		}
	}

	return
}

func unpackPhase(data *[]pkginfo.UnpackTask, vv *vars.Vars) (err error) {
	for _, item := range *data {
		from := path.Join(vv.ExpandVars(&item.Src)...)
		log.Debug("Expanded from '%v'\nto '%v'", item.Src, from)

		log.Info("Unpacking '%s'", from)

		to := path.Join(vv.ExpandVars(&item.Dest)...)
		log.Debug("Expanded from '%v'\nto '%v'", item.Dest, to)

		err = unpack.Unpack(item.Type, from, to, item.SelectedItems)

		if err != nil {
			log.Errorln("Unpack phase failed:")
			log.Error("Failed to unpack '%s': '%s", from, err)
			break
		}
	}
	return
}

func taskPhase(phase string, data *[]pkginfo.CommandTask, vv *vars.Vars) (err error) {

	if !(phase == "build" || phase == "install" || phase == "uninstall") {
		log.Error("Uknown phase '%s'", phase)
		return
	}

	for _, item := range *data {
		if err = task.Do(phase, item, vv); err != nil {
			log.Error("%s phase failed:", strings.ToTitle(phase))
			log.Errorln(err)
			break
		}
	}
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
