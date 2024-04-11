package eskolog

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCollectionParsing(t *testing.T) {
	_, currentFile, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(currentFile)
	targetFilePath := filepath.Join(basePath, "../../test/log.txt")

	op := ParserOptions{
		ParseTimeStamp: true,
		Tags:           []string{"OVIS"},
	}
	coll, err := ReadCollection(targetFilePath, &op)
	if err != nil {
		t.Error(err)
		return
	}

	if len(coll.Sessions) != 1 {
		t.Errorf("expected 1 session")
		return
	}

	if len(coll.Sessions[0].Groups) != 2 {
		t.Errorf("expected 2 groups")
		return
	}

	if !coll.Sessions[0].HasBounds {
		t.Errorf("expected bounds to be set")
		return
	}

	attrs, ok := coll.Sessions[0].Attributes["0x134835000"]
	if !ok {
		t.Errorf("attribute 0x134835000 not found")
	}

	regionAttempt := strings.Split(attrs["RegionAttempts"], ",")

	if regionAttempt[0] != "resolver:only-one-station" {
		t.Errorf("expected other attribute")
		return
	}
	if regionAttempt[1] != "resolver:blocked-by-station" {
		t.Errorf("expected other attribute")
		return
	}
	if regionAttempt[2] != "resolver:claimed-by-station" {
		t.Errorf("expected other attribute")
		return
	}
	if regionAttempt[3] != "valid:0x12380a888:0x600001a71b58" {
		t.Errorf("expected other attribute")
		return
	}
	if regionAttempt[4] != "overlap-dfs-claim:0x12380a888:0x600001a71b58" {
		t.Errorf("expected other attribute")
		return
	}

}
