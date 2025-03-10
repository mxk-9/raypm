package deptree

import (
	"fmt"
	"io"
	"os"
	"path"
	"raypm/internal/dbpkg"
	"raypm/pkg/progress"
	log "raypm/pkg/slog"
	"runtime"
	"testing"
)

func TestInstall(t *testing.T) {
	var (
		tmpRaypm string
		err      error
		db       *dbpkg.PkgDb
	)
	log.Init(false)

	tmpRaypm = path.Join(os.TempDir())
	tmpRaypm, err = os.MkdirTemp(tmpRaypm, "install_test_*")
	if err != nil {
		t.Errorf("Cannot create temp directory: %s\n", err)
		t.FailNow()
	}

	if err = copyTestPkgs(tmpRaypm, "pkgs"); err != nil {
		t.Errorf("Failed to copy files:\n%s\n", err)
		t.FailNow()
	}

	dbPathJson := path.Join(tmpRaypm, "db.json")
	db = dbpkg.NewDb(dbPathJson)
	defer db.WriteData()

	t.Run("install a package", func(t *testing.T) {
		localTree, err := NewDepTree(
			tmpRaypm, "testdep", runtime.GOOS, db,
		)

		if err != nil {
			t.Errorf("Failed to resolve dependencies:\n%s\n", err)
		}

		if err = localTree.Install(); err != nil {
			t.Error(err)
		}

		wantPkgs := dbpkg.PkgsRel{
			"testdep": {
				DependsOn: []string{
					"another", "testpackage",
				},
			},

			"testpackage": {
				RequiredFor: []string{
					"testdep",
				},
			},

			"another": {
				RequiredFor: []string{
					"testdep",
				},
			},
		}

		storeBase := path.Join(tmpRaypm, "store")
		wantFiles := []string{
			path.Join(storeBase, "testdep"),
			path.Join(storeBase, "another"),
			path.Join(storeBase, "testpackage"),
		}

		if !wantPkgs.IsEqual(db.Pkgs) {
			t.Error(mismatchMaps(&wantPkgs, &db.Pkgs))
		}

		for _, item := range wantFiles {
			if _, err = os.Stat(item); err != nil {
				t.Error(err)
			}
		}
	})
}

func TestUnistallPackages(t *testing.T) {
	// install "testdep"
	var (
		tmpRaypm string
		err      error
		db       *dbpkg.PkgDb
	)
	log.Init(false)

	tmpRaypm = path.Join(os.TempDir())
	tmpRaypm, err = os.MkdirTemp(tmpRaypm, "uninstall_test_*")
	if err != nil {
		t.Errorf("Cannot create temp directory: %s\n", err)
		t.FailNow()
	}

	if err = copyTestPkgs(tmpRaypm, "pkgs"); err != nil {
		t.Errorf("Failed to copy files:\n%s\n", err)
		t.FailNow()
	}

	dbPathJson := path.Join(tmpRaypm, "db.json")
	db = dbpkg.NewDb(dbPathJson)
	defer db.WriteData()

	depTree, err := NewDepTree(tmpRaypm, "testdep", runtime.GOOS, db)
	if err != nil {
		t.Errorf("Failed to resolve dependencies:\n%s\n", err)
		t.FailNow()
	}

	if err = depTree.Install(); err != nil {
		t.FailNow()
	}

	// Test uninstalling a package, that is dependency for other (try to uninstall "another")
	t.Run("uninstall a package, that is dependency for other", func(t *testing.T) {
		storeBase := path.Join(tmpRaypm, "store")
		wantFiles := []string{
			path.Join(storeBase, "testdep"),
			path.Join(storeBase, "another"),
			path.Join(storeBase, "testpackage"),
		}

		wantPkgs := dbpkg.PkgsRel{
			"testdep": {
				DependsOn: []string{
					"another", "testpackage",
				},
			},

			"testpackage": {
				RequiredFor: []string{
					"testdep",
				},
			},

			"another": {
				RequiredFor: []string{
					"testdep",
				},
			},
		}

		localTree, err := NewDepTree(tmpRaypm, "another", runtime.GOOS, db)
		if err != nil {
			t.Errorf("Failed to resolve dependencies: %s\n", err)
		}

		err = localTree.Uninstall()
		if err == nil {
			dbLocal, err := dbpkg.Open(db.PathToDb)
			if err != nil {
				t.Errorf("Cannot read '%s'\n%s\n", db.PathToDb, err)
			}

			filesInStore, err := os.ReadDir(storeBase)
			if err != nil {
				t.Error(err)
			}

			t.Errorf(
				"%s\n%s%v\n%s%s%s\n%v",
				"Somehow pm deleted a package, that is a dependency for other:",
				"Database containing: ", dbLocal.Pkgs,
				"Files in '", storeBase, "':",
				filesInStore,
			)
		}

		for _, item := range wantFiles {
			if _, err = os.Stat(item); err != nil {
				t.Error(err)
			}
		}

		if !wantPkgs.IsEqual(db.Pkgs) {
			t.Error(mismatchMaps(&wantPkgs, &db.Pkgs))
		}
	})

	// Test uninstalling one package (try to uninstall "testdep")
	t.Run("uninstall one package", func(t *testing.T) {
		storeBase := path.Join(tmpRaypm, "store")
		wantFiles := []string{
			path.Join(storeBase, "another"),
			path.Join(storeBase, "testpackage"),
		}

		wantPkgs := dbpkg.PkgsRel{
			"testpackage": {},
			"another":     {},
		}

		localTree, err := NewDepTree(tmpRaypm, "testdep", runtime.GOOS, db)
		if err != nil {
			t.Errorf("Failed to resolve dependencies: %s\n", err)
		}

		err = localTree.Uninstall()
		if err != nil {
			t.Error(err)
		}

		for _, item := range wantFiles {
			if _, err = os.Stat(item); err != nil {
				t.Error(err)
			}
		}

		if !wantPkgs.IsEqual(db.Pkgs) {
			t.Error(mismatchMaps(&wantPkgs, &db.Pkgs))
		}
	})
}

func copyTestPkgs(dst, src string) (err error) {
	var (
		fInfo    os.FileInfo
		contents []os.DirEntry
	)

	fInfo, err = os.Stat(src)
	if err != nil {
		return
	}

	if fInfo.IsDir() {
		ldest := path.Join(dst, fInfo.Name())
		if err = os.MkdirAll(ldest, 0754); err != nil {
			return
		}

		contents, err = os.ReadDir(src)
		if err != nil {
			return
		}
		for _, item := range contents {
			copyTestPkgs(ldest, path.Join(src, item.Name()))
		}
	} else {
		var (
			inFile  *os.File
			outFile *os.File
			ldest   string = path.Join(dst, fInfo.Name())
		)

		inFile, err = os.Open(src)
		if err != nil {
			return
		}
		defer inFile.Close()

		outFile, err = os.Create(ldest)
		if err != nil {
			return
		}
		defer outFile.Close()

		prog := progress.NewProgress(true, fInfo.Name(), inFile)
		prog.CountFileSize(inFile)

		_, err = io.Copy(outFile, prog)
		fmt.Println()
	}

	return
}

func mismatchMaps(expect, got *dbpkg.PkgsRel) string {
	return fmt.Sprintf(
		"\nExpect:\n%v\nGot:\n%v\n",
		expect, got,
	)
}
