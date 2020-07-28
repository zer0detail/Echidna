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
		t.Errorf("deleteCurrentDir() returned an error\n%s", err)
	}

	if _, err := os.Stat(current); !os.IsNotExist(err) {
		t.Errorf("deleteCurrentDir() failed to remove the current/ directory, os.stat() returned error\n%s", err)
	}
}

func TestCreateEchidnaDirs(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("TestCreateEchidnaDirs(): Failed to get current working directory with error: %s", err)
	}
	inspect := filepath.Join(dir, "inspect")
	current := filepath.Join(dir, "current")
	err = os.RemoveAll(current)
	if err != nil {
		t.Errorf("TestCreateEchidnaDirs(): Failed to remove current/ directory with error: %s ", err)
	}

	err = os.RemoveAll(inspect)
	if err != nil {
		t.Errorf("TestCreateEchidnaDirs(): Failed to remove inspect/ directory with error: %s ", err)
	}

	err = createEchidnaDirs()
	if err != nil {
		t.Errorf("CreateEchidnaDirs failed test with error %s", err)
	}
}
