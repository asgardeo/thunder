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
	"fmt"
	"log"
	"sync"
	"time"
)

// SharedMockServers provides a centralized manager for all mock servers used in integration tests.
// It ensures that each mock server is started only once and shared across all test suites,
// preventing port conflicts and race conditions.
type SharedMockServers struct {
	notificationServer *MockNotificationServer
	httpServer         *MockHTTPServer
	githubServer       *MockGithubOAuthServer
	googleServer       *MockGoogleOIDCServer
	oauthServer        *MockOAuthServer
	oidcServer         *MockOIDCServer

	// Track initialization state
	notificationStarted bool
	httpStarted         bool
	githubStarted       bool
	googleStarted       bool
	oauthStarted        bool
	oidcStarted         bool

	mutex sync.Mutex
}

// Default ports for mock servers
const (
	DefaultNotificationPort = 8098
	DefaultHTTPPort         = 9091
	DefaultGithubPort       = 8092
	DefaultGooglePort       = 8093
	DefaultOAuthPort        = 8092
	DefaultOIDCPort         = 8093
)

// Singleton instance
var (
	sharedServers     *SharedMockServers
	sharedServersOnce sync.Once
)

// GetSharedMockServers returns the singleton instance of SharedMockServers.
// This ensures all test suites use the same mock server instances.
func GetSharedMockServers() *SharedMockServers {
	sharedServersOnce.Do(func() {
		sharedServers = &SharedMockServers{}
	})
	return sharedServers
}

// GetNotificationServer returns the shared notification server, starting it if necessary.
func (s *SharedMockServers) GetNotificationServer() (*MockNotificationServer, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.notificationStarted {
		return s.notificationServer, nil
	}

	s.notificationServer = NewMockNotificationServer(DefaultNotificationPort)
	if err := s.notificationServer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start notification server: %w", err)
	}

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	s.notificationStarted = true
	log.Printf("Shared notification server started on port %d", DefaultNotificationPort)

	return s.notificationServer, nil
}

// GetHTTPServer returns the shared HTTP mock server, starting it if necessary.
func (s *SharedMockServers) GetHTTPServer() (*MockHTTPServer, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.httpStarted {
		return s.httpServer, nil
	}

	s.httpServer = NewMockHTTPServer(DefaultHTTPPort)
	if err := s.httpServer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start HTTP server: %w", err)
	}

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	s.httpStarted = true
	log.Printf("Shared HTTP mock server started on port %d", DefaultHTTPPort)

	return s.httpServer, nil
}

// GetGithubServer returns the shared GitHub OAuth server, starting it if necessary.
// The clientID and clientSecret are used for the first initialization only.
// Note: This server and GetOAuthServer use the same port (8092). Only one can be active at a time.
func (s *SharedMockServers) GetGithubServer(clientID, clientSecret string) (*MockGithubOAuthServer, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.githubStarted {
		return s.githubServer, nil
	}

	// Check if OAuth server is running on the same port
	if s.oauthStarted {
		return nil, fmt.Errorf("cannot start GitHub server: OAuth server already running on port %d", DefaultOAuthPort)
	}

	s.githubServer = NewMockGithubOAuthServer(DefaultGithubPort, clientID, clientSecret)
	if err := s.githubServer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start GitHub server: %w", err)
	}

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	s.githubStarted = true
	log.Printf("Shared GitHub OAuth server started on port %d", DefaultGithubPort)

	return s.githubServer, nil
}

// GetGoogleServer returns the shared Google OIDC server, starting it if necessary.
// The clientID and clientSecret are used for the first initialization only.
// Note: This server and GetOIDCServer use the same port (8093). Only one can be active at a time.
func (s *SharedMockServers) GetGoogleServer(clientID, clientSecret string) (*MockGoogleOIDCServer, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.googleStarted {
		return s.googleServer, nil
	}

	// Check if OIDC server is running on the same port
	if s.oidcStarted {
		return nil, fmt.Errorf("cannot start Google server: OIDC server already running on port %d", DefaultOIDCPort)
	}

	var err error
	s.googleServer, err = NewMockGoogleOIDCServer(DefaultGooglePort, clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google server: %w", err)
	}

	if err := s.googleServer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Google server: %w", err)
	}

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	s.googleStarted = true
	log.Printf("Shared Google OIDC server started on port %d", DefaultGooglePort)

	return s.googleServer, nil
}

