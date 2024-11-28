package ftp

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/KacperMalachowski/study/internet-protocols/ftp/server/pkg/config"
)

func startServerForTests() (net.Conn, net.Conn) {
	server, client := net.Pipe()

	return client, server
}

func readFromServer(server net.Conn) string {
	buf := make([]byte, 1024)
	n, _ := server.Read(buf)
	return string(buf[:n])
}

func writeToServer(server net.Conn, data string) {
	server.Write([]byte(data + "\r\n"))
}

type Command struct {
	Name             string
	Args             string
	expectedResponse func(t *testing.T, data string)
}

func TestServer(t *testing.T) {
	tc := []struct {
		name     string
		commands []Command
	}{
		{
			name: "valid user authentication",
			commands: []Command{
				{
					Name: "USER",
					Args: "test",
					expectedResponse: func(t *testing.T, data string) {
						if !strings.HasPrefix(data, "331") {
							t.Errorf("expected 331 User name okay, need password, got %s", data)
						}
					},
				},
				{
					Name: "PASS",
					Args: "test",
					expectedResponse: func(t *testing.T, data string) {
						if !strings.HasPrefix(data, "230") {
							t.Errorf("expected 230 User logged in, proceed, got %s", data)
						}
					},
				},
			},
		},
		{
			name: "invalid user authentication",
			commands: []Command{
				{
					Name: "USER",
					Args: "test1",
					expectedResponse: func(t *testing.T, data string) {
						if !strings.HasPrefix(data, "530") {
							t.Errorf("expected 530 Invalid username, got %s", data)
						}
					},
				},
			},
		},
		{
			name: "invalid password",
			commands: []Command{
				{
					Name: "USER",
					Args: "test",
					expectedResponse: func(t *testing.T, data string) {
						if !strings.HasPrefix(data, "331") {
							t.Errorf("expected 331 User name okay, need password, got %s", data)
						}
					},
				},
				{
					Name: "PASS",
					Args: "test1",
					expectedResponse: func(t *testing.T, data string) {
						if !strings.HasPrefix(data, "530") {
							t.Errorf("expected 530 Invalid password, got %s", data)
						}
					},
				},
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			server, client := startServerForTests()

			users := config.Users{
				config.User{
					Username: "test",
					Password: "test",
					HomeDir:  "/",
				},
			}
			s := NewServer("localhost", 2121, users, 0, 0, false, "./tmp")

			go s.handleConnection(server)

			data := readFromServer(client)
			if !strings.HasPrefix(data, "220") {
				t.Errorf("expected 220 Service ready for new user, got %s", data)
			}

			for _, cmd := range c.commands {
				writeToServer(client, fmt.Sprintf("%s %s", cmd.Name, cmd.Args))
				data = readFromServer(client)
				cmd.expectedResponse(t, data)
			}
		})
	}
}
