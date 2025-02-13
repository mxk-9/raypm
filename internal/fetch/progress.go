package fetch

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

type PassThru struct {
	io.Reader
	Total int
	All   int
}

func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)

	if err == nil {
		pt.Total += n

		total, measure := getFormattedData(pt.Total)

		for i := 0; i < 30; i++ {
			fmt.Printf(" ")
		}

		fmt.Print("\r\033[1A\033[K")
		fmt.Printf("Downloading: %2.2f %s\n", total, measure)
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
