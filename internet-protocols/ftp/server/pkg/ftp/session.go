package ftp

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/KacperMalachowski/study/internet-protocols/ftp/server/pkg/config"
)

var (
	ErrDirectoryNotFound      = errors.New("directory not found")
	ErrNotDirectory           = errors.New("not a directory")
	ErrAuthenticaitonRequired = errors.New("authentication required")
)

type Session struct {
	ctrlConn   net.Conn
	dataServer net.Listener
	user       config.User

	// Current directory of the session relative to the user's home directory
	currentDir string
}

func NewAuthenticatedSession(ctrlConn net.Conn, user config.User) *Session {
	return &Session{
		ctrlConn: ctrlConn,
		user:     user,

		currentDir: "/",
	}
}

func NewSession(ctrlConn net.Conn) *Session {
	return &Session{
		ctrlConn: ctrlConn,

		currentDir: "/",
	}
}

func (s *Session) Close() {
	s.ctrlConn.Close()
	if s.dataServer != nil {
		s.dataServer.Close()
	}
}

func (s *Session) GetCurrentDirectory() string {
	return s.currentDir
}

func (s *Session) ChangeDirectoryUp() error {
	if s.currentDir == "/" {
		return nil
	}

	// Move to the parent directory
	s.currentDir = filepath.Dir(s.currentDir)

	return nil
}

func (s *Session) ChangeDirectory(dir string) error {
	if !strings.HasPrefix(dir, "/") {
		dir = filepath.Join(s.currentDir, dir)
	}

	newDirectory := filepath.Clean(dir)

	// Check if the directory exists
	info, err := os.Stat(s.getRealPath(newDirectory))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrDirectoryNotFound
		}

		return fmt.Errorf("error checking directory: %s", err)
	}

	if !info.IsDir() {
		return ErrNotDirectory
	}

	s.currentDir = newDirectory

	return nil
}

func (s *Session) getRealPath(path string) string {
	return filepath.Join(s.user.HomeDir, path)
}
