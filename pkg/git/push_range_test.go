package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

// A commit id no repository in these tests contains.
const unknownCommit = "1111111111111111111111111111111111111111"

const appFile = "create/app.txt"

// pushRangeOrigin is a repository whose master carries a commit of its own
// (hotfix) while develop has moved on by two commits (a, then b).
type pushRangeOrigin struct {
	path string

	base   string
	a      string
	b      string
	hotfix string
}

func setupPushRangeOrigin(t *testing.T) pushRangeOrigin {
	t.Helper()

	isolateGitConfig(t)

	path := filepath.Join(t.TempDir(), "origin")

	runScript(t, fmt.Sprintf(`
mkdir -p %[1]q && cd %[1]q
git init -q -b master .
mkdir core create docs
echo v0 > core/lib.txt && echo v0 > create/app.txt && echo v0 > docs/readme.txt
git add -A && git commit -qm base
git checkout -q -b develop
echo a > create/app.txt && git commit -qam a
echo b > create/app.txt && git commit -qam b
git checkout -q master
echo hotfix > core/lib.txt && git commit -qam hotfix
`, path))

	return pushRangeOrigin{
		path:   path,
		base:   revParse(t, path, "master~1"),
		a:      revParse(t, path, "develop~1"),
		b:      revParse(t, path, "develop"),
		hotfix: revParse(t, path, "master"),
	}
}

// moveMaster points master at a commit the way a force push does, leaving
// the previous commit in the origin's object store but on no branch.
func (o pushRangeOrigin) moveMaster(t *testing.T, commit string) {
	t.Helper()

	gitOutput(t, o.path, "update-ref", "refs/heads/master", commit)
}

// cloneMaster clones master the way the compilation job does.
func (o pushRangeOrigin) cloneMaster(t *testing.T) string {
	t.Helper()

	clone := filepath.Join(t.TempDir(), "clone")

	runScript(t, fmt.Sprintf(
		`git clone -q --depth 50 --single-branch --branch master "file://%s" %q`,
		o.path, clone,
	))

	t.Chdir(clone)
	resetMemoizedGitState()

	return clone
}

func revParse(t *testing.T, dir string, rev string) string {
	t.Helper()

	return gitOutput(t, dir, "rev-parse", rev)
}

func pushRange(base string, head string) string {
	return base + threeDots + head
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()

	output, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}

	return strings.TrimSpace(string(output))
}

// Moving master onto develop drops the hotfix. The undone hotfix is a change
// of master, and a diff through the merge base would miss it.
func TestPushRangeDiffListComparesTipsWhenMasterMovesOntoAnotherBranch(t *testing.T) {
	origin := setupPushRangeOrigin(t)
	origin.moveMaster(t, origin.b)
	clone := origin.cloneMaster(t)

	historyBefore := gitOutput(t, clone, "rev-list", "--count", "HEAD")

	diff, err := PushRangeDiffList(pushRange(origin.hotfix, origin.b))

	assert.NoError(t, err)
	assert.True(t, diff.BaseAvailable)
	assert.Equal(t, []string{"core/lib.txt", appFile}, diff.Files)

	// Fetching the previous commit must not cut the history of the clone.
	assert.Equal(t, historyBefore, gitOutput(t, clone, "rev-list", "--count", "HEAD"))
}

// Moving master back one commit reverts that commit's changes. Through the
// merge base, which is the new commit itself, the diff would be empty.
func TestPushRangeDiffListReportsChangesRevertedByMovingMasterBack(t *testing.T) {
	origin := setupPushRangeOrigin(t)
	origin.moveMaster(t, origin.a)
	origin.cloneMaster(t)

	diff, err := PushRangeDiffList(pushRange(origin.b, origin.a))

	assert.NoError(t, err)
	assert.True(t, diff.BaseAvailable)
	assert.Equal(t, []string{appFile}, diff.Files)
}

// For a fast-forward the merge base is the previous commit, so both ways of
// comparing agree and nothing changes for a branch that is never rewritten.
func TestPushRangeDiffListMatchesTheMergeBaseDiffForAFastForward(t *testing.T) {
	origin := setupPushRangeOrigin(t)
	origin.moveMaster(t, origin.b)
	origin.cloneMaster(t)

	commitRange := pushRange(origin.a, origin.b)

	diff, err := PushRangeDiffList(commitRange)
	assert.NoError(t, err)

	mergeBaseDiff, err := DiffList(commitRange)
	assert.NoError(t, err)

	assert.True(t, diff.BaseAvailable)
	assert.Equal(t, []string{appFile}, diff.Files)
	assert.Equal(t, mergeBaseDiff, diff.Files)
}

// The ranges the push itself carries also come as 'first^...head'.
func TestPushRangeDiffListResolvesAParentRevision(t *testing.T) {
	origin := setupPushRangeOrigin(t)
	origin.moveMaster(t, origin.b)
	origin.cloneMaster(t)

	diff, err := PushRangeDiffList(pushRange(origin.a+"^", origin.b))

	assert.NoError(t, err)
	assert.True(t, diff.BaseAvailable)
	assert.Equal(t, []string{appFile}, diff.Files)
}

// A previous commit that cannot be obtained at all leaves nothing to compare
// against, so every file of the new commit counts as changed.
func TestPushRangeDiffListListsEveryFileWhenThePreviousCommitIsGone(t *testing.T) {
	origin := setupPushRangeOrigin(t)
	origin.moveMaster(t, origin.b)
	origin.cloneMaster(t)

	diff, err := PushRangeDiffList(pushRange(unknownCommit, origin.b))

	assert.NoError(t, err)
	assert.False(t, diff.BaseAvailable)
	assert.Equal(t, []string{"core/lib.txt", appFile, "docs/readme.txt"}, diff.Files)
}

// A remote that refuses to serve a commit by its id still lets deepening
// find a previous commit that is an ancestor beyond the clone's depth.
func TestPushRangeDiffListDeepensWhenTheRemoteRefusesFetchingById(t *testing.T) {
	isolateGitConfig(t)

	root := t.TempDir()
	origin := filepath.Join(root, "origin")
	clone := filepath.Join(root, "clone")

	runScript(t, fmt.Sprintf(`
mkdir -p %[1]q && cd %[1]q
git init -q -b master .
mkdir lib && echo base > lib/base.txt
git add -A && git commit -qm base
echo changed > lib/feature.txt
git add -A && git commit -qm "change in lib"
for i in $(seq 1 5); do git commit -q --allow-empty -m "master $i"; done
git clone -q --depth 1 --single-branch --branch master "file://%[1]s" %[2]q
# Protocol v0 refuses requests for objects the remote did not advertise.
git -C %[2]q config protocol.version 0
`, origin, clone))

	t.Chdir(clone)
	resetMemoizedGitState()

	base := revParse(t, origin, "master~6")
	head := revParse(t, origin, "master")

	diff, err := PushRangeDiffList(pushRange(base, head))

	assert.NoError(t, err)
	assert.True(t, diff.BaseAvailable)
	assert.Equal(t, []string{"lib/feature.txt"}, diff.Files)
}

func TestPushRangeDiffListFailsWhenTheNewCommitIsMissing(t *testing.T) {
	origin := setupPushRangeOrigin(t)
	origin.cloneMaster(t)

	_, err := PushRangeDiffList(pushRange(origin.hotfix, unknownCommit))

	assert.Error(t, err)
}
