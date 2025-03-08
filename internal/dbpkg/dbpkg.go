/*
This is a temporary solution for using database.
TODO: use mysql instead of json
*/
package dbpkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	log "raypm/pkg/slog"
)

type pkg struct {
	DependsOn   []string `json:"depends_on"`
	RequiredFor []string `json:"required_for"`
}

type PkgDb struct {
	Pkgs     map[string]pkg
	PathToDb string
}

func NewDb(pathToDb string) *PkgDb {
	return &PkgDb{
		PathToDb: pathToDb,
		Pkgs:     make(map[string]pkg),
	}
}

func Open(pathToDb string) (pd *PkgDb, err error) {
	if _, err = os.Stat(pathToDb); err != nil {
		log.Errorln(err)
		return
	}
	pd = &PkgDb{
		Pkgs:     make(map[string]pkg, 0),
		PathToDb: pathToDb,
	}

	fDb, err := os.Open(pd.PathToDb)
	if err != nil {
		log.Error("Failed to open database:")
		log.Errorln(err)
		return
	}
	defer fDb.Close()

	if err = json.NewDecoder(fDb).Decode(&pd.Pkgs); err != nil {
		log.Errorln("Cannot decode database:")
		log.Errorln(err)
		return
	}

	return
}

func (pd *PkgDb) WriteData() (err error) {
	dir := path.Dir(pd.PathToDb)
	if _, err = os.Stat(dir); err != nil {
		if err = os.MkdirAll(dir, 0754); err != nil {
			log.Error("Failed to create '%s':", dir)
			log.Errorln(err)
			return
		}
	}

	fDb, err := os.Create(pd.PathToDb)
	if err != nil {
		log.Errorln("Cannot create database file:")
		log.Errorln(err)
	}
	defer fDb.Close()

	if err = json.NewEncoder(fDb).Encode(&pd.Pkgs); err != nil {
		log.Errorln("Cannot encode to database file:")
		log.Errorln(err)
	}
	return
}

func (pd *PkgDb) IsExists(pkgName string) bool {
	_, ok := pd.Pkgs[pkgName]

	return ok
}

func (pd *PkgDb) Add(pkgName string) {
	if _, ok := pd.Pkgs[pkgName]; !ok {
		pd.Pkgs[pkgName] = pkg{}
	}
}

func (pd *PkgDb) AddDep(pkgName, depName string) {
	addingTo, okTo := pd.Pkgs[pkgName]
	dep, okDep := pd.Pkgs[depName]

	if okDep && okTo {
		addingTo.DependsOn = append(addingTo.DependsOn, depName)
		dep.RequiredFor = append(dep.RequiredFor, pkgName)

		pd.Pkgs[pkgName] = addingTo
		pd.Pkgs[depName] = dep
	}
}

func (pd *PkgDb) Del(pkgName string) (err error) {
	if _, ok := pd.Pkgs[pkgName]; ok {
		req := pd.Pkgs[pkgName]

		log.Debug("Looking if '%s' need for other package", pkgName)

		if len(req.RequiredFor) > 0 {
			err = fmt.Errorf("PackageIsRequiredByOther")
			for _, item := range req.RequiredFor {
				log.Error("Package '%s' depends on '%s'", item, pkgName)
			}
		} else {
			delete(pd.Pkgs, pkgName)

			for k, v := range pd.Pkgs {
				ignoreInd := -1
				left := make([]string, 0)
				right := make([]string, 0)

				for i, item := range v.RequiredFor {
					if item == pkgName {
						ignoreInd = i
						break
					}
				}

				if ignoreInd >= 0 && ignoreInd < len(v.RequiredFor)-1 {
					left = v.RequiredFor[:ignoreInd]
					right = v.RequiredFor[ignoreInd+1:]

					left = append(left, right...)
					v.RequiredFor = left
					pd.Pkgs[k] = v
				} else if ignoreInd == 0 && len(v.RequiredFor) == 1 {
					v.RequiredFor = []string{}
					pd.Pkgs[k] = v
				}
			}
		}
	}

	return
}
