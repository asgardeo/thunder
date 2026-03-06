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

	// Single action with no inputs - should NOT auto-complete (confirmation prompts wait for explicit action)
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
	s.Equal(common.NodeStatusIncomplete, resp.Status, "Confirmation prompt should wait for explicit action")
	s.Len(resp.Actions, 1, "Should return the action for user to select")
	s.Equal("continue", resp.Actions[0].Ref)
	s.Equal("", ctx.CurrentAction, "Context should NOT have auto-selected action for confirmation prompts")
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

func (s *PromptOnlyNodeTestSuite) TestExecuteWithFailureReason() {
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

	// Context with failure reason in runtime data
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		RuntimeData: map[string]string{
			"failureReason": "Authentication failed",
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal("Authentication failed", resp.FailureReason, "Should include failure reason in response")
	s.NotContains(ctx.RuntimeData, "failureReason", "Should delete failure reason from runtime data")
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithEmptyFailureReason() {
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

	// Context with empty failure reason
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		RuntimeData: map[string]string{
			"failureReason": "",
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal("", resp.FailureReason, "Should not set failure reason when empty")
	s.Contains(ctx.RuntimeData, "failureReason", "Should not delete empty failure reason from runtime data")
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithNilRuntimeData() {
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

	// Context with nil runtime data
	ctx := &NodeContext{
		FlowID:      "test-flow",
		UserInputs:  map[string]string{},
		RuntimeData: nil,
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal("", resp.FailureReason, "Should handle nil runtime data gracefully")
}

func (s *PromptOnlyNodeTestSuite) TestExecuteInvalidActionReturnsFailure() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	// Setup prompts with specific actions
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "password", Required: true},
			},
			Action: &common.Action{Ref: "valid_action", NextNode: "next_node"},
		},
	})

	// Provide all required inputs but with an action that matches but has no nextNode
	// This simulates when getNextNodeForActionRef returns empty string
	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "valid_action",
		UserInputs: map[string]string{
			"username": "testuser",
			"password": "testpass",
		},
	}

	// Temporarily modify the prompt to have empty nextNode
	prompts := promptNode.GetPrompts()
	originalNextNode := prompts[0].Action.NextNode
	prompts[0].Action.NextNode = ""
	promptNode.SetPrompts(prompts)

	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusFailure, resp.Status, "Should return failure status")
	s.Equal("Invalid action selected", resp.FailureReason, "Should set failure reason")

	// Restore for other tests
	prompts[0].Action.NextNode = originalNextNode
	promptNode.SetPrompts(prompts)
}

func (s *PromptOnlyNodeTestSuite) TestGetAndSetPrompts() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	// Initially should be empty
	prompts := promptNode.GetPrompts()
	s.NotNil(prompts)
	s.Len(prompts, 0)

	// Set prompts
	testPrompts := []common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "password", Required: true},
			},
			Action: &common.Action{Ref: "login", NextNode: "auth_node"},
		},
		{
			Action: &common.Action{Ref: "signup", NextNode: "register_node"},
		},
	}
	promptNode.SetPrompts(testPrompts)

	// Verify prompts are set
	retrievedPrompts := promptNode.GetPrompts()
	s.Len(retrievedPrompts, 2)
	s.Equal("username", retrievedPrompts[0].Inputs[0].Identifier)
	s.Equal("login", retrievedPrompts[0].Action.Ref)
	s.Equal("signup", retrievedPrompts[1].Action.Ref)
}

func (s *PromptOnlyNodeTestSuite) TestGetAllInputs() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(*promptNode)

	// Set multiple prompts with various inputs
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "password", Required: true},
			},
			Action: &common.Action{Ref: "login", NextNode: "auth_node"},
		},
		{
			Inputs: []common.Input{
				{Identifier: "email", Required: true},
			},
			Action: &common.Action{Ref: "signup", NextNode: "register_node"},
		},
		{
			// Prompt with no inputs
			Action: &common.Action{Ref: "cancel", NextNode: "exit_node"},
		},
	})

	// Test getAllInputs
	allInputs := promptNode.getAllInputs()
	s.Len(allInputs, 3, "Should return all inputs from all prompts")
	s.Equal("username", allInputs[0].Identifier)
	s.Equal("password", allInputs[1].Identifier)
	s.Equal("email", allInputs[2].Identifier)
}

