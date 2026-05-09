-- Create databases
CREATE DATABASE runtimedb;
CREATE DATABASE configdb;
CREATE DATABASE userdb;

-- Run db1 initialization
\connect runtimedb
\i /docker-entrypoint-initdb.d/runtime-postgres.sql

-- Run db2 initialization
\connect configdb
\i /docker-entrypoint-initdb.d/config-postgres.sql

-- Run db3 initialization
\connect userdb
\i /docker-entrypoint-initdb.d/user-postgres.sql
