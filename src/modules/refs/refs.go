package refs

import (
	"goit/src/modules/config"
	filesmodule "goit/src/modules/files"
	"goit/src/modules/objects"
	"goit/src/modules/utils"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

func Hash(refOrHash string) interface{} {
	if objects.Exists(refOrHash) {
		return refOrHash
	} else {
		terminalRef := TerminalRef(refOrHash)
		if terminalRef == "FETCH_HEAD" {
			return fetchHeadBranchToMerge(HeadBranchName())
		}
		if Exists(terminalRef) {
			return filesmodule.Read(filesmodule.GoitPath(terminalRef))
		}
	}

	return ""
}

func TerminalRef(ref string) string {
	if ref == "HEAD" && !IsHeadDetached() {
		headPath := filesmodule.GoitPath("HEAD")
		content := filesmodule.Read(headPath)

		re := regexp.MustCompile(`ref: (refs/heads/.+)`)
		matches := re.FindStringSubmatch(content)

		if len(matches) > 1 {
			return matches[1]
		}

		return ""
	}
	if IsRef(ref) {
		return ref
	} else {
		return ToLocalRef(ref)
	}
}

func IsHeadDetached() bool {
	headPath := filesmodule.GoitPath("HEAD")
	content := filesmodule.Read(headPath)

	re := regexp.MustCompile(`refs`)
	matches := re.FindString(content)

	if matches != "" {
		return true
	}
	return false
}

func HeadBranchName() string {
	if !IsHeadDetached() {
		content := filesmodule.Read(filesmodule.GoitPath("HEAD"))
		re1 := regexp.MustCompile(`refs/heads/(.+)`)
		matches := re1.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}


func IsRef(ref string) bool {
	re1 := regexp.MustCompile(`^refs/heads/[A-Za-z-]+$`)
	matches1 := re1.FindStringSubmatch(ref)
	re2 := regexp.MustCompile(`^refs/remotes/[A-Za-z-]+/[A-Za-z-]+$`)
	matches2 := re2.FindStringSubmatch(ref)

	specialRefs := []string{"HEAD", "FETCH_HEAD", "MERGE_HEAD"}
	for _, s := range specialRefs {
		if ref == s {
			return true
		}
	}

	return len(matches1) > 0 || len(matches2) > 0
}

func ToLocalRef(name string) string {
	return "refs/heads/" + name
}

func fetchHeadBranchToMerge(branchName string) []string {
	lines := utils.Lines(filesmodule.Read(filesmodule.GoitPath("FETCH_HEAD")))

	re := regexp.MustCompile(`^.+ branch ` + branchName + ` of`)
	var filteredLines []string
	for _, line := range lines {
		if re.MatchString(line) {
			filteredLines = append(filteredLines, line)
		}
	}

	re1 := regexp.MustCompile(`^([^ ]+) `)
	var result []string
	for _, fLine := range filteredLines {
		matches := re1.FindStringSubmatch(fLine)
		if len(matches) > 1 {
			result = append(result, matches[1])
		}
	}

	return result
}

func Exists(ref string) bool {
	return IsRef(ref) && filesmodule.Exists(filesmodule.GoitPath(ref))
}

func CommitParentHashes() []string {
	headHash := Hash("HEAD")

	if IsMergeInProgress() != "" {
		return []string{headHash.(string), Hash("MERGE_HEAD").(string)}
	}
	if headHash == "" {
		return []string{}
	} else {
		return []string{headHash.(string)}
	}
}

func IsMergeInProgress() string {
	return Hash("MERGE_HEAD").(string)
}

func Write(ref, content string) {
	if IsRef(ref) {
		filesmodule.Write(filesmodule.GoitPath(filepath.Clean(ref)), content)
	}
}

func LocalHeads() map[string]interface{} {
	entries, err := os.ReadDir(filepath.Join(filesmodule.GoitPath(""), "refs", "heads"))
	if err != nil {
		log.Fatal(err)
	}

	heads := make(map[string]interface{})
	for _, entry := range entries {
		heads[entry.Name()] = Hash(entry.Name())
	}

	return heads
}

func ToRemoteRef(remote, name string) string {
	return "refs/remotes/" + remote + "/" + name
}

func IsCheckedOut(branch string) bool {
	return !config.IsBare() && HeadBranchName() == branch
}