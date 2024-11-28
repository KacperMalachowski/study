package ftp

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/KacperMalachowski/study/internet-protocols/ftp/server/pkg/config"
)

type Server struct {
	address        string
	port           int
	users          config.Users
	minPassivePort int
	maxPassivePort int
	allowAnonymous bool
	rootDir        string
	exit           chan struct{}

	sessions map[string]*Session
}

func NewServer(address string, port int, users config.Users, minPassivePort, maxPassivePort int, allowAnonymous bool, rootDir string) *Server {
	return &Server{
		address:        address,
		port:           port,
		users:          users,
		minPassivePort: minPassivePort,
		maxPassivePort: maxPassivePort,
		allowAnonymous: allowAnonymous,
		rootDir:        rootDir,
		exit:           make(chan struct{}),
	}
}

func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", s.address, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	defer listener.Close()

	log.Printf("Listening on %s", addr)

	go func() {
		<-s.exit
		log.Printf("Shutting down server")
		return
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		log.Printf("Accepted connection from %s", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) Close() {
	s.exit <- struct{}{}
}

func (s *Server) handleConnection(conn net.Conn) {
	session := NewSession(conn, s.rootDir)
	defer session.Close()
	defer conn.Close()

	s.sendResponse(session, 220, "Service ready for new user")

	for {
		command, err := s.readCommand(session)
		if err != nil {
			log.Printf("Failed to read command: %s", err)
			return
		}

		s.handleCommand(session, command)
	}
}

func (s *Server) readCommand(session *Session) (string, error) {
	buf := make([]byte, 1024)
	n, err := session.ctrlConn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read command: %w", err)
	}

	command := string(buf[:n])
	log.Printf("Received command: %s", command)
	return command, nil
}

func (s *Server) sendResponse(session *Session, code int, message string) {
	response := fmt.Sprintf("%d %s\r\n", code, message)
	log.Printf("Sending response: %s", response)
	session.ctrlConn.Write([]byte(response))
}

func (s *Server) handleCommand(session *Session, command string) {
	command = strings.TrimSpace(command)
	parts := strings.SplitN(command, " ", 2)
	cmd := strings.ToUpper(parts[0])
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	switch cmd {
	case "USER":
		s.handleUserCommand(session, args)
	case "PASS":
		s.handlePassCommand(session, args)
	case "SYST":
		s.sendResponse(session, 215, "UNIX Type: L8")
	case "TYPE":
		s.sendResponse(session, 200, "Type set")
	case "PWD":
		s.handlePwdCommand(session)
	case "CWD":
		s.handleCwdCommand(session, args)
	case "PASV":
		s.handlePasvCommand(session)
	case "LIST":
		s.handleListCommand(session, args)
	case "EPSV":
		s.handleEpsvCommand(session)
	case "CDUP":
		s.handleCwdCommand(session, "..")
	case "RETR":
		s.handleRetrCommand(session, args)
	case "STOR":
		s.handleStorCommand(session, args)
	case "MKD":
		s.handleMkdCommand(session, args)
	case "RMD":
		s.handleRmdCommand(session, args)
	case "DELE":
		s.handleDeleCommand(session, args)
	case "QUIT":
		s.handleQuitCommand(session)
	default:
		s.sendResponse(session, 502, "Command not implemented")
	}
}

func (s *Server) handleUserCommand(session *Session, username string) {
	if s.allowAnonymous && username == "anonymous" {
		session.user = &config.User{
			Username: "anonymous",
			HomeDir:  "/public",
		}

		s.sendResponse(session, 230, "User logged in, proceed")
	} else {
		user, ok := s.users.FindByUsername(username)
		if !ok {
			s.sendResponse(session, 530, "Invalid username")
			return
		}

		session.user = user
		s.sendResponse(session, 331, "User name okay, need password")
	}
}

func (s *Server) handlePassCommand(session *Session, password string) {
	if session.user == nil {
		s.sendResponse(session, 503, "Login with USER first")
		return
	}

	if session.user.Password != password {
		s.sendResponse(session, 530, "Invalid password")
		return
	}

	s.sendResponse(session, 230, "User logged in, proceed")
}

func (s *Server) handlePwdCommand(session *Session) {
	if !session.IsAuthenticated() {
		s.sendResponse(session, 530, "Not logged in")
		return
	}

	s.sendResponse(session, 257, fmt.Sprintf("\"%s\" is the current directory", session.GetCurrentDirectory()))
}

func (s *Server) handlePasvCommand(session *Session) {
	if !session.IsAuthenticated() {
		s.sendResponse(session, 530, "Not logged in")
		return
	}

	if session.dataServer != nil {
		session.dataServer.Close()
	}

	minPort := s.minPassivePort
	maxPort := s.maxPassivePort

	listener, err := s.listenPassiveDataConnection(minPort, maxPort)
	if err != nil {
		s.sendResponse(session, 421, "Failed to open data connection")
		return
	}

	session.dataServer = listener

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		s.sendResponse(session, 421, "Failed to open data connection")
		return
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		s.sendResponse(session, 421, "Failed to open data connection")
		return
	}

	ipParts := strings.Split(host, ".")
	p1Int := portInt / 256
	p2Int := portInt % 256

	s.sendResponse(session, 227, fmt.Sprintf("Entering Passive Mode (%s,%s,%s,%s,%d,%d)", ipParts[0], ipParts[1], ipParts[2], ipParts[3], p1Int, p2Int))
}

