-- Table to store Roles
CREATE TABLE "ROLE" (
    ID                  INTEGER PRIMARY KEY AUTOINCREMENT,
    ROLE_ID             VARCHAR(36) UNIQUE NOT NULL,
    OU_ID               VARCHAR(36) NOT NULL,
    NAME                VARCHAR(50) NOT NULL,
    DESCRIPTION         VARCHAR(255),
    CREATED_AT          TEXT DEFAULT (datetime('now')),
    UPDATED_AT          TEXT DEFAULT (datetime('now')),
    CONSTRAINT unique_role_ou_name UNIQUE (OU_ID, NAME)
);

-- Table to store Role permissions
CREATE TABLE ROLE_PERMISSION (
    ID              INTEGER PRIMARY KEY AUTOINCREMENT,
    ROLE_ID         VARCHAR(36) NOT NULL,
    PERMISSION      VARCHAR(100) NOT NULL,
    CREATED_AT      TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (ROLE_ID) REFERENCES "ROLE" (ROLE_ID) ON DELETE CASCADE,
    CONSTRAINT unique_role_permission UNIQUE (ROLE_ID, PERMISSION)
);

-- Table to store Role assignments (to users and groups)
CREATE TABLE ROLE_ASSIGNMENT (
    ID              INTEGER PRIMARY KEY AUTOINCREMENT,
    ROLE_ID         VARCHAR(36) NOT NULL,
    ASSIGNEE_TYPE   VARCHAR(5)  NOT NULL CHECK (ASSIGNEE_TYPE IN ('user', 'group')),
    ASSIGNEE_ID     VARCHAR(36) NOT NULL,
    CREATED_AT      TEXT DEFAULT (datetime('now')),
    UPDATED_AT      TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (ROLE_ID) REFERENCES "ROLE" (ROLE_ID) ON DELETE CASCADE,
    CONSTRAINT unique_role_assignment UNIQUE (ROLE_ID, ASSIGNEE_TYPE, ASSIGNEE_ID)
);

-- Indexes for authorization queries

-- Index for finding all roles assigned to a specific assignee
CREATE INDEX idx_role_assignment_assignee 
ON ROLE_ASSIGNMENT (ASSIGNEE_ID, ASSIGNEE_TYPE);

-- Index for finding all permissions for a specific role
CREATE INDEX idx_role_permission_role 
ON ROLE_PERMISSION (ROLE_ID);

-- Table to store basic service provider (app) details.
CREATE TABLE SP_APP (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    APP_ID VARCHAR(36) UNIQUE NOT NULL,
    APP_NAME VARCHAR(255) NOT NULL,
    DESCRIPTION VARCHAR(50) NOT NULL,
    AUTH_FLOW_GRAPH_ID VARCHAR(50) NOT NULL,
    REGISTRATION_FLOW_GRAPH_ID VARCHAR(50) NOT NULL,
    IS_REGISTRATION_FLOW_ENABLED CHAR(1) DEFAULT '1',
    APP_JSON TEXT
);

-- Table to store OAuth configurations for SP apps.
CREATE TABLE IDN_OAUTH_CONSUMER_APPS (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    CONSUMER_KEY VARCHAR(255) NOT NULL,
    CONSUMER_SECRET VARCHAR(255) NOT NULL,
    APP_ID VARCHAR(36) NOT NULL,
    OAUTH_CONFIG_JSON TEXT,
    FOREIGN KEY (APP_ID) REFERENCES SP_APP(APP_ID) ON DELETE CASCADE
);

-- Table to store inbound auth configs (e.g., OAuth, SAML) for SP apps.
CREATE TABLE SP_INBOUND_AUTH (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    INBOUND_AUTH_KEY VARCHAR(255) NOT NULL,
    INBOUND_AUTH_TYPE VARCHAR(50) NOT NULL,
    APP_ID VARCHAR(36) NOT NULL,
    FOREIGN KEY (APP_ID) REFERENCES SP_APP(APP_ID) ON DELETE CASCADE
);

