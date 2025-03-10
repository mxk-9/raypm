/*
This is a temporary solution for using database.
TODO: use mysql instead of json
*/
package dbpkg

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path"
	log "raypm/pkg/slog"
	"slices"
)

type Relations struct {
	DependsOn   []string `json:"depends_on"`
	RequiredFor []string `json:"required_for"`
}

func IsRelEqual(a, b Relations) bool {
	aDep := a.DependsOn
	aReq := a.RequiredFor
	bDep := b.DependsOn
	bReq := b.RequiredFor

	if len(aDep) != len(bDep) || len(aReq) != len(bReq) {
		return false
	}

	for i := range aDep {
		if aDep[i] != bDep[i] {
			return false
		}
	}

	for i := range aReq {
		if aReq[i] != bReq[i] {
			return false
		}
	}

	return true
}

type PkgsRel map[string]Relations

func (a PkgsRel) IsEqual(b PkgsRel) bool {
	return maps.EqualFunc(a, b, IsRelEqual)
}

type PkgDb struct {
	Pkgs     PkgsRel
	PathToDb string
}

func NewDb(pathToDb string) *PkgDb {
	return &PkgDb{
		PathToDb: pathToDb,
		Pkgs:     make(map[string]Relations),
	}
}

func Open(pathToDb string) (pd *PkgDb, err error) {

	if _, err = os.Stat(pathToDb); err != nil {
		log.Errorln(err)
		return
	}
	pd = &PkgDb{
		Pkgs:     make(map[string]Relations, 0),
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

func (pd *PkgDb) IsExists(RelationsName string) bool {
	_, ok := pd.Pkgs[RelationsName]

	return ok
}

func (pd *PkgDb) Add(RelationsName string) {
	if _, ok := pd.Pkgs[RelationsName]; !ok {
		pd.Pkgs[RelationsName] = Relations{}
	}
}

func (pd *PkgDb) AddDep(RelationsName, depName string) {
	addingTo, okTo := pd.Pkgs[RelationsName]
	dep, okDep := pd.Pkgs[depName]

	if okDep && okTo {
		addingTo.DependsOn = alphIns(addingTo.DependsOn, depName)
		dep.RequiredFor = alphIns(dep.RequiredFor, RelationsName)

		pd.Pkgs[RelationsName] = addingTo
		pd.Pkgs[depName] = dep
	}
}

func (pd *PkgDb) Del(RelationsName string) (err error) {
	if _, ok := pd.Pkgs[RelationsName]; ok {
		req := pd.Pkgs[RelationsName]

		if len(req.RequiredFor) > 0 {
			err = fmt.Errorf("PackageIsRequiredByOther")
			for _, item := range req.RequiredFor {
				log.Error("Package '%s' depends on '%s'", item, RelationsName)
			}
		} else {
			delete(pd.Pkgs, RelationsName)

			for k, v := range pd.Pkgs {
				ignoreInd := -1
				left := make([]string, 0)
				right := make([]string, 0)

				for i, item := range v.RequiredFor {
					if item == RelationsName {
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

func alphIns(s []string, item string) (newSlice []string) {
	if len(s) == 0 {
		newSlice = append(s, item)
		return
	} else if len(s) == 1 {
		if s[0] < item {
			newSlice = []string{s[0], item}
		} else if item < s[0] {
			newSlice = []string{item, s[0]}
		}
		return
	}

	for i := range s {
		if i == len(s)-1 {
			break
		}

		left := s[i]
		right := s[i+1]

		if item < left {
			newSlice = slices.Insert(s, 0, item)
			break
		} else if left <= item && item <= right {
			newSlice = slices.Insert(s, i+1, item)
			break
		} else if right < item {
			newSlice = append(s, item)
		}
	}

	return
}
