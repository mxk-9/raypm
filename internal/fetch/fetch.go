package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	log "raypm/pkg/slog"
)

type Downloader struct {
	Client *http.Client
}

func NewClient() (d *Downloader) {
	d = &Downloader{Client: &http.Client{}}
	return
}

func (d *Downloader) GetFile(link, destPath string) (err error) {
	if _, err = os.Stat(destPath); err == nil {
		log.Warn("File '%s' exists, skip downloading\n", destPath)
		err = nil
		return
	}

	var resp *http.Response
	if resp, err = d.Client.Get(link); err != nil {
		return
	}
	defer resp.Body.Close()

	src := &PassThru{Reader: resp.Body}

	currDir := filepath.Dir(destPath)

	if currDir != "" && currDir != "." && currDir != "."+string(os.PathSeparator) {
		err = os.MkdirAll(currDir, 0754)
		if err != nil {
			err = fmt.Errorf("Failed to create a folder:\n%s\n", err)
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

