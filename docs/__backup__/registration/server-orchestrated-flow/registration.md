# Server Orchestrated Registration Flows (Self Registration)

Thunder supports self-registration flows where the server orchestrates the entire registration process. This allows users to execute registration flows via REST APIs, making it suitable for native applications, mobile apps, and single-page applications (SPAs).

> [!TIP]
> To customize registration flows, see the [Flow Guides](../../flows/) for creating and managing flows using the Visual Flow Builder or the Flow Management API.

## Enabling Self Registration

Thunder allows you to control whether users can self-register for your application using the application management API. By default, self-registration is disabled when you create an application.

To enable self-registration, set the `is_registration_flow_enabled` property to `true` when creating or updating an application.

## Register with Username and Password

1. **Create or Update the Registration Flow**

    Use the Flow Management API or Visual Flow Builder to create a username/password registration flow. See the [Flow Examples](/docs/guides/flows/flow-examples.md#username-and-password-registration) for a complete example.

    Note the `id` of the created registration flow for the next step.

2. **Configure an Application with Username/Password Registration**

    Create an application to use the username/password registration template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -H 'Authorization: Bearer <token>' \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_id": "edc013d0-e893-4dc0-990c-3e1d203e005b",
        "registration_flow_id": "<registration_flow_id>",
        "is_registration_flow_enabled": true
    }'
    ```

3. **Start the Registration Flow**

    Start registration flow for the application with the following cURL command:

    ```bash
    curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "applicationId": "<application_id>",
        "flowType": "REGISTRATION"
    }'
    ```

    You'll receive a response similar to the following:

    ```json
    {
        "flowId": "db93a19e-c23f-4cfc-a45f-0e0bc157f6d5",
        "flowStatus": "PROMPT_ONLY",
        "type": "VIEW",
        "data": {
          "inputs": [
            {
                "identifier": "username",
                "type": "TEXT_INPUT",
                "required": true
            },
            {
                "identifier": "password",
                "type": "PASSWORD_INPUT",
                "required": true
            }
          ],
          "actions": [
            {
                "ref": "action_001",
                "nextNode": "basic_auth"
            }
          ]
        }
    }
    ```

    The `actions` array contains the available actions the user can select. Use the `ref` value from the selected action in subsequent requests.

4. **Continue the Registration Flow**

    Make the second cURL request to continue the registration flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response. Also, replace the `username` and `password` with the desired credentials for the new user.

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "action": "<action_ref>",
        "inputs": {
            "username": "thor",
            "password": "<password>"
        }
    }'
    ```

    If the registration is successful, you will receive a response prompting for additional user attributes.

    ```json
    {
        "flowId": "db93a19e-c23f-4cfc-a45f-0e0bc157f6d5",
        "flowStatus": "PROMPT_ONLY",
        "type": "VIEW",
        "data": {
        "inputs": [
            {
                "name": "email",
                "type": "string",
                "required": true
            },
            {
                "name": "firstName",
                "type": "string",
                "required": true
            },
            {
                "name": "lastName",
                "type": "string",
                "required": true
            }
        ]
        }
    }
    ```

5. **Complete the Registration Flow**

    Make the third cURL request to complete the registration flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response.

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "action": "<action_ref>",
        "inputs": {
            "email": "thor@thunder.sky",
            "firstName": "Thor",
            "lastName": "Odinson"
        }
    }'
    ```

    If the registration is successful, you will be automatically logged in and receive a response with the auth assertion.

## Register with SMS OTP

1. **Configure a Message Sender**

    Configure a message sender to send SMS messages. You can use services like Twilio, Vonage, or a custom service of your choice.

2. **Create a Message Sender**

    You can create a message sender using the following cURL command. Make sure to replace properties according to the configured message provider.

    ```bash
    curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/notification-senders/message \
    -d '{
      "name": "Custom SMS Sender",
      "description": "Sender for sending SMS messages",
      "provider": "custom",
      "properties": [
        {
          "name": "url",
          "value": "<custom_sms_provider_url>"
        },
        {
          "name": "http_method",
          "value": "POST"
        },
        {
          "name": "content_type",
          "value": "JSON"
        }
      ]
    }'
    ```

    > Note: Refer [Configuring Message Senders](../../notification-sender/configure-message-senders.md) for more details on configuring message senders.

3. **Create or Update the Registration Flow**

    Use the Flow Management API or Visual Flow Builder to create an SMS OTP registration flow with your `senderId`. See the [Flow Examples](/docs/guides/flows/flow-examples.md#sms-otp-registration) for a complete example.

    Note the `id` of the created registration flow for the next step.

4. **Configure an Application with SMS OTP Registration**

    Create an application to use the SMS OTP registration template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -H 'Authorization: Bearer <token>' \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_id": "edc013d0-e893-4dc0-990c-3e1d203e005b",
        "registration_flow_id": "<registration_flow_id>",
        "is_registration_flow_enabled": true
    }'
    ```