func (s *PromptOnlyNodeTestSuite) TestGetAllInputsEmpty() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(*promptNode)

	// No prompts set
	allInputs := promptNode.getAllInputs()
	s.NotNil(allInputs)
	s.Len(allInputs, 0, "Should return empty slice when no prompts")
}

func (s *PromptOnlyNodeTestSuite) TestGetAllActions() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(*promptNode)

	// Set multiple prompts with actions
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
			},
			Action: &common.Action{Ref: "login", NextNode: "auth_node"},
		},
		{
			Action: &common.Action{Ref: "signup", NextNode: "register_node"},
		},
		{
			Inputs: []common.Input{
				{Identifier: "email", Required: true},
			},
			Action: &common.Action{Ref: "reset", NextNode: "reset_node"},
		},
	})

	// Test getAllActions
	allActions := promptNode.getAllActions()
	s.Len(allActions, 3, "Should return all actions from all prompts")
	s.Equal("login", allActions[0].Ref)
	s.Equal("signup", allActions[1].Ref)
	s.Equal("reset", allActions[2].Ref)
}

func (s *PromptOnlyNodeTestSuite) TestGetAllActionsWithNilAction() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(*promptNode)

	// Set prompts with some nil actions
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
			},
			Action: &common.Action{Ref: "login", NextNode: "auth_node"},
		},
		{
			Inputs: []common.Input{
				{Identifier: "email", Required: true},
			},
			Action: nil, // No action
		},
		{
			Action: &common.Action{Ref: "signup", NextNode: "register_node"},
		},
	})

	// Test getAllActions - should only return non-nil actions
	allActions := promptNode.getAllActions()
	s.Len(allActions, 2, "Should only return non-nil actions")
	s.Equal("login", allActions[0].Ref)
	s.Equal("signup", allActions[1].Ref)
}

func (s *PromptOnlyNodeTestSuite) TestGetAllActionsEmpty() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(*promptNode)

	// No prompts set
	allActions := promptNode.getAllActions()
	s.NotNil(allActions)
	s.Len(allActions, 0, "Should return empty slice when no prompts")
}

func (s *PromptOnlyNodeTestSuite) TestGetNextNodeForActionRef() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(*promptNode)

	// Set prompts with multiple actions
	promptNode.SetPrompts([]common.Prompt{
		{
			Action: &common.Action{Ref: "login", NextNode: "auth_node"},
		},
		{
			Action: &common.Action{Ref: "signup", NextNode: "register_node"},
		},
		{
			Action: &common.Action{Ref: "cancel", NextNode: "exit_node"},
		},
	})

	// Test finding existing actions
	nextNode := promptNode.getNextNodeForActionRef("login", promptNode.logger)
	s.Equal("auth_node", nextNode)

	nextNode = promptNode.getNextNodeForActionRef("signup", promptNode.logger)
	s.Equal("register_node", nextNode)

	nextNode = promptNode.getNextNodeForActionRef("cancel", promptNode.logger)
	s.Equal("exit_node", nextNode)
}

func (s *PromptOnlyNodeTestSuite) TestGetNextNodeForActionRefNotFound() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(*promptNode)

	// Set prompts with actions
	promptNode.SetPrompts([]common.Prompt{
		{
			Action: &common.Action{Ref: "login", NextNode: "auth_node"},
		},
		{
			Action: &common.Action{Ref: "signup", NextNode: "register_node"},
		},
	})

	// Test finding non-existent action
	nextNode := promptNode.getNextNodeForActionRef("nonexistent", promptNode.logger)
	s.Equal("", nextNode, "Should return empty string when action not found")
}

