package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"raypm/internal/deptree"
	"raypm/internal/fetch"
	"raypm/internal/pkginfo"
	"raypm/internal/unpack"
	log "raypm/pkg/slog"
	"runtime"

	"github.com/fatih/color"
)

type Operation uint8

const (
	ListPackages Operation = iota + 1
	SyncPkgs
	FetchPkgInfo
	InstallPkg
	UninstallPkg
)

var (
	Debug           bool
	ProgramTask     Operation = 0
	PathToPkgs      string    = path.Join("third_party", "raypm", "internals")
	SelectedPackage string
	Target          string
)

func init() {
	var (
		listPkgs     bool
		fetchPkgInfo string
		installPkg   string
		removePkg    string
		syncPkgs     bool
	)

	flag.BoolVar(&listPkgs, "list", false, "List all available packages")
	flag.BoolVar(&Debug, "d", false, "Print debug logs")
	flag.BoolVar(&syncPkgs, "sync", false, "Get latest package's database")
	flag.StringVar(&fetchPkgInfo, "info", "", "Show information about package")
	flag.StringVar(&installPkg, "install", "", "Install a package")
	flag.StringVar(&removePkg, "remove", "", "Remove a package")
	flag.Parse()

	log.Init(Debug)

	if flag.NFlag() > 1 && !Debug || flag.NFlag() > 2 && Debug {
		log.Fatalln("Choose only one operation. Type 'raypm -h' to see them")
	}

	if flag.NFlag() == 0 {
		log.Warnln("There's nothing to do. Type 'raypm -h'")
	}

	if listPkgs {
		ProgramTask = ListPackages
	} else if fetchPkgInfo != "" {
		ProgramTask = FetchPkgInfo
		SelectedPackage = fetchPkgInfo
	} else if installPkg != "" {
		ProgramTask = InstallPkg
		SelectedPackage = installPkg
	} else if syncPkgs {
		ProgramTask = SyncPkgs
	}

	if os.Getenv("GOOS") == "" {
		Target = runtime.GOOS
	} else {
		Target = os.Getenv("GOOS")
	}

}

func main() {
	var (
		err error
	)

	switch ProgramTask {
	case SyncPkgs:
		log.Infoln("Synchronization...")
		var pathToArchive string

		raypmPkgs := ".raypm"
		log.Debugln("Creating .raypm directory")
		if _, err = os.Stat(raypmPkgs); err != nil {
			if err = os.MkdirAll(raypmPkgs, 0754); err != nil {
				log.Fatal("Failed to create '%s': %s", raypmPkgs, err)
			}
			log.Debugln("Directory created")
		} else {
			log.Debugln("Directory already exists")
		}

		if pathToArchive, err = fetch.Sync(); err != nil {
			log.Fatalln("Failed to sync:", err)
		}

		if pathToArchive == "" {
			log.Infoln("There is nothing to do")
			os.Exit(0)
		}

		if err = unpack.Unpack("zip", []string{pathToArchive}, []string{".raypm"}, nil); err != nil {
			log.Fatalln("Failed to unpack", err)
		}
	case InstallPkg:
		var (
			lockPath string = path.Join("third_party", "raypm", "lock")
			fileLock *os.File
			deps     *deptree.Tree
		)

		log.Info("Target:%s", Target)

		if _, err = os.Stat(lockPath); err == nil {
			log.Fatal("Another process is using '%s', exiting...", lockPath)
		} else {
			if fileLock, err = os.Create(lockPath); err != nil {
				log.Fatal("Cannot create lock file: %s\n", err)
			}
			log.Debugln("Created lock file")
			fileLock.Chmod(0000)
			log.Debugln("Change mod to 0000")
			fileLock.Close()
			log.Debugln("Closed file")
		}
		defer func() {
			if err = os.Chmod(lockPath, 0754); err != nil {
				log.Fatal("Cannot change mod for lock file:%s\n", err)
			}
			log.Debugln("Changed mod to 0754. Deleting file...")

			if err = os.Remove(lockPath); err != nil {
				log.Fatal("Cannot delete file:%s\n", err)
			}

			log.Debugln("Deleted lock file")
		}()

		if deps, err = deptree.NewDepTree(PathToPkgs, SelectedPackage, Target); err != nil {
			log.Error("Failed to resolve dependencies:\n%s\n", err)
		} else {
			deps.Install()
		}
	case ListPackages:
		var (
			dirs []os.DirEntry
		)

		log.Debug("Going to %s\n", PathToPkgs)
		if err = os.Chdir(PathToPkgs); err != nil {
			log.Fatal("Failed to open '%s':\n%s\n", PathToPkgs, err)
		}

		defer func() {
			os.Chdir(path.Join("..", "..", ".."))
			log.Debugln("Change directory back")
		}()

		if dirs, err = os.ReadDir("."); err != nil {
			log.Fatal("Failed to read '%s':\n%s\n", PathToPkgs, err)
		}

		log.Debug("Readed:\n%v\n", dirs)

		for _, item := range dirs {
			if item.IsDir() {
				currentPackage, err := pkginfo.NewPackageItem(item.Name(), Target)

				if err != nil {
					log.Fatalln(err)
				}

				color.Magenta(currentPackage.Name)
				fmt.Print("\t", currentPackage.Description, "\n")
			}
		}
	case FetchPkgInfo:
		var currentPackage *pkginfo.Package

		log.Info("Target:%s", Target)

		if err = os.Chdir(path.Join(PathToPkgs, SelectedPackage)); err != nil {
			log.Fatal("Package '%s' does not exists", SelectedPackage)
		}

		defer func() {
			os.Chdir(path.Join("..", "..", "..", ".."))
		}()

		if currentPackage, err = pkginfo.NewPackageItem(".", Target); err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Package Information:")
		fmt.Println(currentPackage.Info())
	}
}
