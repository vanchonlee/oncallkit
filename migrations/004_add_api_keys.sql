-- Migration: Add API Keys System
-- Created: 2024-01-15
-- Description: Add tables for API key authentication system

-- API Keys table
CREATE TABLE api_keys (
    id TEXT PRIMARY KEY DEFAULT 'apikey_' || substr(md5(random()::text), 1, 16),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,                    -- Friendly name: "Production Monitoring", "Dev Alerts"
    api_key TEXT UNIQUE NOT NULL,          -- Generated key: "slar_prod_abc123xyz456def789"
    api_key_hash TEXT NOT NULL,            -- Hashed version for security
    permissions TEXT[] DEFAULT '{"create_alerts"}', -- ["create_alerts", "read_alerts", "manage_oncall"]
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,                  -- Optional expiration
    
    -- Rate limiting
    rate_limit_per_hour INTEGER DEFAULT 1000,
    rate_limit_per_day INTEGER DEFAULT 10000,
    
    -- Usage tracking
    total_requests INTEGER DEFAULT 0,
    total_alerts_created INTEGER DEFAULT 0,
    
    -- Metadata
    description TEXT DEFAULT '',
    environment TEXT DEFAULT 'prod',       -- prod, dev, test
    created_by TEXT REFERENCES users(id),
    
    CONSTRAINT valid_environment CHECK (environment IN ('prod', 'dev', 'test')),
    CONSTRAINT valid_permissions CHECK (
        permissions <@ ARRAY['create_alerts', 'read_alerts', 'manage_oncall', 'view_dashboard', 'manage_services']
    )
);

-- API Key Usage Logs table
CREATE TABLE api_key_usage_logs (
    id TEXT PRIMARY KEY DEFAULT 'log_' || substr(md5(random()::text), 1, 16),
    api_key_id TEXT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,               -- "/alert/webhook"
    method TEXT NOT NULL,                 -- "POST"
    ip_address TEXT,
    user_agent TEXT,
    request_size INTEGER DEFAULT 0,
    response_status INTEGER,
    response_time_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Alert specific
    alert_id TEXT,                        -- If alert was created
    alert_title TEXT,
    alert_severity TEXT,
    
    -- Request metadata
    request_id TEXT,
    error_message TEXT,
    
    CONSTRAINT valid_method CHECK (method IN ('GET', 'POST', 'PUT', 'DELETE', 'PATCH')),
    CONSTRAINT valid_status CHECK (response_status >= 100 AND response_status < 600)
);

-- Rate limiting tracking table
CREATE TABLE api_key_rate_limits (
    id TEXT PRIMARY KEY DEFAULT 'rate_' || substr(md5(random()::text), 1, 16),
    api_key_id TEXT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    window_start TIMESTAMP NOT NULL,      -- Start of rate limit window
    window_type TEXT NOT NULL,            -- 'hour' or 'day'
    request_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_window_type CHECK (window_type IN ('hour', 'day')),
    UNIQUE(api_key_id, window_start, window_type)
);

-- Indexes for performance
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(api_key_hash);
CREATE INDEX idx_api_keys_active ON api_keys(is_active) WHERE is_active = true;
CREATE INDEX idx_api_keys_environment ON api_keys(environment);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at) WHERE expires_at IS NOT NULL;

CREATE INDEX idx_usage_logs_api_key ON api_key_usage_logs(api_key_id);
CREATE INDEX idx_usage_logs_created_at ON api_key_usage_logs(created_at);
CREATE INDEX idx_usage_logs_endpoint ON api_key_usage_logs(endpoint);
CREATE INDEX idx_usage_logs_status ON api_key_usage_logs(response_status);

CREATE INDEX idx_rate_limits_api_key ON api_key_rate_limits(api_key_id);
CREATE INDEX idx_rate_limits_window ON api_key_rate_limits(window_start, window_type);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_api_keys_updated_at 
    BEFORE UPDATE ON api_keys 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rate_limits_updated_at 
    BEFORE UPDATE ON api_key_rate_limits 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample API key for admin user (for testing)
INSERT INTO api_keys (
    user_id, 
    name, 
    api_key, 
    api_key_hash,
    description,
    environment,
    permissions,
    rate_limit_per_hour,
    rate_limit_per_day
) VALUES (
    'admin-user-id-001',
    'Admin Test Key',
    'slar_prod_admin123test456key789',
    '$2a$10$example_hash_for_admin_key',  -- This will be properly hashed in code
    'Admin API key for testing webhook alerts',
    'prod',
    ARRAY['create_alerts', 'read_alerts', 'manage_oncall', 'view_dashboard'],
    2000,
    20000
) ON CONFLICT (user_id) DO NOTHING;

-- Add comments for documentation
COMMENT ON TABLE api_keys IS 'API keys for webhook and external system authentication';
COMMENT ON TABLE api_key_usage_logs IS 'Logs of API key usage for analytics and monitoring';
COMMENT ON TABLE api_key_rate_limits IS 'Rate limiting tracking for API keys';

COMMENT ON COLUMN api_keys.api_key IS 'Plain text API key (shown only once during creation)';
COMMENT ON COLUMN api_keys.api_key_hash IS 'Bcrypt hash of the API key for secure storage';
COMMENT ON COLUMN api_keys.permissions IS 'Array of permissions granted to this API key';
COMMENT ON COLUMN api_keys.environment IS 'Environment this key is intended for (prod/dev/test)';

-- Create view for API key statistics
CREATE VIEW api_key_stats AS
SELECT 
    ak.id,
    ak.name,
    ak.user_id,
    u.name as user_name,
    u.email as user_email,
    ak.environment,
    ak.is_active,
    ak.created_at,
    ak.last_used_at,
    ak.total_requests,
    ak.total_alerts_created,
    ak.rate_limit_per_hour,
    ak.rate_limit_per_day,
    
    -- Usage stats from logs (last 24 hours)
    COALESCE(recent_stats.requests_24h, 0) as requests_last_24h,
    COALESCE(recent_stats.alerts_24h, 0) as alerts_last_24h,
    COALESCE(recent_stats.errors_24h, 0) as errors_last_24h,
    COALESCE(recent_stats.avg_response_time, 0) as avg_response_time_ms,
    
    -- Rate limit status
    CASE 
        WHEN ak.expires_at IS NOT NULL AND ak.expires_at < NOW() THEN 'expired'
        WHEN NOT ak.is_active THEN 'disabled'
        ELSE 'active'
    END as status
    
FROM api_keys ak
LEFT JOIN users u ON ak.user_id = u.id
LEFT JOIN (
    SELECT 
        api_key_id,
        COUNT(*) as requests_24h,
        COUNT(*) FILTER (WHERE alert_id IS NOT NULL) as alerts_24h,
        COUNT(*) FILTER (WHERE response_status >= 400) as errors_24h,
        AVG(response_time_ms) as avg_response_time
    FROM api_key_usage_logs 
    WHERE created_at > NOW() - INTERVAL '24 hours'
    GROUP BY api_key_id
) recent_stats ON ak.id = recent_stats.api_key_id;

COMMENT ON VIEW api_key_stats IS 'Comprehensive statistics view for API keys including usage metrics'; 