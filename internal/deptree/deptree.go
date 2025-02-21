package deptree

import (
	"fmt"
	"path"
	"raypm/internal/fetch"
	"raypm/internal/pkginfo"
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
	}

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
	}
	log.Debug("Success fetching dependencies for package '%s'", internalName)

	depNode.Pkg = internal

	return
}

func (dn *Node) Append(data *PkgData, internalName string) (err error) {
	// Expand dependency nodes array
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
		err = fmt.Errorf("Empty node")
		return
	}

	for _, item:= range dn.Depends {
		item.InstallNode()
	}

	// TODO: Fetch phase
	for _, item := range dn.Pkg.FetchPhase {
		log.Info("Getting '%s'", item.From)
		to := ""
		for _, dest := range item.To {
			to = path.Join(to, dest)
		}
		if err = fetch.GetFile(item.From, to); err != nil {
			log.Error("Failed to get '%s'", item.From)
			return
		}
	}
	
	// TODO: Unpack phase
	
	// TODO: Install phase
	// OPTIONAL: Uninstall phase
	return
}

func (dp *Tree) ShowTree() {
	log.Infoln("Packages path:", dp.Data.PkgsPath)
	log.Infoln("Target:", dp.Data.Target)

	dp.Nodes.ShowNode()
}

func (dp *Tree) Install() (err error) {
	// Going through all dependencies packages until len(Depends) == 0 then
	// recursively install everything.
	// For each phase we have one package

	return
}
