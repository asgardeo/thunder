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

package core

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
)

type PromptOnlyNodeTestSuite struct {
	suite.Suite
}

func TestPromptOnlyNodeTestSuite(t *testing.T) {
	suite.Run(t, new(PromptOnlyNodeTestSuite))
}

func (s *PromptOnlyNodeTestSuite) TestNewPromptOnlyNode() {
	node := newPromptNode("prompt-1", map[string]interface{}{"key": "value"}, true, false)

	s.NotNil(node)
	s.Equal("prompt-1", node.GetID())
	s.Equal(common.NodeTypePrompt, node.GetType())
	s.True(node.IsStartNode())
	s.False(node.IsFinalNode())
}

func (s *PromptOnlyNodeTestSuite) TestExecuteNoInputs() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	ctx := &NodeContext{FlowID: "test-flow", UserInputs: map[string]string{}}

	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status)
	s.Equal(common.NodeResponseType(""), resp.Type)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithRequiredData() {
	tests := []struct {
		name           string
		userInputs     map[string]string
		expectComplete bool
		requiredCount  int
	}{
		{"No user input provided", map[string]string{}, false, 2},
		{
			"All required data provided",
			map[string]string{"username": "testuser", "email": "test@example.com"},
			true,
			0,
		},
		{"Partial data provided", map[string]string{"username": "testuser"}, false, 1},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
			promptNode := node.(PromptNodeInterface)
			promptNode.SetPrompts([]common.Prompt{
				{
					Inputs: []common.Input{
						{Identifier: "username", Required: true},
						{Identifier: "email", Required: true},
					},
					Action: &common.Action{Ref: "submit", NextNode: "next"},
				},
			})

			ctx := &NodeContext{FlowID: "test-flow", CurrentAction: "submit", UserInputs: tt.userInputs}
			resp, err := node.Execute(ctx)

			s.Nil(err)
			s.NotNil(resp)

			if tt.expectComplete {
				s.Equal(common.NodeStatusComplete, resp.Status)
				s.Equal(common.NodeResponseType(""), resp.Type)
			} else {
				s.Equal(common.NodeStatusIncomplete, resp.Status)
				s.Equal(common.NodeResponseTypeView, resp.Type)
				s.Len(resp.Inputs, tt.requiredCount)
			}
		})
	}
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithOptionalData() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "nickname", Required: false},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "submit",
		UserInputs:    map[string]string{"username": "testuser"},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status)
	s.Equal(common.NodeResponseType(""), resp.Type)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteMissingRequiredOnly() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "nickname", Required: false},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "submit",
		UserInputs:    map[string]string{"nickname": "testnick"},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal(common.NodeResponseTypeView, resp.Type)
	s.Len(resp.Inputs, 1)

	foundRequired := false
	for _, data := range resp.Inputs {
		if data.Identifier == "username" && data.Required {
			foundRequired = true
		}
	}
	s.True(foundRequired)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithVerboseModeEnabled() {
	meta := map[string]interface{}{
		"components": []interface{}{
			map[string]interface{}{
				"type":  "TEXT",
				"id":    "text_001",
				"label": "Welcome",
			},
		},
	}

	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetMeta(meta)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	// Test with verbose mode enabled
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		Verbose:    true,
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal(common.NodeResponseTypeView, resp.Type)
	s.NotNil(resp.Meta)
	s.Equal(meta, resp.Meta)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithVerboseModeDisabled() {
	meta := map[string]interface{}{
		"components": []interface{}{
			map[string]interface{}{
				"type":  "TEXT",
				"id":    "text_001",
				"label": "Welcome",
			},
		},
	}

	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetMeta(meta)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	// Test with verbose mode disabled (default)
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		Verbose:    false,
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal(common.NodeResponseTypeView, resp.Type)
	s.Nil(resp.Meta)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteVerboseModeNoMeta() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	// Test with verbose mode enabled but no meta defined
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		Verbose:    true,
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal(common.NodeResponseTypeView, resp.Type)
	s.Nil(resp.Meta)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithSets_ActionWithInputs() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "password", Required: true},
			},
			Action: &common.Action{Ref: "action_001", NextNode: "basic_auth"},
		},
		{
			Action: &common.Action{Ref: "action_002", NextNode: "google_auth"},
		},
	})

	// Select action_001 but don't provide inputs
	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "action_001",
		UserInputs:    map[string]string{},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Len(resp.Inputs, 2)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithSets_ActionWithoutInputs() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "password", Required: true},
			},
			Action: &common.Action{Ref: "action_001", NextNode: "basic_auth"},
		},
		{
			Action: &common.Action{Ref: "action_002", NextNode: "google_auth"},
		},
	})

	// Select action_002 which has no inputs
	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "action_002",
		UserInputs:    map[string]string{},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status)
	s.Equal("google_auth", resp.NextNodeID)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithSets_ActionWithInputsProvided() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "password", Required: true},
			},
			Action: &common.Action{Ref: "action_001", NextNode: "basic_auth"},
		},
	})

	// Select action_001 with all inputs provided
	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "action_001",
		UserInputs: map[string]string{
			"username": "testuser",
			"password": "testpass",
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status)
	s.Equal("basic_auth", resp.NextNodeID)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithSets_NoActionSelected() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{{Identifier: "username", Required: true}},
			Action: &common.Action{Ref: "action_001", NextNode: "basic_auth"},
		},
		{
			Action: &common.Action{Ref: "action_002", NextNode: "google_auth"},
		},
	})

	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "",
		UserInputs:    map[string]string{},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Len(resp.Actions, 2)
	s.Len(resp.Inputs, 1, "Should return all inputs from sets when no action selected")
	s.Equal("username", resp.Inputs[0].Identifier)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithInvalidAction() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
			},
			Action: &common.Action{Ref: "login", NextNode: "auth"},
		},
	})

	// Select an action that doesn't exist
	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "unknown_action",
		UserInputs:    map[string]string{},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	// Should treat as no action selected - return both inputs and actions
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Len(resp.Inputs, 1)
	s.Equal("username", resp.Inputs[0].Identifier)
	s.Len(resp.Actions, 1, "Should return actions when invalid action is provided")
	s.Equal("login", resp.Actions[0].Ref)
}

