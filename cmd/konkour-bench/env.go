package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Read a literal key assignment, not shell code. launchd never receives the key
// in its plist or command line. No interpolation, command substitution, or logs.
func keyFromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	key := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.TrimPrefix(line, "export ")
		name, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(name) != "OPENROUTER_API_KEY" {
			continue
		}
		if key != "" {
			return "", fmt.Errorf("duplicate OPENROUTER_API_KEY assignment")
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"')) {
			value = value[1 : len(value)-1]
		}
		if value == "" || strings.ContainsAny(value, " \t\r\n$`\"'") {
			return "", fmt.Errorf("OPENROUTER_API_KEY must be a nonempty literal value")
		}
		key = value
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if key == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY not found in env file")
	}
	return key, nil
}
