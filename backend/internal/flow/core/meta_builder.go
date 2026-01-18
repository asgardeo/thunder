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

package core

import "github.com/asgardeo/thunder/internal/flow/common"

// MetaComponent represents a component in the meta structure for UI rendering.
type MetaComponent struct {
	Type        string          `json:"type"`
	ID          string          `json:"id"`
	Ref         string          `json:"ref,omitempty"`
	Label       string          `json:"label,omitempty"`
	Placeholder string          `json:"placeholder,omitempty"`
	Variant     string          `json:"variant,omitempty"`
	Required    bool            `json:"required,omitempty"`
	EventType   string          `json:"eventType,omitempty"`
	Options     []string        `json:"options,omitempty"`
	Components  []MetaComponent `json:"components,omitempty"`
}

// MetaStructure represents the complete meta structure for a node response.
type MetaStructure struct {
	Components []MetaComponent `json:"components"`
}

// MetaInputConfig contains configuration for a single input component.
type MetaInputConfig struct {
	Label       string
	Placeholder string
}

// MetaBuilder provides a fluent API for constructing meta structures.
// This allows executors to build meta in a readable, chainable way.
type MetaBuilder struct {
	idPrefix    string
	heading     string
	subtitle    string
	submitLabel string
	submitID    string
	inputs      []common.Input
	configs     map[string]MetaInputConfig
}

// NewMetaBuilder creates a new MetaBuilder instance with default settings.
func NewMetaBuilder() *MetaBuilder {
	return &MetaBuilder{
		idPrefix: "dynamic",
		configs:  make(map[string]MetaInputConfig),
	}
}

// WithIDPrefix sets the prefix for component IDs.
func (b *MetaBuilder) WithIDPrefix(prefix string) *MetaBuilder {
	b.idPrefix = prefix
	return b
}

// WithHeading sets the heading text.
func (b *MetaBuilder) WithHeading(heading string) *MetaBuilder {
	b.heading = heading
	return b
}

// WithSubtitle sets the subtitle text.
func (b *MetaBuilder) WithSubtitle(subtitle string) *MetaBuilder {
	b.subtitle = subtitle
	return b
}

// WithInput adds a single input along with its configurations.
func (b *MetaBuilder) WithInput(input common.Input, config MetaInputConfig) *MetaBuilder {
	b.inputs = append(b.inputs, input)
	b.configs[input.Identifier] = config
	return b
}

// WithInputs adds multiple inputs. Use WithInputConfig to set labels for each.
func (b *MetaBuilder) WithInputs(inputs []common.Input) *MetaBuilder {
	b.inputs = append(b.inputs, inputs...)
	return b
}

// WithInputConfig sets the configurations for a specific input identifier.
func (b *MetaBuilder) WithInputConfig(identifier string, config MetaInputConfig) *MetaBuilder {
	b.configs[identifier] = config
	return b
}

// WithSubmitButton configures the submit action button.
func (b *MetaBuilder) WithSubmitButton(label string) *MetaBuilder {
	b.submitLabel = label
	return b
}

// WithSubmitButtonID sets a custom ID for the submit button.
func (b *MetaBuilder) WithSubmitButtonID(id string) *MetaBuilder {
	b.submitID = id
	return b
}

// Build constructs the MetaStructure based on the provided configurations.
func (b *MetaBuilder) Build() MetaStructure {
	// Build block components from inputs
	blockComponents := make([]MetaComponent, 0, len(b.inputs)+1)
	for _, input := range b.inputs {
		config := b.configs[input.Identifier]

		component := MetaComponent{
			Type:        input.Type,
			ID:          input.Ref,        // Component ID = input ref (for frontend mapping)
			Ref:         input.Identifier, // Component ref = identifier (field name)
			Label:       config.Label,
			Placeholder: config.Placeholder,
			Required:    input.Required,
			Options:     input.Options, // Transfer options for SELECT components
		}

		// Use default labels if not provided
		if component.Label == "" {
			component.Label = input.Identifier
		}

		blockComponents = append(blockComponents, component)
	}

	// Add submit action if configured
	if b.submitLabel != "" {
		submitID := b.submitID
		if submitID == "" {
			submitID = b.idPrefix + "_submit"
		}

		blockComponents = append(blockComponents, MetaComponent{
			Type:      "ACTION",
			ID:        submitID,
			Label:     b.submitLabel,
			Variant:   "PRIMARY",
			EventType: "SUBMIT",
		})
	}

	// Build top-level components
	components := make([]MetaComponent, 0, 3)

	// Add heading if provided
	if b.heading != "" {
		components = append(components, MetaComponent{
			Type:    "TEXT",
			ID:      b.idPrefix + "_heading",
			Label:   b.heading,
			Variant: "HEADING_2",
		})
	}

	// Add subtitle if provided
	if b.subtitle != "" {
		components = append(components, MetaComponent{
			Type:    "TEXT",
			ID:      b.idPrefix + "_subtitle",
			Label:   b.subtitle,
			Variant: "BODY_1",
		})
	}

	// Add block containing inputs
	components = append(components, MetaComponent{
		Type:       "BLOCK",
		ID:         b.idPrefix + "_block",
		Components: blockComponents,
	})

	return MetaStructure{Components: components}
}
