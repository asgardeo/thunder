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

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ExtractedMessage represents an i18n message extracted from source code.
type ExtractedMessage struct {
	Key          string
	DefaultValue string
	SourceFile   string
	Line         int
}

// Extractor scans Go source files for core.Message literals.
type Extractor struct {
	verbose bool
}

// NewExtractor creates a new Extractor instance.
func NewExtractor(verbose bool) *Extractor {
	return &Extractor{verbose: verbose}
}

// ExtractFromDirectory scans all Go files in the given directory and extracts core.Message literals.
func (e *Extractor) ExtractFromDirectory(dir string) ([]ExtractedMessage, error) {
	var messages []ExtractedMessage

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() {
			// Skip vendor, testdata, and hidden directories
			if info.Name() == "vendor" || info.Name() == "testdata" || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip the generated defaults.go file itself
		if strings.HasSuffix(path, "core/defaults.go") {
			return nil
		}

		fileMessages, err := e.extractFromFile(path)
		if err != nil {
			return fmt.Errorf("error processing %s: %w", path, err)
		}

		messages = append(messages, fileMessages...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return messages, nil
}

// extractFromFile parses a single Go file and extracts core.Message literals.
func (e *Extractor) extractFromFile(filePath string) ([]ExtractedMessage, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var messages []ExtractedMessage

	// Walk the AST to find core.Message composite literals
	ast.Inspect(node, func(n ast.Node) bool {
		// Look for composite literals (struct instantiation)
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Check if it's a core.Message type
		if !e.isI18nMessage(compLit.Type) {
			return true
		}

		// Extract Key and DefaultValue from the composite literal
		msg := e.extractMessageFromLiteral(compLit, fset, filePath)
		if msg != nil {
			if e.verbose {
				fmt.Printf("  Found: %s = %q at %s:%d\n", msg.Key, msg.DefaultValue, msg.SourceFile, msg.Line)
			}
			messages = append(messages, *msg)
		}

		return true
	})

	return messages, nil
}

// isI18nMessage checks if the type expression represents core.Message.
func (e *Extractor) isI18nMessage(typeExpr ast.Expr) bool {
	// Check for selector expression: core.Message
	sel, ok := typeExpr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	return sel.Sel.Name == "I18nMessage"
}

// extractMessageFromLiteral extracts Key and DefaultValue from a core.Message composite literal.
func (e *Extractor) extractMessageFromLiteral(lit *ast.CompositeLit, fset *token.FileSet, filePath string) *ExtractedMessage {
	var key, defaultValue string

	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		// Extract the string value
		strValue := e.extractStringValue(kv.Value)
		if strValue == "" {
			continue
		}

		switch keyIdent.Name {
		case "Key":
			key = strValue
		case "DefaultValue":
			defaultValue = strValue
		}
	}

	// Only return if we found both Key and DefaultValue
	if key == "" || defaultValue == "" {
		return nil
	}

	pos := fset.Position(lit.Pos())
	return &ExtractedMessage{
		Key:          key,
		DefaultValue: defaultValue,
		SourceFile:   filePath,
		Line:         pos.Line,
	}
}

// extractStringValue extracts the string value from an AST expression.
func (e *Extractor) extractStringValue(expr ast.Expr) string {
	basicLit, ok := expr.(*ast.BasicLit)
	if !ok {
		return ""
	}

	if basicLit.Kind != token.STRING {
		return ""
	}

	// Unquote the string literal
	value, err := strconv.Unquote(basicLit.Value)
	if err != nil {
		return ""
	}

	return value
}
