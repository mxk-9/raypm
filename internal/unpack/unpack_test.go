package unpack

import (
	"os"
	"path"
	log "raypm/pkg/slog"
	"testing"
)

func TestUnpack(t *testing.T) {
	simple := path.Join("testdata", "recursive_unpack.zip")

	log.Init(false)

	t.Run("recursive unpack", func(t *testing.T) {
		testZip(t, simple)
	})

}

func TestUnpackWithItems(t *testing.T) {
	selectedItems := path.Join("testdata", "selected_items.zip")
	items := []string{
		"begin/end",
	}

	log.Init(true)

	t.Run("unpack with selected items", func(t *testing.T) {
		testZipWithItems(t, selectedItems, items)
	})

	if err := os.RemoveAll(path.Join(
		os.TempDir(),
		"testunpack",
		"selected_items",
	)); err == nil {
		log.Infoln("Deleted all files for next tet")
	} else {
		log.Errorln("Failed to delete:", err)
	}
}

func testZip(t *testing.T, archFilePath string) {
	out := path.Join(os.TempDir(), "testunpack", "simple")
	err := Unpack("zip", []string{archFilePath}, []string{out}, nil)

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
	wantDir := path.Join(out, "begin", "end")
	err := Unpack("zip", []string{archFilePath}, []string{out}, selectedItems)

	if err != nil {
		t.Errorf("Failed to unpack: %v\n", err)
	}

	wantFiles := []string{
		wantDir,
		path.Join(wantDir, "high1"),
		path.Join(wantDir, "high2"),
		path.Join(wantDir, "high3"),
		path.Join(wantDir, "high4"),
		path.Join(wantDir, "one_file"),
		path.Join(wantDir, "internal", "inside_one"),
		path.Join(wantDir, "internal", "inside_two"),
		path.Join(wantDir, "internal", "inside_three"),
	}

	checkForFiles(t, wantFiles)
}

func checkForFiles(t *testing.T, fileList []string) {
	for _, item := range fileList {
		if _, err := os.Stat(item); err != nil {
			t.Errorf("Want '%s', but it does not exists", item)
		}
	}
}
