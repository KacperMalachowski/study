package ftp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KacperMalachowski/study/internet-protocols/ftp/server/pkg/config"
)

type FileInfo struct {
	Name string
	Data []byte
	Mode os.FileMode
}

func TestGetCurrentDirector(t *testing.T) {
	tmp := t.TempDir()
	s := NewAuthenticatedSession(nil, &config.User{HomeDir: "/"}, tmp)

	if s.GetCurrentDirectory() != "/" {
		t.Errorf("expected %s, got %s", tmp, s.GetCurrentDirectory())
	}
}

func TestChangeDirectoryUp(t *testing.T) {
	tmp := t.TempDir()
	s := NewAuthenticatedSession(nil, &config.User{HomeDir: "/"}, tmp)

	if err := s.ChangeDirectoryUp(); err != nil {
		t.Errorf("expected no error, got %s", err)
	}

	if s.GetCurrentDirectory() != "/" {
		t.Errorf("expected /, got %s", s.GetCurrentDirectory())
	}
}

func TestChangeDirectory(t *testing.T) {
	tc := []struct {
		name      string
		fs        map[string]FileInfo
		dir       string
		expected  string
		expectErr bool
	}{
		{
			name: "relative path",
			fs: map[string]FileInfo{
				"test/": {},
			},
			dir:      "test",
			expected: "/test",
		},
		{
			name: "absolute path",
			fs: map[string]FileInfo{
				"test/": {},
			},
			dir:      "/test",
			expected: "/test",
		},
		{
			name: "parent directory",
			fs: map[string]FileInfo{
				"test/": {},
			},
			dir:      "..",
			expected: "/",
		},
		{
			name: "sibling directory",
			fs: map[string]FileInfo{
				"test/":  {},
				"test2/": {},
			},
			dir:      "test/../test2",
			expected: "/test2",
		},
		{
			name: "sibling beyond home directory",
			fs: map[string]FileInfo{
				"test/": {},
			},
			dir:       "../../test2",
			expected:  "/test2",
			expectErr: true,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			tmp := t.TempDir()
			s := NewAuthenticatedSession(nil, &config.User{HomeDir: tmp}, tmp)

			for name, file := range c.fs {
				if err := os.MkdirAll(filepath.Join(tmp, name), 0755); err != nil {
					t.Fatalf("error creating directory %s: %s", name, err)
				}

				if file.Data != nil {
					if err := os.WriteFile(filepath.Join(tmp, name), file.Data, file.Mode); err != nil {
						t.Fatalf("error creating file %s: %s", name, err)
					}
				}
			}

			err := s.ChangeDirectory(c.dir)
			if c.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !c.expectErr && err != nil {
				t.Errorf("expected no error, got %s", err)
			}

			if s.GetCurrentDirectory() != c.expected && !c.expectErr {
				t.Errorf("expected %s, got %s", c.expected, s.GetCurrentDirectory())
			}
		})
	}
}
