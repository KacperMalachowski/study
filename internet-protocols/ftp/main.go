package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// Flags
	host       string = "localhost"
	port       int    = 21
	user       string
	password   string
	sourcePath string
	destPath   string

	// ls command flags
	printDetails bool
)

var rootCmd = &cobra.Command{
	Use:   "usftp",
	Short: "usftp is a simple FTP client",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var mkdirCmd = &cobra.Command{
	Use:   "mkdir",
	Short: "Create directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if sourcePath == "" {
			return fmt.Errorf("path is required")
		}

		if isRemotePath(sourcePath) {
			path, err := parseRemotePath(sourcePath)
			if err != nil {
				return nil
			}

			client := NewFTPClient(host, port, user, password)
			if err := client.Connect(); err != nil {
				return err
			}

			if err := client.MakeDirectory(path); err != nil {
				return err
			}

			if err := client.Close(); err != nil {
				return err
			}
		} else {
			if err := createDirectory(sourcePath); err != nil {
				return err
			}
		}

		return nil
	},
}

var rmDirCmd = &cobra.Command{
	Use:   "rmdir",
	Short: "Remove directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if sourcePath == "" {
			return fmt.Errorf("path is required")
		}

		if isRemotePath(sourcePath) {
			path, err := parseRemotePath(sourcePath)
			if err != nil {
				return err
			}

			client := NewFTPClient(host, port, user, password)
			if err := client.Connect(); err != nil {
				return err
			}

			if err := client.RemoveDirectory(path); err != nil {
				return err
			}

			if err := client.Close(); err != nil {
				return err
			}
		} else {
			if err := RemoveDirectory(sourcePath); err != nil {
				return err
			}
		}

		return nil
	},
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List files",
	RunE: func(cmd *cobra.Command, args []string) error {
		if sourcePath == "" {
			return fmt.Errorf("path is required")
		}

		if isRemotePath(sourcePath) {
			path, err := parseRemotePath(sourcePath)
			if err != nil {
				return err
			}

			client := NewFTPClient(host, port, user, password)
			if err := client.Connect(); err != nil {
				return err
			}

			if err := client.List(path); err != nil {
				return err
			}
		} else {

			files, err := listLocalFiles(sourcePath)
			if err != nil {
				return err
			}

			if printDetails {
				for _, file := range files {
					fmt.Println(file.Mode(), file.Size(), file.ModTime().Format("Jan 2 15:04"), file.Name())
				}
			} else {
				result := []string{}
				for _, file := range files {
					result = append(result, file.Name())
				}

				fmt.Println(strings.Join(result, " "))
			}
		}

		return nil
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if sourcePath == "" {
			return fmt.Errorf("path is required")
		}

		if isRemotePath(sourcePath) {
			path, err := parseRemotePath(sourcePath)
			if err != nil {
				return err
			}

			client := NewFTPClient(host, port, user, password)
			if err := client.Connect(); err != nil {
				return err
			}

			if err := client.DeleteFile(path); err != nil {
				return err
			}

			if err := client.Close(); err != nil {
				return err
			}

		} else {
			if err := os.Remove(sourcePath); err != nil {
				return err
			}
		}

		return nil
	},
}

