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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/asgardeo/thunder/tests/integration/testutils"
)

const (
	serverPort = "8095"
)

func main() {

	// Step 1: Unzip the product
	err := testutils.UnzipProduct()
	if err != nil {
		fmt.Printf("Failed to unzip product: %v\n", err)
		return
	}

	// Step 2: Replace the resource files in the unzipped directory.
	err = testutils.ReplaceResources()
	if err != nil {
		fmt.Printf("Failed to replace resources: %v\n", err)
		return
	}

	// Step 3: Run the init script to create the SQLite database
	err = testutils.RunInitScript()
	if err != nil {
		fmt.Printf("Failed to run init script: %v\n", err)
		return
	}

	// Step 4: Start the server
	serverCmd, err := testutils.StartServer(serverPort)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	defer testutils.StopServer(serverCmd)

	// Wait for the server to start
	fmt.Println("Waiting for the server to start...")
	time.Sleep(5 * time.Second)

	// Step 5: Run all tests
	err = runTests()
	if err != nil {
		fmt.Printf("there are test failures: %v\n", err)
		testutils.StopServer(serverCmd)
		os.Exit(1)
	}
}

func runTests() error {
	cmd := exec.Command("go", "test", "-p=1", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
