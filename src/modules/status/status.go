package status

import (
	"goit/src/modules/diff"
	filesmodule "goit/src/modules/files"
	"goit/src/modules/index"
	"goit/src/modules/objects"
	"goit/src/modules/refs"
	"goit/src/modules/utils"
	"os"
	"strings"
)

func ToString() string {

	untracked := func() []string {
		files, _ := os.ReadDir(filesmodule.WorkingCopyPath(""))

		var untrackedFiles []string
		for _, file := range files {
			if index.Toc()[file.Name()] == nil && file.Name() != ".goit" {
				untrackedFiles = append(untrackedFiles, file.Name())
			}
		}

		return untrackedFiles
	}

	toBeCommited := func() []string {
		headHash := refs.Hash("HEAD")
		var headToc map[string]interface{}
		if headHash == nil {
			headToc = map[string]interface{}{}
		} else {
			headToc = objects.CommitToc(headHash.(string))
		}

		ns := diff.NameStatus(diff.TocDiff(headToc, index.Toc(), nil))

		var commitedFiles []string
		for _, key := range utils.MapKeys(ns) {
			commitedFiles = append(commitedFiles, ns[key].(string)+" "+key)
		}

		return commitedFiles
	}

	notStagedForCommit := func() []string {
		ns := diff.NameStatus(diff.Diff(nil, nil))
		var notStagedFiles []string
		for _, key := range utils.MapKeys(ns) {
			notStagedFiles = append(notStagedFiles, ns[key].(string)+" "+key)
		}

		return notStagedFiles
	}

	listing := func(heading string, lines []string) []string {
		if len(lines) > 0 {
			return append([]string{heading}, lines...)
		} else {
			return []string{}
		}
	}

	return "on branch " + refs.HeadBranchName() + "\n" + strings.Join(listing("untracked files:", untracked()), ",") + strings.Join(listing("unmerged paths:", index.ConflictedPaths()), ",") + strings.Join(listing("changes to be commited:", toBeCommited()), ",") + strings.Join(listing("changes not staged for commit:", notStagedForCommit()), ",") + "\n"
}