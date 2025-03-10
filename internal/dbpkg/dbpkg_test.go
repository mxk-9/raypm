package dbpkg

import (
	"fmt"
	"maps"
	"os"
	"path"
	log "raypm/pkg/slog"
	"testing"
)

func TestOpenDatabase(t *testing.T) {
	log.Init(false)

	t.Run("open database", func(t *testing.T) {
		pth := path.Join("testdata", "wantDB.json")
		_, err := Open(pth)
		if err != nil {
			t.Errorf("Cannot open '%s': '%s'\n", pth, err)
		}
	})
}

func TestAddDelPackage(t *testing.T) {
	log.Init(false)

	inDb := path.Join("testdata", "wantDB.json")
	tmpDir := path.Join(os.TempDir())
	tmpDir, err := os.MkdirTemp(tmpDir, "db_test_*")

	wantPackages := map[string]pkg{
		"package": pkg{
			RequiredFor: []string{"neco-arc"},
		},
		"packageReq": pkg{},
		"neco-arc": pkg{
			DependsOn: []string{"package"},
		},
	}

	if err != nil {
		t.Errorf("Failed to create tempdir: '%s'\n", err)
		t.FailNow()
	}

	outDb := path.Join(tmpDir, "result.json")

	if _, err := os.Stat(outDb); err == nil {
		if err = os.Remove(outDb); err != nil {
			log.Errorln(err)
		}
	}

	local, err := Open(inDb)
	if err != nil {
		t.Errorf("Cannot open '%s': '%s'\n", inDb, err)
		t.FailNow()
	}

	t.Run("adding package", func(t *testing.T) {
		output := NewDb(outDb)
		defer output.WriteData()

		maps.Copy(output.Pkgs, local.Pkgs)
		wantPkgs := make(map[string]pkg)

		wantPkgs = map[string]pkg{
			"neco-arc": {
				DependsOn: []string{"package"},
			},
			"package": {
				RequiredFor: []string{"neco-arc"},
			},

			"packageDep": {
				DependsOn: []string{"packageReq"},
			},

			"packageReq": {
				RequiredFor: []string{"packageDep"},
			},
		}

		output.Add("neco-arc")
		output.AddDep("neco-arc", "package")

		if !maps.EqualFunc(wantPkgs, output.Pkgs, equalPkg) {
			t.Error(mismatchMaps(&wantPkgs, &output.Pkgs))
		}
	})

	t.Run("delete package", func(t *testing.T) {
		output, err := Open(outDb)
		if err != nil {
			t.Error(err)
		}
		defer output.WriteData()

		err = output.Del("packageDep")
		if err != nil {
			t.Error(err)
		}

	})

	t.Run("delete package that using by other", func(t *testing.T) {
		output, err := Open(outDb)
		if err != nil {
			t.Error(err)
		}
		defer output.WriteData()

		output.Del("package")

		// Use maps.EqualFunc
		if !(maps.EqualFunc(wantPackages, output.Pkgs, equalPkg)) {
			t.Error(mismatchMaps(&wantPackages, &output.Pkgs))
		}
	})

	// Compare two maps.
	t.Run("comparing result with expectation", func(t *testing.T) {
		result, err := Open(outDb)
		if err != nil {
			t.Error(err)
		}

		ok := maps.EqualFunc(wantPackages, result.Pkgs, equalPkg)
		if !ok {
			t.Error(mismatchMaps(&wantPackages, &result.Pkgs))
		}
	})
}

func equalPkg(a, b pkg) bool {
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

func mismatchMaps(expect, got *map[string]pkg) string {
	return fmt.Sprintf(
		"\nExpect:\n%v\nGot:\n%v\n",
		expect, got,
	)
}
