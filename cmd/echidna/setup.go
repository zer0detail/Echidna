package echidna

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gookit/color"
)

func createEchidnaDirs() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not get current working directory in CreateDirs() with error\n%s", err)
	}

	current := filepath.Join(dir, "current")
	err = os.MkdirAll(current, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Could not create directory %s in createdirs() with error\n%s", current, err)
	}
	inspect := filepath.Join(dir, "inspect")
	err = os.MkdirAll(inspect, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Could not create directory %s in createdirs() with error\n%s", inspect, err)
	}

	return nil

}

func deleteCurrentDir() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get current working directory in deleteCurrentDir(). Please manually remove the folder")
	}

	current := filepath.Join(dir, "current")
	err = os.RemoveAll(current)
	if err != nil {
		log.Fatal("Failed to remove Current/ directory and some subfiles. Please remove them manually")
	}
}

func greeting() {
	color.Yellow.Println(`
	   _     _     _             
          | |   (_)   | |            
  ___  ___| |__  _  __| |_ __   __ _ 
 / _ \/ __| '_ \| |/ _' | '_ \ / _' |
|  __/ (__| | | | | (_| | | | | (_| |
 \___|\___|_| |_|_|\__,_|_| |_|\__,_|`)

	color.LightBlue.Println("Echidna Scanner running. Browse to http://127.0.0.1:8080 to view status.")
}