func (s *Server) handleEpsvCommand(session *Session) {
	if !session.IsAuthenticated() {
		s.sendResponse(session, 530, "Not logged in")
		return
	}

	if session.dataServer != nil {
		session.dataServer.Close()
	}

	minPort := s.minPassivePort
	maxPort := s.maxPassivePort

	listener, err := s.listenPassiveDataConnection(minPort, maxPort)
	if err != nil {
		s.sendResponse(session, 421, "Failed to open data connection")
		return
	}

	session.dataServer = listener

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		s.sendResponse(session, 421, "Failed to open data connection")
		return
	}

	s.sendResponse(session, 229, fmt.Sprintf("Entering Extended Passive Mode (|||%s|)", port))
}

func (s *Server) listenPassiveDataConnection(minPort, maxPort int) (net.Listener, error) {
	for port := minPort; port <= maxPort; port++ {
		addr := fmt.Sprintf("%s:%d", s.address, port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			return listener, nil
		}
	}

	return nil, fmt.Errorf("failed to listen on any port in range %d-%d", minPort, maxPort)
}

func (s *Server) handleListCommand(session *Session, _ string) {
	if !session.IsAuthenticated() {
		s.sendResponse(session, 530, "Not logged in")
		return
	}

	if session.dataServer == nil {
		s.sendResponse(session, 425, "Use PASV first")
		return
	}

	s.sendResponse(session, 150, "Here comes the directory listing")

	conn, err := session.dataServer.Accept()
	if err != nil {
		s.sendResponse(session, 425, "Failed to open data connection")
		return
	}
	defer conn.Close()

	files, err := session.ListDirectory()
	if err != nil {
		log.Printf("Failed to list directory: %s", err)
		s.sendResponse(session, 550, "Failed to list directory")
		return
	}

	for _, file := range files {
		prefix := "-rw-r--r--"
		if file.IsDir() {
			prefix = "drwxr-xr-x"
		}
		if file.Type() == fs.ModeSymlink {
			prefix = "lrwxr-xr-x"
		}

		infor, err := file.Info()
		if err != nil {
			log.Printf("Failed to get file info: %s", err)
			s.sendResponse(session, 550, "Failed to list directory")
			return
		}

		formatted := fmt.Sprintf("%s 1 ftp ftp %13d %s %s\r\n", prefix, infor.Size(), infor.ModTime().Format("Jan 02 15:04"), file.Name())
		conn.Write([]byte(formatted))
	}

	s.sendResponse(session, 226, "Directory send OK")
}

func (s *Server) handleCwdCommand(session *Session, dir string) {
	if !session.IsAuthenticated() {
		s.sendResponse(session, 530, "Not logged in")
		return
	}

	if err := session.ChangeDirectory(dir); err != nil {
		if err == ErrDirectoryNotFound {
			s.sendResponse(session, 550, "Directory not found")
			return
		}

		if err == ErrNotDirectory {
			s.sendResponse(session, 550, "Not a directory")
			return
		}

		s.sendResponse(session, 550, "Failed to change directory")
		return
	}

	s.sendResponse(session, 250, "Directory changed")
}

func (s *Server) handleRetrCommand(session *Session, file string) {
	if !session.IsAuthenticated() {
		s.sendResponse(session, 530, "Not logged in")
		return
	}

	if session.dataServer == nil {
		s.sendResponse(session, 425, "Use PASV first")
		return
	}

	s.sendResponse(session, 150, "Opening data connection")

	conn, err := session.dataServer.Accept()
	if err != nil {
		s.sendResponse(session, 425, "Failed to open data connection")
		return
	}
	defer conn.Close()

	fd, err := session.RetrieveFile(file)
	if err != nil {
		s.sendResponse(session, 550, "Failed to get file")
		return
	}
	defer fd.Close()

	data, err := io.ReadAll(fd)
	if err != nil {
		s.sendResponse(session, 550, "Failed to read file")
		return
	}

	conn.Write(data)

	s.sendResponse(session, 226, "File send OK")
}

func (s *Server) handleStorCommand(session *Session, file string) {
	if !session.IsAuthenticated() {
		s.sendResponse(session, 530, "Not logged in")
		return
	}

	if session.dataServer == nil {
		s.sendResponse(session, 425, "Use PASV first")
		return
	}

	s.sendResponse(session, 150, "Opening data connection")

	conn, err := session.dataServer.Accept()
	if err != nil {
		s.sendResponse(session, 425, "Failed to open data connection")
		return
	}
	defer conn.Close()

	data, err := io.ReadAll(conn)
	if err != nil {
		s.sendResponse(session, 550, "Failed to read data")
		return
	}

	if err := session.StoreFile(file, data); err != nil {
		s.sendResponse(session, 550, "Failed to store file")
		return
	}

	s.sendResponse(session, 226, "File stored OK")
}

func (s *Server) handleMkdCommand(session *Session, dir string) {
	if !session.IsAuthenticated() {
		s.sendResponse(session, 530, "Not logged in")
		return
	}

	if err := session.MakeDirectory(dir); err != nil {
		s.sendResponse(session, 550, "Failed to make directory")
		return
	}

	s.sendResponse(session, 257, "Directory created")
}

func (s *Server) handleRmdCommand(session *Session, dir string) {
	if !session.IsAuthenticated() {
		s.sendResponse(session, 530, "Not logged in")
		return
	}

	if err := session.DeleteFile(dir); err != nil {
		s.sendResponse(session, 550, "Failed to delete directory")
		return
	}

	s.sendResponse(session, 250, "Directory deleted")
}

func (s *Server) handleDeleCommand(session *Session, file string) {
	if !session.IsAuthenticated() {
		s.sendResponse(session, 530, "Not logged in")
		return
	}

	if err := session.DeleteFile(file); err != nil {
		s.sendResponse(session, 550, "Failed to delete file")
		return
	}

	s.sendResponse(session, 250, "File deleted")
}

func (s *Server) handleQuitCommand(session *Session) {
	s.sendResponse(session, 221, "Goodbye")
	session.Close()
}
