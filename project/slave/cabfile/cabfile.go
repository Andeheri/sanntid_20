package cabfile

import (
	"fmt"
	"os"
	"strconv"
)

var cabsMain string = "C:/Users/sunes/OneDrive - NTNU/NTNU/3.år/Sanntid/LAB_EXERCISE/sanntid_20/project/slave/cabfile/cabsMain.txt"
var cabsBackup string = "C:/Users/sunes/OneDrive - NTNU/NTNU/3.år/Sanntid/LAB_EXERCISE/sanntid_20/project/slave/cabfile/cabsBackup.txt"

func Read() []int {
	// Read from primary file
	fileData, err := os.ReadFile(cabsMain)
	if err != nil {
		fmt.Println("Error reading from primary file:", err)
		// Try reading from backup file
		fileData, err = os.ReadFile(cabsBackup)
		if err != nil {
			fmt.Println("Error reading from backup file:", err)
			return nil
		}
	}

	cabs := string(fileData)
	result := make([]int, len(cabs))
	for i, char := range cabs {
		val, err := strconv.Atoi(string(char))
		if err != nil {
			fmt.Println("Error converting string to int:", err)
			return nil
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
	if data == nil {
		return fmt.Errorf("unable to read data from files")
	}
	data[floor] = 1
	if err := writeToFiles(data); err != nil {
		return err
	}
	fmt.Println("Data written successfully to files.")
	return nil
}

func Clear(floor int) error {
	data := Read()
	if data == nil {
		return fmt.Errorf("unable to read data from files")
	}
	data[floor] = 0
	if err := writeToFiles(data); err != nil {
		return err
	}
	fmt.Println("Data written successfully to files.")
	return nil
}
