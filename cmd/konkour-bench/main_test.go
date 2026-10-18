package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvFileIsLiteralAndNeverExecuted(t *testing.T) {
	for _, value := range []string{"test-key", "'test-key'", "\"test-key\""} {
		path := filepath.Join(t.TempDir(), ".env")
		os.WriteFile(path, []byte("# comment\nexport OPENROUTER_API_KEY="+value+"\n"), 0600)
		key, err := keyFromFile(path)
		if err != nil || key != "test-key" {
			t.Fatal(key, err)
		}
	}
	for _, text := range []string{"OPENROUTER_API_KEY=$(touch nope)", "OPENROUTER_API_KEY=one\nOPENROUTER_API_KEY=two", "OPENROUTER_API_KEY=", "OTHER=value"} {
		path := filepath.Join(t.TempDir(), ".env")
		os.WriteFile(path, []byte(text), 0600)
		if _, err := keyFromFile(path); err == nil {
			t.Fatal("accepted invalid env assignment")
		}
	}
}

func TestInvalidSafetyFlagsFailBeforeNetwork(t *testing.T) {
	for _, args := range [][]string{{"run", "--budget-usd", "NaN"}, {"run", "--budget-usd", "-1"}, {"run", "--timeout", "-1s"}, {"run", "--warn-after", "-1s"}} {
		err := run(context.Background(), args)
		if err == nil || !strings.Contains(err.Error(), "nonnegative") {
			t.Fatalf("%v: %v", args, err)
		}
	}
}
