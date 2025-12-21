package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigUtilsTestSuite struct {
	suite.Suite
	originalEnvVars map[string]string
}

func TestConfigUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigUtilsTestSuite))
}

func (suite *ConfigUtilsTestSuite) SetupTest() {
	// Store original environment variables
	suite.originalEnvVars = make(map[string]string)
}

func (suite *ConfigUtilsTestSuite) TearDownTest() {
	// Restore original environment variables
	for key, value := range suite.originalEnvVars {
		if value == "" {
			err := os.Unsetenv(key)
			suite.Require().NoError(err, "Failed to unset environment variable")
		} else {
			err := os.Setenv(key, value)
			suite.Require().NoError(err, "Failed to set environment variable")
		}
	}
}

// Helper function to set environment variable and track for cleanup
func (suite *ConfigUtilsTestSuite) setEnvVar(key, value string) {
	if _, exists := suite.originalEnvVars[key]; !exists {
		if originalValue, hasOriginal := os.LookupEnv(key); hasOriginal {
			suite.originalEnvVars[key] = originalValue
		} else {
			suite.originalEnvVars[key] = ""
		}
	}
	err := os.Setenv(key, value)
	suite.Require().NoError(err, "Failed to set environment variable")
}

// Tests for SubstituteFilePaths function

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_SimpleFilePath() {
	tempFile := suite.createTempFile("test content")

	content := []byte("config: file://" + tempFile)
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	suite.Equal("config: test content", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_MultipleFiles() {
	tempFile1 := suite.createTempFile("content1")
	tempFile2 := suite.createTempFile("content2")
	tempFile3 := suite.createTempFile("content3")

	content := []byte("config: file://" + tempFile1 + " file://" + tempFile2 + " file://" + tempFile3)
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	suite.Equal("config: content1 content2 content3", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_EmptyContent() {
	content := []byte("")
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	suite.Equal("", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_NoFilePaths() {
	content := []byte("static config content without file paths")
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	suite.Equal("static config content without file paths", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_NonExistentFile() {
	content := []byte("config: file:///nonexistent/path/to/file.txt")
	_, err := SubstituteFilePaths(content, "")

	suite.Error(err)
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_EmptyFile() {
	tempFile := suite.createTempFile("")

	content := []byte("config: file://" + tempFile)
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	suite.Equal("config: ", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_FileWithNewlines() {
	tempFile := suite.createTempFile("line1\nline2\nline3")

	content := []byte("config: file://" + tempFile)
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	suite.Equal("config: line1\nline2\nline3", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_YAMLConfig() {
	certFile := suite.createTempFile("-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----")
	keyFile := suite.createTempFile("-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----")

	content := []byte(`security:
  cert: file://` + certFile + `
  key: file://` + keyFile)

	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	expected := `security:
  cert: -----BEGIN CERTIFICATE-----
MIIC...
-----END CERTIFICATE-----
  key: -----BEGIN PRIVATE KEY-----
MIIE...
-----END PRIVATE KEY-----`
	suite.Equal(expected, string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_JSONConfig() {
	secretFile := suite.createTempFile("my-secret-key")

	content := []byte(`{
  "security": {
    "secret": "file://` + secretFile + `"
  }
}`)

	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	expected := `{
  "security": {
    "secret": "my-secret-key"
  }
}`
	suite.Equal(expected, string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_MixedWithStaticContent() {
	tempFile := suite.createTempFile("dynamic-value")

	content := []byte("static-start file://" + tempFile + " static-end")
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	suite.Equal("static-start dynamic-value static-end", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_FilePathWithSpaces() {
	tempDir := suite.T().TempDir()
	tempFile := tempDir + "/file with spaces.txt"
	err := os.WriteFile(tempFile, []byte("content with spaces"), 0600)
	suite.Require().NoError(err)

	content := []byte(`config: file://"` + tempFile + `"`)
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	suite.Equal("config: content with spaces", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_MultilineContent() {
	tempFile := suite.createTempFile(`multi
line
content`)

	content := []byte("config:\n  value: file://" + tempFile)
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	expected := "config:\n  value: multi\nline\ncontent"
	suite.Equal(expected, string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_PartialMatch() {
	content := []byte("config: file://")
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	suite.Equal("config: file://", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_InvalidPath() {
	content := []byte("config: file://   ")
	result, err := SubstituteFilePaths(content, "")

	suite.NoError(err)
	suite.Equal("config: file://   ", string(result))
}

// New tests for base directory resolution

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_BaseDir_ResolvesRelativePath() {
	baseDir := suite.T().TempDir()
	relPath := "config/secret.txt"
	fullPath := filepath.Join(baseDir, relPath)

	suite.Require().NoError(os.MkdirAll(filepath.Dir(fullPath), 0o750))
	suite.Require().NoError(os.WriteFile(fullPath, []byte("rel-secret"), 0o600))

	content := []byte("secret: file://" + relPath)
	result, err := SubstituteFilePaths(content, baseDir)

	suite.NoError(err)
	suite.Equal("secret: rel-secret", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_BaseDir_DoesNotAffectAbsolutePath() {
	absFile := suite.createTempFile("abs-secret")

	baseDir := suite.T().TempDir()
	content := []byte("secret: file://" + absFile)
	result, err := SubstituteFilePaths(content, baseDir)

	suite.NoError(err)
	suite.Equal("secret: abs-secret", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_BaseDir_TrailingSlash() {
	baseDir := suite.T().TempDir()
	// ensure trailing slash handling
	baseDirWithSlash := baseDir + string(os.PathSeparator)
	relPath := "nested/val.txt"
	fullPath := filepath.Join(baseDir, relPath)

	suite.Require().NoError(os.MkdirAll(filepath.Dir(fullPath), 0o750))
	suite.Require().NoError(os.WriteFile(fullPath, []byte("joined"), 0o600))

	content := []byte("val: file://" + relPath)
	result, err := SubstituteFilePaths(content, baseDirWithSlash)

	suite.NoError(err)
	suite.Equal("val: joined", string(result))
}

// Additional tests for non-empty base directory

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_BaseDir_MultipleMixedPaths() {
	baseDir := suite.T().TempDir()

	relA := "a/one.txt"
	relB := "b/two.txt"
	absC := suite.createTempFile("ABS-C")

	fullA := filepath.Join(baseDir, relA)
	fullB := filepath.Join(baseDir, relB)

	suite.Require().NoError(os.MkdirAll(filepath.Dir(fullA), 0o750))
	suite.Require().NoError(os.MkdirAll(filepath.Dir(fullB), 0o750))
	suite.Require().NoError(os.WriteFile(fullA, []byte("REL-A"), 0o600))
	suite.Require().NoError(os.WriteFile(fullB, []byte("REL-B"), 0o600))

	content := []byte("vals: file://" + relA + " file://" + absC + " file://" + relB)
	result, err := SubstituteFilePaths(content, baseDir)

	suite.NoError(err)
	suite.Equal("vals: REL-A ABS-C REL-B", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_BaseDir_ResolvesDotAndParentSegments() {
	baseDir := suite.T().TempDir()
	relPath := "./x/../secrets/./key.txt"
	fullPath := filepath.Join(baseDir, "secrets", "key.txt")

	suite.Require().NoError(os.MkdirAll(filepath.Dir(fullPath), 0o750))
	suite.Require().NoError(os.WriteFile(fullPath, []byte("normalized"), 0o600))

	content := []byte("k: file://" + relPath)
	result, err := SubstituteFilePaths(content, baseDir)

	suite.NoError(err)
	suite.Equal("k: normalized", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_BaseDir_RelativeMissingFileReturnsError() {
	baseDir := suite.T().TempDir()
	relMissing := "missing/does-not-exist.txt"

	content := []byte("x: file://" + relMissing)
	_, err := SubstituteFilePaths(content, baseDir)

	suite.Error(err)
	suite.Contains(err.Error(), "one or more file path substitutions failed")
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_BaseDir_JSONWithRelative() {
	baseDir := suite.T().TempDir()
	relFile := "cfg/secret.txt"
	fullFile := filepath.Join(baseDir, relFile)

	suite.Require().NoError(os.MkdirAll(filepath.Dir(fullFile), 0o750))
	suite.Require().NoError(os.WriteFile(fullFile, []byte("json-secret"), 0o600))

	content := []byte(`{
  "security": {
    "secret": "file://` + relFile + `"
  }
}`)
	result, err := SubstituteFilePaths(content, baseDir)

	suite.NoError(err)
	expected := `{
  "security": {
    "secret": "json-secret"
  }
}`
	suite.Equal(expected, string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteFilePaths_BaseDir_YAMLWithRelative() {
	baseDir := suite.T().TempDir()
	relCert := "tls/cert.pem"
	relKey := "tls/key.pem"
	fullCert := filepath.Join(baseDir, relCert)
	fullKey := filepath.Join(baseDir, relKey)

	suite.Require().NoError(os.MkdirAll(filepath.Dir(fullCert), 0o750))
	suite.Require().NoError(os.WriteFile(fullCert, []byte("CERT\nLINE2"), 0o600))
	suite.Require().NoError(os.WriteFile(fullKey, []byte("KEY\nLINE2"), 0o600))

	content := []byte(`security:
  cert: file://` + relCert + `
  key: file://` + relKey)
	result, err := SubstituteFilePaths(content, baseDir)

	suite.NoError(err)
	expected := `security:
  cert: CERT
LINE2
  key: KEY
LINE2`
	suite.Equal(expected, string(result))
}

// Helper function to create a temporary file with content
func (suite *ConfigUtilsTestSuite) createTempFile(content string) string {
	tempFile, err := os.CreateTemp("", "configutil-test-*.txt")
	suite.Require().NoError(err, "Failed to create temp file")
	_, err = tempFile.WriteString(content)
	suite.Require().NoError(err, "Failed to write to temp file")
	suite.Require().NoError(tempFile.Close())

	path := tempFile.Name()

	suite.T().Cleanup(func() {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			suite.T().Logf("failed to remove temp file %s: %v", path, err)
		}
	})

	return path
}

// Tests for SubstituteEnvironmentVariables function

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_SimpleVariable() {
	suite.setEnvVar("TestVar", "test_value")

	content := []byte("config: {{.TestVar}}")
	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	suite.Equal("config: test_value", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_MultipleVariables() {
	suite.setEnvVar("Host", "db.example.com")
	suite.setEnvVar("Port", "5432")
	suite.setEnvVar("Database", "thunder")

	content := []byte("connection: {{.Host}}:{{.Port}}/{{.Database}}")
	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	suite.Equal("connection: db.example.com:5432/thunder", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_EmptyContent() {
	content := []byte("")
	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	suite.Equal("", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_NoVariables() {
	content := []byte("static config content")
	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	suite.Equal("static config content", string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_MissingVariable() {
	content := []byte("config: {{.MissingVar}}")
	result, err := SubstituteEnvironmentVariables(content)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "environment variable MissingVar is not set")
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_ComplexYAMLConfig() {
	suite.setEnvVar("DBHost", "db.internal")
	suite.setEnvVar("DBUser", "admin")
	suite.setEnvVar("DBPass", "secret123")

	content := []byte(`database:
  host: {{.DBHost}}
  user: {{.DBUser}}
  password: {{.DBPass}}
  ssl: true`)

	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	expected := `database:
  host: db.internal
  user: admin
  password: secret123
  ssl: true`
	suite.Equal(expected, string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_ComplexJSONConfig() {
	suite.setEnvVar("DBHost", "db.internal")
	suite.setEnvVar("DBUser", "admin")
	suite.setEnvVar("DBPass", "secret123")

	content := []byte(`{
  "database": {
    "host": "{{.DBHost}}",
    "user": "{{.DBUser}}",
    "password": "{{.DBPass}}",
    "ssl": true
  }
}`)

	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	expected := `{
  "database": {
    "host": "db.internal",
    "user": "admin",
    "password": "secret123",
    "ssl": true
  }
}`
	suite.Equal(expected, string(result))
}

// Tests for Go template syntax

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_GoTemplate_SimpleVariable() {
	suite.setEnvVar("Host", "localhost")

	content := []byte(`server:
  host: {{.Host}}
  port: 8080`)

	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	expected := `server:
  host: localhost
  port: 8080`
	suite.Equal(expected, string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_GoTemplate_Array() {
	suite.setEnvVar("AllowedIPs_0", "192.168.1.1")
	suite.setEnvVar("AllowedIPs_1", "192.168.1.2")
	suite.setEnvVar("AllowedIPs_2", "192.168.1.3")

	content := []byte(`security:
  allowedIPs:
{{- range .AllowedIPs}}
  - {{.}}
{{- end}}`)

	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	expected := `security:
  allowedIPs:
  - 192.168.1.1
  - 192.168.1.2
  - 192.168.1.3`
	suite.Equal(expected, string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_GoTemplate_Mixed() {
	suite.setEnvVar("Host", "localhost")
	suite.setEnvVar("Port", "8080")
	suite.setEnvVar("Endpoints_0", "/api/v1")
	suite.setEnvVar("Endpoints_1", "/api/v2")

	content := []byte(`server:
  host: {{.Host}}
  port: {{.Port}}
  endpoints:
{{- range .Endpoints}}
  - {{.}}
{{- end}}`)

	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	expected := `server:
  host: localhost
  port: 8080
  endpoints:
  - /api/v1
  - /api/v2`
	suite.Equal(expected, string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_GoTemplate_MissingVariable() {
	content := []byte(`server:
  host: {{.NonExistentVar}}`)

	result, err := SubstituteEnvironmentVariables(content)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "environment variable NonExistentVar is not set")
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_GoTemplate_EmptyArray() {
	// No environment variables set for array
	content := []byte(`security:
  allowedIPs:
{{- range .AllowedIPs}}
  - {{.}}
{{- end}}`)

	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	expected := `security:
  allowedIPs:`
	suite.Equal(expected, string(result))
}

func (suite *ConfigUtilsTestSuite) TestSubstituteEnvironmentVariables_GoTemplate_NoTemplateVars() {
	content := []byte(`server:
  host: localhost
  port: 8080`)

	result, err := SubstituteEnvironmentVariables(content)

	suite.NoError(err)
	suite.Equal(string(content), string(result))
}
