-- Table to store User Schemas
CREATE TABLE USER_SCHEMAS (
    DEPLOYMENT_ID   VARCHAR(255) NOT NULL,
    ID          VARCHAR(36) PRIMARY KEY,
    NAME        VARCHAR(100) NOT NULL,
    OU_ID       VARCHAR(36) NOT NULL,
    ALLOW_SELF_REGISTRATION INTEGER NOT NULL DEFAULT 0,
    SCHEMA_DEF  TEXT NOT NULL,
    SYSTEM_ATTRIBUTES TEXT,
    CREATED_AT  TEXT DEFAULT (datetime('now')),
    UPDATED_AT  TEXT DEFAULT (datetime('now')),
    UNIQUE (NAME, DEPLOYMENT_ID)
);

-- Composite index for deployment + OU-based user schema lookups
CREATE INDEX idx_user_schemas_deployment_ou ON USER_SCHEMAS (DEPLOYMENT_ID, OU_ID);

-- Table to store Roles
CREATE TABLE ROLE (
    DEPLOYMENT_ID           VARCHAR(255) NOT NULL,
    ID                  VARCHAR(36) PRIMARY KEY,
    OU_ID               VARCHAR(36) NOT NULL,
    NAME                VARCHAR(50) NOT NULL,
    DESCRIPTION         VARCHAR(255),
    CREATED_AT          TEXT DEFAULT (datetime('now')),
    UPDATED_AT          TEXT DEFAULT (datetime('now')),
    CONSTRAINT unique_role_ou_name UNIQUE (OU_ID, NAME, DEPLOYMENT_ID)
);

-- Composite index for deployment + OU lookups (supports UNIQUE constraint checks)
CREATE INDEX idx_role_ou_deployment ON ROLE (DEPLOYMENT_ID, OU_ID);

-- Table to store Role permissions
CREATE TABLE ROLE_PERMISSION (
    DEPLOYMENT_ID       VARCHAR(255) NOT NULL,
    ROLE_ID             VARCHAR(36) NOT NULL,
    RESOURCE_SERVER_ID  VARCHAR(36) NOT NULL,
    PERMISSION          VARCHAR(1000) NOT NULL,
    CREATED_AT          TEXT DEFAULT (datetime('now')),
    PRIMARY KEY (ROLE_ID, DEPLOYMENT_ID, RESOURCE_SERVER_ID, PERMISSION),
    FOREIGN KEY (ROLE_ID) REFERENCES ROLE (ID) ON DELETE CASCADE
);

-- Index for resource server queries with deployment isolation on ROLE_PERMISSION
CREATE INDEX idx_role_permission_resource_server ON ROLE_PERMISSION (RESOURCE_SERVER_ID, DEPLOYMENT_ID);

-- Table to store Role assignments (to users and groups)
CREATE TABLE ROLE_ASSIGNMENT (
    DEPLOYMENT_ID       VARCHAR(255) NOT NULL,
    ROLE_ID         VARCHAR(36) NOT NULL,
    ASSIGNEE_TYPE   VARCHAR(5)  NOT NULL CHECK (ASSIGNEE_TYPE IN ('user', 'group')),
    ASSIGNEE_ID     VARCHAR(36) NOT NULL,
    CREATED_AT      TEXT DEFAULT (datetime('now')),
    UPDATED_AT      TEXT DEFAULT (datetime('now')),
    PRIMARY KEY (ROLE_ID, DEPLOYMENT_ID, ASSIGNEE_TYPE, ASSIGNEE_ID),
    FOREIGN KEY (ROLE_ID) REFERENCES ROLE (ID) ON DELETE CASCADE
);

-- Table to store theme configurations.
CREATE TABLE THEME (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    ID VARCHAR(36) PRIMARY KEY,
    DISPLAY_NAME VARCHAR(255) NOT NULL,
    DESCRIPTION VARCHAR(512),
    THEME TEXT NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now'))
);

-- Index for deployment isolation on THEME
CREATE INDEX idx_theme_deployment_id ON THEME (DEPLOYMENT_ID);

-- Table to store layout configurations.
CREATE TABLE LAYOUT (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    ID VARCHAR(36) PRIMARY KEY,
    DISPLAY_NAME VARCHAR(255) NOT NULL,
    DESCRIPTION VARCHAR(512),
    LAYOUT TEXT NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now'))
);

-- Index for deployment isolation on LAYOUT
CREATE INDEX idx_layout_deployment_id ON LAYOUT (DEPLOYMENT_ID);

