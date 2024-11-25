package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	Addr           string `json:"addr"`
	Users          []User `json:"users"`
	MinDataPort    int    `json:"min_data_port"`
	MaxDataPort    int    `json:"max_data_port"`
	AllowAnonymous bool   `json:"allow_anonymous"`
	Path           string `json:"path"`
}

var defaultConfig = &Config{
	Addr: "0.0.0.0:21",
	Users: []User{
		{
			Username: "admin",
			Password: "admin",
		},
	},
	MinDataPort:    30000,
	MaxDataPort:    30010,
	AllowAnonymous: false,
	Path:           "./tmp",
}

func loadConfig(path string) (*Config, error) {
	if path == "" {
		return defaultConfig, nil
	}

	var config *Config
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return config, nil
}

type options struct {
	configPath string
}

type FTPServer struct {
	config *Config
}

func NewFTPServer(config *Config) *FTPServer {
	return &FTPServer{config: config}
}

func (s *FTPServer) ListenAndServe() error {
	log.Printf("Starting server on %s", s.config.Addr)
	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		log.Printf("Accepted connection from %s", conn.RemoteAddr())
		go s.handleConn(conn)
	}
}

func (s *FTPServer) handleConn(conn net.Conn) {
	defer conn.Close()

	if err := s.sendResponse(conn, 220, "US FTP server ready"); err != nil {
		log.Printf("failed to send response: %v", err)
		return
	}

	var currentWorkingDirectory = s.config.Path
	var data net.Conn
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("failed to read from connection: %v", err)
			return
		}

		cmd := strings.TrimSpace(string(buf[:n]))
		log.Printf("received command: %s", cmd)

		command, arg := s.parseCommand(cmd)
		switch command {
		case "USER":
			if err := s.handleUserCommand(conn, arg); err != nil {
				log.Printf("failed to handle USER command: %v", err)
				return
			}
		case "CWD":
			newWorkingDirectory := arg

			if newWorkingDirectory == ".." {
				newWorkingDirectoryParts := strings.Split(currentWorkingDirectory, "/")
				newWorkingDirectoryParts = newWorkingDirectoryParts[:len(newWorkingDirectoryParts)-1]
				newWorkingDirectory = strings.Join(newWorkingDirectoryParts, "/")
			}

			// Check if the new working directory is a valid directory
			if _, err := os.Stat(newWorkingDirectory); err != nil {
				if err := s.sendResponse(conn, 550, "Directory not found"); err != nil {
					log.Printf("failed to send response: %v", err)
					return
				}
			}

			// Check if the new working directory is under the root directory
			if !strings.HasPrefix(newWorkingDirectory, s.config.Path) {
				if err := s.sendResponse(conn, 550, "Directory not found"); err != nil {
					log.Printf("failed to send response: %v", err)
					return
				}
			}
		case "LIST":
			if data == nil {
				if err := s.sendResponse(conn, 425, "Use PASV first"); err != nil {
					log.Printf("failed to send response: %v", err)
					return
				}
			}

			if err := s.sendResponse(conn, 150, "Here comes the directory listing"); err != nil {
				log.Printf("failed to send response: %v", err)
				return
			}

			files, err := os.ReadDir(currentWorkingDirectory)
			if err != nil {
				log.Printf("failed to read directory: %v", err)
				return
			}

			for _, file := range files {
				if _, err := data.Write([]byte(file.Name() + "\r\n")); err != nil {
					log.Printf("failed to write to data connection: %v", err)
					return
				}
			}

			if err := s.sendResponse(conn, 226, "Directory send OK"); err != nil {
				log.Printf("failed to send response: %v", err)
				return
			}

			data.Close()
		case "PASV":
			dataConn, port, err := s.openDataConnection()
			if err != nil {
				log.Printf("failed to open data connection: %v", err)
				return
			}

			ip := strings.Replace(strings.Split(conn.LocalAddr().String(), ":")[0], ".", ",", -1)
			p1 := port / 256
			p2 := port % 256

			log.Printf("Port is: %d", (p1<<8)+p2)
			if err := s.sendResponse(conn, 227, fmt.Sprintf("Entering Passive Mode (%s,%d,%d)", ip, p1, p2)); err != nil {
				log.Printf("failed to send response: %v", err)
				return
			}
			data, err = dataConn.Accept()
			if err != nil {
				log.Printf("failed to accept data connection: %v", err)
				return
			}
		case "PWD":
			if err := s.sendResponse(conn, 257, fmt.Sprintf("\"%s\" is the current directory", currentWorkingDirectory)); err != nil {
				log.Printf("failed to send response: %v", err)
				return
			}
		case "QUIT":
			if err := s.sendResponse(conn, 221, "Goodbye"); err != nil {
				log.Printf("failed to send response: %v", err)
				return
			}
		default:
			conn.Write([]byte("502 Command not implemented.\r\n"))
		}
	}
}

func (s *FTPServer) handleUserCommand(conn net.Conn, arg string) error {
	fmt.Println(arg, arg == "anonymous")
	if s.config.AllowAnonymous && arg == "anonymous" {
		err := s.sendResponse(conn, 230, "User logged in")
		if err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}
		return nil
	}
	if !s.config.AllowAnonymous && arg == "anonymous" {
		err := s.sendResponse(conn, 530, "User cannot be anonymous")
		if err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}
		return nil
	}

	username := arg
	err := s.sendResponse(conn, 331, "Password required")
	if err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read from connection: %w", err)
	}

	cmd, arg := s.parseCommand(strings.TrimSpace(string(buf[:n])))
	if cmd != "PASS" {
		return fmt.Errorf("expected PASS command, got %s", cmd)
	}
	password := arg

	if !s.isValidUser(username, password) {
		err := s.sendResponse(conn, 530, "Invalid username or password")
		if err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}
	} else {
		err := s.sendResponse(conn, 230, "User logged in")
		if err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}
	}

	return nil
}

func (s *FTPServer) openDataConnection() (net.Listener, int, error) {
	for port := s.config.MinDataPort; port <= s.config.MaxDataPort; port++ {
		addrWithoutPort := strings.Split(s.config.Addr, ":")[0]
		addr := fmt.Sprintf("%s:%d", addrWithoutPort, port)

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Printf("failed to listen on port %d: %v", port, err)
			continue
		}

		return listener, port, nil
	}

	return nil, -1, fmt.Errorf("no available ports")
}

func (s *FTPServer) sendResponse(conn net.Conn, code int, message string) error {
	log.Printf("sending response: %d %s", code, message)
	_, err := conn.Write([]byte(fmt.Sprintf("%d %s\r\n", code, message)))
	return err
}

func (s *FTPServer) parseCommand(cmd string) (string, string) {
	parts := strings.SplitN(cmd, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}

	log.Printf("parts: %v", parts)
	return parts[0], parts[1]
}

func (s *FTPServer) isValidUser(username, password string) bool {
	for _, user := range s.config.Users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

func gatherOptions(opts *options) {
	flag.StringVar(&opts.configPath, "config", "", "path to the config file")
}

func main() {
	o := options{}
	gatherOptions(&o)
	flag.Parse()

	config, err := loadConfig(o.configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	server := NewFTPServer(config)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

}
