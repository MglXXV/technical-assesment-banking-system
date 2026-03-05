-- Enable the pgcrypto extension to allow UUID generation if not already available
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create the users table to store relational data and authentication credentials
CREATE TABLE IF NOT EXISTS Users (
    -- Unique Identifier (Primary Key)
    uuid_user UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- User Information
    user_fullname VARCHAR(255) NOT NULL,

    -- Access Credentials
    -- Email is used as the unique login identifier
    user_email VARCHAR(255) UNIQUE NOT NULL,
    -- Stored as a bcrypt hash string
    user_password VARCHAR(255) NOT NULL,

    -- TigerBeetle Account Reference
    tb_account_id TEXT NOT NULL DEFAULT '[]',

    -- Audit Timestamps
    -- Using TIMESTAMPTZ to ensure timezone consistency
    user_created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Support for GORM's soft delete feature
    user_deleted_at TIMESTAMP WITH TIME ZONE
);

-- Index for optimized searching during login and email validation
CREATE INDEX IF NOT EXISTS idx_users_email ON User(user_email);

-- Index for optimized soft delete queries (GORM standard)
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON Users(user_deleted_at);