-- Table to store application details.
CREATE TABLE APPLICATION (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    ID VARCHAR(36) PRIMARY KEY,
    APP_NAME VARCHAR(255) NOT NULL,
    DESCRIPTION VARCHAR(255) NOT NULL,
    AUTH_FLOW_ID VARCHAR(100) NOT NULL,
    REGISTRATION_FLOW_ID VARCHAR(100) NOT NULL,
    IS_REGISTRATION_FLOW_ENABLED CHAR(1) DEFAULT '1',
    THEME_ID VARCHAR(36),
    LAYOUT_ID VARCHAR(36),
    APP_JSON TEXT,
    FOREIGN KEY (THEME_ID) REFERENCES THEME(ID) ON DELETE RESTRICT,
    FOREIGN KEY (LAYOUT_ID) REFERENCES LAYOUT(ID) ON DELETE RESTRICT
);

-- Composite index for name-based application lookups
CREATE INDEX idx_application_name_deployment ON APPLICATION (DEPLOYMENT_ID, APP_NAME);

-- Index for efficient lookups of applications by theme.
CREATE INDEX idx_application_theme_id ON APPLICATION(THEME_ID);

-- Index for efficient lookups of applications by layout.
CREATE INDEX idx_application_layout_id ON APPLICATION(LAYOUT_ID);

-- Table to store OAuth configurations for applications.
CREATE TABLE APP_OAUTH_INBOUND_CONFIG (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    CLIENT_ID VARCHAR(255) NOT NULL,
    CLIENT_SECRET VARCHAR(255) NOT NULL,
    APP_ID VARCHAR(36) NOT NULL,
    OAUTH_CONFIG_JSON TEXT,
    PRIMARY KEY (CLIENT_ID, DEPLOYMENT_ID),
    FOREIGN KEY (APP_ID) REFERENCES APPLICATION(ID) ON DELETE CASCADE
);

-- Index for APP_ID lookups on APP_OAUTH_INBOUND_CONFIG (UPDATE/DELETE by app ID, and JOIN in application list)
CREATE INDEX idx_app_oauth_app_id ON APP_OAUTH_INBOUND_CONFIG (APP_ID);

-- Table to store identity providers.
CREATE TABLE IDP (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    ID VARCHAR(36) PRIMARY KEY,
    NAME VARCHAR(255) NOT NULL,
    DESCRIPTION VARCHAR(500),
    TYPE VARCHAR(20) NOT NULL,
    PROPERTIES TEXT,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now'))
);

-- Composite index for name-based IDP lookups
CREATE INDEX idx_idp_name_deployment ON IDP (DEPLOYMENT_ID, NAME);

-- Table to store notification senders.
CREATE TABLE NOTIFICATION_SENDER (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    NAME VARCHAR(255) NOT NULL,
    ID VARCHAR(36) PRIMARY KEY,
    DESCRIPTION VARCHAR(500),
    TYPE VARCHAR(20) NOT NULL,
    PROVIDER VARCHAR(20) NOT NULL,
    PROPERTIES TEXT,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now'))
);

-- Composite index for name-based notification sender lookups
CREATE INDEX idx_notification_sender_name_deployment ON NOTIFICATION_SENDER (DEPLOYMENT_ID, NAME);

-- Table to store certificates associated with various entities.
CREATE TABLE CERTIFICATE (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    ID VARCHAR(36) PRIMARY KEY,
    REF_TYPE VARCHAR(20) NOT NULL,
    REF_ID VARCHAR(36) NOT NULL,
    TYPE VARCHAR(20) NOT NULL,
    VALUE TEXT NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now')),
    UNIQUE (REF_TYPE, REF_ID, DEPLOYMENT_ID)
);

-- Table to store resource servers.
CREATE TABLE RESOURCE_SERVER (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    ID VARCHAR(36) PRIMARY KEY,
    OU_ID VARCHAR(36) NOT NULL,
    NAME VARCHAR(100) NOT NULL,
    DESCRIPTION TEXT,
    IDENTIFIER VARCHAR(100),
    PROPERTIES TEXT,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now')),
    UNIQUE (OU_ID, NAME, DEPLOYMENT_ID)
);

-- Composite index for name-based resource server lookups
CREATE INDEX idx_resource_server_name_deployment ON RESOURCE_SERVER (DEPLOYMENT_ID, NAME);

-- Unique constraint: Resource server identifier must be unique per deployment (when not null)
CREATE UNIQUE INDEX uq_resource_server_identifier
    ON RESOURCE_SERVER(IDENTIFIER, DEPLOYMENT_ID)
    WHERE IDENTIFIER IS NOT NULL;

-- Table to store resources within resource servers.
CREATE TABLE RESOURCE (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    ID VARCHAR(36) PRIMARY KEY,
    RESOURCE_SERVER_ID VARCHAR(36) NOT NULL,
    PARENT_RESOURCE_ID VARCHAR(36),
    NAME VARCHAR(100) NOT NULL,
    HANDLE VARCHAR(100) NOT NULL,
    DESCRIPTION TEXT,
    PROPERTIES TEXT,
    PERMISSION VARCHAR(1000) NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now')),

    FOREIGN KEY (RESOURCE_SERVER_ID)
        REFERENCES RESOURCE_SERVER(ID)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,
    FOREIGN KEY (PARENT_RESOURCE_ID)
        REFERENCES RESOURCE(ID)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);

