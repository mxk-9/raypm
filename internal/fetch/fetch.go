package fetch

import (
	"fmt"
	"io"
	"raypm/pkg/progress"
	"net/http"
	"os"
	"path/filepath"
	log "raypm/pkg/slog"
)

func GetFile(link, destPath string) (err error) {
	downloader := http.DefaultClient

	if _, err = os.Stat(destPath); err == nil {
		log.Warn("File '%s' exists, skip downloading\n", destPath)
		err = nil
		return
	}

	var resp *http.Response
	if resp, err = downloader.Get(link); err != nil {
		return
	}
	defer resp.Body.Close()

	src := &progress.PassThru{Reader: resp.Body}

	currDir := filepath.Dir(destPath)

	if currDir != "" && currDir != "." && currDir != "."+string(os.PathSeparator) {
		err = os.MkdirAll(currDir, 0754)
		if err != nil {
			err = fmt.Errorf("Failed to create a folder:\n%s\n", err)
			log.Errorln(err)
			return
		}
	}

	out, err := os.Create(destPath)
	if err != nil {
		return
	}
	defer out.Close()

	fmt.Println()
	if _, err = io.Copy(out, src); err != nil {
		err = fmt.Errorf("Failed to download a file:\n%s\n", err)
	}
	fmt.Println()

	return
}

