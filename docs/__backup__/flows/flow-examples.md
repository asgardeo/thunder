# Flow Examples

This guide provides example flow configurations for common authentication and registration patterns.

---

## Authentication Flow Examples

### Username and Password Authentication

Attribute-based login using username and password.

```json
{
  "name": "Basic Authentication Flow",
  "handle": "basic-auth-flow",
  "flowType": "AUTHENTICATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "prompt_credentials"
    },
    {
      "id": "prompt_credentials",
      "type": "PROMPT",
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
    },
    {
      "id": "basic_auth",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "BasicAuthExecutor"
      },
      "onSuccess": "authz_check"
    },
    {
      "id": "authz_check",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthorizationExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

> [!TIP]
> Change `username` to `email` or `mobileNumber` to authenticate with different attributes.

> [!TIP]
> You can add multiple identifying inputs for complex authentication scenarios. For example, use `employeeId`, `department`, and `password` for enterprise authentication.

---

### SMS OTP Authentication

Mobile number verification using SMS OTP.

```json
{
  "name": "SMS OTP Authentication",
  "handle": "sms-otp-auth",
  "flowType": "AUTHENTICATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "prompt_mobile"
    },
    {
      "id": "prompt_mobile",
      "type": "PROMPT",
      "inputs": [
        {
          "identifier": "mobileNumber",
          "type": "PHONE_INPUT",
          "required": true
        }
      ],
      "actions": [
        {
          "ref": "action_001",
          "nextNode": "send_otp"
        }
      ]
    },
    {
      "id": "send_otp",
      "type": "TASK_EXECUTION",
      "properties": {
        "senderId": "<sender-id>"
      },
      "executor": {
        "name": "SMSOTPAuthExecutor",
        "mode": "send"
      },
      "onSuccess": "prompt_otp"
    },
    {
      "id": "prompt_otp",
      "type": "PROMPT",
      "inputs": [
        {
          "identifier": "otp",
          "type": "OTP_INPUT",
          "required": true
        }
      ],
      "actions": [
        {
          "ref": "action_001",
          "nextNode": "verify_otp"
        }
      ]
    },
    {
      "id": "verify_otp",
      "type": "TASK_EXECUTION",
      "properties": {
        "senderId": "<sender-id>"
      },
      "executor": {
        "name": "SMSOTPAuthExecutor",
        "mode": "verify"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

---

### Social Login with Google

Social login using Google OIDC provider.

```json
{
  "name": "Google Authentication",
  "handle": "google-auth-flow",
  "flowType": "AUTHENTICATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "google_auth"
    },
    {
      "id": "google_auth",
      "type": "TASK_EXECUTION",
      "properties": {
        "idpId": "<google-idp-id>"
      },
      "inputs": [
        {
          "identifier": "code",
          "type": "TEXT_INPUT",
          "required": true
        }
      ],
      "executor": {
        "name": "GoogleOIDCAuthExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

---

### Social Login with GitHub

Social login using GitHub OAuth provider.

```json
{
  "name": "GitHub Authentication",
  "handle": "github-auth-flow",
  "flowType": "AUTHENTICATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "github_auth"
    },
    {
      "id": "github_auth",
      "type": "TASK_EXECUTION",
      "properties": {
        "idpId": "<github-idp-id>"
      },
      "inputs": [
        {
          "identifier": "code",
          "type": "TEXT_INPUT",
          "required": true
        }
      ],
      "executor": {
        "name": "GithubOAuthExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

---

### Social Login with Google or GitHub

Social login with provider choice between Google and GitHub.

```json
{
  "name": "Social Login Flow",
  "handle": "social-login-flow",
  "flowType": "AUTHENTICATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "choose_provider"
    },
    {
      "id": "choose_provider",
      "type": "PROMPT",
      "actions": [
        {
          "ref": "google_btn",
          "nextNode": "google_auth"
        },
        {
          "ref": "github_btn",
          "nextNode": "github_auth"
        }
      ]
    },
    {
      "id": "google_auth",
      "type": "TASK_EXECUTION",
      "properties": {
        "idpId": "<google-idp-id>"
      },
      "inputs": [
        {
          "identifier": "code",
          "type": "TEXT_INPUT",
          "required": true
        }
      ],
      "executor": {
        "name": "GoogleOIDCAuthExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "github_auth",
      "type": "TASK_EXECUTION",
      "properties": {
        "idpId": "<github-idp-id>"
      },
      "inputs": [
        {
          "identifier": "code",
          "type": "TEXT_INPUT",
          "required": true
        }
      ],
      "executor": {
        "name": "GithubOAuthExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

---

### Multi-Factor Authentication (Password + SMS OTP)

Password authentication followed by OTP verification.

```json
{
  "name": "MFA Flow",
  "handle": "mfa-flow",
  "flowType": "AUTHENTICATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "prompt_credentials"
    },
    {
      "id": "prompt_credentials",
      "type": "PROMPT",
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
    },
    {
      "id": "basic_auth",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "BasicAuthExecutor"
      },
      "onSuccess": "send_otp"
    },
    {
      "id": "send_otp",
      "type": "TASK_EXECUTION",
      "properties": {
        "senderId": "<sender-id>"
      },
      "executor": {
        "name": "SMSOTPAuthExecutor",
        "mode": "send"
      },
      "onSuccess": "prompt_otp"
    },
    {
      "id": "prompt_otp",
      "type": "PROMPT",
      "inputs": [
        {
          "identifier": "otp",
          "type": "OTP_INPUT",
          "required": true
        }
      ],
      "actions": [
        {
          "ref": "action_001",
          "nextNode": "verify_otp"
        }
      ]
    },
    {
      "id": "verify_otp",
      "type": "TASK_EXECUTION",
      "properties": {
        "senderId": "<sender-id>"
      },
      "executor": {
        "name": "SMSOTPAuthExecutor",
        "mode": "verify"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

---

## Registration Flow Examples

### Username and Password Registration

Username password registration with provisioning and auto-login.

```json
{
  "name": "Basic Registration",
  "handle": "basic-registration",
  "flowType": "REGISTRATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "user_type_resolver"
    },
    {
      "id": "user_type_resolver",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "UserTypeResolver"
      },
      "onSuccess": "prompt_credentials"
    },
    {
      "id": "prompt_credentials",
      "type": "PROMPT",
      "inputs": [
        {
          "ref": "input_001",
          "identifier": "username",
          "type": "TEXT_INPUT",
          "required": true
        },
        {
          "ref": "input_002",
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
    },
    {
      "id": "basic_auth",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "BasicAuthExecutor"
      },
      "onSuccess": "provisioning"
    },
    {
      "id": "provisioning",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "ProvisioningExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

> [!TIP]
> Remove `AuthAssertExecutor` if you don't want auto-login after registration.

---

### SMS OTP Registration

Mobile number registration with SMS OTP verification, provisioning, and auto-login.

```json
{
  "name": "SMS Registration",
  "handle": "sms-registration",
  "flowType": "REGISTRATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "user_type_resolver"
    },
    {
      "id": "user_type_resolver",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "UserTypeResolver"
      },
      "onSuccess": "prompt_mobile"
    },
    {
      "id": "prompt_mobile",
      "type": "PROMPT",
      "inputs": [
        {
          "ref": "input_001",
          "identifier": "mobileNumber",
          "type": "PHONE_INPUT",
          "required": true
        }
      ],
      "actions": [
        {
          "ref": "action_001",
          "nextNode": "send_sms"
        }
      ]
    },
    {
      "id": "send_sms",
      "type": "TASK_EXECUTION",
      "properties": {
        "senderId": "<sender-id>"
      },
      "executor": {
        "name": "SMSOTPAuthExecutor",
        "mode": "send"
      },
      "onSuccess": "prompt_otp"
    },
    {
      "id": "prompt_otp",
      "type": "PROMPT",
      "inputs": [
        {
          "ref": "input_002",
          "identifier": "otp",
          "type": "OTP_INPUT",
          "required": true
        }
      ],
      "actions": [
        {
          "ref": "action_001",
          "nextNode": "verify_sms"
        }
      ]
    },
    {
      "id": "verify_sms",
      "type": "TASK_EXECUTION",
      "properties": {
        "senderId": "<sender-id>"
      },
      "executor": {
        "name": "SMSOTPAuthExecutor",
        "mode": "verify"
      },
      "onSuccess": "provisioning"
    },
    {
      "id": "provisioning",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "ProvisioningExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

---

### Social Registration with Google

Social registration using Google OIDC provider.

```json
{
  "name": "Google Registration",
  "handle": "google-registration",
  "flowType": "REGISTRATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "user_type_resolver"
    },
    {
      "id": "user_type_resolver",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "UserTypeResolver"
      },
      "onSuccess": "google_auth"
    },
    {
      "id": "google_auth",
      "type": "TASK_EXECUTION",
      "properties": {
        "idpId": "<google-idp-id>"
      },
      "inputs": [
        {
          "ref": "input_001",
          "identifier": "code",
          "type": "TEXT_INPUT",
          "required": true
        }
      ],
      "executor": {
        "name": "GoogleOIDCAuthExecutor"
      },
      "onSuccess": "provisioning"
    },
    {
      "id": "provisioning",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "ProvisioningExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

---

### Social Registration with GitHub

Social registration using GitHub OAuth provider.

```json
{
  "name": "GitHub Registration",
  "handle": "github-registration",
  "flowType": "REGISTRATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "user_type_resolver"
    },
    {
      "id": "user_type_resolver",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "UserTypeResolver"
      },
      "onSuccess": "github_auth"
    },
    {
      "id": "github_auth",
      "type": "TASK_EXECUTION",
      "properties": {
        "idpId": "<github-idp-id>"
      },
      "inputs": [
        {
          "ref": "input_001",
          "identifier": "code",
          "type": "TEXT_INPUT",
          "required": true
        }
      ],
      "executor": {
        "name": "GithubOAuthExecutor"
      },
      "onSuccess": "provisioning"
    },
    {
      "id": "provisioning",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "ProvisioningExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

---

### Social Registration with Google or GitHub

Social registration with provider choice between Google and GitHub.

```json
{
  "name": "Social Registration",
  "handle": "social-registration",
  "flowType": "REGISTRATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "user_type_resolver"
    },
    {
      "id": "user_type_resolver",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "UserTypeResolver"
      },
      "onSuccess": "choose_provider"
    },
    {
      "id": "choose_provider",
      "type": "PROMPT",
      "actions": [
        {
          "ref": "google_btn",
          "nextNode": "google_auth"
        },
        {
          "ref": "github_btn",
          "nextNode": "github_auth"
        }
      ]
    },
    {
      "id": "google_auth",
      "type": "TASK_EXECUTION",
      "properties": {
        "idpId": "<google-idp-id>"
      },
      "inputs": [
        {
          "ref": "input_001",
          "identifier": "code",
          "type": "TEXT_INPUT",
          "required": true
        }
      ],
      "executor": {
        "name": "GoogleOIDCAuthExecutor"
      },
      "onSuccess": "provisioning"
    },
    {
      "id": "github_auth",
      "type": "TASK_EXECUTION",
      "properties": {
        "idpId": "<github-idp-id>"
      },
      "inputs": [
        {
          "ref": "input_001",
          "identifier": "code",
          "type": "TEXT_INPUT",
          "required": true
        }
      ],
      "executor": {
        "name": "GithubOAuthExecutor"
      },
      "onSuccess": "provisioning"
    },
    {
      "id": "provisioning",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "ProvisioningExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```

---

### Registration with OU Creation

Flow with OUExecutor for creating organization units during registration.

```json
{
  "name": "Registration with OU",
  "handle": "registration-with-ou",
  "flowType": "REGISTRATION",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "onSuccess": "user_type_resolver"
    },
    {
      "id": "user_type_resolver",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "UserTypeResolver"
      },
      "onSuccess": "prompt_credentials"
    },
    {
      "id": "prompt_credentials",
      "type": "PROMPT",
      "inputs": [
        {
          "ref": "input_001",
          "identifier": "username",
          "type": "TEXT_INPUT",
          "required": true
        },
        {
          "ref": "input_002",
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
    },
    {
      "id": "basic_auth",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "BasicAuthExecutor"
      },
      "onSuccess": "ou_creation"
    },
    {
      "id": "ou_creation",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "OUExecutor"
      },
      "onSuccess": "provisioning"
    },
    {
      "id": "provisioning",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "ProvisioningExecutor"
      },
      "onSuccess": "auth_assert"
    },
    {
      "id": "auth_assert",
      "type": "TASK_EXECUTION",
      "executor": {
        "name": "AuthAssertExecutor"
      },
      "onSuccess": "end"
    },
    {
      "id": "end",
      "type": "END"
    }
  ]
}
```