-- Composite index for resource server + deployment queries (list, count, and handle checks)
CREATE INDEX idx_resource_server_deployment ON RESOURCE (RESOURCE_SERVER_ID, DEPLOYMENT_ID);

-- Unique constraint: Resource handle must be unique under the same parent per deployment
CREATE UNIQUE INDEX uq_resource_handle_with_parent
    ON RESOURCE(RESOURCE_SERVER_ID, PARENT_RESOURCE_ID, HANDLE, DEPLOYMENT_ID)
    WHERE PARENT_RESOURCE_ID IS NOT NULL;

-- Unique constraint: Root-level resource handles must be unique per resource server per deployment
CREATE UNIQUE INDEX uq_resource_handle_null_parent
    ON RESOURCE(RESOURCE_SERVER_ID, HANDLE, DEPLOYMENT_ID)
    WHERE PARENT_RESOURCE_ID IS NULL;

-- Table to store actions at resource server or resource level.
CREATE TABLE ACTION (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    ID VARCHAR(36) PRIMARY KEY,
    RESOURCE_SERVER_ID VARCHAR(36) NOT NULL,
    RESOURCE_ID VARCHAR(36),
    NAME VARCHAR(100) NOT NULL,
    HANDLE VARCHAR(100) NOT NULL,
    DESCRIPTION TEXT,
    PERMISSION VARCHAR(1000) NOT NULL,
    PROPERTIES TEXT,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now')),

    FOREIGN KEY (RESOURCE_SERVER_ID)
        REFERENCES RESOURCE_SERVER(ID)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,
    FOREIGN KEY (RESOURCE_ID)
        REFERENCES RESOURCE(ID)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);

-- Composite index for action list/count queries filtered by resource server + deployment + resource
CREATE INDEX idx_action_server_deployment ON ACTION (RESOURCE_SERVER_ID, DEPLOYMENT_ID, RESOURCE_ID);

-- Unique constraint: Server-level action handles must be unique per resource server per deployment
CREATE UNIQUE INDEX uq_action_server_handle
    ON ACTION(RESOURCE_SERVER_ID, HANDLE, DEPLOYMENT_ID)
    WHERE RESOURCE_ID IS NULL;

-- Unique constraint: Resource-level action handles must be unique per resource per deployment
CREATE UNIQUE INDEX uq_action_resource_handle
    ON ACTION(RESOURCE_ID, HANDLE, DEPLOYMENT_ID)
    WHERE RESOURCE_ID IS NOT NULL;

-- Table to store active flow definitions
CREATE TABLE FLOW (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    ID VARCHAR(36) PRIMARY KEY,
    HANDLE VARCHAR(100) NOT NULL,
    NAME VARCHAR(100) NOT NULL,
    FLOW_TYPE VARCHAR(50) NOT NULL,
    ACTIVE_VERSION INTEGER NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now')),
    UNIQUE (HANDLE, FLOW_TYPE, DEPLOYMENT_ID)
);

-- Composite index for flow type + deployment queries
CREATE INDEX idx_flow_type_deployment ON FLOW (DEPLOYMENT_ID, FLOW_TYPE);

-- Table to store flow version history
CREATE TABLE FLOW_VERSION (
    DEPLOYMENT_ID VARCHAR(255) NOT NULL,
    FLOW_ID VARCHAR(36) NOT NULL,
    VERSION INTEGER NOT NULL,
    NODES TEXT NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    PRIMARY KEY (FLOW_ID, VERSION, DEPLOYMENT_ID),
    FOREIGN KEY (FLOW_ID)
        REFERENCES FLOW(ID)
        ON DELETE CASCADE
);

-- Table to store i18n translations
CREATE TABLE TRANSLATION (
    DEPLOYMENT_ID   VARCHAR(255) NOT NULL,
    MESSAGE_KEY     VARCHAR(255) NOT NULL,
    LANGUAGE_CODE   VARCHAR(10) NOT NULL,
    NAMESPACE       VARCHAR(50) NOT NULL DEFAULT 'default',
    VALUE           TEXT NOT NULL,
    CREATED_AT      TEXT DEFAULT (datetime('now')),
    UPDATED_AT      TEXT DEFAULT (datetime('now')),
    PRIMARY KEY (DEPLOYMENT_ID, NAMESPACE, MESSAGE_KEY, LANGUAGE_CODE)
);

-- Index for efficient language and namespace combination lookups
CREATE INDEX idx_translation_lang_namespace ON TRANSLATION (DEPLOYMENT_ID, LANGUAGE_CODE, NAMESPACE);
