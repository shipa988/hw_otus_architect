DROP DATABASE IF EXISTS networkcore;
CREATE DATABASE networkcore;
CREATE USER admin WITH encrypted password 'admin';
GRANT ALL PRIVILEGES ON DATABASE networkcore to admin;