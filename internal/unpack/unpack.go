package unpack

/* TODO:
 * 1. [X] Check if destination already exists
 * 2. [X] How destiantion's name will calculate
 */

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	log "raypm/pkg/slog"
	"strings"

	"github.com/bodgit/sevenzip"
)

type fileInArhive interface {
	Open() (io.ReadCloser, error)
}

type archive interface {
	Close() error
}

var (
	forArchZip = &zip.ReadCloser{}
	forArch7z  = &sevenzip.ReadCloser{}

	forZip = &zip.File{}
	for7z  = &sevenzip.File{}
)

var (
	zipArchType      = reflect.TypeOf(forArchZip)
	sevenzipArchType = reflect.TypeOf(forArch7z)

	zipType      = reflect.TypeOf(forZip)
	sevenzipType = reflect.TypeOf(for7z)
)

func Unpack(archType string, archSrc, dest []string, selectedItems []string) (err error) {
	var (
		r        archive
		arch     string = path.Join(archSrc...)
	)

	switch archType {
	case "7z":
		r, err = sevenzip.OpenReader(arch)
	case "zip":
		r, err = zip.OpenReader(arch)
	default:
		err = fmt.Errorf("Type '%s' is not supported yet.\n", archType)
		return
	}

	if err != nil {
		err = fmt.Errorf("Failed to open archive '%s': %s\n", arch, err)
		return
	}
	defer r.Close()

	log.Debugln("Archive type is", reflect.TypeOf(r))

	switch reflect.TypeOf(r) {
	case zipArchType:
		log.Debugln("Archive is zip")
		for _, f := range r.(*zip.ReadCloser).File {
			if err = extractFile(f, dest, selectedItems); err != nil {
				return err
			}
		}
	case sevenzipArchType:
		log.Debugln("Archive is 7z")
		for _, f := range r.(*sevenzip.ReadCloser).File {
			if err = extractFile(f, dest, selectedItems); err != nil {
				return err
			}
		}
	default:
		err = fmt.Errorf("Something goes wrong: %v\n", reflect.TypeOf(r))
	}

	return err
}

func extractFile(file fileInArhive, dest []string, selectedItems []string) (err error) {
	var (
		fileName  string
		isDir     bool
		itemFound bool = false
	)

	switch reflect.TypeOf(file) {
	case zipType:
		fileName = file.(*zip.File).Name
		isDir = file.(*zip.File).FileInfo().IsDir()
	case sevenzipType:
		fileName = file.(*sevenzip.File).Name
		isDir = file.(*sevenzip.File).FileInfo().IsDir()
	}

	pth := path.Join(dest...)

	for i := 0; i < len(selectedItems) && !itemFound; i++ {
		item := selectedItems[i]

		if strings.HasPrefix(fileName, item) {
			log.Debug("Found item: '%s'", item)
			itemFound = true
			depth := len(selectedItems) - 1
			if !isDir {
				depth--
			}
			endOfPath := strings.Split(fileName, "/")[depth:]
			log.Debug("Second part of path is %v", endOfPath)

			// Maybe it looks like shit
			pth = path.Join(pth, path.Join(endOfPath...))
		}
	}

	if selectedItems == nil {
		endOfPath := strings.Split(fileName, "/")
		pth = path.Join(pth, path.Join(endOfPath...))
	}

	log.Debug("Final path is '%s'", pth)

	if selectedItems != nil && !itemFound {
		log.Debug("Skipping '%s', because it doesn't match with:\n'%#v'", fileName, selectedItems)
		return
	}

	if _, err = os.Stat(pth); err == nil {
		err = fmt.Errorf(
			"File '%s' already exists, seems archive is already unpacked", pth,
		)
		return
	}

	rc, err := file.Open()

	if err != nil {
		return
	}
	defer rc.Close()

	log.Debug("Opened '%s'", fileName)

	if isDir {
		log.Debugln("Item is a directory")
		if err = os.MkdirAll(pth, 0754); err != nil {
			return
		}
		log.Debugln(pth, "created")
	} else {
		log.Debugln("Item is a file")
		baseDest := path.Dir(pth)

		if _, err = os.Stat(baseDest); err != nil {
			if err = os.MkdirAll(baseDest, 0754); err != nil {
				return
			}
		}

		dstFile, err := os.Create(pth)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		log.Debugln("File", pth, "created")

		if _, err = io.Copy(dstFile, rc); err != nil {
			return err
		}
	}

	return
}
