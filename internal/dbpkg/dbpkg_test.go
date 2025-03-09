package dbpkg

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path"
	log "raypm/pkg/slog"
	"reflect"
	"testing"
)

func TestStructToJson(t *testing.T) {
	log.Init(false)

	t.Run("check convertion", func(t *testing.T) {
		dbobj := &PkgDb{
			Pkgs: make(map[string]pkg, 0),
		}

		dbobj.Pkgs["bebra"] = pkg{
			DependsOn:   []string{"snus"},
			RequiredFor: []string{"notest"},
		}

		dbobj.Pkgs["amogus"] = pkg{
			DependsOn:   []string{"bebra"},
			RequiredFor: []string{"package", "abobus"},
		}

		err := json.NewEncoder(os.Stdout).Encode(&dbobj.Pkgs)
		if err != nil {
			t.Errorf("Failed to encode: %s", err)
		}

	})
}

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
	log.Init(true)

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

		output.Add("neco-arc")
		output.AddDep("neco-arc", "package")

		fmt.Printf("%v\n", output.Pkgs)
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

		if !(reflect.DeepEqual(wantPackages, output.Pkgs)) {
			t.Errorf(
				"\nExpect:\n%v\nGot:\n%v\n",
				wantPackages, output.Pkgs,
			)
		}
	})

	// Compare two maps.
	t.Run("comparing result with expectation", func(t *testing.T) {
		result, err := Open(outDb)
		if err != nil {
			t.Error(err)
		}

		ok := reflect.DeepEqual(result.Pkgs, wantPackages)
		if !ok {
			t.Errorf(
				"%s\n\ngot: %v\n\nwant: %v",
				"Failed to compare result with expectation:",
				result.Pkgs,
				wantPackages,
			)
		}
	})
}
