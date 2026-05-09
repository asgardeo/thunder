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

// Package reactsdk provides MCP tools for integrating with the Asgardeo React SDK.
package reactsdk

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTools registers all React SDK tools with the MCP server.
func RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "thunderid_integrate_react_sdk",
		Description: "Provides instructions and code snippets for integrating ThunderID authentication via the " +
			"Asgardeo React SDK into a React application. Supports two modes: Mode 1 (default) - ThunderID-hosted " +
			"login pages with redirect-based OAuth 2.0/OIDC flow. Mode 2 - Self-hosted login pages using Flow API " +
			"or direct API calls for custom authentication UI.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Integrate React SDK",
			IdempotentHint: true,
		},
	}, integrateReactSDK)
}

// integrateReactSDK handles the integrate_react_sdk tool call.
func integrateReactSDK(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input integrateReactSDKInput,
) (*mcp.CallToolResult, integrateReactSDKOutput, error) {
	rawInstructions := `
# ThunderID Authentication – React Integration Instructions

## Two Integration Modes

ThunderID supports two ways to integrate React authentication:

- **Mode 1 (Default - Recommended)**: ThunderID-hosted login pages with redirect-based
  OAuth 2.0/OIDC flow using the Asgardeo React SDK. Minimal configuration,
  recommended for most use cases.
- **Mode 2**: Self-hosted custom login pages using direct API calls or flow
  orchestration for maximum control. Refer to the React Vanilla Sample or React
  API-Based Sample for implementation details.

This guide covers both integration modes:
- **Mode 1** (default): ThunderID-hosted login with redirect-based OAuth 2.0/OIDC.
- **Mode 2**: Self-hosted app-native login using alternate API-driven patterns.

Mode 1 remains the recommended default path and uses the **Asgardeo React SDK**
for a minimal, declarative authentication flow.

## Project Context
This project is a **React application** that integrates **ThunderID
authentication** using the **Asgardeo React SDK**, covering both
ThunderID-hosted login (Mode 1) and app-native/self-hosted flows (Mode 2).

The goal is to demonstrate a **minimal, declarative authentication flow** with minimal setup.

This project intentionally avoids custom logic, hooks, and advanced configuration.

---

## SDK & Platform
- **SDK**: @asgardeo/react
- **Platform**: AsgardeoV2
- **Authentication Mode**: ThunderID-hosted login (redirect-based)
- **Base URL**: https://localhost:8090 (or ThunderID instance URL)
- **Client ID**: <CREATED_APP'S_CLIENT_ID>

---

## CRITICAL: Mode 1 Provider Configuration Rules (MUST FOLLOW EXACTLY)

For **Mode 1 only**, ~AsgardeoProvider~ **MUST** be configured using **ONLY**
the following three props.
Always create a NEW public SPA application with token endpoint method as none and use the 
client id. Use the application's URL as the redirect URL.
Use the thunderid_get_application_templates tool to get the template and defaults(e.g. user attributes and scopes)
before creating the application.
No variations, no abstractions, no helper objects.

**Mode 2 note:** Method 2 uses an alternate ~AsgardeoProvider~ shape with
~applicationId~ (instead of ~clientId~), as shown in the Mode 2 examples.

### ✅ REQUIRED Provider Configuration
~~~jsx
import { AsgardeoProvider } from '@asgardeo/react'

<AsgardeoProvider
  clientId="<client-id>"
  baseUrl="https://localhost:8090"
  platform="AsgardeoV2"
>
  <App />
</AsgardeoProvider>
~~~

### 🚨 FORBIDDEN Patterns
**For Mode 1, NEVER** do any of the following:
- ❌ ~const config = { ... }; <AsgardeoProvider {...config} />~
- ❌ Extract props to variables
- ❌ Add any other props beyond the three required ones
- ❌ Use different prop names or aliases

---

## Application Structure

### Entry Point (main.jsx or index.jsx)
~~~jsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { AsgardeoProvider } from '@asgardeo/react'
import App from './App.jsx'
import './index.css'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <AsgardeoProvider
      clientId="<client-id>"
      baseUrl="https://localhost:8090"
      platform="AsgardeoV2"
    >
      <App />
    </AsgardeoProvider>
  </StrictMode>
)
~~~

---

## Authentication Components

### Using Pre-built Components (Recommended for Simplicity)

The SDK provides declarative components for handling auth states:

#### 1. Sign In/Out Buttons
~~~jsx
import { SignInButton, SignOutButton } from '@asgardeo/react'

function Navigation() {
  return (
    <nav>
      <SignInButton>Sign In</SignInButton>
      <SignOutButton>Sign Out</SignOutButton>
    </nav>
  )
}
~~~

#### 2. Conditional Rendering Based on Auth State
~~~jsx
import { SignedIn, SignedOut, Loading } from '@asgardeo/react'

function App() {
  return (
    <>
      <Loading>
        <div>Loading authentication...</div>
      </Loading>

      <SignedOut>
        <h1>Welcome! Please sign in.</h1>
        <SignInButton>Sign In</SignInButton>
      </SignedOut>

      <SignedIn>
        <h1>Welcome back!</h1>
        <SignOutButton>Sign Out</SignOutButton>
      </SignedIn>
    </>
  )
}
~~~

#### 3. Display User Information

**PREFERRED:** Use the ~User~ component from ~@asgardeo/react~ with render props pattern:

~~~jsx
import { SignedIn, User } from '@asgardeo/react'

function UserProfile() {
  return (
    <SignedIn>
      <div>
        <h2>User Profile</h2>
        <User>
          {(user) => user && (
            <>
              {user.picture && (
                <img 
                  src={user.picture} 
                  alt={user.name || 'User avatar'} 
                  style={{ width: '80px', height: '80px', borderRadius: '50%' }}
                />
              )}
              <p>Name: {user?.name}</p>
              <p>Email: {user.email}</p>
              <p>First Name: {user.given_name}</p>
              <p>Last Name: {user.family_name}</p>
            </>
          )}
        </User>
      </div>
    </SignedIn>
  )
}
~~~

---

## Using the Hook (Advanced/Programmatic Control Only)

The ~useAsgardeo~ hook should only be used when you need programmatic control:

~~~jsx
import { useAsgardeo } from '@asgardeo/react'

function CustomComponent() {
  const { isSignedIn, user, signIn, signOut, loading, error } = useAsgardeo()

  if (loading) {
    return <div>Loading...</div>
  }

  if (error) {
    return <div>Error: {error.message}</div>
  }

  return (
    <div>
      {isSignedIn ? (
        <>
          <p>Welcome, {user?.displayName}!</p>
          <button onClick={signOut}>Sign Out</button>
        </>
      ) : (
        <button onClick={signIn}>Sign In</button>
      )}
    </div>
  )
}
~~~

**Important:** The ~useAsgardeo~ hook must be used within a component that is a descendant of ~AsgardeoProvider~.

---

## Route Protection

### Option 1: Using SDK Control Components
~~~jsx
import { SignedIn, SignedOut } from '@asgardeo/react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/signin" element={<SignInPage />} />
        <Route 
          path="/dashboard" 
          element={
            <SignedIn fallback={<Navigate to="/signin" />}>
              <Dashboard />
            </SignedIn>
          } 
        />
      </Routes>
    </BrowserRouter>
  )
}
~~~

### Option 2: Using React Router Integration
~~~bash
npm install @asgardeo/react-router
~~~

~~~jsx
import { ProtectedRoute } from '@asgardeo/react-router'

<Route 
  path="/dashboard" 
  element={
    <ProtectedRoute redirectTo="/signin">
      <Dashboard />
    </ProtectedRoute>
  } 
/>
~~~

### Option 3: Custom Implementation
~~~jsx
import { useAsgardeo } from '@asgardeo/react'
import { Navigate } from 'react-router-dom'

function ProtectedRoute({ children }) {
  const { isSignedIn, loading } = useAsgardeo()

  if (loading) {
    return <div>Loading...</div>
  }

  if (!isSignedIn) {
    return <Navigate to="/signin" replace />
  }

  return children
}
~~~

---

## Accessing Protected APIs

### Using SDK Built-in HTTP Client (webWorker storage)
~~~jsx
import { useAsgardeo } from '@asgardeo/react'
import { useEffect, useState } from 'react'

function UserData() {
  const { http, isSignedIn } = useAsgardeo()
  const [data, setData] = useState(null)

  useEffect(() => {
    if (!isSignedIn) return

    (async () => {
      try {
        const response = await http.request({
          url: 'https://localhost:8090/scim2/Me',
          method: 'GET',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/scim+json'
          }
        })
        setData(response.data)
      } catch (error) {
        console.error('API Error:', error)
      }
    })()
  }, [http, isSignedIn])

  return <div>{data && <pre>{JSON.stringify(data, null, 2)}</pre>}</div>
}
~~~

**Note:** The ~http~ module automatically attaches the access token to requests.

### Using Custom HTTP Client (sessionStorage/localStorage)
~~~jsx
import { useAsgardeo } from '@asgardeo/react'

async function fetchUserData() {
  const { getAccessToken, isSignedIn } = useAsgardeo()
  
  if (!isSignedIn) return

  const token = await getAccessToken()
  
  const response = await fetch('https://localhost:8090/scim2/Me', {
    headers: {
      'Authorization': ~Bearer ${token}~,
      'Accept': 'application/json'
    }
  })
  
  return response.json()
}
~~~

---

## Complete Example

~~~jsx
// main.jsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { AsgardeoProvider } from '@asgardeo/react'
import App from './App.jsx'
import './index.css'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <AsgardeoProvider
      clientId="<client-id>"
      baseUrl="https://localhost:8090"
      platform="AsgardeoV2"
    >
      <App />
    </AsgardeoProvider>
  </StrictMode>
)
~~~

~~~jsx
// App.jsx
import { SignedIn, SignedOut, SignInButton, SignOutButton, Loading } from '@asgardeo/react'
import User from './components/User'

function App() {
  return (
    <div className="app">
      <header>
        <h1>ThunderID Auth Demo</h1>
        <Loading>
          <div>Loading...</div>
        </Loading>
      </header>

      <main>
        <SignedOut>
          <div className="welcome">
            <h2>Welcome!</h2>
            <p>Please sign in to continue</p>
            <SignInButton>Sign In</SignInButton>
          </div>
        </SignedOut>

        <SignedIn>
          <div className="dashboard">
            <User>
              {(user) => (
                <>
                  <h2>Welcome, {user?.displayName}!</h2>
                  <div className="user-info">
                    <p><strong>Email:</strong> {user?.email}</p>
                    <p><strong>Username:</strong> {user?.username}</p>
                  </div>
                </>
              )}
            </User>
            <SignOutButton>Sign Out</SignOutButton>
          </div>
        </SignedIn>
      </main>
    </div>
  )
}

export default App
~~~

---

## Method 2: ThunderID App Native Authentication with React (Vite)

This guide shows how to integrate ThunderID App Native authentication into a React
app using ~@asgardeo/react~, based on this sample project.

### Prerequisites

- A ThunderID application already created.
- Your ThunderID **Application ID (UUID)**.
- Node.js and npm installed.

### 1) Create a new Vite project (or use existing)

If starting fresh, create a new Vite React app:

~~~bash
npm create vite@latest my-app -- --template react
cd my-app
npm install
~~~

### 2) Install dependencies

~~~bash
npm install @asgardeo/react
~~~

If the dependency already exists in ~package.json~, you can skip the above steps.

### 3) Wrap your app with ~AsgardeoProvider~

Update ~src/main.jsx~ to configure the authentication provider.

~~~jsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App.jsx';
import { AsgardeoProvider } from '@asgardeo/react';
import './index.css';

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <AsgardeoProvider
      baseUrl="https://localhost:8090"
      applicationId="<THUNDERID_APP_ID>"
      platform="AsgardeoV2"
    >
      <App />
    </AsgardeoProvider>
  </StrictMode>
);
~~~

Replace ~<THUNDERID_APP_ID>~ with your real application UUID from ThunderID.

### 4) Add Sign-In UI

You can quickly enable login by using the built-in ~SignIn~ component.

~~~jsx
import { SignIn } from '@asgardeo/react';

function App() {
  return (
    <div>
      <h1>My React App</h1>
      <SignIn />
    </div>
  );
}

export default App;
~~~

This component starts the App Native authentication flow when the user signs in.

### 5) Run the app

~~~bash
npm run dev
~~~

Open the app in your browser and click the sign-in action.

### Optional: Move auth config to environment variables

For cleaner configuration, store values in ~.env~:

~~~bash
VITE_THUNDERID_BASE_URL=https://localhost:8090
VITE_THUNDERID_APP_ID=<THUNDERID_APP_ID>
~~~

Then use them in ~src/main.jsx~:

~~~jsx
<AsgardeoProvider
  baseUrl={import.meta.env.VITE_THUNDERID_BASE_URL}
  applicationId={import.meta.env.VITE_THUNDERID_APP_ID}
  platform="AsgardeoV2"
>
  <App />
</AsgardeoProvider>
~~~

### Do's and Don'ts

#### ✅ Do

- Do replace ~<THUNDERID_APP_ID>~ with the exact UUID from your ThunderID app registration.
- Do keep auth settings (~baseUrl~, app ID) environment-specific (dev/stage/prod).
- Do keep ~AsgardeoProvider~ high in the component tree (usually in ~src/main.jsx~).
- Do validate the ~baseUrl~ and cert setup when running locally over HTTPS.
- Do use the latest compatible ~@asgardeo/react~ version for your app.

#### ❌ Don't

- Don't hard-code production credentials across multiple files.
- Don't commit secret or environment-specific values into source control.
- Don't initialize multiple auth providers in parallel for the same React tree.
- Don't assume localhost settings will work unchanged in production.
- Don't bypass proper sign-in flow with manual token handling unless absolutely necessary.

### Quick troubleshooting

- **Sign-in not starting:** Verify ~applicationId~ and ~baseUrl~.
- **Invalid app/client errors:** Re-check the ThunderID app registration and copied UUID.
- **Local HTTPS issues:** Confirm your local endpoint and certificate trust setup.

---

For complete implementation examples of Method 2 authentication, refer to:

- **React Vanilla Sample** - Demonstrates both:
  - App-native authentication using Flow Orchestration API
  - Standard OAuth 2.0 / OIDC with custom UI

Both samples show how to build custom authentication UIs while leveraging ThunderID's authentication capabilities.

---

## Best Practices

### ✅ DO:
- Use declarative components (~<SignedIn>~, ~<SignedOut>~, ~<Loading>~) for UI state
- Use pre-built action components (~<SignInButton>~, ~<SignOutButton>~)
- Keep the provider configuration minimal and explicit
- Use the ~useAsgardeo~ hook only when programmatic control is needed
- Handle loading and error states properly

### ❌ DON'T:
- Don't create custom authentication logic unless absolutely necessary
- Don't manipulate tokens manually
- Don't store tokens in localStorage unless using the SDK's storage mechanism
- Don't add unnecessary configuration to the provider
- Don't use the hook outside of components wrapped by ~AsgardeoProvider~

---

## Common Patterns

### Pattern 1: Simple Auth-Gated App
~~~jsx
function App() {
  return (
    <>
      <SignedOut>
        <LandingPage />
      </SignedOut>
      <SignedIn>
        <Dashboard />
      </SignedIn>
    </>
  )
}
~~~

### Pattern 2: Navigation Bar with Conditional Auth
~~~jsx
function NavBar() {
  return (
    <nav>
      <Logo />
      <SignedOut>
        <SignInButton>Login</SignInButton>
      </SignedOut>
      <SignedIn>
        <UserMenu />
        <SignOutButton>Sign Out</SignOutButton>
      </SignedIn>
    </nav>
  )
}
~~~

### Pattern 3: Loading State Handling
~~~jsx
function App() {
  return (
    <>
      <Loading fallback={null}>
        <div className="spinner">Authenticating...</div>
      </Loading>
      
      <SignedIn>
        <Dashboard />
      </SignedIn>
    </>
  )
}
~~~

---

## Troubleshooting

### Issue: Hook Error "useAsgardeo must be used within AsgardeoProvider"
**Solution:** Ensure the component using ~useAsgardeo~ is a child of ~<AsgardeoProvider>~

### Issue: Infinite redirect loop
**Solution:** Check that ~baseUrl~ and ~clientId~ are correct. Verify token validation settings.

### Issue: User object is null after sign in
**Solution:** Ensure authentication has completed. Check for any errors in the console.

### Issue: CORS errors
**Solution:** Configure CORS settings in ThunderID/Asgardeo to allow your app's origin.

---

## References
- [Asgardeo React SDK Docs](https://wso2.com/asgardeo/docs/sdks/react/overview/)
- [AsgardeoProvider Configuration](https://wso2.com/asgardeo/docs/sdks/react/contexts/asgardeo-provider/)
- [SDK Components](https://wso2.com/asgardeo/docs/sdks/react/components/action-components/sign-in-button/)
- [useAsgardeo Hook](https://wso2.com/asgardeo/docs/sdks/react/hooks/use-asgardeo/)
- [Protecting Routes](https://wso2.com/asgardeo/docs/sdks/react/guides/protecting-routes/)
- [Accessing Protected APIs](https://wso2.com/asgardeo/docs/sdks/react/guides/accessing-protected-apis/)
`
	instructions := strings.ReplaceAll(rawInstructions, "~", "`")

	snippets := `
import { AsgardeoProvider } from '@asgardeo/react';

// Main Provider Setup
<AsgardeoProvider
  clientId="<client-id>"
  baseUrl="https://localhost:8090"
  platform="AsgardeoV2"
>
  <App />
</AsgardeoProvider>
`

	// Template the URL if provided
	if input.ServerURL != "" {
		instructions = strings.ReplaceAll(instructions, "https://localhost:8090", input.ServerURL)
		snippets = strings.ReplaceAll(snippets, "https://localhost:8090", input.ServerURL)
	}

	return nil, integrateReactSDKOutput{
		Instructions: instructions,
		CodeSnippets: snippets,
	}, nil
}