// GetOAuthServer returns the shared generic OAuth server, starting it if necessary.
// Note: This server and GetGithubServer use the same port (8092). Only one can be active at a time.
func (s *SharedMockServers) GetOAuthServer(clientID, clientSecret string) (*MockOAuthServer, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.oauthStarted {
		return s.oauthServer, nil
	}

	// Check if GitHub server is running on the same port
	if s.githubStarted {
		return nil, fmt.Errorf("cannot start OAuth server: GitHub server already running on port %d", DefaultGithubPort)
	}

	s.oauthServer = NewMockOAuthServer(DefaultOAuthPort, clientID, clientSecret)
	if err := s.oauthServer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start OAuth server: %w", err)
	}

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	s.oauthStarted = true
	log.Printf("Shared OAuth server started on port %d", DefaultOAuthPort)

	return s.oauthServer, nil
}

// GetOIDCServer returns the shared generic OIDC server, starting it if necessary.
// Note: This server and GetGoogleServer use the same port (8093). Only one can be active at a time.
func (s *SharedMockServers) GetOIDCServer(clientID, clientSecret string) (*MockOIDCServer, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.oidcStarted {
		return s.oidcServer, nil
	}

	// Check if Google server is running on the same port
	if s.googleStarted {
		return nil, fmt.Errorf("cannot start OIDC server: Google server already running on port %d", DefaultGooglePort)
	}

	var err error
	s.oidcServer, err = NewMockOIDCServer(DefaultOIDCPort, clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC server: %w", err)
	}

	if err := s.oidcServer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start OIDC server: %w", err)
	}

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	s.oidcStarted = true
	log.Printf("Shared OIDC server started on port %d", DefaultOIDCPort)

	return s.oidcServer, nil
}

// StopAll stops all mock servers. This should be called once when the entire test suite finishes.
// Note: In practice, the servers will be stopped when the test process exits.
// This method is provided for explicit cleanup if needed.
func (s *SharedMockServers) StopAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.notificationStarted && s.notificationServer != nil {
		if err := s.notificationServer.Stop(); err != nil {
			log.Printf("Error stopping notification server: %v", err)
		}
		s.notificationStarted = false
	}

	if s.httpStarted && s.httpServer != nil {
		if err := s.httpServer.Stop(); err != nil {
			log.Printf("Error stopping HTTP server: %v", err)
		}
		s.httpStarted = false
	}

	if s.githubStarted && s.githubServer != nil {
		if err := s.githubServer.Stop(); err != nil {
			log.Printf("Error stopping GitHub server: %v", err)
		}
		s.githubStarted = false
	}

	if s.googleStarted && s.googleServer != nil {
		if err := s.googleServer.Stop(); err != nil {
			log.Printf("Error stopping Google server: %v", err)
		}
		s.googleStarted = false
	}

	if s.oauthStarted && s.oauthServer != nil {
		if err := s.oauthServer.Stop(); err != nil {
			log.Printf("Error stopping OAuth server: %v", err)
		}
		s.oauthStarted = false
	}

	if s.oidcStarted && s.oidcServer != nil {
		if err := s.oidcServer.Stop(); err != nil {
			log.Printf("Error stopping OIDC server: %v", err)
		}
		s.oidcStarted = false
	}

	log.Println("All shared mock servers stopped")
}

// IsNotificationServerRunning returns true if the notification server is running.
func (s *SharedMockServers) IsNotificationServerRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.notificationStarted
}

// IsHTTPServerRunning returns true if the HTTP server is running.
func (s *SharedMockServers) IsHTTPServerRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.httpStarted
}

// IsGithubServerRunning returns true if the GitHub server is running.
func (s *SharedMockServers) IsGithubServerRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.githubStarted
}

// IsGoogleServerRunning returns true if the Google server is running.
func (s *SharedMockServers) IsGoogleServerRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.googleStarted
}

// IsOAuthServerRunning returns true if the OAuth server is running.
func (s *SharedMockServers) IsOAuthServerRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.oauthStarted
}

// IsOIDCServerRunning returns true if the OIDC server is running.
func (s *SharedMockServers) IsOIDCServerRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.oidcStarted
}