var copyCmd = &cobra.Command{
	Use:   "cp",
	Short: "Copy file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if sourcePath == "" || destPath == "" {
			return fmt.Errorf("source and destination paths are required")
		}

		if isRemotePath(sourcePath) && isRemotePath(destPath) {
			source, err := parseRemotePath(sourcePath)
			if err != nil {
				return err
			}

			srcClient := NewFTPClient(host, port, user, password)
			if err := srcClient.Connect(); err != nil {
				return err
			}
			defer srcClient.Close()

			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			temp, err := os.MkdirTemp(dir, "temp")
			if err != nil {
				return err
			}
			defer os.RemoveAll(temp)

			if err := srcClient.DownloadFile(source, temp); err != nil {
				return err
			}

			dest, err := parseRemotePath(destPath)
			if err != nil {
				return err
			}

			destClient := NewFTPClient(host, port, user, password)
			if err := srcClient.Connect(); err != nil {
				return err
			}
			defer destClient.Close()

			if err := destClient.UploadFile(temp, dest); err != nil {
				return err
			}
		} else if !isRemotePath(sourcePath) && !isRemotePath(destPath) {
			file, err := os.ReadFile(sourcePath)
			if err != nil {
				return err
			}

			err = os.WriteFile(destPath, file, 0644)
			if err != nil {
				return err
			}
		} else if isRemotePath(sourcePath) && !isRemotePath(destPath) {
			source, err := parseRemotePath(sourcePath)
			if err != nil {
				return err
			}

			client := NewFTPClient(host, port, user, password)
			if err := client.Connect(); err != nil {
				return err
			}

			if err := client.DownloadFile(source, destPath); err != nil {
				return err
			}

			if err := client.Close(); err != nil {
				return err
			}

		} else if !isRemotePath(sourcePath) && isRemotePath(destPath) {
			dest, err := parseRemotePath(destPath)
			if err != nil {
				return err
			}

			client := NewFTPClient(host, port, user, password)
			if err := client.Connect(); err != nil {
				return err
			}

			if err := client.UploadFile(sourcePath, dest); err != nil {
				return err
			}

			if err := client.Close(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("source and destination paths must be both local or remote")
		}

		return nil
	},
}

var moveCmd = &cobra.Command{
	Use:   "mv",
	Short: "Move file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if sourcePath == "" || destPath == "" {
			return fmt.Errorf("source and destination paths are required")
		}

		if isRemotePath(sourcePath) && isRemotePath(destPath) {
			source, err := parseRemotePath(sourcePath)
			if err != nil {
				return err
			}

			srcClient := NewFTPClient(host, port, user, password)
			if err := srcClient.Connect(); err != nil {
				return err
			}
			defer srcClient.Close()

			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			temp, err := os.MkdirTemp(dir, "temp")
			if err != nil {
				return err
			}
			defer os.RemoveAll(temp)

			if err := srcClient.DownloadFile(source, temp); err != nil {
				return err
			}

			dest, err := parseRemotePath(destPath)
			if err != nil {
				return err
			}

			destClient := NewFTPClient(host, port, user, password)
			if err := srcClient.Connect(); err != nil {
				return err
			}
			defer destClient.Close()

			if err := destClient.UploadFile(temp, dest); err != nil {
				return err
			}

			if err := srcClient.DeleteFile(source); err != nil {
				return err
			}
		} else if !isRemotePath(sourcePath) && !isRemotePath(destPath) {
			err := os.Rename(sourcePath, destPath)
			if err != nil {
				return err
			}
		} else if isRemotePath(sourcePath) && !isRemotePath(destPath) {
			source, err := parseRemotePath(sourcePath)
			if err != nil {
				return err
			}

			client := NewFTPClient(host, port, user, password)
			if err := client.Connect(); err != nil {
				return err
			}
			defer client.Close()

			if err := client.DownloadFile(source, destPath); err != nil {
				return err
			}

			if err := client.DeleteFile(source); err != nil {
				return err
			}
		} else if !isRemotePath(sourcePath) && isRemotePath(destPath) {
			dest, err := parseRemotePath(destPath)
			if err != nil {
				return err
			}

			client := NewFTPClient(host, port, user, password)
			if err := client.Connect(); err != nil {
				return err
			}

			if err := client.UploadFile(sourcePath, dest); err != nil {
				return err
			}

			if err := os.Remove(sourcePath); err != nil {
				return err
			}

			if err := client.Close(); err != nil {
				return err
			}
		}

		return nil
	},
}

func listLocalFiles(rootPath string) ([]os.FileInfo, error) {
	var files []os.FileInfo
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		fileInfo, err := entry.Info()
		if err != nil {
			return nil, err
		}
		files = append(files, fileInfo)
	}

	return files, nil
}

// isRemotePath checks if the path is a remote path
func isRemotePath(path string) bool {
	return strings.HasPrefix(path, "ftp://")
}

