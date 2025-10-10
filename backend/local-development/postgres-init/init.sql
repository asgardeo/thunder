-- Create databases
CREATE DATABASE runtimedb;
CREATE DATABASE thunderdb;

-- Run db1 initialization
\connect runtimedb
\i /docker-entrypoint-initdb.d/runtime-postgres.sql

-- Run db2 initialization
\connect thunderdb
\i /docker-entrypoint-initdb.d/thunder-postgres.sql
