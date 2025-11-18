package filesmodule

import (
	"fmt"
	"goit/src/modules/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func InRepo() bool {
	return GoitPath("") != ""
}

func AssertInRepo() error {
	if InRepo() {
		return fmt.Errorf("not a Goit repository")
	}

	return nil
}

func Read(path string) string {
	if Exists(path) {
		if path == "" {
			return ""
		}

		data, err := os.ReadFile(path)
		if err != nil {

			fmt.Println("Error reading file:", err)
		}
		return string(data)
	}

	return ""
}

func GoitPath(path string) string {
	goitDir := func(dir string) string {
		if Exists(dir) {
			potentialGoitFile := filepath.Join(dir, ".goit")
			potentialConfigFile := filepath.Join(dir, "config")

			if Exists(potentialGoitFile) {
				return potentialGoitFile
			}
			if Exists(potentialConfigFile) && !Exists(potentialGoitFile) {
				return dir
			}
		}

		return ""
	}

	dir, _ := os.Getwd()
	gDir := goitDir(dir)
	if gDir != "" {
		return filepath.Join(gDir, path)
	}

	return ""
}

func LsRecursive(path string) []string {
	if !Exists(path) {
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

func Exists(path string) bool {
	if path == "" {
		return true
	}
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func PathFromRepoRoot(path string) string {
	dir, _ := os.Getwd()

	relativePath, err := filepath.Rel(WorkingCopyPath(""), filepath.Join(dir, path))
	if err != nil {
		fmt.Printf("error")
	}

	return relativePath
}

func WorkingCopyPath(path string) string {
	return filepath.Join(filepath.Join(GoitPath(""), ".."), path)
}

func Write(path string, content string) {

	parts := strings.Split(path, string(filepath.Separator))
	result := append(parts, content)

	WriteFilesFromTree(utils.SetIn(map[string]interface{}{}, stringToInterface(result)), "")
}

func stringToInterface(strs []string) []interface{} {
	res := make([]interface{}, len(strs))
	for i, s := range strs {
		res[i] = s
	}
	return res
}

func WriteFilesFromTree(tree map[string]interface{}, basePath string) {
	treeKeys := utils.MapKeys(tree)

	for _, name := range treeKeys {
		path := filepath.Join(basePath, name)
		// hacky way
		if len(path) > 2 && path[1] == ':' && path[2] != '\\' {
			path = path[:2] + `\` + path[2:]
		}

		data := tree[name]

		if str, ok := data.(string); ok {
			os.WriteFile(path, []byte(str), 0644)
		} else {
			if !Exists(path) {
				os.Mkdir(path, 0777)
			}

			WriteFilesFromTree(tree[name].(map[string]interface{}), path)
		}
	}
}

func FlattenNestedTree(tree, obj map[string]interface{}, basePath string) map[string]interface{} {
	if obj == nil {
		return FlattenNestedTree(tree, map[string]interface{}{}, "")
	}

	treeKeys := utils.MapKeys(tree)

	for _, key := range treeKeys {
		path := filepath.Join(basePath, key)
		if str, ok := tree[key].(string); ok {
			obj[path] = str
		} else {
			FlattenNestedTree(tree, obj, path)
		}
	}

	return obj
}

func NestFlatTree(obj map[string]interface{}) map[string]interface{} {
	objKeys := utils.MapKeys(obj)

	nestedObj := map[string]interface{}{objKeys[0]: obj[objKeys[0]]}
	for _, key := range objKeys {
		_, ok := obj[key].([]interface{})
		if ok {
			nestedObj = utils.SetIn(nestedObj, obj[key].([]interface{}))
		} else {
			nestedObj = utils.SetIn(nestedObj, append([]interface{}{}, obj[key]))
		}
	}

	return nestedObj
}

func RmEmptyDirs(path string) {
	info, _ := os.Stat(path)

	if info.IsDir() {
		entries, _ := os.ReadDir(path)
		for _, entry := range entries {
			RmEmptyDirs(filepath.Join(path, entry.Name()))
		}
		if len(entries) == 0 {
			err := os.Remove(path)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
