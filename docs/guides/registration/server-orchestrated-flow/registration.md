# Server Orchestrated Registration Flows (Self Registration)

Thunder supports self-registration flows where the server orchestrates the entire registration process. This allows users to execute registration flows via REST APIs, making it suitable for native applications, mobile apps, and single-page applications (SPAs).

> Note: Refer [Customizing Registration Flows](./customize-registration-flow.md) for more details on customizing registration flows.

## Enabling Self Registration

Thunder allows you to control whether users can self-register for your application using the application management API. By default, self-registration is disabled when you create an application.

To enable self-registration, set the `is_registration_flow_enabled` property to `true` when creating or updating an application.

## Register with Username and Password

1. **Configure an Application with Username/Password Registration**

    Create an application to use the username/password registration template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_graph_id": "auth_flow_config_basic",
        "registration_flow_graph_id": "registration_flow_config_basic",
        "is_registration_flow_enabled": true
    }'
    ```

2. **Start the Registration Flow**

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
                "name": "username",
                "type": "string",
                "required": true
            },
            {
                "name": "password",
                "type": "string",
                "required": true
            }
        ]
        }
    }
    ```

3. **Continue the Registration Flow**

    Make the second cURL request to continue the registration flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response. Also, replace the `username` and `password` with the desired credentials for the new user.

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
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

4. **Complete the Registration Flow**

    Make the third cURL request to complete the registration flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response.

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
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

    > Note: Refer [Configuring Message Senders](/../../notification-sender/configure-message-senders.md) for more details on configuring message senders.

3. **Update the Registration Flow Graph**

    Update the registration flow graph to use the configured message sender. To do so, open the `registration_flow_config_sms.json` file in the `repository/resources/graph/` directory and update the `senderId` with the unique identifier of the message sender you configured in the previous step.

    If the file doesn't exist, that means the registration flow graph is automatically constructed from the equivalent authentication flow graph. In that case, you can update the `senderId` property in the `auth_flow_config_sms.json` file or create a new `registration_flow_config_sms.json` file and define the registration flow.

    > Note: Refer [Customizing Registration Flows](./customize-registration-flow.md) for more details on automatic registration flow creation.

    Make sure to restart the server after making this change.

4. **Configure an Application with SMS OTP Registration**

    Create an application to use the SMS OTP registration template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_graph_id": "auth_flow_config_sms",
        "registration_flow_graph_id": "registration_flow_config_sms",
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

    > Note: Refer [Configuring Identity Providers](/../../identity-provider/configure-identity-providers.md) for more details on configuring identity providers.

3. **Update the Registration Flow Graph**

    Update the registration flow graph to use the configured identity provider. To do so, open the `registration_flow_config_google.json` file in the `repository/resources/graph/` directory and update the `idpName` with the name of the identity provider you configured in the previous step.

    If the file doesn't exist, that means the registration flow graph is automatically constructed from the equivalent authentication flow graph. In that case, you can update the `idpName` property in the `auth_flow_config_google.json` file or create a new `registration_flow_config_google.json` file and define the registration flow.

    > Note: Refer [Customizing Registration Flows](./customize-registration-flow.md) for more details on automatic registration flow creation.
    
    Make sure to restart the server after making this change.

4. **Configure an Application with Google Sign Up**

    Create an application to use the Google sign up template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_graph_id": "auth_flow_config_google",
        "registration_flow_graph_id": "registration_flow_config_google",
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

    > Note: Refer [Configuring Identity Providers](/../../identity-provider/configure-identity-providers.md) for more details on configuring identity providers.

3. **Update the Registration Flow Graph**

    Update the registration flow graph to use the configured identity provider. To do so, open the `registration_flow_config_github.json` file in the `repository/resources/graph/` directory and update the `idpName` with the name of the identity provider you configured in the previous step.

    If the file doesn't exist, that means the registration flow graph is automatically constructed from the equivalent authentication flow graph. In that case, you can update the `idpName` property in the `auth_flow_config_github.json` file or create a new `registration_flow_config_github.json` file and define the registration flow.

    > Note: Refer [Customizing Registration Flows](./customize-registration-flow.md) for more details on automatic registration flow creation.
    
    Make sure to restart the server after making this change.

4. **Configure an Application with GitHub Sign Up**

    Create an application to use the GitHub sign up template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_graph_id": "auth_flow_config_github",
        "registration_flow_graph_id": "registration_flow_config_github",
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
