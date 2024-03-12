package cabfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"project/rblog"
	"project/slave/iodevice"
	"runtime"
	"strconv"
)

var cabsMain string = ""
var cabsBackup string = ""
var allCabs []int = []int{1, 1, 1, 1}

func Read() []int {
	if cabsMain == "" {
		// Get the path of the currently running executable
		exeDir, err := getCurrentDirectory()
		if err != nil {
			rblog.Red.Println("Error getting current directory:", err)
			return allCabs
		}
		cabsMain = filepath.Join(exeDir, "cabsMain.txt")
		cabsBackup = filepath.Join(exeDir, "cabsBackup.txt")
	}
	// Read from primary file
	fileData, err := os.ReadFile(cabsMain)
	if err != nil || len(string(fileData)) != iodevice.N_FLOORS {
		rblog.Red.Println("Error reading from primary file:", err)
		// Try reading from backup file
		fileData, err = os.ReadFile(cabsBackup)
		if err != nil || len(string(fileData)) != iodevice.N_FLOORS {
			rblog.Red.Println("Error reading from backup file:", err)
			return allCabs
		}
	}

	cabs := string(fileData)
	result := make([]int, len(cabs))
	for i, char := range cabs {
		val, err := strconv.Atoi(string(char))
		if err != nil {
			rblog.Red.Println("Error converting string to int:", err)
			return allCabs
		}
		result[i] = val
	}

	return result
}

func writeToFiles(data []int) error {
	newString := ""
	for _, val := range data {
		newString += strconv.Itoa(val)
	}
	save := []byte(newString)

	err := os.WriteFile(cabsMain, save, 0644)
	if err != nil {
		return fmt.Errorf("error writing to primary file: %w", err)
	}

	err = os.WriteFile(cabsBackup, save, 0644)
	if err != nil {
		return fmt.Errorf("error writing to backup file: %w", err)
	}

	return nil
}

func Set(floor int) error {
	data := Read()
	data[floor] = 1
	if err := writeToFiles(data); err != nil {
		return err
	}
	return nil
}

func Clear(floor int) error {
	data := Read()
	data[floor] = 0
	if err := writeToFiles(data); err != nil {
		return err
	}
	return nil
}

func getCurrentDirectory() (string, error) {
	// Get the absolute path of the file containing this function
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("unable to get current directory")
	}
	// Get the directory containing the file
	packageDir := filepath.Dir(filename)
	return packageDir, nil
}
