package task

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"raypm/internal/pkginfo"
	"raypm/internal/vars"
	"raypm/pkg/progress"
	log "raypm/pkg/slog"
	"strings"
)

const (
	Exec               string = "exec"
	Mkdir              string = "mkdir"
	Copy               string = "copy"
	CallPackageManager string = "pkgman"
	Overwrite          string = "overwrite"
)

func Do(phaseType string, task pkginfo.CommandTask, vv *vars.Vars) (err error) {
	cmd := task.Command
	if cmd == Exec {
		exe := vv.ExpandVars(&task.ExecBase)
		args := vv.ExpandVars(&task.Args)

		err = external_cmd(exe, args)
	} else if cmd == Mkdir {
		pth := vv.ExpandVars(&task.Path)
		err = mkdir(pth)
	} else if cmd == Copy || cmd == Overwrite{
		from := path.Join(vv.ExpandVars(&task.From)...)
		to := path.Join(vv.ExpandVars(&task.To)...)

		overwrite := cmd == Overwrite

		err = copyItemOverwrite(from, to, overwrite)
	} else if cmd == CallPackageManager {
		err = pkgman(vv.Package, phaseType)
	} else {
		log.Error("Unknown command '%s'", cmd)
		err = fmt.Errorf("UnknownCommand")
	}

	return
}

func external_cmd(exec_base []string, args []string) (err error) {
	exe := path.Join(exec_base...)

	cmd := exec.Command(exe, args...)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	return
}

func mkdir(pth []string) (err error) {
	dir := path.Join(pth...)

	if _, err = os.Stat(dir); err == nil {
		log.Error("Directory '%s' already exists!", dir)
		err = fmt.Errorf("DirectoryAlreadyExists")
		return
	}

	if err = os.MkdirAll(dir, 0754); err != nil {
		log.Error("Failed to create '%s': %s", dir, err)
		return
	}

	return
}

func copyItemOverwrite(from, to string, overwrite bool) (err error) {
	if _, err = os.Stat(to); err == nil && !overwrite {
		log.Error("Cannot copy '%s' to '%s': it already exists", from, to)
		err = fmt.Errorf("FileAlreadyExists")
		return
	}

	fInfo, err := os.Stat(from)
	if err != nil {
		log.Error("Failed to get info of '%s': %s", from, err)
		return
	}

	if !fInfo.IsDir() {
		var (
			inpFile *os.File
			outFile *os.File
		)
		inpFile, err = os.Open(from)
		if err != nil {
			log.Error("Failed to open '%s': %s", from, err)
			return
		}
		defer inpFile.Close()

		outFile, err = os.Create(to)
		if err != nil {
			log.Error("Failed to create '%s': %s", to, err)
			return
		}
		defer outFile.Close()

		src := &progress.PassThru{Reader: inpFile}

		if _, err = io.Copy(outFile, src); err != nil {
			log.Error("Failed to copy '%s': %s", from, err)
		}
	} else {
		var files []os.DirEntry
		if err = os.MkdirAll(to, 0754); err != nil {
			log.Error("Failed to create '%s': %s", to, err)
			return
		}

		if files, err = os.ReadDir(from); err == nil {
			for _, item := range files {
				nFrom := path.Join(from, item.Name())
				nTo := path.Join(to, item.Name())

				if err = copyItemOverwrite(nFrom, nTo, overwrite); err != nil {
					log.Error(
						"Failed to copy '%s' to '%s': %s",
						item.Name(), to, err,
					)
					return
				}
			}
		} else {
			log.Error("Failed to read '%s': %s", from, err)
		}
	}
	log.Debug("'%s' created", to)

	return
}

func pkgman(pkgFiles, phase string) (err error) {
	var (
		osRelease    *os.File
		packagesList *os.File
		distro       string
		pm           []string
	)

	if osRelease, err = os.Open("/etc/os-release"); err != nil {
		log.Errorln("Failed to read /etc/os-release:", err)
		return
	}
	defer osRelease.Close()

	scan := bufio.NewScanner(osRelease)

	for scan.Scan() {
		if strings.HasPrefix(scan.Text(), "ID=") {
			distro = scan.Text()[3:]
			log.Info("Linux distro is '%s'", distro)
			break
		}
	}

	switch distro {
	case "arch", "manjaro":
		pm = []string{"pacman"}
		if phase == "install" {
			pm = append(pm, "-Sy")
		} else if phase == "uninstall" {
			pm = append(pm, "-R")
		}

		pm = append(pm, "--noconfirm")
		pm = append(pm, "--needed")
	case "fedora":
		pm = []string{"dnf"}
		if phase == "install" {
			pm = append(pm, "install")
		} else if phase == "uninstall" {
			pm = append(pm, "remove")
		}
	case "ubuntu", "debian":
		pm = []string{"apt"}
		if phase == "install" {
			pm = append(pm, "install")
		} else if phase == "uninstall" {
			pm = append(pm, "purge")
		}
	case "void":
		if phase == "install" {
			pm = []string{"xbps-install"}
		} else if phase == "uninstall" {
			pm = []string{"xbps-remove"}
		}
	default:
		log.Error("Package manager for '%s' is not implemented yet", distro)
	}

	f := path.Join(pkgFiles, pm[0]) + ".txt"
	if packagesList, err = os.Open(f); err !=
		nil {
		log.Error("Failed to open '%s': %s", f, err)
		return
	}
	defer packagesList.Close()

	buf := bufio.NewScanner(packagesList)

	for buf.Scan() {
		pm = append(pm, buf.Text())
	}

	cmd := exec.Command("sudo", pm...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	return
}
