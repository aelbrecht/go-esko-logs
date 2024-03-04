package eskolog

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp int64
	Time      int64
	Body      string
	Meta      map[string]string
}

type ParserOptions struct {
	ParseTimeStamp bool
	Tags           []string
}

func parseParams(input string) map[string]string {
	result := make(map[string]string)

	// Split the input string by spaces to get individual key-value pairs
	pairs := strings.Split(input, ",")

	// Iterate through the key-value pairs and split by '='
	for _, pair := range pairs {
		keyValue := strings.Split(pair, "=")
		if len(keyValue) == 1 {
			key := keyValue[0]
			result[key] = ""
		} else if len(keyValue) == 2 {
			key := keyValue[0]
			value := keyValue[1]
			result[key] = value
		}
	}

	return result
}

func hasTags(tag string, opts *ParserOptions) bool {
	for _, t := range opts.Tags {
		if tag == t {
			return true
		}
	}
	return false
}

func parseLogEntry(line string, opts *ParserOptions) (*LogEntry, error) {

	t := time.Now().UTC()

	// Extract the timestamp using the third colon
	if opts.ParseTimeStamp {
		colonIndex := strings.Index(line, ":")
		for i := 0; i < 2; i++ {
			if colonIndex == -1 {
				return nil, fmt.Errorf("not enough colons for timestamp")
			}
			colonIndex = strings.Index(line[colonIndex+1:], ":") + colonIndex + 1
		}

		if colonIndex == -1 {
			return nil, fmt.Errorf("not enough colons for timestamp")
		}

		timestampStr := line[:colonIndex]
		layout := "2006-01-02 15:04:05"
		if len(timestampStr) < 19 {
			log.Printf("invalid log entry: %s\n", line)
			return nil, nil
		}

		var err error
		t, err = time.Parse(layout, timestampStr[:19])
		if err != nil {
			return nil, err
		}
	}

	// Find the bracketed tag
	startBracket := strings.Index(line, "[")
	endBracket := strings.Index(line, "]")
	if startBracket == -1 || endBracket == -1 || endBracket < startBracket {
		return nil, fmt.Errorf("no valid tag found")
	}

	if !hasTags(line[startBracket+1:endBracket], opts) {
		return nil, nil
	}

	params := make(map[string]string)
	startPar := strings.Index(line, "(")
	endPar := strings.Index(line, ")")
	if startPar != -1 && endPar != -1 || endPar > startPar {
		params = parseParams(line[startPar+1 : endPar])
	}

	// Find the "at X.XXXs" part and convert to milliseconds
	timeMs := int64(-1)
	if opts.ParseTimeStamp {
		atIndex := strings.Index(line, "at ")
		sIndex := strings.Index(line, "s:")
		if atIndex == -1 || sIndex == -1 || sIndex < atIndex {
			return nil, fmt.Errorf("no valid time found")
		}
		timeStr := line[atIndex+3 : sIndex]
		timeFloat, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			return nil, err
		}
		timeMs = int64(timeFloat * 1000)
	}

	// Extract the body
	body := line[endBracket+1:]
	if endPar != -1 {
		body = line[endPar+1:]
	}

	return &LogEntry{
		Timestamp: t.UTC().Unix(),
		Time:      timeMs,
		Body:      strings.TrimSpace(body),
		Meta:      params,
	}, nil
}

func ReadLog(filePath string, opts *ParserOptions) ([]LogEntry, error) {

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")

	var logEntries []LogEntry

	delimiterIndex := -1
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "--" {
			delimiterIndex = i
			break
		}
	}

	// Parse the lines after the delimiter
	for i := delimiterIndex + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		entry, err := parseLogEntry(line, opts)
		if err != nil {
			continue // Or handle the error as needed
		}
		if entry != nil {
			logEntries = append(logEntries, *entry)
		}
	}

	return logEntries, nil
}
