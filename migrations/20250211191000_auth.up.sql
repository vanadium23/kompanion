CREATE TABLE auth_user (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auth_session (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL REFERENCES auth_user(username),
    session_key TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    ip_address INET NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    deactivated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON COLUMN auth_session.ip_address IS 'ip address that was used to create the session';
COMMENT ON COLUMN auth_session.user_agent IS 'User-Agent that was used to create the session';

CREATE TABLE auth_device (
    id BIGSERIAL PRIMARY KEY,
    device_name VARCHAR(32) NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    deactivated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE auth_device IS 'devices used to auth for webdav, sync and opds';
COMMENT ON COLUMN auth_device.hashed_password IS 'md5 hash of the password, because koreader send it in that format for sync';
