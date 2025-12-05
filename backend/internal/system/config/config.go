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

// Package config provides structures and functions for loading and managing server configurations.
package config

import (
	"encoding/json"
	"os"
	urlpath "path"
	"path/filepath"
	"reflect"
	"time"

	"github.com/asgardeo/thunder/internal/system/log"

	yaml "gopkg.in/yaml.v3"
)

// ServerConfig holds the server configuration details.
type ServerConfig struct {
	Hostname   string `yaml:"hostname" json:"hostname"`
	Port       int    `yaml:"port" json:"port"`
	HTTPOnly   bool   `yaml:"http_only" json:"http_only"`
	PublicURL  string `yaml:"public_url" json:"public_url"`
	Identifier string `yaml:"identifier" json:"identifier"`
}

// GateClientConfig holds the client configuration details.
type GateClientConfig struct {
	Hostname  string `yaml:"hostname" json:"hostname"`
	Port      int    `yaml:"port" json:"port"`
	Scheme    string `yaml:"scheme" json:"scheme"`
	Path      string `yaml:"path" json:"path"`
	LoginPath string `yaml:"login_path" json:"login_path"`
	ErrorPath string `yaml:"error_path" json:"error_path"`
}

// SecurityConfig holds the security configuration details.
type SecurityConfig struct {
	CertFile   string `yaml:"cert_file" json:"cert_file"`
	KeyFile    string `yaml:"key_file" json:"key_file"`
	CryptoFile string `yaml:"crypto_file" json:"crypto_file"`
}

// DataSource holds the individual database connection details.
type DataSource struct {
	Type            string `yaml:"type" json:"type"`
	Hostname        string `yaml:"hostname" json:"hostname"`
	Port            int    `yaml:"port" json:"port"`
	Name            string `yaml:"name" json:"name"`
	Username        string `yaml:"username" json:"username"`
	Password        string `yaml:"password" json:"password"`
	SSLMode         string `yaml:"sslmode" json:"sslmode"`
	Path            string `yaml:"path" json:"path"`
	Options         string `yaml:"options" json:"options"`
	MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// DatabaseConfig holds the different database configuration details.
type DatabaseConfig struct {
	Identity DataSource `yaml:"identity" json:"identity"`
	Runtime  DataSource `yaml:"runtime" json:"runtime"`
	User     DataSource `yaml:"user" json:"user"`
}

// CacheProperty defines the properties for individual caches.
type CacheProperty struct {
	Name           string `yaml:"name" json:"name"`
	Disabled       bool   `yaml:"disabled" json:"disabled"`
	Size           int    `yaml:"size" json:"size"`
	TTL            int    `yaml:"ttl" json:"ttl"`
	EvictionPolicy string `yaml:"eviction_policy" json:"eviction_policy"`
}

// CacheConfig holds the cache configuration details.
type CacheConfig struct {
	Disabled        bool            `yaml:"disabled" json:"disabled"`
	Type            string          `yaml:"type" json:"type"`
	Size            int             `yaml:"size" json:"size"`
	TTL             int             `yaml:"ttl" json:"ttl"`
	EvictionPolicy  string          `yaml:"eviction_policy" json:"eviction_policy"`
	CleanupInterval int             `yaml:"cleanup_interval" json:"cleanup_interval"`
	Properties      []CacheProperty `yaml:"properties,omitempty" json:"properties,omitempty"`
}

// JWTConfig holds the JWT configuration details.
type JWTConfig struct {
	Issuer         string `yaml:"issuer" json:"issuer"`
	ValidityPeriod int64  `yaml:"validity_period" json:"validity_period"`
	Audience       string `yaml:"audience" json:"audience"`
}

// RefreshTokenConfig holds the refresh token configuration details.
type RefreshTokenConfig struct {
	RenewOnGrant   bool  `yaml:"renew_on_grant" json:"renew_on_grant"`
	ValidityPeriod int64 `yaml:"validity_period" json:"validity_period"`
}

// AuthorizationCodeConfig holds the authorization code configuration details.
type AuthorizationCodeConfig struct {
	ValidityPeriod int64 `yaml:"validity_period" json:"validity_period"`
}

// OAuthConfig holds the OAuth configuration details.
type OAuthConfig struct {
	RefreshToken      RefreshTokenConfig      `yaml:"refresh_token" json:"refresh_token"`
	AuthorizationCode AuthorizationCodeConfig `yaml:"authorization_code" json:"authorization_code"`
}

// FlowAuthnConfig holds the configuration details for the authentication flows.
type FlowAuthnConfig struct {
	DefaultFlow string `yaml:"default_flow" json:"default_flow"`
}

// FlowConfig holds the configuration details for the flow service.
type FlowConfig struct {
	GraphDirectory string          `yaml:"graph_directory" json:"graph_directory"`
	Authn          FlowAuthnConfig `yaml:"authn" json:"authn"`
}

// CORSConfig holds the configuration details for the CORS.
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
}

