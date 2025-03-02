package progress

import (
	"fmt"
	"io"
	"os"
	"path"
	"testing"
	"time"
)

func TestPrintingLines(t *testing.T) {
	t.Run("printing 'Hello', than 'Bye'", func(t *testing.T) {
		fmt.Printf("Hello (wait 2 secs)")
		time.Sleep(time.Second * 2)
		fmt.Printf(ClearLine)
		fmt.Printf("\r")
		fmt.Printf("Bye\n")
	})
}

func TestCopyingProgress(t *testing.T) {
	testPath :=path.Join(os.TempDir(), "progress_test")
	if _, err := os.Stat(testPath); err == nil {
		os.RemoveAll(testPath)
	}
	
	t.Run("copying progress", func(t *testing.T) {
		fName := path.Join("testdata", "emptydata")
		f, err := os.Open(fName)
		if err != nil {
			t.Error(err)
		}
		defer f.Close()

		src := NewProgress(true, "Copying "+f.Name(), f)
		src.CountFileSize(f)
		to := path.Join(testPath, "emptytest")

		if err = os.MkdirAll(testPath, 0754); err != nil {
			t.Error(err)
		}

		out, err := os.Create(to)
		if err != nil {
			t.Error(err)
		}
		defer out.Close()

		if _, err = io.Copy(out, src); err != nil {
			t.Error(err)
		}

		fmt.Println()
	})
}
