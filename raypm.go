package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"raypm/internal/dbpkg"
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
	ListPackages Operation = iota
	SyncPkgs
	Clean
	FetchPkgInfo
	InstallPkg
	RemovePkg
)

func main() {
	var (
		ProgramTask     Operation = 0
		LocalRaypmPath  string    = ".raypm"
		RaypmPath       string
		PathToPkgs      string
		lockPath        string
		SelectedPackage string
		dbJson          string
		Target          string

		err error

		listPkgs     bool
		Debug        bool
		syncPkgs     bool
		fetchPkgInfo string
		installPkg   string
		removePkg    string
		cleanStorage string
	)

	flag.BoolVar(&listPkgs, "list", false, "List all available packages")
	flag.BoolVar(&Debug, "d", false, "Print debug logs")
	flag.BoolVar(&syncPkgs, "sync", false, "Get latest package's database")
	flag.StringVar(&fetchPkgInfo, "info", "", "Show information about package")
	flag.StringVar(&installPkg, "install", "", "Install a package")
	flag.StringVar(&removePkg, "remove", "", "Remove a package")
	flag.StringVar(&cleanStorage,
		"clean",
		"",
		"Cleaning raypm's storage. Available options: 'cache', 'all'",
	)
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
	} else if removePkg != "" {
		ProgramTask = RemovePkg
		SelectedPackage = removePkg
	} else if cleanStorage != "" {
		ProgramTask = Clean
	}

	if os.Getenv("GOOS") == "" {
		Target = runtime.GOOS
	} else {
		Target = os.Getenv("GOOS")
	}

	RaypmPath = LocalRaypmPath
	lockPath = path.Join(RaypmPath, "lock")
	PathToPkgs = path.Join(RaypmPath, "pkgs")
	dbJson = path.Join(RaypmPath, "db.json")

	switch ProgramTask {
	case SyncPkgs:
		enableRaypmAccess(RaypmPath, true)
		defer enableRaypmAccess(RaypmPath, false)

		log.Infoln("Synchronization")
		var (
			pathToArchive string
			version       string
		)

		raypmPkgs := path.Join(RaypmPath, "pkgs")
		log.Debugln("Creating .raypm directory")
		if _, err = os.Stat(raypmPkgs); err != nil {
			if err = os.MkdirAll(raypmPkgs, 0754); err != nil {
				log.Error("Failed to create '%s': %s", raypmPkgs, err)
				return
			}
			log.Debugln("Directory created")
		} else {
			log.Debugln("Directory already exists")
		}

		if pathToArchive, version, err = fetch.Sync(RaypmPath); err != nil {
			log.Errorln("Failed to sync:", err)
			return
		}

		if pathToArchive == "" {
			log.Infoln("There is nothing to do")
			return
		}

		log.Infoln("Unpacking sources")
		if err = unpack.Unpack("zip", pathToArchive, ".raypm", nil); err != nil {
			log.Errorln("Failed to unpack", err)
			return
		}

		fInfoPath := path.Join(".raypm", "pkgs", "info.txt")
		fInfo, err := os.Create(fInfoPath)
		if err != nil {
			log.Error("Failed to create info file '%s': '%s'", fInfoPath, err)
			return
		}
		defer fInfo.Close()

		if _, err = fInfo.WriteString(version); err != nil {
			log.Errorln("Failed to write a date:", err)
			return
		}

		log.Infoln("Package's database is up to date now")
	case Clean:
		enableRaypmAccess(RaypmPath, true)
		defer enableRaypmAccess(RaypmPath, false)

		dirToDel := ".raypm"

		switch cleanStorage {
		case "all":
		case "cache":
			dirToDel = path.Join(dirToDel, "cache")
		default:
			log.Error("Undefined option '%s', run 'raypm -h' for more info",
				cleanStorage)
			return
		}

		if _, err = os.Stat(dirToDel); err == nil {
			if err = os.RemoveAll(dirToDel); err != nil {
				log.Error("Failed to remove '%s': %s", dirToDel, err)
				return
			}
			log.Info("Deleted directory '%s'", dirToDel)
		} else {
			log.Infoln("Directory already deleted")
		}

	case InstallPkg, RemovePkg:
		enableRaypmAccess(RaypmPath, true)
		defer enableRaypmAccess(RaypmPath, false)

		var (
			fileLock *os.File
			deps     *deptree.Tree
			db       *dbpkg.PkgDb
		)

		if _, err = os.Stat(lockPath); err == nil {
			log.Error("Another process is using '%s', exiting", lockPath)
			return
		} else {
			if fileLock, err = os.Create(lockPath); err != nil {
				log.Error("Cannot create lock file: %s\n", err)
				return
			}
			log.Debugln("Created lock file")
			fileLock.Chmod(0000)
			log.Debugln("Change mod to 0000")
			fileLock.Close()
			log.Debugln("Closed file")
		}
		defer func() {
			if err = os.Chmod(lockPath, 0754); err != nil {
				log.Error("Cannot change mod for lock file:%s\n", err)
				return
			}
			log.Debugln("Changed mod to 0754. Deleting file...")

			if err = os.Remove(lockPath); err != nil {
				log.Error("Cannot delete file:%s\n", err)
				return
			}

			log.Debugln("Deleted lock file")
		}()

		if ProgramTask == InstallPkg {
			if _, err = os.Stat(dbJson); err != nil {
				db = dbpkg.NewDb(dbJson)
			} else {
				db, err = dbpkg.Open(dbJson)
				if err != nil {
					return
				}
			}
			defer db.WriteData()

			if deps, err = deptree.NewDepTree(RaypmPath, SelectedPackage, Target, db); err != nil {
				log.Error("Failed to resolve dependencies:\n%s\n", err)
				return
			} else {
				deps.Install()
			}
		} else if ProgramTask == RemovePkg {
			if _, err = os.Stat(dbJson); err != nil {
				log.Errorln("The local database not found, perhaps no packages were installed")
				return
			}

			db, err := dbpkg.Open(dbJson)
			if err != nil {
				log.Errorln(err)
				return
			}
			defer db.WriteData()

			if deps, err = deptree.NewDepTree(RaypmPath, SelectedPackage, Target, db); err != nil {
				log.Error("Failed to resolve dependencies:\n%s\n", err)
				return
			} else {
				deps.Uninstall()
			}
		}
	case ListPackages:
		var (
			dirs []os.DirEntry
		)

		log.Debug("Going to %s\n", PathToPkgs)
		if err = os.Chdir(PathToPkgs); err != nil {
			log.Error("Failed to open '%s':\n%s\n", PathToPkgs, err)
			return
		}

		defer func() {
			os.Chdir(path.Join("..", "..", ".."))
			log.Debugln("Change directory back")
		}()

		if dirs, err = os.ReadDir("."); err != nil {
			log.Error("Failed to read '%s':\n%s\n", PathToPkgs, err)
			return
		}

		log.Debug("Readed:\n%v\n", dirs)

		for _, item := range dirs {
			if item.IsDir() {
				currentPackage, err := pkginfo.NewPackageItem(
					item.Name(),
					Target,
				)

				if err != nil {
					log.Errorln(err)
					return
				}

				printLine := color.MagentaString(currentPackage.Name)

				pth := path.Join(RaypmPath, "store", currentPackage.Name)
				pth = os.Getenv("PWD") + string(os.PathSeparator) + pth

				if _, err = os.Stat(pth); err == nil {
					printLine += color.GreenString("\t[Installed]")
				}
				err = nil
				fmt.Print(printLine, "\n\t", currentPackage.Description, "\n")
			}
		}
	case FetchPkgInfo:
		var currentPackage *pkginfo.Package

		if err = os.Chdir(path.Join(PathToPkgs, SelectedPackage)); err != nil {
			log.Error("Package '%s' does not exists", SelectedPackage)
			return
		}

		defer func() {
			os.Chdir(path.Join("..", "..", "..", ".."))
		}()

		if currentPackage, err = pkginfo.NewPackageItem(".", Target); err != nil {
			log.Errorln(err)
		}

		fmt.Println("Package Information:")
		currentPackage.Info()
	}
}

func enableRaypmAccess(raypmPath string, enable bool) {
	var bits fs.FileMode = 0555

	if enable {
		bits = 0754
	}

	log.Debugln("Chaging to", bits)

	recRaypmAccess(raypmPath, bits)
	return
}

func recRaypmAccess(item string, bits fs.FileMode) {
	os.Chmod(item, bits)

	if item == "cache" {
		return
	}

	dirs, _ := os.ReadDir(item)

	for _, entry := range dirs {
		recRaypmAccess(path.Join(item, entry.Name()), bits)
	}

	return
}
