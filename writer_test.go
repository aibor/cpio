package cpio_test

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"sort"
	"testing"

	"github.com/aibor/cpio"
	"github.com/aibor/cpio/internal"
)

func store(w *cpio.Writer, fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	hdr, err := cpio.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	if err := w.WriteHeader(hdr); err != nil {
		return err
	}
	if !fi.IsDir() {
		if _, err := io.Copy(w, f); err != nil {
			return err
		}
	}
	return err
}

func TestWriter(t *testing.T) {
	var buf bytes.Buffer
	w := cpio.NewWriter(&buf)
	if err := store(w, "testdata/etc"); err != nil {
		t.Fatalf("store: %v", err)
	}
	if err := store(w, "testdata/etc/hosts"); err != nil {
		t.Fatalf("store: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestWriter_AddFS(t *testing.T) {
	testFS := internal.ReadLinkDirFS("testdata")

	var archive bytes.Buffer

	w := cpio.NewWriter(&archive)

	if err := w.AddFS(testFS); err != nil {
		t.Fatalf("add fs: %v", err)
	}

	r := cpio.NewReader(&archive)

	expectedFiles := []string{}
	for _, pattern := range []string{"*", "etc/*"} {
		results, err := fs.Glob(testFS, pattern)
		if err != nil {
			t.Fatalf("glob: %v", err)
		}
		expectedFiles = append(expectedFiles, results...)
	}

	sort.Strings(expectedFiles)

	for _, expected := range expectedFiles {
		hdr, err := r.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if expected != hdr.Name {
			t.Errorf("name mismatch: want %s, got %s", expected, hdr.Name)
		}

	}
}
