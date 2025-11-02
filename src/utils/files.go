package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func InRepo() bool {
	return GoitPath("") != ""
}

func AssertInRepo() error {
	if (InRepo()) {
		return fmt.Errorf("not a Goit repository")
	}

	return nil
}

func GoitPath(path string) string {
	goitDir := func(dir string) string {
		if exists(path) {
			potentialGoitFile := filepath.Join(dir, ".goit")

			if (exists(potentialGoitFile)) {
				return potentialGoitFile
			}
		}

		return ""
	}

	dir, _ := os.Getwd()
	gDir := goitDir(dir)
	if (gDir != "") {
		return filepath.Join(gDir, path)
	}

	return ""
}

func LsRecursive(path string) []string {
	if !exists(path) {
		return nil
	}
	
	info, _ := os.Stat(path)
	if !info.IsDir() { // path is a file
		listPath := []string{path}
		return listPath
	}

	if info.IsDir() { // path is a directory
		files, err := os.ReadDir(path)
		if err != nil {
			fmt.Println("error: ", err)
		}

		var fileList []string
		for _, file := range files {
			filePath := filepath.Join(path, file.Name())

			if file.IsDir() {
				fileList = append(fileList, LsRecursive(filePath)...)
			} else {
				fileList = append(fileList, filePath)
			}
		}

		return fileList
	}

	return nil
}

func exists(path string) bool {
	if path == "" {
		return true
	}
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
