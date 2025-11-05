package index

import (
	"fmt"
	filesmodule "goit/src/modules/files"
	"goit/src/modules/objects"
	"goit/src/modules/utils"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func UpdateIndex(path string, cmds []string) string {
	filesmodule.AssertInRepo()

	pathFromRoot := filesmodule.PathFromRepoRoot(path)
	isOnDisk := filesmodule.Exists(path)
	isInIndex := hasFile(path, 0)

	info, _ := os.Stat(path)
	if isOnDisk && info.IsDir() {
		fmt.Println(pathFromRoot + " is a directory - add files inside\n")
		return ""
	}
	if slices.Contains(cmds, "remove") && !isOnDisk && isInIndex {
		if isFileInConflict(path) {
			fmt.Println("unsupported")
		} else {
			WriteRm(path)
			return "\n"
		}
	}
	if slices.Contains(cmds, "remove") && !isOnDisk && !isInIndex {
		return "\n"
	}
	if !slices.Contains(cmds, "add") && isOnDisk && !isInIndex {
		fmt.Println("cannot add " + pathFromRoot + " to index - use --add option\n")
		return ""
	}
	if isOnDisk && (slices.Contains(cmds, "add") || isInIndex) {
		writeNonConflict(path, filesmodule.Read(filesmodule.WorkingCopyPath(path)))
		return "\n"
	}
	if !slices.Contains(cmds, "remove") && !isOnDisk {
		fmt.Println(pathFromRoot + " does not exist and --remove not passed\n")
	}

	return "\n"
}

func hasFile(path string, stage int) bool {
	idx := Read()

	_, exists := idx[key(path, stage)]
	return exists
}

func Read() map[string]interface{} {
	indexFilePath := filesmodule.GoitPath("index")

	var lines []string
	if filesmodule.Exists(indexFilePath) {
		lines = utils.Lines(filesmodule.Read(indexFilePath))
	} else {
		lines = utils.Lines("\n")
	}

	idx := make(map[string]interface{})
	for _, blobStr := range lines {
		blobData := strings.Split(blobStr, " ")
		if len(blobData) < 3 {
			continue
		}

		stageInt, _ := strconv.Atoi(blobData[1])
		key := key(blobData[0], stageInt)
		idx[key] = blobData[2]
	}

	return idx
}

func key(path string, stage int) string {
	return path + "," + strconv.Itoa(stage)
}

func isFileInConflict(path string) bool {
	return hasFile(path, 2)
}

func writeNonConflict(path, content string) {
	WriteRm(path)

	writeStageEntry(path, content, 0)
}

func WriteRm(path string) {
	idx := Read()
	stages := []int{0, 1, 2, 3}
	for stage := range stages {
		delete(idx, key(path, stage))
	}

	Write(idx)
}

func Write(index map[string]interface{}) {
	indexKeys := utils.MapKeys(index)

	var proccesedKeys []string
	for _, k := range indexKeys {
		parts := strings.Split(k, ",")
		line := fmt.Sprintf("%s %s %s", parts[0], parts[1], index[k].(string))
		proccesedKeys = append(proccesedKeys, line)
	}

	indexStr := strings.Join(proccesedKeys, "\n") + "\n"

	filesmodule.Write(filesmodule.GoitPath("index"), indexStr)
}

func writeStageEntry(path, content string, stage int) {
	idx := Read()
	idx[key(path, stage)] = objects.Write(content)
	Write(idx)
}

func MatchingFiles(path string) []string {
	searchPath := filesmodule.PathFromRepoRoot(path)
	searchPathKeys := utils.MapKeys(toc())

	escaped := regexp.QuoteMeta(searchPath)
	re := regexp.MustCompile("^" + escaped)

	var paths []string
	for _, key := range searchPathKeys {
		if re.MatchString(key) {
			paths = append(paths, key)
		}
	}

	return paths
}

func toc() map[string]interface{} {
	idx := Read()

	result := make(map[string]interface{})
	for k, v := range idx {
		parts := strings.Split(k, ",")
		if len(parts) == 0 {
			continue
		}

		path := parts[0]
		result[path] = v
	}

	return result
}

func WorkingCopyToc() map[string]interface{} {
	indexKeys := utils.MapKeys(Read())

	result := make(map[string]interface{})
	for _, key := range indexKeys {
		key = strings.Split(key, ",")[0]

		if filesmodule.Exists(filesmodule.WorkingCopyPath(key)) {
			hashed := utils.Hash(filesmodule.Read(filesmodule.WorkingCopyPath(key)))
			result[key] = hashed
		}
	}

	return result
}
