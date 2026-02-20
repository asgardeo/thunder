-- Table to store Organization Units
CREATE TABLE ORGANIZATION_UNIT (
    ID          INTEGER PRIMARY KEY AUTOINCREMENT,
    DEPLOYMENT_ID   VARCHAR(255) NOT NULL,
    OU_ID       VARCHAR(36) NOT NULL,
    PARENT_ID   VARCHAR(36),
    HANDLE      VARCHAR(100)        NOT NULL,
    NAME        VARCHAR(100)        NOT NULL,
    DESCRIPTION VARCHAR(255),
    THEME_ID    VARCHAR(36),
    LAYOUT_ID   VARCHAR(36),
    METADATA     TEXT,
    CREATED_AT  TEXT DEFAULT (datetime('now')),
    UPDATED_AT  TEXT DEFAULT (datetime('now')),
    UNIQUE (OU_ID, DEPLOYMENT_ID)
);

-- Index for deployment isolation on ORGANIZATION_UNIT
CREATE INDEX idx_organization_unit_deployment_id ON ORGANIZATION_UNIT (DEPLOYMENT_ID);

-- Table to store Users
CREATE TABLE USER (
    ID          INTEGER PRIMARY KEY AUTOINCREMENT,
    DEPLOYMENT_ID   VARCHAR(255) NOT NULL,
    USER_ID     VARCHAR(36)        NOT NULL,
    OU_ID       VARCHAR(36)        NOT NULL,
    TYPE        TEXT               NOT NULL,
    ATTRIBUTES  TEXT,
    CREDENTIALS TEXT,
    CREATED_AT  TEXT DEFAULT (datetime('now')),
    UPDATED_AT  TEXT DEFAULT (datetime('now')),
    UNIQUE (USER_ID, DEPLOYMENT_ID)
);

-- Index for deployment isolation on USER
CREATE INDEX idx_user_deployment_id ON USER (DEPLOYMENT_ID);

-- Table to store Groups
CREATE TABLE "GROUP" (
    ID          INTEGER PRIMARY KEY AUTOINCREMENT,
    DEPLOYMENT_ID   VARCHAR(255) NOT NULL,
    GROUP_ID    VARCHAR(36)        NOT NULL,
    OU_ID       VARCHAR(36)        NOT NULL,
    NAME        VARCHAR(50)        NOT NULL,
    DESCRIPTION VARCHAR(255),
    CREATED_AT  TEXT DEFAULT (datetime('now')),
    UPDATED_AT  TEXT DEFAULT (datetime('now')),
    UNIQUE (GROUP_ID, DEPLOYMENT_ID)
);

-- Index for deployment isolation on GROUP
CREATE INDEX idx_group_deployment_id ON "GROUP" (DEPLOYMENT_ID);

-- Table to store Group member assignments
CREATE TABLE GROUP_MEMBER_REFERENCE (
    ID          INTEGER PRIMARY KEY AUTOINCREMENT,
    DEPLOYMENT_ID   VARCHAR(255) NOT NULL,
    GROUP_ID    VARCHAR(36) NOT NULL,
    MEMBER_TYPE VARCHAR(7)  NOT NULL,
    MEMBER_ID   VARCHAR(36) NOT NULL,
    CREATED_AT  TEXT DEFAULT (datetime('now')),
    UPDATED_AT  TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (GROUP_ID, DEPLOYMENT_ID) REFERENCES "GROUP" (GROUP_ID, DEPLOYMENT_ID) ON DELETE CASCADE,
    CONSTRAINT unique_group_member UNIQUE (GROUP_ID, MEMBER_TYPE, MEMBER_ID, DEPLOYMENT_ID)
);

-- Index for deployment isolation on GROUP_MEMBER_REFERENCE
CREATE INDEX idx_group_member_reference_deployment_id ON GROUP_MEMBER_REFERENCE (DEPLOYMENT_ID);

-- Table to store indexed user attributes for fast lookups
CREATE TABLE USER_INDEXED_ATTRIBUTES (
    ID              INTEGER PRIMARY KEY AUTOINCREMENT,
    DEPLOYMENT_ID   VARCHAR(255) NOT NULL,
    USER_ID         VARCHAR(36) NOT NULL,
    ATTRIBUTE_NAME  VARCHAR(255) NOT NULL,
    ATTRIBUTE_VALUE TEXT NOT NULL,
    CREATED_AT      TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (USER_ID, DEPLOYMENT_ID) REFERENCES USER (USER_ID, DEPLOYMENT_ID) ON DELETE CASCADE,
    UNIQUE (USER_ID, DEPLOYMENT_ID, ATTRIBUTE_NAME)
);

-- Index for fast attribute lookups (primary use case for authentication)
CREATE INDEX idx_user_indexed_attributes_lookup ON USER_INDEXED_ATTRIBUTES (ATTRIBUTE_NAME, ATTRIBUTE_VALUE, DEPLOYMENT_ID);

-- Index for user-based queries and cleanup
CREATE INDEX idx_user_indexed_attributes_user ON USER_INDEXED_ATTRIBUTES (USER_ID, DEPLOYMENT_ID);

-- Index for deployment isolation
CREATE INDEX idx_user_indexed_attributes_deployment_id ON USER_INDEXED_ATTRIBUTES (DEPLOYMENT_ID);
