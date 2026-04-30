/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// --- DTOs that mirror the API contract (matching default_client.go expectations) ---

type mockConsentElementDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Properties  map[string]string `json:"properties,omitempty"`
}

type mockConsentElementCreateDTO struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Type        string            `json:"type"`
	Properties  map[string]string `json:"properties,omitempty"`
}

type mockConsentPurposeElementDTO struct {
	Name        string `json:"name"`
	IsMandatory bool   `json:"isMandatory"`
}

type mockConsentPurposeCreateDTO struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description,omitempty"`
	Elements    []mockConsentPurposeElementDTO `json:"elements"`
}

type mockConsentPurposeDTO struct {
	ID          string                         `json:"id"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	ClientID    string                         `json:"clientId"`
	Elements    []mockConsentPurposeElementDTO `json:"elements"`
	CreatedTime int64                          `json:"createdTime"`
	UpdatedTime int64                          `json:"updatedTime"`
}

// --- Internal mock state types ---

type mockConsentElement struct {
	mockConsentElementDTO
}

// mockConsentPurpose is the internal storage type for a consent purpose in the mock server.
type mockConsentPurpose = mockConsentPurposeDTO

// MockConsentPurpose holds the in-memory state of a consent purpose for test inspection.
type MockConsentPurpose struct {
	ID       string
	Name     string
	ClientID string
	Elements []mockConsentPurposeElementDTO
}

// MockConsentServer provides a lightweight mock of the default consent management REST API.
// It stores consent elements and purposes in memory, allowing integration tests to verify
// that Thunder correctly syncs consent state on application lifecycle events.
type MockConsentServer struct {
	server   *http.Server
	port     int
	mu       sync.Mutex
	elements map[string]*mockConsentElement // elementID -> element
	purposes map[string]*mockConsentPurpose // purposeID -> purpose
	idSeq    int
}

// NewMockConsentServer creates a new mock consent server that listens on the given port.
func NewMockConsentServer(port int) *MockConsentServer {
	return &MockConsentServer{
		port:     port,
		elements: make(map[string]*mockConsentElement),
		purposes: make(map[string]*mockConsentPurpose),
	}
}

// nextIDLocked generates the next mock ID. Must be called with mu held.
func (s *MockConsentServer) nextIDLocked() string {
	s.idSeq++
	return fmt.Sprintf("mock-consent-%04d", s.idSeq)
}

// GetURL returns the base API URL of the mock server.
func (s *MockConsentServer) GetURL() string {
	return fmt.Sprintf("http://localhost:%d/api/v1", s.port)
}

// GetPurposesForClient returns all purposes stored for the given clientID (application ID).
func (s *MockConsentServer) GetPurposesForClient(clientID string) []MockConsentPurpose {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []MockConsentPurpose
	for _, p := range s.purposes {
		if p.ClientID == clientID {
			elements := make([]mockConsentPurposeElementDTO, len(p.Elements))
			copy(elements, p.Elements)
			result = append(result, MockConsentPurpose{
				ID:       p.ID,
				Name:     p.Name,
				ClientID: p.ClientID,
				Elements: elements,
			})
		}
	}

	return result
}

// GetTotalPurposeCount returns the total number of consent purposes currently stored.
func (s *MockConsentServer) GetTotalPurposeCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.purposes)
}

// Reset clears all stored elements and purposes.
func (s *MockConsentServer) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.elements = make(map[string]*mockConsentElement)
	s.purposes = make(map[string]*mockConsentPurpose)
	s.idSeq = 0
}

// Start starts the mock consent server in the background.
func (s *MockConsentServer) Start() error {
	mux := http.NewServeMux()

	// Register /validate before the subtree handler to ensure it takes precedence.
	mux.HandleFunc("/api/v1/consent-elements/validate", s.handleElementsValidate)
	mux.HandleFunc("/api/v1/consent-elements/", s.handleElementByID)
	mux.HandleFunc("/api/v1/consent-elements", s.handleElements)
	mux.HandleFunc("/api/v1/consent-purposes/", s.handlePurposeByID)
	mux.HandleFunc("/api/v1/consent-purposes", s.handlePurposes)

	// Test inspection endpoints — NOT part of the real OpenFGC API.
	// These allow test suites to query and reset the mock state without holding
	// a reference to the server struct (tests only have the base URL).
	mux.HandleFunc("/test/purposes", s.handleTestPurposes)
	mux.HandleFunc("/test/reset", s.handleTestReset)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}

	go func() {
		log.Printf("Starting mock consent server on port %d", s.port)
		if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("Mock consent server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the mock consent server.
func (s *MockConsentServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}

	return nil
}

// writeJSON writes a JSON response with the given status code.
func (s *MockConsentServer) writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeError writes a plain-text error response.
func (s *MockConsentServer) writeError(w http.ResponseWriter, status int, msg string) {
	http.Error(w, msg, status)
}

// --- Consent Elements handlers ---

// handleElements handles POST (batch create) and GET (list) on /api/v1/consent-elements.
func (s *MockConsentServer) handleElements(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleElementsCreate(w, r)
	case http.MethodGet:
		s.handleElementsList(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleElementsCreate handles POST /api/v1/consent-elements.
// Batch-creates consent elements; returns `elementsCreateResponseDTO`.
func (s *MockConsentServer) handleElementsCreate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var inputs []mockConsentElementCreateDTO
	if err := json.Unmarshal(body, &inputs); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.mu.Lock()
	created := make([]mockConsentElementDTO, 0, len(inputs))
	for _, inp := range inputs {
		el := mockConsentElement{
			mockConsentElementDTO: mockConsentElementDTO{
				ID:          s.nextIDLocked(),
				Name:        inp.Name,
				Description: inp.Description,
				Type:        inp.Type,
				Properties:  inp.Properties,
			},
		}
		s.elements[el.ID] = &el
		created = append(created, el.mockConsentElementDTO)
	}
	s.mu.Unlock()

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    created,
		"message": "Consent elements created successfully",
	})
}

// handleElementsList handles GET /api/v1/consent-elements with optional ?name= filter.
func (s *MockConsentServer) handleElementsList(w http.ResponseWriter, r *http.Request) {
	nameFilter := r.URL.Query().Get("name")

	s.mu.Lock()
	list := make([]mockConsentElementDTO, 0, len(s.elements))
	for _, el := range s.elements {
		if nameFilter == "" || el.Name == nameFilter {
			list = append(list, el.mockConsentElementDTO)
		}
	}
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": list})
}

// handleElementsValidate handles POST /api/v1/consent-elements/validate.
// Accepts a list of element names and returns the subset that already exist.
func (s *MockConsentServer) handleElementsValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var names []string
	if err := json.Unmarshal(body, &names); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build a name lookup from existing elements.
	s.mu.Lock()
	existingNames := make(map[string]bool, len(s.elements))
	for _, el := range s.elements {
		existingNames[el.Name] = true
	}
	s.mu.Unlock()

	// Return the intersection.
	found := make([]string, 0, len(names))
	for _, name := range names {
		if existingNames[name] {
			found = append(found, name)
		}
	}

	if len(found) == 0 {
		// The client treats 400 as "no elements matched" and returns an empty list.
		s.writeError(w, http.StatusBadRequest, "no matching consent elements found")
		return
	}

	s.writeJSON(w, http.StatusOK, found)
}

// handleElementByID handles PUT and DELETE on /api/v1/consent-elements/{id}.
func (s *MockConsentServer) handleElementByID(w http.ResponseWriter, r *http.Request) {
	elementID := strings.TrimPrefix(r.URL.Path, "/api/v1/consent-elements/")
	if elementID == "" {
		s.writeError(w, http.StatusBadRequest, "missing element ID")
		return
	}

	switch r.Method {
	case http.MethodPut:
		s.handleElementUpdate(w, r, elementID)
	case http.MethodDelete:
		s.handleElementDelete(w, r, elementID)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleElementUpdate handles PUT /api/v1/consent-elements/{id}.
func (s *MockConsentServer) handleElementUpdate(w http.ResponseWriter, r *http.Request, elementID string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var inp mockConsentElementCreateDTO
	if err := json.Unmarshal(body, &inp); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.mu.Lock()
	el, exists := s.elements[elementID]
	if !exists {
		s.mu.Unlock()
		s.writeError(w, http.StatusNotFound, "consent element not found")
		return
	}

	el.Name = inp.Name
	el.Description = inp.Description
	el.Type = inp.Type
	el.Properties = inp.Properties
	resp := el.mockConsentElementDTO
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, resp)
}

// handleElementDelete handles DELETE /api/v1/consent-elements/{id}.
func (s *MockConsentServer) handleElementDelete(w http.ResponseWriter, r *http.Request, elementID string) {
	s.mu.Lock()
	_, exists := s.elements[elementID]
	if !exists {
		s.mu.Unlock()
		s.writeError(w, http.StatusNotFound, "consent element not found")
		return
	}

	delete(s.elements, elementID)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// --- Consent Purposes handlers ---

// handlePurposes handles POST (create) and GET (list) on /api/v1/consent-purposes.
func (s *MockConsentServer) handlePurposes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handlePurposeCreate(w, r)
	case http.MethodGet:
		s.handlePurposesList(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handlePurposeCreate handles POST /api/v1/consent-purposes.
// The TPP-client-id header carries the application (group) ID.
func (s *MockConsentServer) handlePurposeCreate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var inp mockConsentPurposeCreateDTO
	if err := json.Unmarshal(body, &inp); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	clientID := r.Header.Get("TPP-client-id")
	now := time.Now().UnixMilli()

	s.mu.Lock()
	p := &mockConsentPurpose{
		ID:          s.nextIDLocked(),
		Name:        inp.Name,
		Description: inp.Description,
		ClientID:    clientID,
		Elements:    inp.Elements,
		CreatedTime: now,
		UpdatedTime: now,
	}
	s.purposes[p.ID] = p
	resp := *p
	s.mu.Unlock()

	s.writeJSON(w, http.StatusCreated, resp)
}

// handlePurposesList handles GET /api/v1/consent-purposes with optional ?clientIds= filter.
func (s *MockConsentServer) handlePurposesList(w http.ResponseWriter, r *http.Request) {
	clientIDFilter := r.URL.Query().Get("clientIds")

	s.mu.Lock()
	list := make([]mockConsentPurposeDTO, 0, len(s.purposes))
	for _, p := range s.purposes {
		if clientIDFilter == "" || p.ClientID == clientIDFilter {
			list = append(list, *p)
		}
	}
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": list})
}

// handlePurposeByID handles PUT and DELETE on /api/v1/consent-purposes/{id}.
func (s *MockConsentServer) handlePurposeByID(w http.ResponseWriter, r *http.Request) {
	purposeID := strings.TrimPrefix(r.URL.Path, "/api/v1/consent-purposes/")
	if purposeID == "" {
		s.writeError(w, http.StatusBadRequest, "missing purpose ID")
		return
	}

	switch r.Method {
	case http.MethodPut:
		s.handlePurposeUpdate(w, r, purposeID)
	case http.MethodDelete:
		s.handlePurposeDelete(w, r, purposeID)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handlePurposeUpdate handles PUT /api/v1/consent-purposes/{id}.
func (s *MockConsentServer) handlePurposeUpdate(w http.ResponseWriter, r *http.Request, purposeID string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var inp mockConsentPurposeCreateDTO
	if err := json.Unmarshal(body, &inp); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	clientID := r.Header.Get("TPP-client-id")

	s.mu.Lock()
	p, exists := s.purposes[purposeID]
	if !exists {
		s.mu.Unlock()
		s.writeError(w, http.StatusNotFound, "consent purpose not found")
		return
	}

	p.Name = inp.Name
	p.Description = inp.Description
	p.Elements = inp.Elements
	p.UpdatedTime = time.Now().UnixMilli()
	if clientID != "" {
		p.ClientID = clientID
	}
	resp := *p
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, resp)
}

// handlePurposeDelete handles DELETE /api/v1/consent-purposes/{id}.
func (s *MockConsentServer) handlePurposeDelete(w http.ResponseWriter, r *http.Request, purposeID string) {
	s.mu.Lock()
	_, exists := s.purposes[purposeID]
	if !exists {
		s.mu.Unlock()
		s.writeError(w, http.StatusNotFound, "consent purpose not found")
		return
	}

	delete(s.purposes, purposeID)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// --- Test inspection endpoints (not part of OpenFGC API) ---

// handleTestPurposes handles GET /test/purposes?clientIds=<appID>.
// Returns all purposes stored for the given clientID so tests can verify state.
func (s *MockConsentServer) handleTestPurposes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	clientIDFilter := r.URL.Query().Get("clientIds")

	s.mu.Lock()
	list := make([]mockConsentPurposeDTO, 0)
	for _, p := range s.purposes {
		if clientIDFilter == "" || p.ClientID == clientIDFilter {
			list = append(list, *p)
		}
	}
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, list)
}

// handleTestReset handles POST /test/reset.
// Clears all stored elements and purposes.
func (s *MockConsentServer) handleTestReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.mu.Lock()
	s.elements = make(map[string]*mockConsentElement)
	s.purposes = make(map[string]*mockConsentPurpose)
	s.idSeq = 0
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}
