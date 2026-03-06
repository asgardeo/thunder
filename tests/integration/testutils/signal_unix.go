//go:build !windows

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
	"os"
	"syscall"
)

// sendStopSignal sends SIGTERM to the process for graceful shutdown on Unix.
func sendStopSignal(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}

// isProcessAlive reports whether the process identified by proc is still running.
// On Unix the null signal (signal 0) is used: syscall.Kill returns ESRCH when the
// PID no longer exists, which also protects against accidentally killing a recycled
// PID after a grace-period sleep.
func isProcessAlive(proc *os.Process) bool {
	return proc.Signal(syscall.Signal(0)) == nil
}
