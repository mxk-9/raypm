package progress

import (
	"fmt"
	"io"
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
}

// Add settings for copying output:
// Enable/Disable showing progress
// Count filesize
// Print file name
func NewProgress() (pt *PassThru) {

	return
}

func (pt *PassThru) SetText() {

}

func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)

	if err == nil {
		pt.Total += n

		total, measure := getFormattedData(pt.Total)

		for range 30 {
			fmt.Printf(" ")
		}

		fmt.Printf(ClearLine)
		fmt.Printf("\r")
		fmt.Printf("Downloading: %2.2f %s", total, measure)
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