// HashConfig holds the hashing configuration details.
type HashConfig struct {
	Algorithm string `yaml:"algorithm"`
}

// ImmutableResources holds the configuration details for the immutable resources.
type ImmutableResources struct {
	Enabled bool `yaml:"enabled" json:"enabled" default:"false"`
}

// ObservabilityConfig holds the observability configuration details.
type ObservabilityConfig struct {
	Enabled     bool                      `yaml:"enabled" json:"enabled"`
	Output      ObservabilityOutputConfig `yaml:"output" json:"output"`
	FailureMode string                    `yaml:"failure_mode" json:"failure_mode"`
}

// ObservabilityOutputConfig holds observability output configuration.
type ObservabilityOutputConfig struct {
	File          ObservabilityFileConfig    `yaml:"file" json:"file"`
	Console       ObservabilityConsoleConfig `yaml:"console" json:"console"`
	OpenTelemetry ObservabilityOTelConfig    `yaml:"opentelemetry" json:"opentelemetry"`
}

// ObservabilityFileConfig captures file sink settings for observability events.
type ObservabilityFileConfig struct {
	Enabled       bool          `yaml:"enabled" json:"enabled"`
	FilePath      string        `yaml:"file_path" json:"file_path"`
	Format        string        `yaml:"format" json:"format"`
	BufferSize    int           `yaml:"buffer_size" json:"buffer_size"`
	FlushInterval time.Duration `yaml:"flush_interval" json:"flush_interval"`
	Categories    []string      `yaml:"categories" json:"categories"`
}

// ObservabilityConsoleConfig captures console sink settings for observability events.
type ObservabilityConsoleConfig struct {
	Enabled    bool     `yaml:"enabled" json:"enabled"`
	Format     string   `yaml:"format" json:"format"`
	Categories []string `yaml:"categories" json:"categories"`
}

// ObservabilityOTelConfig holds OpenTelemetry configuration.
type ObservabilityOTelConfig struct {
	Enabled        bool     `yaml:"enabled" json:"enabled"`
	ExporterType   string   `yaml:"exporter_type" json:"exporter_type"`
	OTLPEndpoint   string   `yaml:"otlp_endpoint" json:"otlp_endpoint"`
	ServiceName    string   `yaml:"service_name" json:"service_name"`
	ServiceVersion string   `yaml:"service_version" json:"service_version"`
	Environment    string   `yaml:"environment" json:"environment"`
	SampleRate     float64  `yaml:"sample_rate" json:"sample_rate"`
	Categories     []string `yaml:"categories" json:"categories"`
	// Insecure disables TLS for OTLP (not recommended for production)
	Insecure bool `yaml:"insecure" json:"insecure"`
}

