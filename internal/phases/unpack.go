package phases

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	log "raypm/pkg/slog"
	"reflect"
	"strings"

	"github.com/bodgit/sevenzip"
)

type fileInArchive interface {
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

func Unpack(archType, archSrc, dest string, selectedItems []string) (err error) {
	var (
		r archive
	)

	switch archType {
	case "7z":
		r, err = sevenzip.OpenReader(archSrc)
	case "zip":
		r, err = zip.OpenReader(archSrc)
	default:
		log.Error("Type '%s' is not supported yet.\n", archType)
		err = fmt.Errorf("ArchiveFormatIsNotSupported")
		return
	}

	if err != nil {
		log.Error("Failed to open archive '%s': %s\n", archSrc, err)
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
		log.Error("Something goes wrong: %v\n", reflect.TypeOf(r))
	}

	return err
}

// Main problem, that this function just copy files and not recreating all
// folders. For ex., file $fetch/bebra/touchme.c will copied as $src/touchme.c
// TODO: Dest must be just string
func extractFile(file fileInArchive, destDir string, selectedItems []string) (err error) {
	var (
		fileName   string
		isDir      bool
		itemFound  bool = false
		checkItems bool = selectedItems != nil || len(selectedItems) > 0
	)

	switch reflect.TypeOf(file) {
	case zipType:
		fileName = file.(*zip.File).Name
		isDir = file.(*zip.File).FileInfo().IsDir()
	case sevenzipType:
		fileName = file.(*sevenzip.File).Name
		isDir = file.(*sevenzip.File).FileInfo().IsDir()
	}

	log.Debug("Fname: %s; isDir: %t", fileName, isDir)

	destPath := destDir

	for i := 0; checkItems && i < len(selectedItems) && !itemFound; i++ {
		item := selectedItems[i]

		if strings.HasPrefix(fileName, item) {
			log.Debug("Found item: '%s'", item)
			itemFound = true
			splited := strings.Split(item, "/")
			depth := len(splited)
			depth--
			log.Debug("Depth is %d", depth)
			endOfPath := (strings.Split(fileName, "/"))[depth:]
			log.Debug("Second part of path is %v", endOfPath)

			// Maybe it looks like a shit
			log.Debug("Full %s %v", destPath, endOfPath)
			destPath = path.Join(destPath, path.Join(endOfPath...))
		}
	}

	if !checkItems {
		endOfPath := strings.Split(fileName, "/")
		destPath = path.Join(destPath, path.Join(endOfPath...))
	}

	log.Debug("Final path is '%s'", destPath)

	if checkItems && !itemFound {
		log.Debug("Skipping '%s', because it doesn't match with: '%#v'", fileName, selectedItems)
		return
	}

	if _, err = os.Stat(destPath); err == nil {
		log.Warn(
			"File '%s' already exists, seems archive is already unpacked", destPath,
		)
		return
	}

	if isDir {
		log.Debugln("Item is a directory")
		if err = os.MkdirAll(destPath, 0754); err != nil {
			return
		}
		log.Debugln(destPath, "created")
	} else {
		rc, lerr := file.Open()

		if lerr != nil {
			err = lerr
			return
		}
		defer rc.Close()
		log.Debug("Opened '%s'", fileName)

		log.Debugln("Item is a file")
		baseDest := path.Dir(destPath)

		if _, err = os.Stat(baseDest); err != nil {
			if err = os.MkdirAll(baseDest, 0754); err != nil {
				return
			}
		}

		dstFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		log.Debugln("File", destPath, "created")

		if _, err = io.Copy(dstFile, rc); err != nil {
			return err
		}
	}

	return
}