5. **Start the Registration Flow**

    Start registration flow for the application with the following cURL command:

    ```bash
    curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "applicationId": "<application_id>",
        "flowType": "REGISTRATION"
    }'
    ```

    You'll receive a response similar to the following:

    ```json
    {
        "flowId": "db93a19e-c23f-4cfc-a45f-0e0bc157f6d5",
        "flowStatus": "PROMPT_ONLY",
        "type": "VIEW",
        "data": {
        "inputs": [
            {
                "name": "mobileNumber",
                "type": "string",
                "required": true
            }
        ]
        }
    }
    ```

6. **Continue the Registration Flow**

    Make the second cURL request to continue the registration flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response. Also, replace the `mobileNumber` with the desired mobile number for the new user.

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "action": "<action_ref>",
        "inputs": {
            "mobileNumber": "+94xxxxxxxxx"
        }
    }'
    ```

    An OTP will be sent to the provided mobile number.

7. **Verify the Mobile Number with OTP**

    Continue the registration by providing the OTP received on the mobile number:

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "action": "<action_ref>",
        "inputs": {
            "otp": "696546"
        }
    }'
    ```

    If the verification is successful, you will receive a response prompting for additional user attributes.

    ```json
    {
        "flowId": "db93a19e-c23f-4cfc-a45f-0e0bc157f6d5",
        "flowStatus": "PROMPT_ONLY",
        "type": "VIEW",
        "data": {
        "inputs": [
            {
                "name": "email",
                "type": "string",
                "required": true
            },
            {
                "name": "firstName",
                "type": "string",
                "required": true
            },
            {
                "name": "lastName",
                "type": "string",
                "required": true
            }
        ]
        }
    }
    ```

8. **Complete the Registration Flow**

    Make the third cURL request to complete the registration flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response.

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "action": "<action_ref>",
        "inputs": {
            "email": "thor@thunder.sky",
            "firstName": "Thor",
            "lastName": "Odinson"
        }
    }'
    ```

    If the registration is successful, you will receive a response with the auth assertion.

## Google Sign Up

1. **Create a Google OAuth Application**

    Create an OAuth application in your Google account following the instructions given in the [Google documentation](https://developers.google.com/identity/protocols/oauth2/web-server#creatingcred).
    - Configure the Authorized origin and Redirect URI as per your application.
    - Copy the **Client ID** and **Client Secret**.

2. **Configure the Google Identity Provider**

    Create a Google IDP by invoking the IDP management API with the following cURL command. Make sure to replace `<client_id>`, `<client_secret>`, and `<app_callback_url>` with the values you copied from your Google OAuth application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers \
    -H 'Authorization: Bearer <token>' \
    -d '{
        "name": "Google",
        "description": "Login with Google",
        "type": "GOOGLE",
        "properties": [
            {
                "name": "client_id",
                "value": "<client_id>",
                "is_secret": false
            },
            {
                "name": "client_secret",
                "value": "<client_secret>",
                "is_secret": true
            },
            {
                "name": "redirect_uri",
                "value": "<app_callback_url>",
                "is_secret": false
            },
            {
                "name": "scopes",
                "value": "openid,email,profile",
                "is_secret": false
            }
        ]
    }'
    ```

    > Note: Refer [Configuring Identity Providers](./../../identity-provider/configure-identity-providers.md) for more details on configuring identity providers.

