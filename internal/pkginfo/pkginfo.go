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

type TaskType string

type CommandTask struct {
	Command  TaskType `json:"command"`
	ExecBase []string `json:"exec_base"`
	Args     []string `json:"args"`
	From     []string `json:"from"`
	To       []string `json:"to"`
	Path     []string `json:"path"`
}

const (
	Exec               TaskType = "exec"
	Mkdir              TaskType = "mkdir"
	Copy               TaskType = "copy"
	Remove             TaskType = "rm"
	CallPackageManager TaskType = "pkgman"
)

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

func phasesInfo(data *string, phase ...any) {
	if len(phase) != 0 {
		for _, item := range phase {
			*data = fmt.Sprintf("%s\t%v\n", *data, item)
		}
	}
}

func (internal *Package) Info() (data string) {
	data = fmt.Sprintf("Name: %s\n", internal.Name)
	data += fmt.Sprintf("Description: %s\n", internal.Description)
	data += fmt.Sprintf("Depends on: %v\n", internal.Dependendencies)

	data += fmt.Sprintln("Fetching:")
	phasesInfo(&data, internal.FetchPhase)

	data += fmt.Sprintln("Unpacking:")
	phasesInfo(&data, internal.UnpackPhase)

	data += fmt.Sprintln("Installation:")
	phasesInfo(&data, internal.InstallPhase)

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

	if len(phases.InstallPhase) > 0 {
		log.Debugln("Install phase found")
		internal.InstallPhase = append(internal.InstallPhase, phases.InstallPhase...)
	}

	return
}
