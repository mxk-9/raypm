package progress

import (
	"fmt"
	"io"
	"os"
	log "raypm/pkg/slog"
)

const (
	B uint8 = iota
	Kb
	Mb
	Gb
)

const SI float32 = 1000

const ClearLine = "\033[2K"

type PassThru struct {
	io.Reader
	Total int
	All   int
	Prefix string
	Show bool
	FileSize int64
}

// Add settings for copying output:
// Enable/Disable showing progress
// Count filesize
// Print file name
func NewProgress(show bool, prefix string, reader io.Reader) (pt *PassThru) {
	pt = &PassThru{
		Reader: reader,
		Prefix: prefix,
		Show: show,
	}

	return
}

func (pt *PassThru) CountFileSize(f *os.File) (err error) {
	fInfo, err := f.Stat()
	if err != nil {
		log.Errorln("Could not read file data:", err)
		return
	}

	pt.FileSize = fInfo.Size()
	return
}

func (pt *PassThru) fmtPrint() {
	if !pt.Show {
		return
	}
	total, measure := getFormattedData(pt.Total)

	fmt.Printf(ClearLine)
	fmt.Printf("\r")

	fmt.Printf("%s: %2.2f %s", pt.Prefix, total, measure)

	if pt.FileSize != 0 {
		total, measure := getFormattedData(int(pt.FileSize))
		fmt.Printf("/%2.2f %s", total, measure)
	}
}

func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)

	if err == nil {
		pt.Total += n
		pt.fmtPrint()
	}

	return n, err
}

func getFormattedData(lengthInBytes int) (total float32, measure string) {
	totalM := B
	total = float32(lengthInBytes)

	for total > SI {
		total /= SI
		totalM++
	}

	switch totalM {
	case B:
		measure = "Bytes"
	case Kb:
		measure = "Kb"
	case Mb:
		measure = "Mb"
	case Gb:
		measure = "Gb"
	}

	return
}
