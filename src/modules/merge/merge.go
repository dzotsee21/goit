package merge

import (
	"fmt"
	"goit/src/modules/config"
	"goit/src/modules/diff"
	filesmodule "goit/src/modules/files"
	"goit/src/modules/index"
	"goit/src/modules/objects"
	"goit/src/modules/refs"
	"goit/src/modules/utils"
	workingcopy "goit/src/modules/working_copy"
	"sort"
	"strings"
)

func IsAForceFetch(receiverHash, giverHash string) bool {
	return !objects.IsAncestor(giverHash, receiverHash)
}

func CanFastForward(receiverHash, giverHash string) bool {
	return receiverHash == "" || objects.IsAncestor(giverHash, receiverHash)
}

func WriteFastForwardMerge(receiverHash, giverHash interface{}) {
	refs.Write(refs.ToLocalRef(refs.HeadBranchName()), giverHash.(string))

	index.Write(index.TocToIndex(objects.CommitToc(giverHash.(string))))

	if !config.IsBare() {
		var receiverToc map[string]interface{}
		if receiverHash == nil {
			receiverToc = map[string]interface{}{}
		} else {
			receiverToc = objects.CommitToc(receiverHash.(string))
		}

		workingcopy.Write(diff.TocDiff(receiverToc, objects.CommitToc(giverHash.(string)), nil))
	}
}

func WriteNonFastForwardMerge(receiverHash, giverHash, giverRef string) {
	refs.Write("MERGE_HEAD", giverHash)

	writeMergeMsg(receiverHash, giverHash, giverRef)

	writeIndex(receiverHash, giverHash)

	if !config.IsBare() {
		workingcopy.Write(mergeDiff(receiverHash, giverHash))
	}

}

func writeMergeMsg(receiverHash, giverHash, ref string) {
	msg := "merge " + ref + " into " + refs.HeadBranchName()

	_mergeDiff := mergeDiff(receiverHash, giverHash)
	var conflicts []string
	for _, key := range utils.MapKeys(_mergeDiff) {
		if _mergeDiff[key].(map[string]string)["status"] == diff.FILE_STATUS["CONFLICT"] {
			conflicts = append(conflicts, key)
		}
	}

	if len(conflicts) > 0 {
		msg += "\nconflicts:\n" + strings.Join(conflicts, "\n")
	}

	filesmodule.Write(filesmodule.GoitPath("MERGE_MSG"), msg)
}

func mergeDiff(receiverHash, giverHash string) map[string]interface{} {
	return diff.TocDiff(objects.CommitToc(receiverHash),
		objects.CommitToc(giverHash),
		objects.CommitToc(commonAncestor(receiverHash, giverHash)))
}

func commonAncestor(aHash, bHash string) string {
	sorted := []string{aHash, bHash}
	sort.Strings(sorted)

	aHash = sorted[0]
	bHash = sorted[1]
	aAncestors := append([]string{aHash}, toStringSlice(objects.Ancestors(aHash))...)
	bAncestors := append([]string{aHash}, toStringSlice(objects.Ancestors(bHash))...)

	return utils.Intersection(aAncestors, bAncestors)[0]
}

func toStringSlice(in []interface{}) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = fmt.Sprint(v)
	}
	return out
}

func writeIndex(receiverHash, giverHash string) {
	_mergeDiff := mergeDiff(receiverHash, giverHash)
	index.Write(map[string]interface{}{})

	for _, p := range utils.MapKeys(_mergeDiff) {
		if _mergeDiff[p].(map[string]string)["status"] == diff.FILE_STATUS["CONFLICT"] {
			index.WriteConflict(p,
				objects.Read(_mergeDiff[p].(map[string]string)["receiver"]),
				objects.Read(_mergeDiff[p].(map[string]string)["giver"]),
				objects.Read(_mergeDiff[p].(map[string]string)["base"]),
			)
		}
		if _mergeDiff[p].(map[string]string)["status"] == diff.FILE_STATUS["MODIFY"] {
			index.WriteNonConflict(p, objects.Read(_mergeDiff[p].(map[string]string)["giver"]))
		}
		if _mergeDiff[p].(map[string]string)["status"] == diff.FILE_STATUS["ADD"] || _mergeDiff[p].(map[string]string)["status"] == diff.FILE_STATUS["SAME"] {
			_, exists := _mergeDiff[p].(map[string]string)["receiver"]

			var content string
			if exists {
				content = objects.Read(_mergeDiff[p].(map[string]string)["receiver"])
			}
			_, exists = _mergeDiff[p].(map[string]string)["giver"]

			if exists {
				content = objects.Read(_mergeDiff[p].(map[string]string)["giver"])
			}

			index.WriteNonConflict(p, content)
		}
	}
}

func HasConflicts(receiverHash, giverHash string) bool {
	_mergeDiff := mergeDiff(receiverHash, giverHash)
	var conflicts []string
	for _, key := range utils.MapKeys(_mergeDiff) {
		if _mergeDiff[key].(map[string]string)["status"] == diff.FILE_STATUS["CONFLICT"] {
			conflicts = append(conflicts, key)
		}
	}

	return len(conflicts) > 0
}
