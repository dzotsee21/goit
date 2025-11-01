package commands

import (
	"log"
	"os"
	"path/filepath"
	"slices"
)

func Init() {
	var currentFiles []string

	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		currentFiles = append(currentFiles, file.Name())
	}

	containsGoit := slices.Contains(currentFiles, ".goit")

	if !containsGoit {
		goitStructure := map[string]interface{}{
			"HEAD": "ref: refs/heads/master\n",
			"config": "",
			"objects": map[string]interface{}{},
			"refs": map[string]interface{}{
				"heads": map[string]interface{}{},
			},
		}

		base := ".goit"

		createStructure(base, goitStructure)
	}
}

func createStructure(basePath string, structure map[string]interface{}) error {
	err := os.Mkdir(basePath, os.FileMode(0755))
	if err != nil {
		log.Fatal(err)
	}

	for name, value := range structure {
		filePath := filepath.Join(basePath, name)

		switch v := value.(type) {
		case string:
			if err := os.WriteFile(filePath, []byte(v), 0644); err != nil {
				return err
			}
		case map[string]interface{}:
			if err := createStructure(filePath, v); err != nil {
				return err
			}
		}
	}

	return nil
}