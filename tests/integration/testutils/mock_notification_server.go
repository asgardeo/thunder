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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// MockNotificationServer provides a mock HTTP server for testing SMS notifications
type MockNotificationServer struct {
	server   *http.Server
	messages []SMSMessage
	mutex    sync.RWMutex
	port     int
}

// SMSMessage represents a received SMS message
type SMSMessage struct {
	Message string `json:"message"`
	OTP     string `json:"otp,omitempty"`
}

// SMSRequest represents the expected SMS request format
type SMSRequest struct {
	Body string `json:"body"`
}

// NewMockNotificationServer creates a new mock notification server
func NewMockNotificationServer(port int) *MockNotificationServer {
	return &MockNotificationServer{
		port:     port,
		messages: make([]SMSMessage, 0),
	}
}

// Start starts the mock notification server
func (m *MockNotificationServer) Start() error {
	mux := http.NewServeMux()

	// Handle SMS sending endpoint
	mux.HandleFunc("/send-sms", m.handleSendSMS)

	// Handle message retrieval endpoint for testing
	mux.HandleFunc("/messages", m.handleGetMessages)

	// Handle clear messages endpoint for testing
	mux.HandleFunc("/clear", m.handleClearMessages)

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.port),
		Handler: mux,
	}

	go func() {
		log.Printf("Starting mock notification server on port %d", m.port)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Mock notification server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the mock notification server
func (m *MockNotificationServer) Stop() error {
	if m.server != nil {
		return m.server.Close()
	}
	return nil
}

// GetURL returns the base URL of the mock server
func (m *MockNotificationServer) GetURL() string {
	return fmt.Sprintf("http://localhost:%d", m.port)
}

// GetSendSMSURL returns the SMS sending endpoint URL
func (m *MockNotificationServer) GetSendSMSURL() string {
	return fmt.Sprintf("%s/send-sms", m.GetURL())
}

// handleSendSMS handles SMS sending requests
func (m *MockNotificationServer) handleSendSMS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the raw body as the SMS message content
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	messageBody := string(bodyBytes)

	// Extract OTP from message (assuming it's a numeric sequence)
	otp := extractOTPFromMessage(messageBody)

	message := SMSMessage{
		Message: messageBody,
		OTP:     otp,
	}

	m.mutex.Lock()
	m.messages = append(m.messages, message)
	m.mutex.Unlock()

	log.Printf("Mock SMS received: %s (OTP: %s)", messageBody, otp)

	// Return success response
	response := map[string]interface{}{
		"success":   true,
		"messageId": fmt.Sprintf("mock-msg-%d", len(m.messages)),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleGetMessages handles requests to retrieve sent messages
func (m *MockNotificationServer) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	m.mutex.RLock()
	messages := make([]SMSMessage, len(m.messages))
	copy(messages, m.messages)
	m.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
}

// handleClearMessages handles requests to clear all messages
func (m *MockNotificationServer) handleClearMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	m.mutex.Lock()
	m.messages = make([]SMSMessage, 0)
	m.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}

// GetLastMessage returns the last received message
func (m *MockNotificationServer) GetLastMessage() *SMSMessage {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.messages) == 0 {
		return nil
	}
	lastMessage := m.messages[len(m.messages)-1]
	m.messages = m.messages[:len(m.messages)-1]
	return &lastMessage
}

// WaitForMessage waits for a message to arrive with a timeout.
// It polls every 50ms until a message is available or the timeout is reached.
// This is more reliable than a fixed sleep for integration tests.
func (m *MockNotificationServer) WaitForMessage(timeoutMs int) *SMSMessage {
	pollInterval := 50 // milliseconds
	elapsed := 0

	for elapsed < timeoutMs {
		m.mutex.RLock()
		hasMessages := len(m.messages) > 0
		m.mutex.RUnlock()

		if hasMessages {
			return m.GetLastMessage()
		}

		time.Sleep(time.Duration(pollInterval) * time.Millisecond)
		elapsed += pollInterval
	}

	return nil
}

// ClearMessages clears all stored messages
func (m *MockNotificationServer) ClearMessages() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = make([]SMSMessage, 0)
}

// extractOTPFromMessage extracts numeric OTP from the message text
// Handles formats like "Your verification code is: 657786. This code is valid for 2 minutes."
func extractOTPFromMessage(message string) string {
	var currentNumber string
	var bestSequence string
	var bestScore int

	for _, char := range message {
		if char >= '0' && char <= '9' {
			currentNumber += string(char)
		} else {
			// When we hit a non-digit, check if current sequence is a valid OTP length
			if len(currentNumber) >= 4 && len(currentNumber) <= 8 {
				score := calculateOTPScore(currentNumber)
				if score > bestScore {
					bestSequence = currentNumber
					bestScore = score
				}
			}
			currentNumber = ""
		}
	}

	// Check the last sequence too
	if len(currentNumber) >= 4 && len(currentNumber) <= 8 {
		score := calculateOTPScore(currentNumber)
		if score > bestScore {
			bestSequence = currentNumber
			bestScore = score
		}
	}

	return bestSequence
}

// calculateOTPScore assigns a score to potential OTP sequences
// Prioritizes 6-digit codes, then length, to find the most likely OTP
func calculateOTPScore(sequence string) int {
	length := len(sequence)

	// 6-digit codes are most common for SMS OTP
	if length == 6 {
		return 100
	}
	// 4-digit codes are second most common
	if length == 4 {
		return 80
	}
	// 5-digit codes
	if length == 5 {
		return 70
	}
	// 8-digit codes (less common but valid)
	if length == 8 {
		return 60
	}
	// 7-digit codes (least common)
	if length == 7 {
		return 50
	}

	return 0
}
