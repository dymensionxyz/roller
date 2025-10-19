package filesystem

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a malicious tar.gz file for testing
func createTestTarGz(t *testing.T, entries map[string]string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-*.tar.gz")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	gzw := gzip.NewWriter(tmpFile)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	for name, content := range entries {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if content == "" { // Directory
			hdr.Typeflag = tar.TypeDir
			hdr.Mode = 0o755
			hdr.Size = 0
		} else {
			hdr.Typeflag = tar.TypeReg
		}

		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("Failed to write header: %v", err)
		}

		if content != "" {
			if _, err := tw.Write([]byte(content)); err != nil {
				t.Fatalf("Failed to write content: %v", err)
			}
		}
	}

	return tmpFile.Name()
}

func TestExtractTarGz_PathTraversal(t *testing.T) {
	tests := []struct {
		name        string
		entries     map[string]string
		shouldFail  bool
		description string
	}{
		{
			name: "legitimate data directory",
			entries: map[string]string{
				"data":          "",
				"data/file.txt": "legitimate content",
			},
			shouldFail:  false,
			description: "Should extract legitimate files in data directory",
		},
		{
			name: "path traversal with ../",
			entries: map[string]string{
				"data/../../../etc/malicious.txt": "malicious content",
			},
			shouldFail:  true,
			description: "Should reject path traversal using ../",
		},
		{
			name: "path traversal within data prefix",
			entries: map[string]string{
				"data/../../outside.txt": "malicious content",
			},
			shouldFail:  true,
			description: "Should reject path traversal from within data directory",
		},
		{
			name: "absolute path",
			entries: map[string]string{
				"/etc/passwd": "malicious content",
			},
			shouldFail:  false,
			description: "Absolute paths without data/ prefix are silently skipped (no extraction occurs)",
		},
		{
			name: "nested data directory",
			entries: map[string]string{
				"data":                   "",
				"data/subdir":            "",
				"data/subdir/nested.txt": "nested content",
			},
			shouldFail:  false,
			description: "Should handle nested directories correctly",
		},
		{
			name: "path with dot segments",
			entries: map[string]string{
				"data/./file.txt": "content",
			},
			shouldFail:  false,
			description: "Should handle path with ./ segments",
		},
		{
			name: "symlink escape attempt",
			entries: map[string]string{
				"data":          "",
				"data/link.txt": "regular file, not symlink in this test",
			},
			shouldFail:  false,
			description: "Regular files should work fine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directories
			tmpDir, err := os.MkdirTemp("", "extract-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create test tar.gz file
			tarPath := createTestTarGz(t, tt.entries)
			defer os.Remove(tarPath)

			// Attempt extraction
			err = ExtractTarGz(tarPath, tmpDir)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.description)

					// Check if malicious file was created outside the destination
					for name := range tt.entries {
						target := filepath.Join(tmpDir, name)
						cleanTarget := filepath.Clean(target)
						cleanDest := filepath.Clean(tmpDir)

						// If the file exists outside the destination directory, that's a security issue
						if !filepath.HasPrefix(cleanTarget, cleanDest) {
							if _, statErr := os.Stat(target); statErr == nil {
								t.Errorf("Security breach: file was created outside destination: %s", target)
							}
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				}

				// Verify legitimate files were extracted correctly
				for name, expectedContent := range tt.entries {
					if expectedContent == "" {
						continue // Skip directories
					}

					// Only verify files that should have been extracted (data/ prefix)
					if name != "data" && !filepath.HasPrefix(name, "data/") {
						continue // Skip files that were filtered out
					}

					extractedPath := filepath.Join(tmpDir, name)
					content, readErr := os.ReadFile(extractedPath)
					if readErr != nil {
						t.Errorf("Failed to read extracted file %s: %v", name, readErr)
						continue
					}

					if string(content) != expectedContent {
						t.Errorf("Content mismatch for %s: got %q, want %q", name, string(content), expectedContent)
					}
				}
			}
		})
	}
}

func TestExtractTarGz_NonExistentFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extract-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = ExtractTarGz("/non/existent/file.tar.gz", tmpDir)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestExtractTarGz_InvalidGzip(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid-*.tar.gz")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid gzip content
	_, err = tmpFile.Write([]byte("not a gzip file"))
	tmpFile.Close()
	if err != nil {
		t.Fatalf("Failed to write invalid content: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "extract-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = ExtractTarGz(tmpFile.Name(), tmpDir)
	if err == nil {
		t.Error("Expected error for invalid gzip file")
	}
}
