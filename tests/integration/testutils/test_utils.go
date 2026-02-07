/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package testutils

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	TargetDir                   = "../../target/dist"
	ExtractedDir                = "../../target/out/.test"
	TestDeploymentYamlPath      = "./resources/deployment.yaml"
	TestDatabaseSchemaDirectory = "resources/dbscripts"
	DatabaseFileBasePath        = "repository/database/"
)

// ServerBinary is the name of the server binary, platform-dependent.
var ServerBinary string

func init() {
	if runtime.GOOS == "windows" {
		ServerBinary = "thunder.exe"
	} else {
		ServerBinary = "thunder"
	}
}

// Package-level variables for server configuration
var (
	serverPort           string
	zipFilePattern       string
	extractedProductHome string
	serverCmd            *exec.Cmd
	isInitialized        bool
	dbType               string
)

// InitializeTestContext initializes the package-level variables for server configuration.
func InitializeTestContext(port string, zipPattern string, databaseType string) {
	serverPort = port
	zipFilePattern = zipPattern
	dbType = databaseType
	isInitialized = true
}

// ensureInitialized checks if the test context has been initialized and panics if not.
func ensureInitialized() {
	if !isInitialized {
		panic("Test context not initialized. Call InitializeTestContext() first.")
	}
}

func UnzipProduct() error {
	// Find the zip file.
	files, err := findMatchingZipFile(zipFilePattern)
	if err != nil || len(files) == 0 {
		return fmt.Errorf("zip file not found in target directory")
	}

	// Unzip the file
	zipFile := files[0]
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(ExtractedDir, os.ModePerm)
	for _, f := range r.File {
		err := extractFile(f, ExtractedDir)
		if err != nil {
			return err
		}
	}

	productHome, err := getExtractedProductHome()
	if err != nil {
		return err
	}
	extractedProductHome = productHome

	// Set executable permissions for the server binary (not needed on Windows)
	if runtime.GOOS != "windows" {
		serverPath := filepath.Join(productHome, ServerBinary)
		if err := os.Chmod(serverPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions for server binary: %v", err)
		}
	}

	return nil
}

func extractFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(dest, f.Name)
	if f.FileInfo().IsDir() {
		return os.MkdirAll(path, os.ModePerm)
	}

	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)

	return err
}

// getExtractedProductHome constructs the path to the unzipped folder.
func getExtractedProductHome() (string, error) {
	files, err := findMatchingZipFile(zipFilePattern)
	if err != nil || len(files) == 0 {
		return "", fmt.Errorf("zip file not found in target directory")
	}
	zipFile := files[0]

	return filepath.Join(ExtractedDir, filepath.Base(zipFile[:len(zipFile)-4])), nil
}

// findMatchingZipFile finds zip files that match our specific version pattern criteria
func findMatchingZipFile(zipFilePattern string) ([]string, error) {
	path := filepath.Join(TargetDir, zipFilePattern)
	files, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}

	// Filter the files to only include those that have a version number or 'v' after 'thunder-'
	var matchingFiles []string
	for _, file := range files {
		baseName := filepath.Base(file)
		parts := strings.Split(baseName, "-")
		if len(parts) >= 3 {
			// Check if the second part starts with a number or 'v'
			secondPart := parts[1]
			if len(secondPart) > 0 && (secondPart[0] == 'v' || (secondPart[0] >= '0' && secondPart[0] <= '9')) {
				matchingFiles = append(matchingFiles, file)
			}
		}
	}

	return matchingFiles, nil
}

func ReplaceResources(zipFilePattern string) error {
	log.Println("Replacing resources...")

	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %v", err)
	} else {
		log.Printf("Current working directory: %s", cwd)
	}

	destPath := filepath.Join(extractedProductHome, "repository", "conf", "deployment.yaml")

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create conf directory: %v", err)
	}

	err = copyFile(TestDeploymentYamlPath, destPath)
	if err != nil {
		return fmt.Errorf("failed to replace deployment.yaml: %v", err)
	}

	return nil
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)

	return err
}

func copyDirectory(src, dest string) error {
	srcDir, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcDir.Close()

	entries, err := srcDir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			err = os.MkdirAll(destPath, os.ModePerm)
			if err != nil {
				return err
			}
			err = copyDirectory(srcPath, destPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, destPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RunInitScript(zipFilePattern string) error {
	log.Println("Running init script...")

	// Skip database initialization for PostgreSQL
	if dbType == "postgres" {
		log.Println("Skipping database initialization for PostgreSQL (already initialized in workflow)")
		return nil
	}

	// Ensure the database directory exists
	dbDir := filepath.Join(extractedProductHome, DatabaseFileBasePath)
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	// Initialize each SQLite database
	databases := []struct {
		name       string
		schemaDir  string
		dbFileName string
	}{
		{"thunderdb", "dbscripts/thunderdb", "thunderdb.db"},
		{"runtimedb", "dbscripts/runtimedb", "runtimedb.db"},
		{"userdb", "dbscripts/userdb", "userdb.db"},
	}

	for _, db := range databases {
		schemaPath := filepath.Join(extractedProductHome, db.schemaDir, "sqlite.sql")
		dbPath := filepath.Join(extractedProductHome, DatabaseFileBasePath, db.dbFileName)

		if err := initSQLiteDB(db.name, schemaPath, dbPath); err != nil {
			return err
		}
	}

	return nil
}

// initSQLiteDB creates a SQLite database from a schema file using the sqlite3 CLI.
func initSQLiteDB(name, schemaPath, dbPath string) error {
	log.Printf("Initializing SQLite database: %s", name)

	// Remove existing database file for a clean start
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing %s database: %v", name, err)
	}

	// Use forward slashes for sqlite3's .read command (works cross-platform)
	readCmd := ".read " + filepath.ToSlash(schemaPath)
	cmd := exec.Command("sqlite3", dbPath, readCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize %s database: %v", name, err)
	}

	// Enable WAL mode
	cmd = exec.Command("sqlite3", dbPath, "PRAGMA journal_mode=WAL;")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable WAL mode for %s: %v", name, err)
	}

	log.Printf("Successfully initialized %s database", name)
	return nil
}

