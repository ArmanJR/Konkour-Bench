// Package supervisor prepares a launchd job without loading or starting it.
package supervisor

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const Label = "ai.konkour-bench.runner"

// Prepare uses launchd to own the process, rather than an interactive terminal
// or tool session. KeepAlive is deliberately false: crashes must not trigger
// automatic inference retries. Secrets are read by the child from EnvFile.
func Prepare(path, binary, directory string, args []string) error {
	for _, p := range []string{path, binary, directory} {
		if !filepath.IsAbs(p) {
			return fmt.Errorf("supervisor paths must be absolute")
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	var b bytes.Buffer
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n<plist version=\"1.0\"><dict>\n")
	element := func(tag, value string) {
		b.WriteString("<" + tag + ">")
		_ = xml.EscapeText(&b, []byte(value))
		b.WriteString("</" + tag + ">\n")
	}
	element("key", "Label")
	element("string", Label)
	element("key", "ProgramArguments")
	b.WriteString("<array>\n")
	element("string", binary)
	for _, a := range args {
		if strings.ContainsRune(a, 0) {
			return fmt.Errorf("invalid supervisor argument")
		}
		element("string", a)
	}
	b.WriteString("</array>\n")
	element("key", "WorkingDirectory")
	element("string", directory)
	element("key", "RunAtLoad")
	b.WriteString("<true/>\n")
	element("key", "KeepAlive")
	b.WriteString("<false/>\n")
	element("key", "ExitTimeOut")
	b.WriteString("<integer>30</integer>\n")
	element("key", "StandardOutPath")
	element("string", filepath.Join(filepath.Dir(path), "runner-summary.log"))
	element("key", "StandardErrorPath")
	element("string", filepath.Join(filepath.Dir(path), "runner.log"))
	b.WriteString("</dict></plist>\n")
	// Never overwrite an existing service configuration silently.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(b.Bytes())
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
