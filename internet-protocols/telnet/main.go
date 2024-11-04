package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	IAC  = 255
	DONT = 254
	DO   = 253
	WONT = 252
	WILL = 251
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: telnet [remote_ip] [remote_port]")
		return
	}

	remoteIP := os.Args[1]
	remotePort := os.Args[2]
	address := remoteIP + ":" + remotePort

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Cannot connect to server: %s", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s\n", address)

	go handleMessages(conn)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text() + "\r\n"
		_, err := conn.Write([]byte(input))
		if err != nil {
			fmt.Println("Error sending data:", err)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from stdin:", err)
	}
}

func handleMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			fmt.Println("Connection closed by server")
			os.Exit(0)
		}
		output := negotiateOptions(conn, buffer[:n])
		fmt.Print(string(output))
	}
}

func negotiateOptions(conn net.Conn, buf []byte) []byte {
	i := 0
	output := []byte{}

	for i < len(buf) {
		if buf[i] == IAC {
			if i+2 < len(buf) {
				command := buf[i+1]
				opt := buf[i+2]

				switch command {
				case DO:
					conn.Write([]byte{IAC, WONT, opt})
				case DONT:
					conn.Write([]byte{IAC, WONT, opt})
				case WILL:
					conn.Write([]byte{IAC, DONT, opt})
				case WONT:
					conn.Write([]byte{IAC, DONT, opt})
				}
				i += 3
			} else {
				break
			}
		} else {
			output = append(output, buf[i])
			i++
		}
	}

	return output
}
