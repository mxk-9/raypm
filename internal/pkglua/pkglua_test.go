package pkglua

import (
	"fmt"
	"maps"
	"path"
	log "raypm/pkg/slog"
	"slices"
	"testing"
)

func TestGoLua(t *testing.T) {
	log.Init(false)

	t.Run("generics", func(t *testing.T) {
		pd1 := new(Package)
		pd1.MData = make(MData)
		pd1.TargetSpec = make(TargetSpec)
		pd1.MData["name"] = "hello"
		pd1.TargetSpec["build_phase"] = []string{"hi", "b"}

		pd2 := new(Package)
		pd2.MData = make(MData)
		pd2.TargetSpec = make(TargetSpec)
		pd2.MData["name"] = "hello"
		pd2.TargetSpec["build_phase"] = []string{"hi", "b"}

		if !cmpPackage(pd1, pd2) {
			t.Error("My generic's usage is broken")
		}
	})

	t.Run("creating pkgdata", func(t *testing.T) {
		wantedPd := &Package{
			MData: map[string]string{
				"name":        "snake",
				"version":     "0.2.1",
				"description": "Simple snake on golang",
				"src_path":    ".",
				"build_path":  "build",
			},
			TargetSpec: map[string][]string{
				"dependencies": []string{
					"go", "mingw", "base",
				},
				"build_phase": []string{
					"${setenv CGO_ENABLED 1}",
					"${setenv CC x86_64-w64-mingw32-gcc}",
					"${setenv GOOS windows}",
					"${setenv GOARCH amd64}",
					"go build -x -ldflags '-s -w' -o build .",
				},
			},
		}

		pd, err := NewPackage(path.Join("testdata", "snake.lua"), "linux", "windows")
		if err != nil {
			t.Error(err)
		}

		if !cmpPackage(pd, wantedPd) {
			t.Errorf(
				"Expect:\n%v\n\nGot:\n%v\n",
				wantedPd, pd,
			)
		}
	})

	t.Run("pkgdata with linux packages", func(t *testing.T) {
		pd, err := NewPackage(path.Join("testdata", "base.lua"), "linux", "linux")
		if err != nil {
			t.Error(err)
		}

		fmt.Println(pd)
	})
}

func cmpPackage(pd1, pd2 *Package) bool {
	if !maps.Equal(pd1.MData, pd2.MData) {
		return false
	}

	if !maps.EqualFunc(pd1.TargetSpec, pd2.TargetSpec, slices.Equal) {
		return false
	}

	return true
}
