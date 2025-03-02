package vars

import (
	"path"
	log "raypm/pkg/slog"
	"strings"
)

type Vars struct {
	Base    string
	Src     string
	Out     string
	Fetch   string
	Cache   string
	Package string
	Dep     []string
}

// 'base' is a path to '.raypm'
func NewVars(base, packageName string) (vv *Vars) {
	vv = &Vars{
		Base: base,
	}

	vv.Cache = path.Join(vv.Base, "cache", packageName)
	vv.Src = path.Join(vv.Cache, "src")
	vv.Fetch = path.Join(vv.Cache, "fetch")
	vv.Out = path.Join(vv.Base, "store", packageName)
	vv.Package = path.Join(vv.Base, "pkgs", packageName)
	return
}

func (vv *Vars) ExpandVars(line *[]string) (changedStr []string) {
	var word string

	changedStr = make([]string, 0)

	for _, item := range *line {
		word = vv.matchAndReplace(item)
		changedStr = append(changedStr, word)
	}

	log.Debug("Expanded from '%v' to '%v'", *line, changedStr)

	return
}

func (vv *Vars) matchAndReplace(text string) (word string) {
	word = ""

	word = strings.ReplaceAll(text, "$src", vv.Src)
	word = strings.ReplaceAll(word, "$out", vv.Out)
	word = strings.ReplaceAll(word, "$fetch", vv.Fetch)
	word = strings.ReplaceAll(word, "$cache", vv.Cache)
	word = strings.ReplaceAll(word, "$pkg", vv.Package)
	word = strings.ReplaceAll(word, "$dep", path.Join(".raypm", "store"))

	return
}
