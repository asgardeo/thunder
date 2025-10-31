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

// Package main is the entry point for starting the Thunder server.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/internal/system/cert"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// shutdownTimeout defines the timeout duration for graceful shutdown.
const shutdownTimeout = 5 * time.Second

func main() {
	logger := log.GetLogger()

	thunderHome := getThunderHome(logger)

	cfg := initThunderConfigurations(logger, thunderHome)
	if cfg == nil {
		logger.Fatal("Failed to initialize configurations")
	}

	// Initialize the cache manager.
	initCacheManager(logger)

	// Create a new HTTP multiplexer.
	mux := http.NewServeMux()
	if mux == nil {
		logger.Fatal("Failed to initialize multiplexer")
	}

	// Register the services.
	registerServices(mux)

	// Register static file handlers for frontend applications.
	registerStaticFileHandlers(logger, mux, thunderHome)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Load the certificate configuration.
	tlsConfig := loadCertConfig(logger, cfg, thunderHome)

	var server *http.Server
	if cfg.Server.HTTPOnly {
		logger.Info("TLS is not enabled, starting server without TLS")
		server = startHTTPServer(logger, cfg, mux)
	} else {
		server = startTLSServer(logger, cfg, mux, tlsConfig)
	}

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Shutting down server...")
	gracefulShutdown(logger, server)
}

// getThunderHome retrieves and return the Thunder home directory.
func getThunderHome(logger *log.Logger) string {
	// Parse project directory from command line arguments.
	projectHome := ""
	projectHomeFlag := flag.String("thunderHome", "", "Path to Thunder home directory")
	flag.Parse()

	if *projectHomeFlag != "" {
		logger.Info("Using thunderHome from command line argument", log.String("thunderHome", *projectHomeFlag))
		projectHome = *projectHomeFlag
	} else {
		// If no command line argument is provided, use the current working directory.
		dir, dirErr := os.Getwd()
		if dirErr != nil {
			logger.Fatal("Failed to get current working directory", log.Error(dirErr))
		}
		projectHome = dir
	}

	return projectHome
}

// initThunderConfigurations initializes the Thunder configurations.
func initThunderConfigurations(logger *log.Logger, thunderHome string) *config.Config {
	// Load the configurations.
	configFilePath := path.Join(thunderHome, "repository/conf/deployment.yaml")
	defaultConfigPath := path.Join(thunderHome, "repository/resources/conf/default.json")
	cfg, err := config.LoadConfig(configFilePath, defaultConfigPath)
	if err != nil {
		logger.Fatal("Failed to load configurations", log.Error(err))
	}

	// Initialize runtime configurations.
	if err := config.InitializeThunderRuntime(thunderHome, cfg); err != nil {
		logger.Fatal("Failed to initialize thunder runtime", log.Error(err))
	}

	return cfg
}

// initCacheManager initializes the cache manager with centralized cleanup.
func initCacheManager(logger *log.Logger) {
	cm := cache.GetCacheManager()
	if cm == nil {
		logger.Fatal("Failed to get cache manager instance")
	}
	cm.Init()
}

// loadCertConfig loads the certificate configuration and extracts the Key ID (kid).
func loadCertConfig(logger *log.Logger, cfg *config.Config, thunderHome string) *tls.Config {
	sysCertSvc := cert.NewSystemCertificateService()
	tlsConfig, err := sysCertSvc.GetTLSConfig(cfg, thunderHome)
	if err != nil {
		logger.Fatal("Failed to load TLS configuration", log.Error(err))
	}

	// Extract and set the certificate Key ID (kid).
	kid, err := sysCertSvc.GetCertificateKid(tlsConfig)
	if err != nil {
		logger.Fatal("Failed to extract certificate kid", log.Error(err))
	}

	certConfig := config.CertConfig{
		TLSConfig: tlsConfig,
		CertKid:   kid,
	}
	config.GetThunderRuntime().SetCertConfig(certConfig)

	return tlsConfig
}

