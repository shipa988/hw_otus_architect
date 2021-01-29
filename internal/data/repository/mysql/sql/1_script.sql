--DROP DATABASE IF EXISTS networkcore;
CREATE DATABASE networkcore;
CREATE USER user WITH encrypted password 'userpw';
GRANT ALL PRIVILEGES ON DATABASE networkcore to user;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";