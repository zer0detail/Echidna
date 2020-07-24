package echidna

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteCurrent(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("TestDeleteCurrent() failed to get working directory with error\n%s", err)
	}

	current := filepath.Join(dir, "current")
	err = os.MkdirAll(current, os.ModePerm)
	if err != nil {
		t.Errorf("TestDeleteCurrent() failed to create current/ directory with error\n%s", err)
	}

	err = deleteCurrentDir()
	if err != nil {
		t.Errorf("deleteCurrentDir() failed to delete current/ directory with error\n%s", err)
	}
}
