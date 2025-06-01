package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func TestSSG(t *testing.T) {
	destDir := "dist"

	if os.RemoveAll(destDir) != nil {
		t.Fatalf("Failed to remove old destination dir")
	}

	sourceDir := "./testdata"
	cmd := exec.Command("../bin/goose-ssg", sourceDir, "--destination", destDir)

	output, cmdErr := cmd.CombinedOutput()

	if cmdErr != nil {
		t.Errorf("Error: %v\nOutput:\n%s", cmdErr, output)
	}

	if !fileExists(filepath.Join(destDir, "CNAME")) {
		t.Errorf("Static file 'CNAME' does not exist")
	}

	if !fileExists(filepath.Join(destDir, "style.css")) {
		t.Errorf("Static file 'style.css' does not exist")
	}

	if !fileExists(filepath.Join(destDir, "img", "favicon.ico")) {
		t.Errorf("Static file 'img/favicon.ico' does not exist")
	}

	if !fileExists(filepath.Join(destDir, "index.html")) {
		t.Errorf("'index.html' does not exist")
	}

	if !fileExists(filepath.Join(destDir, "about", "index.html")) {
		t.Errorf("'about/index.html' does not exist")
	}

	if !fileExists(filepath.Join(destDir, "about", "nested", "index.html")) {
		t.Errorf("'about/nested/index.html' does not exist")
	}
}
