CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- Users table to store user information
CREATE TABLE IF NOT EXISTS users (
    user_id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- Folder

CREATE TABLE IF NOT EXISTS folders (
    folder_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(255) NOT NULL,

    parent_folder_id UUID REFERENCES folders(folder_id),

    owner_id BIGINT NOT NULL REFERENCES users(user_id),

    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);


-- Files Metadata

CREATE TABLE IF NOT EXISTS file_metadata (
		file_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

		owner_id BIGINT NOT NULL REFERENCES users(user_id),

		parent_folder_id UUID REFERENCES folders(folder_id),

		file_name VARCHAR(255) NOT NULL,
		extension VARCHAR(50),
		mime_type VARCHAR(100),

		size_bytes BIGINT NOT NULL,

		checksum VARCHAR(255) NOT NULL,

		total_chunks INT NOT NULL,
		uploaded_chunks_count INT NOT NULL DEFAULT 0,

		s3_object_key VARCHAR(255) NOT NULL,

		multipart_upload_id VARCHAR(255),

		compression_algorithm VARCHAR(50) NOT NULL DEFAULT 'none',

		status VARCHAR(20) NOT NULL DEFAULT 'pending',

        is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMPTZ
	);


-- file chunks

CREATE TABLE IF NOT EXISTS file_chunks (
    file_id UUID NOT NULL REFERENCES file_metadata(file_id) ON DELETE CASCADE,

    chunk_index INTEGER NOT NULL,

    checksum VARCHAR(128) NOT NULL,

    chunk_size BIGINT NOT NULL,

    etag VARCHAR(255),

    status VARCHAR(50) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (file_id, chunk_index)
);


-- changes table for sync
CREATE TABLE IF NOT EXISTS changes (
    change_id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL REFERENCES users(user_id),

    entity_type VARCHAR(20) NOT NULL,
    entity_id UUID NOT NULL,
    parent_folder_id UUID,
    action VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