func (s *PromptOnlyNodeTestSuite) TestGetNextNodeForActionRefEmptyRef() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(*promptNode)

	// Set prompts with actions
	promptNode.SetPrompts([]common.Prompt{
		{
			Action: &common.Action{Ref: "login", NextNode: "auth_node"},
		},
	})

	// Test with empty action ref
	nextNode := promptNode.getNextNodeForActionRef("", promptNode.logger)
	s.Equal("", nextNode, "Should return empty string for empty action ref")
}

func (s *PromptOnlyNodeTestSuite) TestAutoSelectClearsActionsFromResponse() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)

	// Single action with inputs - should auto-select
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
			},
			Action: &common.Action{Ref: "submit", NextNode: "auth_node"},
		},
	})

	ctx := &NodeContext{
		FlowID:        "test-flow",
		CurrentAction: "", // No action selected
		UserInputs: map[string]string{
			"username": "testuser",
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status)
	s.Len(resp.Actions, 0, "Actions should be cleared after auto-selection")
	s.Equal("submit", ctx.CurrentAction, "Action should be auto-selected in context")
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithFailureAndRecovery() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{Identifier: "password", Required: true},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	// First execution with failure
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		RuntimeData: map[string]string{
			"failureReason": "Invalid credentials",
			"otherData":     "should remain",
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal("Invalid credentials", resp.FailureReason)
	s.NotContains(ctx.RuntimeData, "failureReason", "Failure reason should be removed")
	s.Contains(ctx.RuntimeData, "otherData", "Other runtime data should remain")

	// Second execution with correct inputs (recovery)
	ctx.CurrentAction = "submit"
	ctx.UserInputs = map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	resp, err = node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status)
	s.Equal("", resp.FailureReason, "Should not have failure reason on success")
	s.Equal("next", resp.NextNodeID)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithForwardedDataOptions() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{
					Ref:        "usertype_input",
					Identifier: "userType",
					Type:       "SELECT",
					Required:   true,
					Options:    []string{}, // Empty in prompt definition
				},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	// Execute with ForwardedData containing inputs with options
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		ForwardedData: map[string]interface{}{
			common.ForwardedDataKeyInputs: []common.Input{
				{
					Identifier: "userType",
					Type:       "SELECT",
					Options:    []string{"employee", "customer", "partner"},
				},
			},
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Len(resp.Inputs, 1)

	// Verify the input is enriched with options from ForwardedData
	enrichedInput := resp.Inputs[0]
	s.Equal("userType", enrichedInput.Identifier)
	s.Equal("usertype_input", enrichedInput.Ref, "Ref from prompt definition should be preserved")
	s.Equal("SELECT", enrichedInput.Type, "Type from prompt definition should be preserved")
	s.True(enrichedInput.Required, "Required from prompt definition should be preserved")
	s.ElementsMatch([]string{"employee", "customer", "partner"}, enrichedInput.Options,
		"Options should be enriched from ForwardedData")
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithForwardedDataNoMatch() {
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

	// ForwardedData has inputs but different Identifier
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		ForwardedData: map[string]interface{}{
			common.ForwardedDataKeyInputs: []common.Input{
				{
					Identifier: "userType", // Different identifier
					Options:    []string{"option1", "option2"},
				},
			},
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Len(resp.Inputs, 1)

	// Verify prompt input is unchanged since no match
	promptInput := resp.Inputs[0]
	s.Equal("username", promptInput.Identifier)
	s.Empty(promptInput.Options, "Options should remain empty when no matching forwarded input")
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithNoForwardedData() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "userType", Type: "SELECT", Required: true, Options: []string{}},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	// No ForwardedData
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Len(resp.Inputs, 1)

	// Verify options remain empty
	promptInput := resp.Inputs[0]
	s.Equal("userType", promptInput.Identifier)
	s.Empty(promptInput.Options, "Options should remain empty without ForwardedData")
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithForwardedDataMultipleInputs() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "userType", Type: "SELECT", Required: true, Options: []string{}},
				{Identifier: "region", Type: "SELECT", Required: true, Options: []string{}},
				{Identifier: "username", Type: "TEXT", Required: true},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	// ForwardedData has options for only userType
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		ForwardedData: map[string]interface{}{
			common.ForwardedDataKeyInputs: []common.Input{
				{
					Identifier: "userType",
					Options:    []string{"employee", "customer"},
				},
			},
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Len(resp.Inputs, 3)

	// Find each input and verify
	var userTypeInput, regionInput, usernameInput *common.Input
	for i := range resp.Inputs {
		switch resp.Inputs[i].Identifier {
		case "userType":
			userTypeInput = &resp.Inputs[i]
		case "region":
			regionInput = &resp.Inputs[i]
		case "username":
			usernameInput = &resp.Inputs[i]
		}
	}

	s.NotNil(userTypeInput)
	s.NotNil(regionInput)
	s.NotNil(usernameInput)

	// Only userType should be enriched
	s.ElementsMatch([]string{"employee", "customer"}, userTypeInput.Options)
	s.Empty(regionInput.Options, "Region options should remain empty")
	s.Empty(usernameInput.Options, "Username should have no options")
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithForwardedDataNonInputType() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "userType", Required: true},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	// ForwardedData has wrong type (string instead of []common.Input)
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		ForwardedData: map[string]interface{}{
			common.ForwardedDataKeyInputs: "not-an-input-slice",
		},
	}

	// Should not panic, should handle gracefully
	s.NotPanics(func() {
		resp, err := node.Execute(ctx)
		s.Nil(err)
		s.NotNil(resp)
		s.Len(resp.Inputs, 1)
		s.Empty(resp.Inputs[0].Options, "Options should remain empty with invalid ForwardedData type")
	})
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithForwardedDataPreservesPromptFields() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{
					Ref:        "usertype_input_custom",
					Identifier: "userType",
					Type:       "SELECT",
					Required:   true,
					Options:    []string{},
				},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	// ForwardedData has different Ref and Type (should NOT overwrite)
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		ForwardedData: map[string]interface{}{
			common.ForwardedDataKeyInputs: []common.Input{
				{
					Ref:        "different_ref",     // Should NOT overwrite
					Identifier: "userType",          // Match by this
					Type:       "DIFFERENT_TYPE",    // Should NOT overwrite
					Required:   false,               // Should NOT overwrite
					Options:    []string{"option1"}, // Should enrich
				},
			},
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Len(resp.Inputs, 1)

	// Verify only Options is enriched, other fields preserved from prompt definition
	enrichedInput := resp.Inputs[0]
	s.Equal("usertype_input_custom", enrichedInput.Ref, "Ref should NOT be overwritten")
	s.Equal("userType", enrichedInput.Identifier)
	s.Equal("SELECT", enrichedInput.Type, "Type should NOT be overwritten")
	s.True(enrichedInput.Required, "Required should NOT be overwritten")
	s.ElementsMatch([]string{"option1"}, enrichedInput.Options, "Only Options should be enriched")
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithForwardedDataEmptyOptions() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	promptNode := node.(PromptNodeInterface)
	promptNode.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "userType", Type: "SELECT", Required: true, Options: []string{"default"}},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	// ForwardedData has matching input but with empty options
	ctx := &NodeContext{
		FlowID:     "test-flow",
		UserInputs: map[string]string{},
		ForwardedData: map[string]interface{}{
			common.ForwardedDataKeyInputs: []common.Input{
				{
					Identifier: "userType",
					Options:    []string{}, // Empty options
				},
			},
		},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Len(resp.Inputs, 1)

	// Verify options are NOT enriched when ForwardedData has empty options
	promptInput := resp.Inputs[0]
	s.Equal("userType", promptInput.Identifier)
	s.ElementsMatch([]string{"default"}, promptInput.Options,
		"Options should not be overwritten with empty options from ForwardedData")
}
