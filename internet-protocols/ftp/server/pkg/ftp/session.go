package ftp

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
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
	ctrlConn      net.Conn
	dataServer    net.Listener
	user          *config.User
	serverRootDir string

	// Current directory of the session relative to the user's home directory
	currentDir string
}

func NewAuthenticatedSession(ctrlConn net.Conn, user *config.User, serverRootDir string) *Session {
	return &Session{
		ctrlConn:      ctrlConn,
		user:          user,
		serverRootDir: serverRootDir,

		currentDir: "/",
	}
}

func NewSession(ctrlConn net.Conn, serverRootDir string) *Session {
	return &Session{
		ctrlConn:      ctrlConn,
		serverRootDir: serverRootDir,

		currentDir: "/",
	}
}

func (s *Session) Close() {
	s.ctrlConn.Close()
	if s.dataServer != nil {
		s.dataServer.Close()
	}
}

func (s *Session) IsAuthenticated() bool {
	log.Printf("user: %v", s.user)
	return s.user.Username != "" && s.user.HomeDir != ""
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

	log.Printf("Changed directory to %s", s.currentDir)

	return nil
}

func (s *Session) ListDirectory() ([]fs.DirEntry, error) {
	dir := s.getRealPath(s.currentDir)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %s", err)
	}

	return files, nil
}

func (s *Session) RetrieveFile(file string) (fs.File, error) {
	return os.Open(s.getRealPath(file))
}

func (s *Session) StoreFile(file string, data []byte) error {
	return os.WriteFile(s.getRealPath(file), data, 0644)
}

func (s *Session) DeleteFile(file string) error {
	return os.Remove(s.getRealPath(file))
}

func (s *Session) MakeDirectory(dir string) error {
	return os.Mkdir(s.getRealPath(dir), 0755)
}

func (s *Session) RemoveDirectory(dir string) error {
	return os.Remove(s.getRealPath(dir))
}

func (s *Session) getRealPath(path string) string {
	if strings.HasPrefix(path, "/public") {
		return filepath.Join(s.serverRootDir, path)
	}

	if strings.HasPrefix(s.currentDir, "/public") {
		return filepath.Join(s.serverRootDir, s.currentDir, path)
	}

	if !strings.HasPrefix(path, "/") {
		path = filepath.Join(s.currentDir, path)
	}
	return filepath.Join(s.serverRootDir, s.user.HomeDir, path)
}
