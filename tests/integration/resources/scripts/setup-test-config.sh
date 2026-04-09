#!/bin/bash

DB_TYPE=${DB_TYPE:-sqlite}

cat > tests/integration/resources/deployment.yaml <<EOF
server:
  hostname: localhost
  port: 8095


tls:
  cert_file: "repository/resources/security/server.cert"
  key_file: "repository/resources/security/server.key"

database:
EOF

if [ "$DB_TYPE" = "postgres" ]; then
  cat >> tests/integration/resources/deployment.yaml <<EOF
  config:
    type: postgres
    postgres:
      hostname: localhost
      port: 5432
      name: configdb
      username: asgthunder
      password: asgthunder
      sslmode: disable

  runtime:
    type: postgres
    postgres:
      hostname: localhost
      port: 5432
      name: runtimedb
      username: asgthunder
      password: asgthunder
      sslmode: disable

  user:
    type: postgres
    postgres:
      hostname: localhost
      port: 5432
      name: userdb
      username: asgthunder
      password: asgthunder
      sslmode: disable
EOF
elif [ "$DB_TYPE" = "redis" ]; then
  cat >> tests/integration/resources/deployment.yaml <<EOF
  config:
    type: sqlite
    sqlite:
      path: "repository/database/configdb.db"
      options: "cache=shared"

  runtime:
    type: redis
    redis:
      address: "localhost:6379"
      db: 0
      key_prefix: "thunder"

  user:
    type: sqlite
    sqlite:
      path: "repository/database/userdb.db"
      options: "cache=shared"
EOF
else
  cat >> tests/integration/resources/deployment.yaml <<EOF
  config:
    type: sqlite
    sqlite:
      path: "repository/database/configdb.db"
      options: "cache=shared"

  runtime:
    type: sqlite
    sqlite:
      path: "repository/database/runtimedb.db"
      options: "cache=shared"

  user:
    type: sqlite
    sqlite:
      path: "repository/database/userdb.db"
      options: "cache=shared"
EOF
fi

cat >> tests/integration/resources/deployment.yaml <<EOF


flow:
  max_version_history: 3
EOF
