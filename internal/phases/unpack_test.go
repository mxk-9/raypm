package phases

import (
	"fmt"
	"os"
	"path"
	log "raypm/pkg/slog"
	"testing"
)

func TestUnpack(t *testing.T) {
	log.Init(false)

	deleteTmpFiles(path.Join("testunpack", "simple"))

	simple := path.Join("testdata", "recursive_unpack.zip")

	t.Run("recursive unpack", func(t *testing.T) {
		testZip(t, simple)
	})
}

func TestUnpackWithItems(t *testing.T) {
	log.Init(false)

	deleteTmpFiles(path.Join("testunpack", "selected_items"))

	selectedItems := path.Join("testdata", "selected_items.zip")
	items := []string{
		"begin/end",
	}

	t.Run("unpack with selected items", func(t *testing.T) {
		testZipWithItems(t, selectedItems, items)
	})

}

func testZip(t *testing.T, archFilePath string) {
	out := path.Join(os.TempDir(), "testunpack", "simple")
	err := Unpack("zip", archFilePath, out, nil)

	if err != nil {
		t.Errorf("Failed to unpack: %v\n", err)
	}

	wantFiles := []string{
		path.Join(out, "arch"),
		path.Join(out, "arch", "high"),
		path.Join(out, "arch", "first"),
		path.Join(out, "arch", "first", "second"),
		path.Join(out, "arch", "first", "second", "low.txt"),
	}

	checkForFiles(t, wantFiles)
}

func testZipWithItems(t *testing.T, archFilePath string, selectedItems []string) {
	out := path.Join(os.TempDir(), "testunpack", "selected_items")
	wantDir := path.Join(out, "end")
	err := Unpack("zip", archFilePath, out, selectedItems)

	if err != nil {
		t.Errorf("Failed to unpack: %v\n", err)
	}

	wantFiles := []string{
		wantDir,
		path.Join(wantDir, "high_1"),
		path.Join(wantDir, "high_2"),
		path.Join(wantDir, "high_3"),
		path.Join(wantDir, "high_4"),
		path.Join(wantDir, "high_5"),
		path.Join(wantDir, "one_file"),
		path.Join(wantDir, "internal", "inside_one"),
		path.Join(wantDir, "internal", "inside_two"),
		path.Join(wantDir, "internal", "inside_three"),
	}

	treeDirView(out, 0)
	checkForFiles(t, wantFiles)
}

func checkForFiles(t *testing.T, fileList []string) {
	for _, item := range fileList {
		if _, err := os.Stat(item); err != nil {
			t.Errorf("'%s' error=%s", item, err)
		}
	}
}

func treeDirView(folder string, depth int) (err error) {
	dirItems, err := os.ReadDir(folder)
	if err != nil {
		log.Errorln("Error:", err)
		return
	}

	for _, item := range dirItems {
		for range depth {
			fmt.Print(" ")
		}

		fmt.Println(item.Name())
		if item.IsDir() {
			if err = treeDirView(path.Join(folder, item.Name()), depth+2); err != nil {
				return
			}
		}
	}

	return
}

func deleteTmpFiles(endOfPath string) {
	if err := os.RemoveAll(path.Join(os.TempDir(), endOfPath)); err == nil {
		log.Infoln("Deleted all files for next test")
	} else {
		log.Errorln("Failed to delete:", err)
	}
}
