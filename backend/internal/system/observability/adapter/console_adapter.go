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

package adapter

import (
	"fmt"
	"os"
	"sync"
)

// ConsoleAdapter writes events to stdout/stderr.
type consoleAdapter struct {
	mu     sync.Mutex
	closed bool
}

var _ OutputAdapterInterface = (*consoleAdapter)(nil)

// newConsoleAdapter creates a new console-based output adapter.
func newConsoleAdapter() *consoleAdapter {
	return &consoleAdapter{
		closed: false,
	}
}

// Write writes data to stdout.
func (ca *consoleAdapter) Write(data []byte) error {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	if ca.closed {
		return fmt.Errorf("console adapter is closed")
	}

	_, err := os.Stdout.Write(append(data, '\n'))
	return err
}

// Flush is a no-op for console adapter as stdout is unbuffered.
func (ca *consoleAdapter) Flush() error {
	return nil
}

// Close closes the console adapter.
func (ca *consoleAdapter) Close() error {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	ca.closed = true
	return nil
}

// GetName returns the name of this adapter.
func (ca *consoleAdapter) GetName() string {
	return "ConsoleAdapter"
}
