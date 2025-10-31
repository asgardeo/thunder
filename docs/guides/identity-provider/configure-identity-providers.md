# Configuring Identity Providers

This guide provides instructions on how to configure and manage identity providers (IDPs) in WSO2 Thunder. You can manage IDPs through the REST API, allowing you to create, update, retrieve, and delete identity providers.

---

## Managing Identity Providers

You can manage identity providers using the following REST API endpoints.

### List Identity Providers

Retrieve a list of all configured identity providers.

```bash
curl -kL -H 'Accept: application/json' https://localhost:8090/identity-providers
```

### Create an Identity Provider

Create a new identity provider. See provider-specific examples below for the request body.

```bash
curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers \
-d '{ ... }'
```

Refer to [Identity Provider Configuration](#identity-provider-configuration) for details on the supported providers and their required properties.

### Get an Identity Provider by ID

Retrieve details of a specific identity provider by its unique identifier.

```bash
curl -kL -H 'Accept: application/json' https://localhost:8090/identity-providers/<idp_id>
```

### Update an Identity Provider

Update an existing identity provider by its unique identifier.

```bash
curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers/<idp_id> \
-d '{ ... }'
```

### Delete an Identity Provider

Delete an identity provider by its unique identifier.

```bash
curl -kL -X DELETE https://localhost:8090/identity-providers/<idp_id>
```

---

## Identity Provider Configuration

To configure an identity provider, you need to specify the required properties for that provider. The following are common types of identity providers supported by Thunder.

### Google

**Required Properties:**

| Property        | Description                        | Example Value                |
|-----------------|------------------------------------|------------------------------|
| client_id       | Google OAuth Client ID             | your_client_id               |
| client_secret   | Google OAuth Client Secret         | your_client_secret           |
| redirect_uri    | Redirect URI for your application  | https://localhost:3000       |
| scopes          | OAuth scopes                       | openid,email,profile         |

> Note: You can also configure any other additional properties supported by the Google IDP that can be passed in the authorization request.

**Example cURL:**

```bash
curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers \
-d '{
  "name": "Google",
  "description": "Login with Google",
  "type": "GOOGLE",
  "properties": [
    {
      "name": "client_id",
      "value": "your_client_id",
      "is_secret": false
    },
    {
      "name": "client_secret",
      "value": "your_client_secret",
      "is_secret": true
    },
    {
      "name": "redirect_uri",
      "value": "https://localhost:3000",
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

### GitHub

**Required Properties:**

| Property        | Description                        | Example Value                |
|-----------------|------------------------------------|------------------------------|
| client_id       | GitHub OAuth Client ID             | your_client_id               |
| client_secret   | GitHub OAuth Client Secret         | your_client_secret           |
| redirect_uri    | Redirect URI for your application  | https://localhost:3000       |
| scopes          | OAuth scopes                       | user:email,read:user         |

> Note: You can also configure any other additional properties supported by the GitHub IDP that can be passed in the authorization request.

**Example cURL:**

```bash
curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers \
-d '{
  "name": "Github",
  "description": "Login with Github",
  "type": "GITHUB",
  "properties": [
    {
      "name": "client_id",
      "value": "your_client_id",
      "is_secret": false
    },
    {
      "name": "client_secret",
      "value": "your_client_secret",
      "is_secret": true
    },
    {
      "name": "redirect_uri",
      "value": "https://localhost:3000",
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

---

For more details on the API, refer to the [OpenAPI specification](../apis/idp.yaml).
