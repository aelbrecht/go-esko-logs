package eskolog

import (
	"path/filepath"
	"runtime"
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
	}

	if len(coll.Sessions) != 1 {
		t.Errorf("expected 1 session")
	}

	if len(coll.Sessions[0].Groups) != 2 {
		t.Errorf("expected 2 groups")
	}

	if !coll.Sessions[0].HasBounds {
		t.Errorf("expected bounds to be set")
	}
}
