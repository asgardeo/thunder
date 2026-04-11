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

package log

const (
	// LoggerKeyComponentName is the key used to identify the component name in the logger.
	LoggerKeyComponentName = "component"
	// LoggerKeyExecutorName is the key used to identify the executor ID in the logger.
	LoggerKeyExecutorName = "executorName"
	// LoggerKeyExecutionID is the key used to identify the flow execution ID in the logger.
	LoggerKeyExecutionID = "executionID"
	// LoggerKeyNodeID is the key used to identify the node ID in the logger.
	LoggerKeyNodeID = "nodeId"
	// LoggerKeyTraceID is the key used to identify the trace ID (correlation ID) in the logger.
	LoggerKeyTraceID = "trace_id"
)