3. **Create or Update the Registration Flow**

    Use the Flow Management API or Visual Flow Builder to create a Google registration flow with your `idpId`. See the [Flow Examples](./../../flows/flow-examples.md#google-oidc-registration) for a complete example.
    Note the `id` of the created registration flow for the next step.

4. **Configure an Application with Google Sign Up**

    Create an application to use the Google sign up template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_id": "edc013d0-e893-4dc0-990c-3e1d203e005b",
        "registration_flow_id": "<registration_flow_id>",
        "is_registration_flow_enabled": true
    }'
    ```

5. **Start the Registration Flow**

    Start registration flow for the application with the following cURL command:

    ```bash
    curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "applicationId": "<application_id>",
        "flowType": "REGISTRATION"
    }'
    ```

    You'll receive a response similar to the following:

    ```json
    {
        "flowId": "db93a19e-c23f-4cfc-a45f-0e0bc157f6d5",
        "flowStatus": "PROMPT_ONLY",
        "type": "REDIRECTION",
        "data": {
            "redirectURL": "<google_auth_redirect_url>",
            "inputs": [
              {
                  "name": "code",
                  "type": "string",
                  "required": true
              },
              {
                  "name": "nonce",
                  "type": "string",
                  "required": false
              }
            ],
            "additionalData": {
              "idpName": "Google"
            }
        }
    }
    ```

6. **Login with Google Account**

    Open the `redirect_url` in your browser. You will be redirected to the Google login page. Enter your Google credentials and authorize the application.

    After successful authentication, you will be redirected to the redirect URI with the authorization code, state and other parameters.

    ```bash
    https://localhost:3000/?code=<code>&state=db93a19e-c23f-4cfc-a45f-0e0bc157f6d5
    ```

7. **Complete the Registration Flow**

    Copy the authorization code and make the second cURL request to complete the registration flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response.

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "inputs": {
            "code": "<code>"
        }
    }'
    ```

    If the registration is successful, you will receive a response with the auth assertion.

## GitHub Sign Up

1. **Create a GitHub OAuth Application**

    Create an OAuth application in your Github account following the instructions given in the [Github documentation](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app).
      - Configure home page and callback URLs as per your application.
      - Copy the **Client ID** and **Client Secret**.

2. **Configure the GitHub Identity Provider**

    Create a GitHub IDP by invoking the IDP management API with the following cURL command. Make sure to replace `<client_id>`, `<client_secret>`, and `<app_callback_url>` with the values you copied from your GitHub OAuth application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers \
    -H 'Authorization: Bearer <token>' \
    -d '{
        "name": "Github",
        "description": "Login with Github",
        "type": "GITHUB",
        "properties": [
            {
                "name": "client_id",
                "value": "<client_id>",
                "is_secret": false
            },
            {
                "name": "client_secret",
                "value": "<client_secret>",
                "is_secret": true
            },
            {
                "name": "redirect_uri",
                "value": "<app_callback_url>",
                "is_secret": false
            },
            {
                "name": "scopes",
                "value": "user:email,read:user",
                "is_secret": false
            }
        ]
    }'
    ```

    > Note: Refer [Configuring Identity Providers](./../../identity-provider/configure-identity-providers.md) for more details on configuring identity providers.

3. **Create or Update the Registration Flow**

    Use the Flow Management API or Visual Flow Builder to create a GitHub registration flow with your `idpId`. See the [Flow Examples](../../flows/flow-examples.md#github-oauth-registration) for a complete example.

    Note the `id` of the created registration flow for the next step.

4. **Configure an Application with GitHub Sign Up**

    Create an application to use the GitHub sign up template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -H 'Authorization: Bearer <token>' \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_id": "edc013d0-e893-4dc0-990c-3e1d203e005b",
        "registration_flow_id": "<registration_flow_id>",
        "is_registration_flow_enabled": true
    }'
    ```

5. **Start the Registration Flow**

    Start registration flow for the application with the following cURL command:

    ```bash
    curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "applicationId": "<application_id>",
        "flowType": "REGISTRATION"
    }'
    ```

    You'll receive a response similar to the following:

    ```json
    {
        "flowId": "db93a19e-c23f-4cfc-a45f-0e0bc157f6d5",
        "flowStatus": "PROMPT_ONLY",
        "type": "REDIRECTION",
        "data": {
            "redirectURL": "<github_auth_redirect_url>",
            "inputs": [
              {
                  "name": "code",
                  "type": "string",
                  "required": true
              }
            ],
            "additionalData": {
              "idpName": "Github"
            }
        }
    }
    ```

6. **Login with GitHub Account**

    Open the `redirect_url` in your browser. You will be redirected to the GitHub login page. Enter your GitHub credentials and authorize the application.

    After successful authentication, you will be redirected to the redirect URI with the authorization code and state.

    ```bash
    https://localhost:3000/?code=<code>&state=db93a19e-c23f-4cfc-a45f-0e0bc157f6d5
    ```

7. **Complete the Registration Flow**

    Copy the authorization code and make the second cURL request to complete the registration flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response.

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "inputs": {
            "code": "<code>"
        }
    }'
    ```

    If the registration is successful, you will receive a response with the auth assertion.
