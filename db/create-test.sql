CREATE ROLE db_user LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE test_db
  WITH OWNER = db_user
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;

\connect test_db;

 CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

 CREATE TABLE "test_table" (
   "id" uuid DEFAULT uuid_generate_v4(),
   PRIMARY KEY ("id")
 );