// Config holds the complete configuration details of the server.
type Config struct {
	Server             ServerConfig        `yaml:"server" json:"server"`
	GateClient         GateClientConfig    `yaml:"gate_client" json:"gate_client"`
	Security           SecurityConfig      `yaml:"security" json:"security"`
	Database           DatabaseConfig      `yaml:"database" json:"database"`
	Cache              CacheConfig         `yaml:"cache" json:"cache"`
	JWT                JWTConfig           `yaml:"jwt" json:"jwt"`
	OAuth              OAuthConfig         `yaml:"oauth" json:"oauth"`
	Flow               FlowConfig          `yaml:"flow" json:"flow"`
	Hash               HashConfig          `yaml:"hash"`
	CORS               CORSConfig          `yaml:"cors" json:"cors"`
	ImmutableResources ImmutableResources  `yaml:"immutable_resources" json:"immutable_resources"`
	Observability      ObservabilityConfig `yaml:"observability" json:"observability"`
}

// LoadConfig loads the configurations from the specified YAML file and applies defaults.
func LoadConfig(path string, defaultsPath string) (*Config, error) {
	var cfg Config
	path = filepath.Clean(path)

	// Load default configuration if provided
	if defaultsPath != "" {
		defaultCfg, err := loadDefaultConfig(defaultsPath)
		if err != nil {
			return nil, err
		}
		cfg = *defaultCfg
	}

	// Load user configuration
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
	var userCfg Config
	if err := decoder.Decode(&userCfg); err != nil {
		return nil, err
	}

	// Merge user configuration with defaults
	mergeConfigs(&cfg, &userCfg)
	// Derive login_path and error_path from path if not explicitly set
	if cfg.GateClient.Path != "" {
		if cfg.GateClient.LoginPath == "" {
			cfg.GateClient.LoginPath = urlpath.Join(cfg.GateClient.Path, "signin")
		}
		if cfg.GateClient.ErrorPath == "" {
			cfg.GateClient.ErrorPath = urlpath.Join(cfg.GateClient.Path, "error")
		}
	}

	return &cfg, nil
}

// loadDefaultConfig loads the default configuration from a JSON file.
func loadDefaultConfig(path string) (*Config, error) {
	var cfg Config
	path = filepath.Clean(path)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if ferr := file.Close(); ferr != nil {
			log.GetLogger().Error("Failed to close default config file", log.Error(ferr))
		}
	}()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// mergeConfigs merges user configuration into the base configuration.
// Non-zero values from userCfg will override corresponding values in baseCfg.
func mergeConfigs(baseCfg, userCfg *Config) {
	mergeStructs(reflect.ValueOf(baseCfg).Elem(), reflect.ValueOf(userCfg).Elem())
}

// mergeStructs recursively merges struct fields.
func mergeStructs(base, user reflect.Value) {
	if !base.IsValid() || !user.IsValid() {
		return
	}

	switch base.Kind() {
	case reflect.Struct:
		for i := 0; i < base.NumField(); i++ {
			baseField := base.Field(i)
			userField := user.Field(i)
			if baseField.CanSet() && userField.IsValid() {
				// For structs, we need to recursively merge even if the user struct is zero value
				// to ensure defaults are preserved
				if baseField.Kind() == reflect.Struct && userField.Kind() == reflect.Struct {
					mergeStructs(baseField, userField)
				} else {
					// For non-struct fields, only override if user value is non-zero
					if !isZeroValue(userField) {
						baseField.Set(userField)
					}
				}
			}
		}
	case reflect.Slice:
		// For slices, if user has values, use them. Otherwise keep base values
		if user.Len() > 0 {
			base.Set(user)
		}
	case reflect.Map:
		// For maps, merge key-value pairs
		if !user.IsNil() && user.Len() > 0 {
			if base.IsNil() {
				base.Set(reflect.MakeMap(base.Type()))
			}
			for _, key := range user.MapKeys() {
				base.SetMapIndex(key, user.MapIndex(key))
			}
		}
	default:
		// For primitive types, use user value if it's not zero value
		if !isZeroValue(user) {
			base.Set(user)
		}
	}
}

// isZeroValue checks if a reflect.Value represents the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Slice, reflect.Map, reflect.Chan:
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}