// startTLSServer starts the HTTPS server with TLS configuration.
func startTLSServer(logger *log.Logger, cfg *config.Config, mux *http.ServeMux, tlsConfig *tls.Config) *http.Server {
	server, serverAddr := createHTTPServer(logger, cfg, mux)

	ln, err := tls.Listen("tcp", serverAddr, tlsConfig)
	if err != nil {
		logger.Fatal("Failed to start TLS listener", log.Error(err))
	}

	logger.Info("WSO2 Thunder server started (HTTPS)...", log.String("address", serverAddr))

	// Start server in a goroutine
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to serve requests", log.Error(err))
		}
	}()

	return server
}

// startHTTPServer starts the HTTP server without TLS.
func startHTTPServer(logger *log.Logger, cfg *config.Config, mux *http.ServeMux) *http.Server {
	server, serverAddr := createHTTPServer(logger, cfg, mux)

	logger.Info("WSO2 Thunder server started (HTTP)...", log.String("address", serverAddr))

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to serve HTTP requests", log.Error(err))
		}
	}()

	return server
}

// createHTTPServer creates and configures an HTTP server with common settings.
func createHTTPServer(logger *log.Logger, cfg *config.Config, mux *http.ServeMux) (*http.Server, string) {
	// Wrap the multiplexer with AccessLogHandler.
	wrappedMux := log.AccessLogHandler(logger, mux)

	// Build the server address using hostname and port from the configurations.
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Hostname, cfg.Server.Port)

	server := &http.Server{
		Addr:              serverAddr,
		Handler:           wrappedMux,
		ReadHeaderTimeout: 10 * time.Second, // Mitigate Slowloris attacks
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return server, serverAddr
}

// gracefulShutdown handles the graceful shutdown of all components.
func gracefulShutdown(logger *log.Logger, server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error during server shutdown", log.Error(err))
	} else {
		logger.Debug("HTTP server shutdown completed")
	}

	// Close database connections
	dbCloser := provider.GetDBProviderCloser()
	if err := dbCloser.Close(); err != nil {
		logger.Error("Error closing database connections", log.Error(err))
	} else {
		logger.Debug("Database connections closed successfully")
	}

	logger.Info("Server shutdown completed")
}

// registerStaticFileHandlers registers static file handlers for frontend applications.
func registerStaticFileHandlers(logger *log.Logger, mux *http.ServeMux, thunderHome string) {
	// Serve gate application from /signin
	gateDir := path.Join(thunderHome, "apps", "gate")
	if directoryExists(gateDir) {
		logger.Debug("Registering static file handler for Gate application",
			log.String("path", "/signin/"), log.String("directory", gateDir))
		mux.Handle("/signin/", createStaticFileHandler("/signin/", gateDir, logger))
	} else {
		logger.Warn("Gate application directory not found", log.String("directory", gateDir))
	}

	// Serve develop application from /develop
	developDir := path.Join(thunderHome, "apps", "develop")
	if directoryExists(developDir) {
		logger.Debug("Registering static file handler for Develop application",
			log.String("path", "/develop/"), log.String("directory", developDir))
		mux.Handle("/develop/", createStaticFileHandler("/develop/", developDir, logger))
	} else {
		logger.Warn("Develop application directory not found", log.String("directory", developDir))
	}
}

// createStaticFileHandler creates a handler for serving static files with SPA fallback.
func createStaticFileHandler(routePrefix, directory string, logger *log.Logger) http.Handler {
	fileServer := http.FileServer(http.Dir(directory))

	return http.StripPrefix(routePrefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the file path
		filePath := path.Join(directory, r.URL.Path)

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// For SPA routing, serve index.html for non-existent files
			indexPath := path.Join(directory, "index.html")
			if fileExists(indexPath) {
				logger.Debug("Serving index.html for SPA routing",
					log.String("requested_path", r.URL.Path),
					log.String("route_prefix", routePrefix))
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		// Serve the requested file or directory listing
		fileServer.ServeHTTP(w, r)
	}))
}

// directoryExists checks if a directory exists.
func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
