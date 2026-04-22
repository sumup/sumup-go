package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSamplesToStdout(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	if err := writeSamples("", []byte("samples\n"), &stdout); err != nil {
		t.Fatalf("write samples: %v", err)
	}
	if stdout.String() != "samples\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "samples\\n")
	}
}

func TestWriteSamplesToFile(t *testing.T) {
	t.Parallel()

	filename := filepath.Join(t.TempDir(), "nested", "samples.json")
	if err := writeSamples(filename, []byte("samples\n"), &bytes.Buffer{}); err != nil {
		t.Fatalf("write samples: %v", err)
	}
	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read samples: %v", err)
	}
	if string(contents) != "samples\n" {
		t.Fatalf("contents = %q, want %q", contents, "samples\\n")
	}
}

func TestReadSDKVersion(t *testing.T) {
	t.Parallel()

	filename := filepath.Join(t.TempDir(), "version.go")
	source := []byte("package internal\n\nconst Version = \"1.2.3\" // x-release-please-version\n")
	if err := os.WriteFile(filename, source, 0o600); err != nil {
		t.Fatalf("write version file: %v", err)
	}

	version, err := readSDKVersion(filename)
	if err != nil {
		t.Fatalf("read SDK version: %v", err)
	}
	if version != "1.2.3" {
		t.Fatalf("version = %q, want 1.2.3", version)
	}
}
