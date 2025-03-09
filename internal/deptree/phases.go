package deptree

import (
	"path"
	"raypm/internal/fetch"
	"raypm/internal/pkginfo"
	"raypm/internal/task"
	"raypm/internal/unpack"
	"raypm/internal/vars"
	log "raypm/pkg/slog"
	"strings"
)

func fetchPhase(data *[]pkginfo.FetchPath, vv *vars.Vars) (err error) {
	for _, item := range *data {
		log.Info("Getting '%s'", item.From)
		to := ""

		expanded := vv.ExpandVars(&item.To)

		for _, dest := range expanded {
			to = path.Join(to, dest)
		}

		if err = fetch.GetFile(item.From, to); err != nil {
			log.Errorln("Fetch phase failed:")
			log.Error("Failed to get '%s': %s", item.From, err)
			break
		}
	}

	return
}

func unpackPhase(data *[]pkginfo.UnpackTask, vv *vars.Vars) (err error) {
	for _, item := range *data {
		from := path.Join(vv.ExpandVars(&item.Src)...)
		log.Debug("Expanded from '%v'\nto '%v'", item.Src, from)

		log.Info("Unpacking '%s'", from)

		to := path.Join(vv.ExpandVars(&item.Dest)...)
		log.Debug("Expanded from '%v'\nto '%v'", item.Dest, to)

		err = unpack.Unpack(item.Type, from, to, item.SelectedItems)

		if err != nil {
			log.Errorln("Unpack phase failed:")
			log.Error("Failed to unpack '%s': '%s", from, err)
			break
		}
	}
	return
}

func taskPhase(phase string, data *[]pkginfo.CommandTask, vv *vars.Vars) (err error) {

	if !(phase == "build" || phase == "install" || phase == "uninstall") {
		log.Error("Uknown phase '%s'", phase)
		return
	}

	for _, item := range *data {
		if err = task.Do(phase, item, vv); err != nil {
			log.Error("%s phase failed:", strings.ToTitle(phase))
			log.Errorln(err)
			break
		}
	}
	return
}
