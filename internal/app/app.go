package app

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	log "raypm/pkg/slog"
	"runtime"
)

type Operation uint8

const (
	ListPackages Operation = iota
	SyncPkgs
	Clean
	FetchPkgInfo
	InstallPkg
	RemovePkg
	BuildPkg
)

type Settings struct {
	RaypmPath  string
	PathToPkgs string
	LockPath   string
	DbJson     string
	Build      Build
}

type Build struct {
	Host   string
	Target string
	Cross  bool
}

type Options struct {
	ListPkgs      bool
	Debug         bool
	SyncPkgs      bool
	BuildPackage  bool
	FetchPkgInfo  string
	InstallPkg    string
	RemovePkg     string
	CleanStorage  string
	PackageTarget string
	OutputPath    string
	CustomPkgs    string
}

func NewOptions() (o *Options, err error) {
	o = &Options{}

	flag.BoolVar(&o.ListPkgs, "list", false, "List all available packages")
	flag.BoolVar(&o.Debug, "d", false, "Print debug logs")
	flag.BoolVar(&o.SyncPkgs, "sync", false, "Get latest package's database")
	flag.BoolVar(&o.BuildPackage, "build", false, "Build a package")
	flag.StringVar(&o.FetchPkgInfo, "info", "", "Show information about package")
	flag.StringVar(&o.InstallPkg, "install", "", "Install a package")
	flag.StringVar(&o.RemovePkg, "remove", "", "Remove a package")
	flag.StringVar(&o.PackageTarget, "target", "", "Set target OS")
	flag.StringVar(&o.OutputPath, "o", "", "Set custom output path(for -build and -install)")
	flag.StringVar(&o.CustomPkgs, "pkgs", "", "Set custom pkgs path")
	flag.StringVar(&o.CleanStorage,
		"clean",
		"",
		"Cleaning raypm's storage. Available options: 'cache', 'all'",
	)
	flag.Parse()

	if flag.NFlag() == 0 {
		log.Warnln("There's nothing to do. Type 'raypm -h'")
		err = fmt.Errorf("NoOperations")
	}

	if _, err = os.Stat(o.CustomPkgs); err != nil {

	}

	return
}

func (o *Options) SetProgramTask() (programTask Operation, selectedPackage string, err error) {

	return
}

func InitApp(opts *Options, raypmPath, target string) (*Settings, error) {
	app := &Settings{
		RaypmPath:  raypmPath,
		PathToPkgs: path.Join(raypmPath, "pkgs"),
		LockPath:   path.Join(raypmPath, "lock"),
		DbJson:     path.Join(raypmPath, "db.json"),
	}

	if opts.CustomPkgs != "" {
		app.PathToPkgs = opts.CustomPkgs
	}

	if target == "linux" || target == "windows" || target == "android" {
		app.Build = Build{
			Target: target,
		}

		app.Build.Host = runtime.GOOS
		app.Build.Cross = app.Build.Target != app.Build.Host
	} else if target == "" {
		app.Build.Target = runtime.GOOS
		app.Build.Host = runtime.GOOS
		app.Build.Cross = false
	} else {
		log.Error("Undefined system '%s'", target)
		return nil, fmt.Errorf("UndefinedSystem")
	}

	return app, nil
}

// Need for experimenting with packages
func (s *Settings) SetPkgs(pkgsPath string) (err error) {
	if _, err = os.Stat(pkgsPath); err != nil {
		log.Errorln(err)
		return
	}

	s.PathToPkgs = pkgsPath
	return
}

func (s *Settings) EnableAccess() {
	var bits fs.FileMode = 0754

	log.Debugln("Chaging to", bits)

	recRaypmAccess(s.RaypmPath, bits)
}

func (s *Settings) DisableAccess() {
	var bits fs.FileMode = 0550

	log.Debugln("Chaging to", bits)

	recRaypmAccess(s.RaypmPath, bits)
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