-- Table to store identity providers.
CREATE TABLE IDP (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    IDP_ID VARCHAR(36) UNIQUE NOT NULL,
    NAME VARCHAR(255) NOT NULL,
    DESCRIPTION VARCHAR(500),
    TYPE VARCHAR(20) NOT NULL,
    PROPERTIES TEXT,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now'))
);

-- Table to store notification senders.
CREATE TABLE NOTIFICATION_SENDER (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    NAME VARCHAR(255) NOT NULL,
    SENDER_ID VARCHAR(36) UNIQUE NOT NULL,
    DESCRIPTION VARCHAR(500),
    TYPE VARCHAR(20) NOT NULL,
    PROVIDER VARCHAR(20) NOT NULL,
    PROPERTIES TEXT,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now'))
);

-- Table to store certificates associated with various entities.
CREATE TABLE CERTIFICATE (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    CERT_ID VARCHAR(36) UNIQUE NOT NULL,
    REF_TYPE VARCHAR(20) NOT NULL,
    REF_ID VARCHAR(36) NOT NULL,
    TYPE VARCHAR(20) NOT NULL,
    VALUE TEXT NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now')),
    UNIQUE (REF_TYPE, REF_ID)
);

-- Table to store user schemas.
CREATE TABLE USER_SCHEMAS (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    SCHEMA_ID VARCHAR(36) UNIQUE NOT NULL,
    NAME VARCHAR(255) UNIQUE NOT NULL,
    SCHEMA_DEF TEXT NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now'))
);

-- Insert a pre-configured notification sender for SMS OTP tests
INSERT INTO NOTIFICATION_SENDER (NAME, SENDER_ID, DESCRIPTION, TYPE, PROVIDER, PROPERTIES) VALUES
('Custom SMS Sender', 'test-sms-sender-id', 'Custom SMS sender for integration tests', 'MESSAGE', 'custom', 
'[{"name":"url","value":"http://localhost:8098/send-sms","is_secret":false},{"name":"http_method","value":"POST","is_secret":false},{"name":"content_type","value":"JSON","is_secret":false}]');

-- Insert pre-configured IDPs for flow authentication tests
INSERT INTO IDP (IDP_ID, NAME, DESCRIPTION, TYPE, PROPERTIES) VALUES
('test-google-idp-id', 'Google', 'Google Identity Provider for integration tests', 'GOOGLE',
'[{"name":"client_id","value":"test_google_client","is_secret":false},{"name":"client_secret","value":"test_google_secret","is_secret":false},{"name":"authorization_endpoint","value":"http://localhost:8093/o/oauth2/v2/auth","is_secret":false},{"name":"token_endpoint","value":"http://localhost:8093/token","is_secret":false},{"name":"userinfo_endpoint","value":"http://localhost:8093/v1/userinfo","is_secret":false},{"name":"jwks_endpoint","value":"http://localhost:8093/oauth2/v3/certs","is_secret":false},{"name":"redirect_uri","value":"https://localhost:3000/google/callback","is_secret":false},{"name":"scopes","value":"openid email profile","is_secret":false}]');

INSERT INTO IDP (IDP_ID, NAME, DESCRIPTION, TYPE, PROPERTIES) VALUES
('test-github-idp-id', 'Github', 'GitHub Identity Provider for integration tests', 'GITHUB',
'[{"name":"client_id","value":"test_github_client","is_secret":false},{"name":"client_secret","value":"test_github_secret","is_secret":false},{"name":"authorization_endpoint","value":"http://localhost:8092/login/oauth/authorize","is_secret":false},{"name":"token_endpoint","value":"http://localhost:8092/login/oauth/access_token","is_secret":false},{"name":"userinfo_endpoint","value":"http://localhost:8092/user","is_secret":false},{"name":"redirect_uri","value":"https://localhost:3000/github/callback","is_secret":false},{"name":"scopes","value":"user:email,read:user","is_secret":false}]');
