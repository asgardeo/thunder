/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package model

// ExecutorResponse represents the response from an executor
type ExecutorResponse struct{}

// ExecutorInterface defines the interface for executors.
type ExecutorInterface interface {
	Execute(ctx *FlowContext) (*ExecutorResponse, error)
}

// // Executor represents the basic implementation of an executor.
// type Executor struct {
// 	ID   string
// 	Name string
// }

// // NewExecutor creates a new executor with the given ID and name.
// func NewExecutor(id string, name string) ExecutorInterface {
// 	return &Executor{
// 		ID:   id,
// 		Name: name,
// 	}
// }

// // Execute executes the executor logic.
// func (e *Executor) Execute(ctx *model.FlowContext) (*model.ExecutorResponse, error) {
// 	// Implement the logic for executing the executor
// 	// This is just a placeholder implementation
// 	return &model.ExecutorResponse{}, nil
// }
