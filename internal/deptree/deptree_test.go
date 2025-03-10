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
	tmpRaypm, err = os.MkdirTemp(tmpRaypm, "raypm_*")
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

		localTree.Install()
	})
}

func TestUnistallPackages(t *testing.T) {
	// Test for uninstalling one package
	// Test for uninstalling package with dependencies, but withDeps is false
	// Test for uninstalling package with dependencies, but withDeps is true
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
