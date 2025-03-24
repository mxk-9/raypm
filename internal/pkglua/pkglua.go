package pkglua

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	log "raypm/pkg/slog"
	"strings"

	"github.com/Shopify/go-lua"
)

//go:embed lib.lua
var lualib string

type Package struct {
	MData
	TargetSpec
}

// Contains:
//   - name
//   - version
//   - description
//   - src_path
//   - build_path
type MData map[string]string

// Contains:
//   - supported_systems
//   - dependencies
//   - pkgman_install
//   - pkgman_uninstall
//   - packages
//   - phases*
type TargetSpec map[string][]string

func NewPackage(pathToPackageFile, host, target string) (pd *Package, err error) {
	mdata := make(map[string]string)
	tspec := make(map[string][]string)
	ok := true
	phaseStr := ""
	l := lua.NewState()
	lua.OpenLibraries(l)

	if err = lua.DoString(l, lualib); err != nil {
		log.Errorln("Failed to execute internal lib in Lua:", err)
		return
	}

	l.Global("Get_Metadata")
	l.PushString(pathToPackageFile)
	l.Call(1, 5)
	tableInd := 1

	if l.IsNil(tableInd) {
		err = &LuaTableError{
			Err:   FieldIsNil,
			Value: showStack(l, "Get_Metadata:"),
		}
		return
	}

	l.Field(tableInd, "name")
	mdata["name"], ok = l.ToString(-1)
	if !ok {
		err = &LuaTableError{
			Err:   ParseStringFailed,
			Value: l.ToValue(-1),
		}
		log.Errorln("'name' field is required!")
		return
	}

	l.Field(tableInd, "version")
	mdata["version"], ok = l.ToString(-1)
	if !ok {
		err = &LuaTableError{
			Err:   ParseStringFailed,
			Value: l.ToValue(-1),
		}
		log.Errorln("'version' field is required!")
		return
	}

	l.Field(tableInd, "description")
	mdata["description"], _ = l.ToString(-1)

	l.Field(tableInd, "src_path")
	mdata["src_path"], _ = l.ToString(-1)

	l.Field(tableInd, "build_path")
	mdata["build_path"], _ = l.ToString(-1)

	l.Pop(l.Top() + 1)
	l.SetTop(0)

	l.Global("Get_Phases")
	l.PushString(pathToPackageFile)
	l.PushString(host)
	l.PushString(target)
	l.Call(3, 7)

	pkgArrSpecs := []string{
		"fetch_phase", "unpack_phase", "prepare_phase", "build_phase", "install_phase", "uninstall_phase",
	}

	for _, item := range pkgArrSpecs {
		l.Field(1, item)
		phaseStr, ok = l.ToString(l.Top())
		if ok {
			sStr := splitString(phaseStr)
			if len(sStr) > 0 {
				tspec[item] = sStr
			}
		}
	}

	// Parsing dependencies
	l.Field(1, "dependencies")
	if !l.IsNil(l.Top()) {
		ok = true
		depInd := l.Top()
		deps := make([]string, 0)

		for i := 1; ok; i++ {
			l.RawGetInt(depInd, i)
			ind := l.Top()

			if l.IsNil(ind) {
				break
			}
			phaseStr, ok = l.ToString(ind)
			if ok {
				deps = append(deps, phaseStr)
			}
		}

		tspec["dependencies"] = deps
	}

	l.Pop(l.Top() + 1)
	l.SetTop(0)

	pd = new(Package)
	pd.MData = mdata
	pd.TargetSpec = tspec

	if host == "linux" {
		var (
			osRelease *os.File
			distro    string
			pm        []string
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
				break
			}
		}

		l.Global("Get_Pkgman_Cmd")
		l.PushString(pathToPackageFile)
		l.PushString(distro)
		l.Call(3, 3)

		log.Debugln(showStack(l, "pkgs"))
		if l.IsNil(1) {
			log.Debugln("Packages is nil")
			return
		}

		ok := true
		str := ""
		for ok {
			l.RawGetInt(1, l.Top()-2)
			str, ok = l.ToString(l.Top())
			if ok {
				pm = append(pm, str)
			}
		}

		pd.TargetSpec["packages"] = pm
	}

	return
}

func splitString(phaseStr string) (splitted []string) {
	if phaseStr == "" {
		return
	}

	splitted = make([]string, 0)
	before := strings.Split(phaseStr, "\n")

	for _, item := range before {
		s := strings.Trim(item, "\r")
		s = strings.TrimSpace(item)

		if len(s) > 0 {
			splitted = append(splitted, s)
		}
	}
	return
}

func showStack(l *lua.State, prefix string) (output string) {
	output = ""
	for i := range l.Top() + 1 {
		output += fmt.Sprintln(prefix, l.ToValue(i))
	}

	return
}
