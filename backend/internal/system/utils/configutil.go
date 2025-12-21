package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/asgardeo/thunder/internal/system/log"
)

var (
	// Pattern to match Go template variables like {{.Variable}}
	varPattern = regexp.MustCompile(`\{\{\s*\.\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

	// Pattern to match Go template range patterns like {{- range .ArrayVar}}
	rangePattern = regexp.MustCompile(`\{\{-\s*range\s+\.\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

	// Pattern to match Go template file path patterns
	filePattern = regexp.MustCompile(`file://(?:"([^"]*)"|([^\s"]+))`)
)

// SubstituteFilePaths replaces file path placeholders in the given content with the actual file contents.
//
// Supported patterns:
//  1. file://path/to/file - Unquoted file path (no spaces allowed)
//  2. file://"path/with/spaces" - Quoted file path (spaces allowed)
//  3. file:/relative/path - Relative file path (resolved against base directory)
//
// If a file cannot be read, an error is returned.
func SubstituteFilePaths(content []byte, baseDir string) ([]byte, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ConfigUtil"))
	isError := false

	out := filePattern.ReplaceAllStringFunc(string(content), func(match string) string {
		sub := filePattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}

		// Group 1: quoted path (file://"path"), Group 2: unquoted path (file://path)
		path := sub[1]
		if path == "" {
			path = sub[2]
		}
		if path == "" {
			logger.Warn("Empty file path in placeholder", log.String("placeholder", match))
			return ""
		}

		// Convert relative paths to absolute
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}

		data, err := readFileContent(path)
		if err != nil {
			logger.Error("Failed to read file content", log.String("filePath", path), log.Error(err))
			isError = true
			return ""
		}

		return data
	})

	return []byte(out), func() error {
		if isError {
			return fmt.Errorf("one or more file path substitutions failed")
		}
		return nil
	}()
}

// readFileContent reads the content of the file at the given path.
func readFileContent(path string) (string, error) {
	path = filepath.Clean(path)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// SubstituteEnvironmentVariables replaces Go template variable placeholders in the given content.
//
// Supported patterns:
//  1. {{.Variable}} - Simple variable substitution from environment variables
//  2. {{- range .ArrayVariable}} - Array iteration using VARIABLE_NAME_0, VARIABLE_NAME_1, ... pattern
//
// If an environment variable is not set, an error is returned.
func SubstituteEnvironmentVariables(content []byte) ([]byte, error) {
	contentStr := string(content)

	// Find all variables referenced in the template
	templateVars := make(map[string]interface{})

	// Extract simple variables
	varMatches := varPattern.FindAllStringSubmatch(contentStr, -1)
	for _, match := range varMatches {
		if len(match) > 1 {
			varName := match[1]
			envValue, exists := os.LookupEnv(varName)
			if !exists {
				return nil, fmt.Errorf("environment variable %s is not set", varName)
			}
			templateVars[varName] = envValue
		}
	}

	// Extract array variables from range statements
	rangeMatches := rangePattern.FindAllStringSubmatch(contentStr, -1)
	for _, match := range rangeMatches {
		if len(match) > 1 {
			varName := match[1]
			arrayElements := buildArrayFromEnvVars(varName)
			templateVars[varName] = arrayElements
		}
	}

	// If no template variables found, return original content
	if len(templateVars) == 0 {
		return content, nil
	}

	// Create and execute the template
	tmpl, err := template.New("config").Parse(contentStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// buildArrayFromEnvVars builds an array by reading environment variables with indexed suffixes
// starting from VARNAME_0, VARNAME_1, etc., until an empty or non-existent variable is found.
func buildArrayFromEnvVars(varName string) []string {
	var elements []string
	index := 0

	for {
		indexedVarName := fmt.Sprintf("%s_%d", varName, index)
		value, exists := os.LookupEnv(indexedVarName)

		// Stop if the variable doesn't exist or is empty
		if !exists || value == "" {
			break
		}

		elements = append(elements, value)
		index++
	}

	return elements
}
