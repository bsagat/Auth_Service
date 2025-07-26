SET TIMEZONE = 'Asia/Almaty';

CREATE TYPE role as enum ('admin', 'user');

CREATE TABLE IF NOT EXISTS Users (
    ID SERIAL PRIMARY KEY,
    Name VARCHAR(100) NOT NULL,
    Email VARCHAR(255) UNIQUE NOT NULL,
    PassHash VARCHAR(255) NOT NULL,
    Created_At TIMESTAMPTZ DEFAULT NOW(),
    Updated_At TIMESTAMPTZ,
    IsAdmin Bool DEFAULT false,
    Role role NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_email ON Users (Email);