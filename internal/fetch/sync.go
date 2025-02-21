package fetch

import (
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
func Sync() (pathToArchive string, err error) {
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

	log.Debug("Latest version: %s\nLink: %s", latestTagName, latestLink)
	log.Debugln("Creating cache directory")
	cache := path.Join(".raypm", "cache")

	if _, err = os.Stat(cache); err != nil {
		if err = os.MkdirAll(cache, 0754); err != nil {
			err = fmt.Errorf("Failed to create '%s': '%s'", cache, err)
		}
		log.Debugln("Created")
	} else {
		log.Debugln("Directory already created")
	}

	pathToArchive = path.Join(cache, latestTagName + ".zip")

	if _, err = os.Stat(pathToArchive); err == nil {
		log.Infoln("Packages database is up to date")
		pathToArchive = ""
		return
	}

	err = GetFile(latestLink, pathToArchive)

	return
}

