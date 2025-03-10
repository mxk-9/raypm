package deptree

import (
	"fmt"
	"os"
	"path"
	"raypm/internal/dbpkg"
	"raypm/internal/pkginfo"
	"raypm/internal/vars"
	log "raypm/pkg/slog"
)

type Node struct {
	Data    *PkgData
	Db      *dbpkg.PkgDb
	Pkg     *pkginfo.Package
	Depends []*Node // If len(Depends) is 0, we reach the end
	// Predefined variables
	Vars *vars.Vars
}

func NewNode(data *PkgData, db *dbpkg.PkgDb, internalName string) (depNode *Node, err error) {
	// Creates simple dependency node
	depNode = &Node{
		Data: data,
		Vars: &vars.Vars{},
		Db:   db,
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
	if localNode, err = NewNode(dn.Data, dn.Db, internalName); err != nil {
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

	inDb, inStore := checkExisting(dn.Pkg.Name, dn.Db, dn.Vars.Out)

	if inDb && inStore {
		log.Info("Package '%s' already installed", dn.Vars.Out)
		return
	} else if inDb != inStore {
		log.Errorln(
			"Seems there was an error while installation/uninstallation packages:",
		)
		err = fmt.Errorf("DatabaseError")
		return
	}

	for _, item := range dn.Depends {
		if err = item.InstallNode(); err != nil {
			return
		}
		dn.Db.Add(dn.Pkg.Name)
		dn.Db.AddDep(dn.Pkg.Name, item.Pkg.Name)
	}

	log.Infoln("Installing", dn.Pkg.Name)

	if err = fetchPhase(&dn.Pkg.FetchPhase, dn.Vars); err != nil {
		dn.Db.Del(dn.Pkg.Name)
		return
	}

	if err = unpackPhase(&dn.Pkg.UnpackPhase, dn.Vars); err != nil {
		dn.Db.Del(dn.Pkg.Name)
		return
	}

	if err = taskPhase("build", &dn.Pkg.BuildPhase, dn.Vars); err != nil {
		dn.Db.Del(dn.Pkg.Name)
		return
	}

	outDir := dn.Vars.Out
	if err = os.MkdirAll(outDir, 0754); err != nil {
		log.Error("Failed to create directory '%s': %s", outDir, err)
		dn.Db.Del(dn.Pkg.Name)
		return
	}

	if err = taskPhase("install", &dn.Pkg.InstallPhase, dn.Vars); err != nil {
		dn.Db.Del(dn.Pkg.Name)
		return
	}

	log.Info("Package '%s' installed", dn.Pkg.Name)

	dn.Db.Add(dn.Pkg.Name)

	return
}

// TODO
func (dn *Node) BuildNode() (err error) {
	return
}

// TODO
func (dn *Node) UninstallNode() (err error) {
	if dn.Pkg == nil {
		return
	}

	inDb, inStore := checkExisting(dn.Pkg.Name, dn.Db, dn.Vars.Out)

	if inDb != inStore {
		log.Errorln(
			"Seems there was an error while installation/uninstallation packages:",
		)
		err = fmt.Errorf("DatabaseError")
		return
	} else if !inDb && !inStore {
		log.Warn("Package '%s' is not installed", dn.Pkg.Name)
		return
	}

	if err = dn.Db.Del(dn.Pkg.Name); err != nil {
		return
	}

	log.Infoln("Uninstalling", dn.Pkg.Name)
	if err = taskPhase("uninstall", &dn.Pkg.UninstallPhase, dn.Vars); err != nil {
		return
	}

	if err = os.RemoveAll(dn.Vars.Out); err != nil {
		log.Errorln("Failed to remove package's directory")
		log.Errorln(err)
		return
	}

	os.RemoveAll(dn.Vars.Cache)

	log.Info("Package '%s' removed", dn.Pkg.Name)
	return
}

func checkExisting(pkgName string, db *dbpkg.PkgDb, dir string) (inDataBase, inStoreDir bool) {
	inDataBase = db.IsExists(pkgName)

	_, err := os.Stat(dir)
	inStoreDir = err == nil

	return
}
