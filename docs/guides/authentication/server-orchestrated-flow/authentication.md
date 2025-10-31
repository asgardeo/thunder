# Server Orchestrated Login Flows (App Native Authentication)

Thunder supports app native authentication flows where the server orchestrates the entire authentication process. This allows users to execute login flows via REST APIs, making it suitable for native applications, mobile apps, and single-page applications (SPAs).

> Note: Refer [Customizing Authentication Flows](./customize-auth-flow.md) for more details on customizing authentication flows.

## Login with Username and Password

Follow the steps below to configure and execute a login flow using username and password authentication.

1. **Create a User**

    Create a user in the system if you haven't already. First, create an organization unit, then create a user with the required attributes.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/organization-units \
    -d '{
        "name": "Asgard",
        "description": "Realm of the gods",
        "handle": "asgard"
    }'
    ```

    Note the `id` from the response above, then use it in the user creation:

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/users \
    -d '{
        "organizationUnit": "{ou-id-from-above}",
        "type": "superhuman",
        "attributes": {
            "username": "thor",
            "password": "<password>",
            "email": "thor@thunder.sky",
            "given_name": "Thor",
            "family_name": "Odinson",
            "age": 1534,
            "abilities": [
                "strength",
                "speed",
                "healing"
            ],
            "address": {
                "city": "Asgard",
                "zip": "00100"
            }
        }
    }'
    ```

2. **Configure an Application with Username/Password Login**

    Create an application to use the username/password login template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_graph_id": "auth_flow_config_basic"
    }'
    ```

3. **Start the Login Flow**

    Start login flow for the application with the following cURL command:

    ```bash
    curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "applicationId": "<application_id>",
        "flowType": "AUTHENTICATION"
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

4. **Complete the Login Flow**

    Make the second cURL request to complete the login flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response. Also, replace the `username` and `password` with the credentials of the user you created in the first step.

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

    If the login is successful, you will receive a response with the auth assertion.

## Login with SMS OTP

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

3. **Update the Authentication Flow Graph**

    Update the authentication flow graph to use the configured message sender. To do so, open the `auth_flow_config_sms.json` file in the `repository/resources/graph/` directory and update the `senderId` with the unique identifier of the message sender you configured in the previous step. Make sure to restart the server after making this change.

4. **Configure an Application with SMS OTP Login**

    Create an application to use the SMS OTP login template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_graph_id": "auth_flow_config_sms"
    }'
    ```

5. **Create a User with Mobile Number**

    Create a user in the system with a mobile number attribute to receive SMS OTP. First, create an organization unit, then create a user with the required attributes.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/organization-units \
    -d '{
        "name": "Asgard",
        "description": "Realm of the gods",
        "handle": "asgard"
    }'
    ```

    Note the `id` from the response above, then use it in the user creation:

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/users \
    -d '{
        "organizationUnit": "{ou-id-from-above}",
        "type": "superhuman",
        "attributes": {
            "username": "thor",
            "password": "<password>",
            "email": "thor@thunder.sky",
            "given_name": "Thor",
            "family_name": "Odinson",
            "mobileNumber": "+94xxxxxxxxx"
        }
    }'
    ```

6. **Start the Login Flow**

    Start the login flow for the application:

    ```bash
    curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "applicationId": "<application_id>",
        "flowType": "AUTHENTICATION"
    }'
    ```

    You'll receive a response prompting for username input.

7. **Continue the Login Flow**

    Provide the username to continue the flow:

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "inputs": {
            "username": "thor"
        }
    }'
    ```

    An OTP will be sent to the user's mobile number.

8. **Complete the Authentication with OTP**

    Complete authentication by providing the OTP:

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "inputs": {
            "otp": "696546"
        }
    }'
    ```

    If the OTP is valid, you will receive a response with the auth assertion.

## Login with Google

Follow the steps below to configure and execute a login flow using Google OAuth authentication.

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

3. **Update the Authentication Flow Graph**

    Update the authentication flow graph to use the configured identity provider. To do so, open the `auth_flow_config_github.json` file in the `repository/resources/graph/` directory and update the `idpName` with the name of the identity provider you configured in the previous step. Make sure to restart the server after making this change.

4. **Configure an Application with Google Login**

    Create an application to use the Google login template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_graph_id": "auth_flow_config_google"
    }'
    ```

5. **Start the Login Flow**

    Start login flow for the application with the following cURL command:

    ```bash
    curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "applicationId": "<application_id>",
        "flowType": "AUTHENTICATION"
    }'
    ```

    You'll receive a response similar to the following:

    ```json
    {
        "flowId": "80d57e64-8082-4096-bb0e-22b2187f8265",
        "flowStatus": "INCOMPLETE",
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
    https://localhost:3000/?code=<code>&state=80d57e64-8082-4096-bb0e-22b2187f8265
    ```

7. **Complete the Login Flow**

    Copy the authorization code and make the second cURL request to complete the login flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response.

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "inputs": {
            "code": "<code>"
        }
    }'
    ```

    If the login is successful, you will receive a response with the auth assertion.

## Login with GitHub

Follow the steps below to configure and execute a login flow using GitHub OAuth authentication.

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

3. **Update the Authentication Flow Graph**

    Update the authentication flow graph to use the configured identity provider. To do so, open the `auth_flow_config_github.json` file in the `repository/resources/graph/` directory and update the `idpName` with the name of the identity provider you configured in the previous step. Make sure to restart the server after making this change.

4. **Configure an Application with GitHub Login**

    Create an application to use the Github login template. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "App Native Login 1",
        "description": "Sample application for App native login",
        "auth_flow_graph_id": "auth_flow_config_github"
    }'
    ```

5. **Start the Login Flow**

    Start login flow for the application with the following cURL command:

    ```bash
    curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "applicationId": "<application_id>",
        "flowType": "AUTHENTICATION"
    }'
    ```

    You'll receive a response similar to the following:

    ```json
    {
        "flowId": "80d57e64-8082-4096-bb0e-22b2187f8265",
        "flowStatus": "INCOMPLETE",
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
    https://localhost:3000/?code=<code>&state=80d57e64-8082-4096-bb0e-22b2187f8265
    ```

7. **Complete the Login Flow**

    Copy the authorization code and make the second cURL request to complete the login flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response.

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
    -d '{
        "flowId": "<flow_id>",
        "inputs": {
            "code": "<code>"
        }
    }'
    ```

    If the login is successful, you will receive a response with the auth assertion.
