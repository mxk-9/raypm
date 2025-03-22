package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"raypm/internal/app"
	"raypm/internal/dbpkg"
	"raypm/internal/deptree"
	"raypm/internal/fetch"
	"raypm/internal/pkginfo"
	"raypm/internal/unpack"
	log "raypm/pkg/slog"
	"runtime"

	"github.com/fatih/color"
)

func main() {
	var (
		ProgramTask     app.Operation = 0
		SelectedPackage string
		settings        *app.Settings

		err error

		// Flags

		operations int = 0
	)

	opts, err := app.NewOptions()
	_ = opts
	// Handle errors

	log.Init(opts.Debug)

	if listPkgs {
		ProgramTask = ListPackages
		operations++
	}

	if fetchPkgInfo != "" {
		ProgramTask = FetchPkgInfo
		SelectedPackage = fetchPkgInfo
		operations++
	}

	if installPkg != "" {
		ProgramTask = InstallPkg
		SelectedPackage = installPkg
		operations++
	}

	if syncPkgs {
		ProgramTask = SyncPkgs
		operations++
	}

	if removePkg != "" {
		ProgramTask = RemovePkg
		SelectedPackage = removePkg
		operations++
	}

	if cleanStorage != "" {
		ProgramTask = Clean
		operations++
	}

	if buildPackage {
		ProgramTask = BuildPkg
		operations++
	}

	if operations > 1 {
		log.Fatalln("Choose only one operation. Type 'raypm -h' to see them")
	}

	if buildPackage {
		settings, err = app.InitApp(".raypm", packageTarget)
	} else {
		var tmpStr string
		if tmpStr, err = os.UserHomeDir(); err != nil {
			log.Errorln(err)
			return
		}

		if runtime.GOOS == "windows" {
			tmpStr = path.Join(tmpStr, "AppData", "Local", "Raypm")
		} else {
			tmpStr = path.Join(tmpStr, ".raypm")
		}

		settings, err = app.InitApp(tmpStr, packageTarget)
	}

	if err != nil {
		log.Errorln(err)
		return
	}

	switch ProgramTask {
	case SyncPkgs:
		settings.EnableAccess()
		defer settings.DisableAccess()

		log.Infoln("Synchronization")
		var (
			pathToArchive string
			version       string
		)

		log.Debugln("Creating .raypm directory")
		if _, err = os.Stat(settings.PathToPkgs); err != nil {
			if err = os.MkdirAll(settings.PathToPkgs, 0754); err != nil {
				log.Error("Failed to create '%s': %s", settings.PathToPkgs, err)
				return
			}
			log.Debugln("Directory created")
		} else {
			log.Debugln("Directory already exists")
		}

		if pathToArchive, version, err = fetch.Sync(settings.RaypmPath); err != nil {
			log.Errorln("Failed to sync:", err)
			return
		}

		if pathToArchive == "" {
			log.Infoln("There is nothing to do")
			return
		}

		log.Infoln("Unpacking sources")
		if err = unpack.Unpack("zip", pathToArchive, settings.RaypmPath, nil); err != nil {
			log.Errorln("Failed to unpack", err)
			return
		}

		fInfoPath := path.Join(settings.PathToPkgs, "info.txt")
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
		settings.EnableAccess()
		defer settings.DisableAccess()

		dirToDel := settings.RaypmPath

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
		settings.EnableAccess()
		defer settings.DisableAccess()

		var (
			fileLock *os.File
			deps     *deptree.Tree
			db       *dbpkg.PkgDb
		)

		if _, err = os.Stat(settings.LockPath); err == nil {
			log.Error("Another process is using '%s', exiting", settings.LockPath)
			return
		} else {
			if fileLock, err = os.Create(settings.LockPath); err != nil {
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
			if err = os.Chmod(settings.LockPath, 0754); err != nil {
				log.Error("Cannot change mod for lock file:%s\n", err)
				return
			}
			log.Debugln("Changed mod to 0754. Deleting file...")

			if err = os.Remove(settings.LockPath); err != nil {
				log.Error("Cannot delete file:%s\n", err)
				return
			}

			log.Debugln("Deleted lock file")
		}()

		if ProgramTask == InstallPkg {
			if _, err = os.Stat(settings.DbJson); err != nil {
				db = dbpkg.NewDb(settings.DbJson)
			} else {
				db, err = dbpkg.Open(settings.DbJson)
				if err != nil {
					return
				}
			}
			defer db.WriteData()

			if deps, err = deptree.NewDepTree(settings.RaypmPath, SelectedPackage, settings.Build.Target, db); err != nil {
				log.Error("Failed to resolve dependencies:\n%s\n", err)
				return
			} else {
				deps.Install()
			}
		} else if ProgramTask == RemovePkg {
			if _, err = os.Stat(settings.DbJson); err != nil {
				log.Errorln("The local database not found, perhaps no packages were installed")
				return
			}

			db, err := dbpkg.Open(settings.DbJson)
			if err != nil {
				log.Errorln(err)
				return
			}
			defer db.WriteData()

			if deps, err = deptree.NewDepTree(settings.RaypmPath, SelectedPackage, settings.Build.Target, db); err != nil {
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

		log.Debug("Going to %s\n", settings.PathToPkgs)
		if err = os.Chdir(settings.PathToPkgs); err != nil {
			log.Error("Failed to open '%s':\n%s\n", settings.PathToPkgs, err)
			return
		}

		defer func() {
			os.Chdir(path.Join("..", "..", ".."))
			log.Debugln("Change directory back")
		}()

		if dirs, err = os.ReadDir("."); err != nil {
			log.Error("Failed to read '%s':\n%s\n", settings.PathToPkgs, err)
			return
		}

		log.Debug("Readed:\n%v\n", dirs)

		for _, item := range dirs {
			if item.IsDir() {
				currentPackage, err := pkginfo.NewPackageItem(
					item.Name(),
					settings.Build.Target,
				)

				if err != nil {
					log.Errorln(err)
					return
				}

				printLine := color.MagentaString(currentPackage.Name)

				pth := path.Join(settings.RaypmPath, "store", currentPackage.Name)
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

		if err = os.Chdir(path.Join(settings.PathToPkgs, SelectedPackage)); err != nil {
			log.Error("Package '%s' does not exists", SelectedPackage)
			return
		}

		defer func() {
			os.Chdir(path.Join("..", "..", "..", ".."))
		}()

		if currentPackage, err = pkginfo.NewPackageItem(".", settings.Build.Target); err != nil {
			log.Errorln(err)
		}

		fmt.Println("Package Information:")
		currentPackage.Info()
	}
}
