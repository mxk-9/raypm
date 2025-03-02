package pkginfo

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	log "raypm/pkg/slog"
)

type FetchPath struct {
	From string   `json:"from"`
	To   []string `json:"to"`
}

type UnpackTask struct {
	Type          string   `json:"type"`
	Src           []string `json:"src"`
	Dest          []string `json:"dest"`
	SelectedItems []string `json:"selected_items"` // Full path to needed object, extacting recursively
}

type CommandTask struct {
	Command  string `json:"command"`
	ExecBase []string `json:"exec_base"`
	Args     []string `json:"args"`
	From     []string `json:"from"`
	To       []string `json:"to"`
	Path     []string `json:"path"`
}

type Package struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Version         string   `json:"version"`
	Dependendencies []string `json:"dependencies"`
	Systems         []string `json:"systems"`

	FetchPhase     []FetchPath   `json:"fetch_phase"`
	UnpackPhase    []UnpackTask  `json:"unpack_phase"`
	BuildPhase     []CommandTask `json:"build_phase"`
	InstallPhase   []CommandTask `json:"install_phase"`
	UninstallPhase []CommandTask `json:"uninstall_phase"`
}

// pathToPackage is a directory where package stores
// target is for what OS package should be builded
func NewPackageItem(pathToPackage, target string) (internal *Package, err error) {
	var (
		filePackage         *os.File
		targetPackage       *os.File
		pathToTargetPackage string = path.Join(pathToPackage, "package_"+target+".json")
	)

	internal = &Package{}

	if filePackage, err = os.Open(path.Join(pathToPackage, "package.json")); err != nil {
		err = fmt.Errorf("Failed to open package's file:\n%s\n", err)
		return
	}
	defer filePackage.Close()

	if err = json.NewDecoder(filePackage).Decode(internal); err != nil {
		err = fmt.Errorf("Failed to decode package.json:\n%s\n", err)
		return
	}

	if _, err = os.Stat(pathToTargetPackage); err == nil && target != "" {
		log.Debugln("OS-depend package found:", pathToTargetPackage)
		phases := &Package{}

		if targetPackage, err = os.Open(pathToTargetPackage); err != nil {
			err = fmt.Errorf("Failed to open '%s':\n%s\n", pathToTargetPackage, err)
			return
		}
		defer targetPackage.Close()

		if err = json.NewDecoder(targetPackage).Decode(phases); err != nil {
			err = fmt.Errorf("Failed to decode '%s':\n%s\n", pathToTargetPackage, err)
			return
		}

		internal.concatPkgPhases(phases)
		log.Debug(
			"F: %v\nU: %v\nI: %v\n",
			internal.FetchPhase, internal.UnpackPhase, internal.InstallPhase,
		)
	} else if err != nil {
		err = nil
	}

	return
}

func (internal *Package) Info() {
	fmt.Printf("Name: %s\n", internal.Name)
	fmt.Printf("Description: %s\n", internal.Description)
	if internal.Dependendencies != nil && len(internal.Dependendencies) > 0 {
		fmt.Printf("Depends on:\n")
		for _, item := range internal.Dependendencies {
			fmt.Printf("\t+ %s\n", item)
		}
	}

	return
}

func (internal *Package) concatPkgPhases(phases *Package) {
	if len(phases.FetchPhase) > 0 {
		log.Debugln("Fetch phase found")
		internal.FetchPhase = append(internal.FetchPhase, phases.FetchPhase...)
	}

	if len(phases.UnpackPhase) > 0 {
		log.Debugln("Unpack phase found")
		internal.UnpackPhase = append(internal.UnpackPhase, phases.UnpackPhase...)
	}

	if len(phases.BuildPhase) > 0 {
		log.Debugln("Build phase found")
		internal.BuildPhase = append(internal.BuildPhase, phases.BuildPhase...)
	}

	if len(phases.InstallPhase) > 0 {
		log.Debugln("Install phase found")
		internal.InstallPhase = append(internal.InstallPhase, phases.InstallPhase...)
	}

	return
}