func (s *PromptOnlyNodeTestSuite) TestAutoSelectSingleAction_NoInputs() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	// Single action with no inputs - should auto-complete
	promptNode.SetPrompts([]common.Prompt{
		{
			Action: &common.Action{Ref: "continue", NextNode: "next_node"},
		},
	})

	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "", // No action selected
		UserInputs:    map[string]string{},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status, "Single action should be auto-selected")
	s.Equal("next_node", resp.NextNodeID, "Should route to the next node for auto-selected action")
	s.Empty(resp.Actions, "No actions should be returned when auto-selected")
	s.Equal("continue", ctx.CurrentAction, "Context should have the auto-selected action")
}

func (s *PromptOnlyNodeTestSuite) TestAutoSelectSingleAction_WithInputsProvided() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	// Single action with inputs - should auto-select and validate inputs
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "password", Required: true},
			},
			Action: &common.Action{Ref: "submit", NextNode: "auth_node"},
		},
	})

	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "", // No action selected
		UserInputs: map[string]string{
			"username": "testuser",
			"password": "testpass",
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status,
		"Should complete when single action auto-selected and inputs provided")
	s.Equal("auth_node", resp.NextNodeID)
	s.Equal("submit", ctx.CurrentAction, "Context should have the auto-selected action")
}

func (s *PromptOnlyNodeTestSuite) TestAutoSelectSingleAction_WithMissingInputs() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	// Single action with required inputs missing - should NOT auto-select
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "password", Required: true},
			},
			Action: &common.Action{Ref: "submit", NextNode: "auth_node"},
		},
	})

	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "",                  // No action selected
		UserInputs:    map[string]string{}, // No inputs
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status, "Should be incomplete when inputs are missing")
	s.Len(resp.Inputs, 2, "Should return missing inputs")
	s.Len(resp.Actions, 1, "Actions should be returned when inputs are missing (no auto-select)")
	s.Equal("submit", resp.Actions[0].Ref, "Action should be in the response")
	s.Equal("", ctx.CurrentAction, "Context should NOT have auto-selected action when inputs missing")
}

func (s *PromptOnlyNodeTestSuite) TestAutoSelectSingleAction_MultipleActionsNoAutoSelect() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	// Multiple actions - should not auto-select
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{{Identifier: "username", Required: true}},
			Action: &common.Action{Ref: "basic_auth", NextNode: "basic_node"},
		},
		{
			Action: &common.Action{Ref: "social_auth", NextNode: "social_node"},
		},
	})

	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "", // No action selected
		UserInputs:    map[string]string{},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status, "Should be incomplete when multiple actions exist")
	s.Len(resp.Actions, 2, "Should return all actions when multiple exist")
	s.Equal("", ctx.CurrentAction, "Context should NOT have an auto-selected action with multiple actions")
}
