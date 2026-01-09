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

package export

import immutableresource "github.com/asgardeo/thunder/internal/system/immutable_resource"

// ExportRequest represents the request structure for exporting resources.
type ExportRequest struct {
	Applications        []string `json:"applications,omitempty"`
	IdentityProviders   []string `json:"identity_providers,omitempty"`
	NotificationSenders []string `json:"notification_senders,omitempty"`
	UserSchemas         []string `json:"user_schemas,omitempty"`
	OrganizationUnits   []string `json:"organization_units,omitempty"`
	Flows               []string `json:"flows,omitempty"`
	Translations        []string `json:"translations,omitempty"`

	Options *ExportOptions `json:"options,omitempty"`
}

// ExportOptions provides configuration for export behavior.
type ExportOptions struct {
	// IncludeMetadata determines whether to include metadata (creation dates, IDs, etc.)
	IncludeMetadata bool `json:"include_metadata,omitempty"`

	// IncludeDependencies automatically exports related resources
	IncludeDependencies bool `json:"include_dependencies,omitempty"`

	// Format specifies the output format for individual files (yaml, json)
	Format string `json:"format,omitempty"` // Default: "yaml"

	// Folder structure options
	FolderStructure *FolderStructureOptions `json:"folder_structure,omitempty"`

	// Pagination for bulk exports
	Pagination *PaginationOptions `json:"pagination,omitempty"`
}

// FolderStructureOptions configures how files are organized in exports.
type FolderStructureOptions struct {
	// GroupByType creates separate folders for each resource type
	GroupByType bool `json:"group_by_type,omitempty"`

	// CustomStructure allows defining custom folder paths
	CustomStructure map[string]string `json:"custom_structure,omitempty"`

	// FileNamingPattern defines how files should be named
	FileNamingPattern string `json:"file_naming_pattern,omitempty"` // e.g., "${name}_${id}", "${type}_${name}"
}

// PaginationOptions configures pagination for bulk exports.
type PaginationOptions struct {
	// Page number (1-based)
	Page int `json:"page,omitempty"`

	// Number of resources per page
	Limit int `json:"limit,omitempty"`
}

// ExportResponse represents the response structure for exporting resources.
type ExportResponse struct {
	Files []ExportFile `json:"files"`

	// Summary information about the export
	Summary *ExportSummary `json:"summary,omitempty"`
}

// ExportSummary provides metadata about the export operation.
type ExportSummary struct {
	TotalFiles    int                             `json:"total_files"`
	TotalSize     int64                           `json:"total_size_bytes,omitempty"`
	ExportedAt    string                          `json:"exported_at,omitempty"`
	ResourceTypes map[string]int                  `json:"resource_types,omitempty"` // Type -> count
	Errors        []immutableresource.ExportError `json:"errors,omitempty"`
	Pagination    *PaginationInfo                 `json:"pagination,omitempty"`
}

// ExportError represents errors that occurred during export.
// Deprecated: Use immutableresource.ExportError instead.
type ExportError = immutableresource.ExportError

// PaginationInfo provides pagination metadata.
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalPages int  `json:"total_pages,omitempty"`
	HasMore    bool `json:"has_more"`
}

// ExportFile represents a single YAML file in the export response.
type ExportFile struct {
	FileName     string `json:"file_name"`
	Content      string `json:"content"`
	FolderPath   string `json:"folder_path,omitempty"`   // Relative path within the export
	ResourceType string `json:"resource_type,omitempty"` // application, group, user, idp
	ResourceID   string `json:"resource_id,omitempty"`   // ID of the exported resource
	Size         int64  `json:"size,omitempty"`          // File size in bytes
}