// parseRemotePath parses the remote path and returns the path without the protocol
func parseRemotePath(path string) (string, error) {
	if !isRemotePath(path) {
		return "", fmt.Errorf("path is not a remote path")
	}

	trimmedPath := strings.TrimPrefix(path, "ftp://")
	if strings.Contains(trimmedPath, "@") {
		pathParts := strings.Split(trimmedPath, "@")
		if len(pathParts) != 2 {
			return "", fmt.Errorf("invalid remote path")
		}

		creds := pathParts[0]

		credsParts := strings.Split(creds, ":")
		if len(credsParts) != 2 {
			return "", fmt.Errorf("invalid credentials")
		}

		user = credsParts[0]
		password = credsParts[1]

		trimmedPath = pathParts[1]
	}

	parts := strings.Split(trimmedPath, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid remote path")
	}

	hostPort := parts[0]
	hostPortParts := strings.Split(hostPort, ":")

	if len(hostPortParts) == 2 {
		var err error
		host = hostPortParts[0]
		port, err = strconv.Atoi(hostPortParts[1])
		if err != nil {
			return "", fmt.Errorf("invalid port")
		}
	} else {
		host = hostPort
	}

	return strings.Join(parts[1:], "/"), nil
}

func createDirectory(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func RemoveDirectory(path string) error {
	return os.RemoveAll(path)
}

type FTPClient struct {
	conn     net.Conn
	dataConn net.Conn
	host     string
	port     int
	user     string
	password string
}

func NewFTPClient(host string, port int, user string, password string) *FTPClient {
	return &FTPClient{
		host:     host,
		port:     port,
		user:     user,
		password: password,
	}
}

func (c *FTPClient) Connect() error {
	address := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to FTP server: %v", err)
	}
	c.conn = conn

	// Read server's welcome message
	welcomeMsg := make([]byte, 1024)
	_, err = conn.Read(welcomeMsg)
	if err != nil {
		return fmt.Errorf("failed to read welcome message: %v", err)
	}
	log.Printf("Welcome message: %s\n", string(welcomeMsg))

	// Send USER command
	err = c.sendCmd(fmt.Sprintf("USER %s", c.user))
	if err != nil {
		return fmt.Errorf("failed to send USER command: %v", err)
	}

	// Read server's response to USER command
	code, message, err := c.readCmdResponse()
	if code >= 400 {
		return fmt.Errorf("failed to authenticate: %s", message)
	} else if code == 331 {
		// Send PASS command
		err = c.sendCmd(fmt.Sprintf("PASS %s", c.password))
		if err != nil {
			return fmt.Errorf("failed to send PASS command: %v", err)
		}

		// Read server's response to PASS command
		code, message, err = c.readCmdResponse()
		if err != nil {
			return fmt.Errorf("failed to read response to PASS command: %v", err)
		}
		if code >= 400 {
			return fmt.Errorf("failed to authenticate: %s", message)
		}
	}

	// Setup settings
	// Set type to 8-bit binary
	err = c.sendCmd("TYPE I")
	if err != nil {
		return fmt.Errorf("failed to send TYPE command: %v", err)
	}

	// Read server's response to TYPE command
	code, message, err = c.readCmdResponse()
	if err != nil {
		return fmt.Errorf("failed to parse response to TYPE command: %v", err)
	}
	if code >= 400 {
		return fmt.Errorf("failed to set type: %s", message)
	}

	// Set stream mode
	err = c.sendCmd("MODE S")
	if err != nil {
		return fmt.Errorf("failed to send MODE command: %v", err)
	}

	// Read server's response to MODE command
	code, message, err = c.readCmdResponse()
	if err != nil {
		return fmt.Errorf("failed to parse response to TYPE command: %v", err)
	}
	if code >= 400 {
		return fmt.Errorf("failed to set type: %s", message)
	}

	// Set transfer mode to file oriented
	err = c.sendCmd("STRU F")
	if err != nil {
		return fmt.Errorf("failed to send STRU command: %v", err)
	}

	// Read server's response to STRU command
	code, message, err = c.readCmdResponse()
	if err != nil {
		return fmt.Errorf("failed to parse response to STRU command: %v", err)
	}
	if code >= 400 {
		return fmt.Errorf("failed to set structure: %s", message)
	}

	return nil
}

