package gitreviewers

import (
	"os/exec"
	rx "regexp"
	"strconv"
	"strings"
)

var countExtractor *rx.Regexp

func init() {
	// Pattern to extract commit count and name/email from git shortlog.
	countExtractor = rx.MustCompile("(\\d+)\\s*(.*)$")
}

// run executes cmd via a shell process and returns
// its output as a string. If the shell returns an error, return
// that instead.
func run(cmd string) (string, error) {
	// TODO Output command in verbose mode
	tokens := strings.Split(cmd, " ")
	out, err := exec.Command(tokens[0], tokens[1:]...).CombinedOutput()

	if err != nil {
		// TODO Output error in verbose mode
		return "", err
	}

	return string(out), nil
}

// commitTimeStamp returns the timestamp of the current commit for
// the object (branch, commit, etc.).
func commitTimeStamp(obj string) (string, error) {
	out, err := run("git show --format=\"%ct\" " + obj)
	if err != nil {
		return "", nil
	}

	line := strings.Split(out, "\n")[0]
	return strings.Trim(line, "\""), nil
}

// changedFiles returns the paths of all files changed in commits between
// master and the current branch.
func changedFiles() ([]string, error) {
	var lines []string
	out, err := run("git diff master HEAD --name-only")

	if err != nil {
		return lines, err
	}

	for _, line := range strings.Split(out, "\n") {
		l := strings.Trim(line, " ")
		if len(l) > 0 {
			lines = append(lines, l)
		}
	}

	return lines, err
}

// committerCounts returns recent committers and commit counts for
// the file at `path`.
func committerCounts(path string) (Stats, error) {
	var stats []Stat

	// TODO Parse "since" date from options or calculate from current
	// date if not specified
	since, err := exec.Command(
		"bash", "-c", "git log --since 2015-01-01 --reverse |"+
			"head -n 1 | awk '{print $2}'").Output()

	if err != nil {
		return stats, err
	}

	cmd := strings.Join(
		[]string{
			"git shortlog -sne --no-merges",
			strings.TrimSpace(string(since)) + "..HEAD",
			path,
		}, " ")

	out, err := run(cmd)
	if err != nil {
		return stats, err
	}

	for _, line := range strings.Split(out, "\n") {
		line = strings.Trim(line, " ")
		matches := countExtractor.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		ct := matches[1]
		rvwr := matches[2]

		cti, err := strconv.Atoi(ct)
		if err != nil {
			continue
		}

		stats = append(stats, Stat{rvwr, cti})
	}

	return stats, nil
}
