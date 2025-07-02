/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package config provides structures and functions for loading and managing server configurations.
package config

import (
	"os"
	"path/filepath"

	"github.com/asgardeo/thunder/internal/system/log"

	yaml "gopkg.in/yaml.v3"
)

// ServerConfig holds the server configuration details.
type ServerConfig struct {
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	HTTPOnly bool   `yaml:"http_only"`
}

// GateClientConfig holds the client configuration details.
type GateClientConfig struct {
	Hostname  string `yaml:"hostname"`
	Port      int    `yaml:"port"`
	Scheme    string `yaml:"scheme"`
	LoginPath string `yaml:"login_path"`
	ErrorPath string `yaml:"error_path"`
}

// SecurityConfig holds the security configuration details.
type SecurityConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// DataSource holds the individual database connection details.
type DataSource struct {
	Type     string `yaml:"type"`
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslmode"`
	Path     string `yaml:"path"`
	Options  string `yaml:"options"`
}

// DatabaseConfig holds the different database configuration details.
type DatabaseConfig struct {
	Identity DataSource `yaml:"identity"`
	Runtime  DataSource `yaml:"runtime"`
}

// JWTConfig holds the JWT configuration details.
type JWTConfig struct {
	Issuer         string `yaml:"issuer"`
	ValidityPeriod int64  `yaml:"validity_period"`
}

// RefreshTokenConfig holds the refresh token configuration details.
type RefreshTokenConfig struct {
	RenewOnGrant   bool  `yaml:"renew_on_grant"`
	ValidityPeriod int64 `yaml:"validity_period"`
}

// OAuthConfig holds the OAuth configuration details.
type OAuthConfig struct {
	JWT          JWTConfig          `yaml:"jwt"`
	RefreshToken RefreshTokenConfig `yaml:"refresh_token"`
}

// Authenticator holds the configuration details for an individual authenticator.
type Authenticator struct {
	Name             string            `yaml:"name"`
	Type             string            `yaml:"type"`
	DisplayName      string            `yaml:"display_name"`
	Description      string            `yaml:"description"`
	ClientID         string            `yaml:"client_id"`
	ClientSecret     string            `yaml:"client_secret"`
	RedirectURI      string            `yaml:"redirect_uri"`
	Scopes           []string          `yaml:"scopes"`
	AdditionalParams map[string]string `yaml:"additional_params"`
}

// AuthenticatorConfig holds the configuration details for the authenticators.
type AuthenticatorConfig struct {
	DefaultAuthenticator string          `yaml:"default"`
	Authenticators       []Authenticator `yaml:"authenticators"`
}

// FlowAuthnConfig holds the configuration details for the authentication flows.
type FlowAuthnConfig struct {
	DefaultFlow string `yaml:"default_flow"`
}

// FlowConfig holds the configuration details for the flow service.
type FlowConfig struct {
	GraphDirectory string          `yaml:"graph_directory"`
	Authn          FlowAuthnConfig `yaml:"authn"`
}

// Config holds the complete configuration details of the server.
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	GateClient    GateClientConfig    `yaml:"gate_client"`
	Security      SecurityConfig      `yaml:"security"`
	Database      DatabaseConfig      `yaml:"database"`
	OAuth         OAuthConfig         `yaml:"oauth"`
	Authenticator AuthenticatorConfig `yaml:"authenticator"`
	Flow          FlowConfig          `yaml:"flow"`
}

// LoadConfig loads the configurations from the specified YAML file.
func LoadConfig(path string) (*Config, error) {
	var cfg Config
	path = filepath.Clean(path)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if ferr := file.Close(); ferr != nil {
			log.GetLogger().Error("Failed to close config file", log.Error(ferr))
		}
	}()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