// Close closes the connection to the FTP server sending the QUIT command
func (c *FTPClient) Close() error {
	quitCmd := "QUIT\r\n"
	_, err := c.conn.Write([]byte(quitCmd))
	if err != nil {
		return fmt.Errorf("failed to send QUIT command: %v", err)
	}

	// Read server's response to QUIT command
	quitResp := make([]byte, 1024)
	_, err = c.conn.Read(quitResp)
	if err != nil {
		return fmt.Errorf("failed to read response to QUIT command: %v", err)
	}

	err = c.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}

	return nil
}

func (c *FTPClient) MakeDirectory(path string) error {
	err := c.sendCmd(fmt.Sprintf("MKD %s", path))
	if err != nil {
		return fmt.Errorf("failed to send MKD command: %v", err)
	}

	// Read server's response to MKD command
	code, msg, err := c.readCmdResponse()
	if err != nil {
		return fmt.Errorf("failed to read response to MKD command: %v", err)
	}
	if code >= 400 {
		return fmt.Errorf("failed to create directory %d: %s", code, msg)
	}

	return nil
}

func (c *FTPClient) DeleteFile(path string) error {
	err := c.sendCmd(fmt.Sprintf("DELE %s", path))
	if err != nil {
		return fmt.Errorf("failed to send DELE command: %v", err)
	}

	// Read server's response to DELE command
	code, msg, err := c.readCmdResponse()
	if err != nil {
		return fmt.Errorf("failed to read response to DELE command: %v", err)
	}
	if code >= 400 {
		return fmt.Errorf("failed to delete file %d: %s", code, msg)
	}

	return nil
}

func (c *FTPClient) UploadFile(localPath, remotePath string) error {
	err := c.openDataConnection()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %v", err)
	}
	defer c.dataConn.Close()

	err = c.sendCmd(fmt.Sprintf("STOR %s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to send STOR command: %v", err)
	}

	// Read server's response to STOR command
	code, msg, err := c.readCmdResponse()
	if err != nil {
		return fmt.Errorf("failed to read response to STOR command: %v", err)
	}
	if code >= 400 {
		return fmt.Errorf("failed to upload file %d: %s", code, msg)
	}

	file, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	_, err = c.dataConn.Write(file)
	if err != nil {
		return fmt.Errorf("failed to write data to data connection: %v", err)
	}

	return nil
}

func (c *FTPClient) DownloadFile(remotePath, localPath string) error {
	err := c.openDataConnection()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %v", err)
	}

	err = c.sendCmd(fmt.Sprintf("RETR %s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to send RETR command: %v", err)
	}

	// Read server's response to RETR command
	code, msg, err := c.readCmdResponse()
	if err != nil {
		return fmt.Errorf("failed to read response to RETR command: %v", err)
	}
	if code >= 400 {
		return fmt.Errorf("failed to download file %d: %s", code, msg)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, c.dataConn)
	if err != nil {
		return fmt.Errorf("failed to copy data to file: %v", err)
	}

	return nil
}

func (c *FTPClient) RemoveDirectory(path string) error {
	err := c.sendCmd(fmt.Sprintf("RMD %s", path))
	if err != nil {
		return fmt.Errorf("failed to send RMD command: %v", err)
	}

	// Read server's response to RMD command
	code, msg, err := c.readCmdResponse()
	if err != nil {
		return fmt.Errorf("failed to read response to RMD command: %v", err)
	}
	if code >= 400 {
		return fmt.Errorf("failed to remove directory %d: %s", code, msg)
	}

	return nil
}

func (c *FTPClient) List(path string) error {
	if err := c.openDataConnection(); err != nil {
		return err
	}

	err := c.sendCmd(fmt.Sprintf("LIST %s", path))
	if err != nil {
		return fmt.Errorf("failed to send LIST command: %v", err)
	}

	// Read server's response to LIST command
	code, msg, err := c.readCmdResponse()
	if err != nil {
		return fmt.Errorf("failed to read response to LIST command: %v", err)
	}
	if code >= 400 {
		return fmt.Errorf("failed to list directory: %s", msg)
	}

	if code == 150 {
		code, msg, err = c.readCmdResponse()
		if err != nil {
			return fmt.Errorf("failed to read response to LIST command: %v", err)
		}
		if code >= 400 {
			return fmt.Errorf("failed to list directory: %s", msg)
		}
	}

	// Read data from data connection
	data := make([]byte, 1024)
	_, err = c.dataConn.Read(data)
	if err != nil {
		// EOF is expected when directory is empty
		if err == io.EOF {
			log.Printf("Remote files:\n%s", string(data))
			return nil
		}
		return fmt.Errorf("failed to read data from data connection: %v", err)
	}
	log.Printf("Remote files:\n%s", string(data))

	return nil
}

