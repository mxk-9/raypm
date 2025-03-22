package phases

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	log "raypm/pkg/slog"

	"github.com/google/go-github/v69/github"
)

const defaultDataBase string = "https://github.com/mxk-9/raypkgs/releases"

type AssetsInfo struct {
	DownloadUrl string `json:"browser_download_url"`
}

type ReleaseInfo struct {
	TagName string       `json:"tag_name"`
	Assets  []AssetsInfo `json:"assets"`
}

type Releases []ReleaseInfo

// It can returns an empty string, that means, package's database is already
// exists
func Sync(raypmPath string) (pathToArchive, version string, err error) {
	var (
		fInfo     *os.File
		pkgsPath  string = path.Join(raypmPath, "pkgs")
		fInfoPath string = path.Join(pkgsPath, "info.txt")
	)
	client := github.NewClient(nil)

	log.Debugln("Creating a request")

	req, err := client.NewRequest(
		"GET",
		"/repos/mxk-9/raypkgs/releases",
		nil,
	)

	if err != nil {
		return
	}

	log.Debugln("Making a request")
	rel := &Releases{}
	res, err := client.Do(context.Background(), req, rel)

	if err != nil {
		return
	}
	log.Debug("Got status: %v", res.Status)
	defer res.Body.Close()

	latestLink := (*rel)[0].Assets[0].DownloadUrl
	latestTagName := (*rel)[0].TagName

	log.Debugln("Latest version:", latestTagName)
	log.Debugln("Link:", latestLink)
	log.Debugln("Creating cache directory")
	cache := path.Join(raypmPath, "cache")

	if _, err = os.Stat(cache); err != nil {
		if err = os.MkdirAll(cache, 0754); err != nil {
			err = fmt.Errorf("Failed to create '%s': '%s'", cache, err)
		}
		log.Debugln("Created")
	} else {
		log.Debugln("Directory already created")
	}

	log.Debugln("Checking installed version")
	if _, err = os.Stat(fInfoPath); err == nil {
		if fInfo, err = os.Open(fInfoPath); err != nil {
			log.Error("Failed to open '%s': %s", fInfoPath, err)
			return
		}
		defer fInfo.Close()

		buf := bufio.NewScanner(fInfo)

		buf.Scan()
		ver := buf.Text()
		if ver != latestTagName {
			log.Warn(
				"Current pkgs version is '%s', new: '%s', removing old",
				ver, latestTagName,
			)

			if err = os.RemoveAll(pkgsPath); err != nil {
				log.Error("Failed to remove '%s': %s", pkgsPath, err)
				return
			}
			log.Info("'pkgs' removed")
		} else {
			log.Warn("Latest package database is already installed")
			return
		}

	} else {
		log.Debugln("Don't find")
	}

	pathToArchive = path.Join(cache, latestTagName+".zip")
	version = latestTagName

	err = GetFile(latestLink, pathToArchive)

	return
}
