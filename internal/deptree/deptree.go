package deptree

import (
	"fmt"
	"path"
	"raypm/internal/dbpkg"
	log "raypm/pkg/slog"
)

type PkgData struct {
	BasePath string
	PkgsPath string
	Target   string
	Host     string
}

type Tree struct {
	Nodes    *Node
	Data     PkgData
	DataBase *dbpkg.PkgDb
}

func NewDepTree(raypmPath, packageName, host, target string, db *dbpkg.PkgDb) (depTree *Tree,
	err error) {
	depTree = &Tree{
		Data: PkgData{
			BasePath: raypmPath,
			PkgsPath: path.Join(raypmPath, "pkgs"),
			Target:   target,
			Host:     host,
		},
		DataBase: db,
	}

	log.Debugln("Creating dependency tree")
	if depTree.Nodes, err = NewNode(&depTree.Data, depTree.DataBase, packageName); err != nil {
		err = fmt.Errorf("Failed to build dependency tree:\n%s\n", err)
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

func (dp *Tree) Uninstall() (err error) {
	if err = dp.Nodes.UninstallNode(); err != nil {
		log.Errorln("Failed to delete package")
		return
	}

	return
}
