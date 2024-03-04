package eskolog

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestLogParsing(t *testing.T) {
	_, currentFile, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(currentFile)
	targetFilePath := filepath.Join(basePath, "../../test/log.txt")

	op := ParserOptions{
		ParseTimeStamp: true,
		Tags:           []string{"OVIS"},
	}
	data, err := ReadLog(targetFilePath, &op)
	if err != nil {
		t.Error(err)
	}

	if len(data) != 126 {
		t.Errorf("expected %d entries but got %d", 126, len(data))
	}
}
