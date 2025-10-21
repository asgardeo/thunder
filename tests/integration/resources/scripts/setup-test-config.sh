#!/bin/bash

DB_TYPE=${DB_TYPE:-sqlite}

cat > tests/integration/resources/deployment.yaml <<EOF
server:
  hostname: localhost
  port: 8095


security:
  cert_file: "repository/resources/security/server.cert"
  key_file: "repository/resources/security/server.key"

database:
EOF

if [ "$DB_TYPE" = "postgres" ]; then
  cat >> tests/integration/resources/deployment.yaml <<EOF
  identity:
    type: postgres
    hostname: localhost
    port: 5432
    name: identitydb
    username: asgthunder
    password: asgthunder
    sslmode: disable
    path: ""
    options: ""

  runtime:
    type: postgres
    hostname: localhost
    port: 5432
    name: runtimedb
    username: asgthunder
    password: asgthunder
    sslmode: disable
    path: ""
    options: ""
EOF
else
  cat >> tests/integration/resources/deployment.yaml <<EOF
  identity:
    type: sqlite
    hostname: ""
    port: 0
    name: ""
    username: ""
    password: ""
    sslmode: ""
    path: "repository/database/thunderdb.db"
    options: "cache=shared"

  runtime:
    type: sqlite
    hostname: ""
    port: 0
    name: ""
    username: ""
    password: ""
    sslmode: ""
    path: "repository/database/runtimedb.db"
    options: "cache=shared"
EOF
fi

cat >> tests/integration/resources/deployment.yaml <<EOF


flow:
  graph_directory: "repository/resources/graphs/"
  authn:
    default_flow: "auth_flow_config_basic"
EOF
