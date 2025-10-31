# OAuth Standards-Based Authentication

Thunder supports OAuth 2.0 and OpenID Connect (OIDC) standards for authentication and authorization. The following flows are supported:

1. Client Credentials Flow
2. Authorization Code Flow
3. Refresh Token Flow

## Client Credentials Flow

The Client Credentials flow is used to obtain an access token for machine-to-machine communication. This flow does not require user interaction and is typically used for server-to-server communication.

To try out the Client Credentials flow, follow these steps:

1. **Create a Client Application**

    Create a client application in the system to use for the Client Credentials flow. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "Test Sample App",
        "description": "Initial testing App",
        "auth_flow_graph_id": "auth_flow_config_basic",
        "inbound_auth_config": [
            {
                "type": "oauth2",
                "config": {
                    "client_id": "<client_id>",
                    "client_secret": "<client_secret>",
                    "redirect_uris": [
                        "https://localhost:3000"
                    ],
                    "grant_types": [
                        "client_credentials"
                    ],
                    "token_endpoint_auth_method": "client_secret_basic",
                    "pkce_required": false,
                    "public_client": false
                }
            }
        ]
    }'
    ```

2. **Obtain an Access Token**

    Use the following cURL command to obtain an access token using the Client Credentials flow. Make sure to replace the `<client_id>` and `<client_secret>` with the values you used when creating the client application.

    ```bash
    curl -k -X POST https://localhost:8090/oauth2/token \
      -d 'grant_type=client_credentials' \
      -u '<client_id>:<client_secret>'
    ```

    If the request is successful, you will receive a response containing the access token:

    ```json
    {
      "access_token": "<access_token>",
      "token_type": "Bearer",
      "expires_in": 3600
    }
    ```

## Authorization Code Flow

The Authorization Code flow is used to obtain an access token after the user authenticates. This flow is typically used for web applications where a user redirection is required to complete the authentication process.

1. **Configure a Gate Client**

    Authorization code flow requires you to setup a gate client to handle the login and error redirection. You can implement your own client following the [sample gate client implementation](./frontend/apps/gate/).
  
    Add the following configurations to the `deployment.yaml` file to configure the gate client.

    ```yaml
    gate_client:
      hostname: "localhost"
      port: 9090
      scheme: "https"
      login_path: "/login"
      error_path: "/error"
    ```

    Make sure to restart the server after making this change.

2. **Create a Client Application**

    Create a client application in the system to use for the Authorization Code flow. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "Test Sample App",
        "description": "Initial testing App",
        "auth_flow_graph_id": "auth_flow_config_basic",
        "inbound_auth_config": [
            {
                "type": "oauth2",
                "config": {
                    "client_id": "<client_id>",
                    "client_secret": "<client_secret>",
                    "redirect_uris": [
                        "https://localhost:3000"
                    ],
                    "grant_types": [
                        "authorization_code"
                    ],
                    "response_types": [
                        "code"
                    ],
                    "token_endpoint_auth_method": "client_secret_basic",
                    "pkce_required": false,
                    "public_client": false,
                    "token": {
                        "issuer": "thunder",
                        "access_token": {
                            "validity_period": 3600,
                            "user_attributes": [
                                "given_name",
                                "family_name",
                                "email",
                                "groups"
                            ]
                        },
                        "id_token": {
                            "validity_period": 3600,
                            "user_attributes": [
                                "given_name",
                                "family_name",
                                "email",
                                "groups"
                            ],
                            "scope_claims": {
                                "email": [
                                    "email",
                                    "email_verified"
                                ],
                                "group": [
                                    "groups"
                                ],
                                "phone": [
                                    "phone_number",
                                    "phone_number_verified"
                                ],
                                "profile": [
                                    "name",
                                    "given_name",
                                    "family_name",
                                    "picture"
                                ]
                            }
                        }
                    }
                }
            }
        ]
    }'
    ```

3. **Create a User**

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

