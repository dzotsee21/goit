package diff

import (
	"goit/src/modules/index"
	"goit/src/modules/objects"
	"goit/src/modules/refs"
	"goit/src/modules/utils"
	"slices"
)

var FILE_STATUS = map[string]string{"ADD": "A", "MODIFY": "M", "DELETE": "D", "SAME": "SAME", "CONFLICT": "CONFLICT"}

func AddedOrModifiedFiles() []string {

	var headToc map[string]interface{}
	if refs.Hash("HEAD") != "" {
		headToc = objects.CommitToc(refs.Hash("HEAD").(string))
	} else {
		headToc = make(map[string]interface{})
	}
	wc := nameStatus(tocDiff(headToc, index.WorkingCopyToc(), nil))

	wcKeys := utils.MapKeys(wc)

	var files []string
	for _, key := range wcKeys {
		if wc[key] != FILE_STATUS["DELETE"] {
			files = append(files, key)
		}
	}

	return files
}

func nameStatus(dif map[string]interface{}) map[string]interface{} {
	difKeys := utils.MapKeys(dif)
	statuses := make(map[string]interface{})

	for _, key := range difKeys {
		info := dif[key].(map[string]interface{})

		if info["status"] != FILE_STATUS["SAME"] {
			statuses[key] = info["status"].(string)
		}
	}

	return statuses
}

func tocDiff(receiver, giver, base map[string]interface{}) map[string]interface{} {
	fileStatus := func(receiver, giver, base map[string]interface{}) string {
		var receiverPresent bool
		var giverPresent bool
		var basePresent bool

		if receiver == nil {
			receiverPresent = false
		} else {
			receiverPresent = true
		}

		if giver == nil {
			giverPresent = false
		} else {
			giverPresent = true
		}

		if base == nil {
			basePresent = false
		} else {
			basePresent = true
		}

		if receiverPresent && giverPresent && !mapsEqual(receiver, giver) {
			if !mapsEqual(receiver, base) && !mapsEqual(giver, base) {
				return FILE_STATUS["CONFLICT"]
			} else {
				return FILE_STATUS["MODIFY"]
			}

		}
		if mapsEqual(receiver, giver) {
			return FILE_STATUS["SAME"]
		}
		if (!receiverPresent && !basePresent && giverPresent) ||
		   (receiverPresent && !basePresent && !giverPresent) {
			return FILE_STATUS["ADD"]
		}
		if (receiverPresent && basePresent && !giverPresent) ||
		   (!receiverPresent && basePresent && giverPresent) {
			return FILE_STATUS["DELETE"]
		}

		return ""
	}

	if base == nil {
		base = receiver
	}

	receiverKeys := utils.MapKeys(receiver)
	baseKeys := utils.MapKeys(base)
	giverKeys := utils.MapKeys(giver)

	paths := slices.Concat(receiverKeys, baseKeys, giverKeys)

	uniquePaths := utils.Unique(paths)

	idx := make(map[string]interface{})

	for _, uPath := range uniquePaths {
		status := fileStatus(receiver[uPath].(map[string]interface{}), giver[uPath].(map[string]interface{}), base[uPath].(map[string]interface{}))

		idx = utils.SetIn(idx, []interface{}{map[string]interface{}{
			"status": status,
			"receiver": receiver[uPath],
			"base": base[uPath],
			"giver": giver[uPath],
		}})
	}

	return idx
}

func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, val := range a {
		bVal, ok := b[k]
		if !ok {
			return false
		} else {
			if bVal != val {
				return false
			}
		}
	}

	return true
}

func ChangedFilesCommitWouldOverwrite(hash string) []string {
	headHash := refs.Hash("HEAD")
	return utils.Intersection(utils.MapKeys(nameStatus(Diff(headHash, nil))), utils.MapKeys(nameStatus(Diff(headHash, hash))))
}

func Diff(hash1, hash2 interface{}) map[string]interface{} {
	var a map[string]interface{}
	if hash1 == nil {
		a = index.Toc()
	} else {
		a = objects.CommitToc(hash1.(string))
	}
	var b map[string]interface{}
	if hash2 == nil {
		b = index.WorkingCopyToc()
	} else {
		b = objects.CommitToc(hash2.(string))
	}

	return tocDiff(a, b, nil)
}