func StartServer(port string, zipFilePattern string) error {
	log.Println("Starting server...")

	serverPath := filepath.Join(extractedProductHome, ServerBinary)
	cmd := exec.Command(serverPath, "-thunderHome="+extractedProductHome)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Preserve GOCOVERDIR environment variable for coverage collection
	envVars := []string{
		"PORT=" + port,
	}

	if goCoverDir := os.Getenv("GOCOVERDIR"); goCoverDir != "" {
		envVars = append(envVars, "GOCOVERDIR="+goCoverDir)
		log.Printf("Coverage collection enabled: GOCOVERDIR=%s\n", goCoverDir)
	}
	cmd.Env = append(os.Environ(), envVars...)

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	serverCmd = cmd

	return nil
}

func StopServer() {
	log.Println("Stopping server...")
	if serverCmd != nil {
		// Send stop signal for graceful shutdown (SIGTERM on Unix, Kill on Windows)
		sendStopSignal(serverCmd.Process)
		// Wait for the process to exit (with timeout)
		done := make(chan error, 1)
		go func() {
			done <- serverCmd.Wait()
		}()

		select {
		case <-done:
			// Process exited gracefully
			// Give a brief moment for coverage files to be fully flushed to disk
			time.Sleep(100 * time.Millisecond)
		case <-time.After(3 * time.Second):
			// Timeout - force kill
			log.Println("Server did not stop gracefully, forcing kill...")
			serverCmd.Process.Kill()
			serverCmd.Wait()
		}
	}
	// Clear the stored command
	serverCmd = nil
}

// RestartServer stops the current server and starts a new one with the same configuration.
func RestartServer() error {
	ensureInitialized()
	log.Println("Restarting server...")

	// Stop the current server if it exists
	StopServer()

	// Wait a moment for the port to be released
	time.Sleep(3 * time.Second)

	// Start a new server instance
	return StartServer(serverPort, zipFilePattern)
}

// RunSetupScript runs the setup script from the extracted product directory.
// This script starts the server without security, runs bootstrap scripts, and stops the server.
func RunSetupScript() error {
	ensureInitialized()

	// Get absolute path to extracted product home
	absProductHome, err := filepath.Abs(extractedProductHome)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		log.Println("Running setup.ps1 from extracted product...")
		setupScript := filepath.Join(absProductHome, "setup.ps1")
		cmd = exec.Command("pwsh", "-File", setupScript)
	} else {
		log.Println("Running setup.sh from extracted product...")
		setupScript := filepath.Join(absProductHome, "setup.sh")
		cmd = exec.Command("bash", setupScript)
	}

	cmd.Dir = absProductHome // Run from product directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	log.Println("Setup script will start server, run bootstrap, and stop server automatically")

	return cmd.Run()
}

func GetZipFilePattern() string {
	goos, goarch := detectOSAndArchitecture()
	// Use a more general pattern, the filtering will happen in findMatchingZipFile
	return fmt.Sprintf("thunder-*-%s-%s.zip", goos, goarch)
}

// detectOSAndArchitecture detects the OS and architecture using Go environment variables
// or falls back to system detection if environment variables are not available
func detectOSAndArchitecture() (string, string) {
	// Try to get from environment variables first
	goos := os.Getenv("GOOS")
	goarch := os.Getenv("GOARCH")

	// If GOOS is not set, try to detect from system
	if goos == "" {
		// Try using go env command first
		cmd := exec.Command("go", "env", "GOOS")
		output, err := cmd.Output()
		if err == nil {
			goos = strings.TrimSpace(string(output))
		}

		// Fallback to uname if go env didn't work
		if goos == "" {
			cmd := exec.Command("uname", "-s")
			output, err := cmd.Output()
			if err == nil {
				osName := strings.TrimSpace(string(output))
				switch {
				case osName == "Darwin":
					goos = "darwin"
				case osName == "Linux":
					goos = "linux"
				case strings.HasPrefix(osName, "MINGW") ||
					strings.HasPrefix(osName, "MSYS") ||
					strings.HasPrefix(osName, "CYGWIN"):
					goos = "windows"
				}
			}
		}
	}

	// If GOARCH is not set, try to detect from system
	if goarch == "" {
		// Try using go env command first
		cmd := exec.Command("go", "env", "GOARCH")
		output, err := cmd.Output()
		if err == nil {
			goarch = strings.TrimSpace(string(output))
		}

		// Fall back to uname if go env didn't work
		if goarch == "" {
			cmd := exec.Command("uname", "-m")
			output, err := cmd.Output()
			if err == nil {
				arch := strings.TrimSpace(string(output))
				switch arch {
				case "x86_64", "amd64":
					goarch = "amd64"
				case "arm64", "aarch64":
					goarch = "arm64"
				}
			}
		}
	}

	// Normalize OS name according to distribution packaging
	if goos == "darwin" {
		goos = "macos"
	} else if goos == "windows" {
		goos = "win"
	}

	// Normalize architecture
	if goarch == "amd64" {
		goarch = "x64"
	}

	return goos, goarch
}