4. **Start the Authorization Code Flow**

    - Open the following URL in your browser to start the authorization code flow. Make sure to replace `<client_id>` with the client ID of the application you created in the previous step.

        ```bash
        https://localhost:8090/oauth2/authorize?response_type=code&client_id=<client_id>&redirect_uri=https://localhost:3000&scope=openid&state=state_1
        ```

    - You will be redirected to the login page of the gate client. If you have configured a custom gate client, you will see the login page of that client.

    - Enter the credentials of the user you created in the first step.

    - After successful authentication, you will be redirected to the redirect URI with the authorization code and state.

        ```bash
        https://localhost:3000/?code=<code>&state=state_1
        ```

    - Copy the authorization code and exchange it for an access token using the following cURL command:

        ```bash
        curl -k -X POST 'https://localhost:8090/oauth2/token' \
        -u '<client_id>:<client_secret>' \
        -d 'grant_type=authorization_code' \
        -d 'redirect_uri=https://localhost:3000' \
        -d 'code=<code>'
        ```

    - If the request is successful, you will receive a response containing the access token:

        ```json
        {
          "access_token": "<access_token>",
          "token_type": "Bearer",
          "expires_in": 3600,
        }
        ```

## Refresh Token Flow

The Refresh Token flow is used to obtain a new access token using a refresh token. This flow is typically used when the access token has expired and the user does not need to re-authenticate.

Refresh token flow can only be try out with the Authorization Code flow. To try out the Refresh Token flow, follow these steps:

1. **Create a Client Application**

    Create a client application in the system to use for the Refresh Token flow. You can use the following cURL command to create a new application.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -d '{
        "name": "Test Sample App",
        "description": "Initial testing App",
        "auth_flow_graph_id": "auth_flow_config_basic",
        "inbound_auth_config": [
            {
                "type": "oauth2",
                "config": {
                    "client_id": "<client_id>",
                    "client_secret": "<client_secret>",
                    "redirect_uris": [
                        "https://localhost:3000"
                    ],
                    "grant_types": [
                        "authorization_code", "refresh_token"
                    ],
                    "response_types": [
                        "code"
                    ],
                    "token_endpoint_auth_method": "client_secret_basic",
                    "pkce_required": false,
                    "public_client": false
                }
            }
        ]
    }'
    ```

2. **Start the Authorization Code Flow**

    Follow the steps in the [Authorization Code Flow](#authorization-code-flow) section to obtain an access token and a refresh token.

    After successful authentication, you will receive a response containing the access token and the refresh token:

    ```json
    {
      "access_token": "<access_token>",
      "token_type": "Bearer",
      "expires_in": 3600,
      "refresh_token": "<refresh_token>"
    }
    ```

3. **Obtain a New Access Token Using the Refresh Token**

    Use the following cURL command to obtain a new access token using the refresh token. Make sure to replace `<client_id>`, `<client_secret>`, and `<refresh_token>` with the values you used when creating the client application and the refresh token you received in the previous step.

    ```bash
    curl -k -X POST 'https://localhost:8090/oauth2/token' \
    -u '<client_id>:<client_secret>' \
    -d 'grant_type=refresh_token' \
    -d 'refresh_token=<refresh_token>'
    ```

    If the request is successful, you will receive a response containing the new access token:

    ```json
    {
      "access_token": "<new_access_token>",
      "token_type": "Bearer",
      "expires_in": 3600,
      "refresh_token": "<new_refresh_token>"
    }
    ```

## Additional Endpoints

Thunder provides the following additional endpoints to support OAuth 2.0 and OpenID Connect functionalities.

### JWKS Endpoint

The JSON Web Key Set (JWKS) endpoint provides the public keys used to verify the signatures of JSON Web Tokens (JWTs) issued by the Thunder server. Clients can use these keys to validate the authenticity of the tokens they receive.

To retrieve the JWKS, you can use the following cURL command:

```bash
curl -kL https://localhost:8090/oauth2/jwks
```