func (c *FTPClient) sendCmd(command string) error {
	log.Printf("Sending command: %s\n", command)
	_, err := c.conn.Write([]byte(command + "\r\n"))
	if err != nil {
		return fmt.Errorf("failed to send command: %v", err)
	}

	return nil
}

func (c *FTPClient) readCmdResponse() (int, string, error) {
	resp := make([]byte, 1024)
	_, err := c.conn.Read(resp)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read response: %v", err)
	}

	code, message, err := c.parseResponse(resp)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse response: %v", err)
	}

	log.Printf("Received code %d, message: %s\n", code, message)
	return code, message, nil
}

func (c *FTPClient) openDataConnection() error {
	pasvCmd := "PASV\r\n"
	_, err := c.conn.Write([]byte(pasvCmd))
	if err != nil {
		return fmt.Errorf("failed to send PASV command: %v", err)
	}

	// Read server's response to PASV command
	pasvResp := make([]byte, 1024)
	_, err = c.conn.Read(pasvResp)
	if err != nil {
		return fmt.Errorf("failed to read response to PASV command: %v", err)
	}
	code, message, err := c.parseResponse(pasvResp)
	if err != nil {
		return fmt.Errorf("failed to parse response to PASV command: %v", err)
	}
	if code >= 400 {
		return fmt.Errorf("failed to open data connection: %s", message)
	}

	// Parse the PASV response using regexp
	regexp := regexp.MustCompile(`\((\d+),(\d+),(\d+),(\d+),(\d+),(\d+)\)`)
	matches := regexp.FindStringSubmatch(message)
	if len(matches) != 7 {
		return fmt.Errorf("failed to parse PASV response: %s", message)
	}

	ip := fmt.Sprintf("%s.%s.%s.%s", matches[1], matches[2], matches[3], matches[4])
	lowerPortBits, err := strconv.Atoi(matches[5])
	if err != nil {
		return fmt.Errorf("failed to parse port: %v", err)
	}
	upperPortBits, err := strconv.Atoi(matches[6])
	if err != nil {
		return fmt.Errorf("failed to parse port: %v", err)
	}
	port := (lowerPortBits << 8) + upperPortBits

	dataConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return fmt.Errorf("failed to open data connection: %v", err)
	}
	c.dataConn = dataConn

	return nil
}

func (c *FTPClient) parseResponse(resp []byte) (int, string, error) {
	codeStr := string(resp[0:3])
	message := string(resp[4:])

	code, err := strconv.Atoi(codeStr)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse response code: %v", err)
	}

	return code, message, nil
}

func init() {
	// Add commands to the root command
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(mkdirCmd)
	rootCmd.AddCommand(rmDirCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(copyCmd)
	rootCmd.AddCommand(moveCmd)

	// Add global flags to the root command
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "FTP server hostname")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 21, "FTP server port")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "anonymous", "FTP username")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "P", "anonymous", "FTP password")

	// Add local flags to the ls command
	lsCmd.Flags().StringVarP(&sourcePath, "path", "s", "", "Path to list files")
	lsCmd.Flags().BoolVarP(&printDetails, "details", "l", false, "Print file details")

	// Add local flags to the mkdir command
	mkdirCmd.Flags().StringVarP(&sourcePath, "path", "s", "", "Path to create directory")

	// Add local flags to the rmdir command
	rmDirCmd.Flags().StringVarP(&sourcePath, "path", "s", "", "Path to remove directory")

	// Add local flags to the rm command
	rmCmd.Flags().StringVarP(&sourcePath, "path", "s", "", "Path to remove file")

	// Add local flags to the cp command
	copyCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Source path")
	copyCmd.Flags().StringVarP(&destPath, "destination", "d", "", "Destination path")

	// Add local flags to the mv command
	moveCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Source path")
	moveCmd.Flags().StringVarP(&destPath, "destination", "d", "", "Destination path")

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}
}
