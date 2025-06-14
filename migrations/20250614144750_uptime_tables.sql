-- Migration: Create uptime monitoring tables
-- Created: 2025-01-14

-- Services table - stores monitored services
CREATE TABLE IF NOT EXISTS services (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'http', -- http, https, tcp, ping
    method VARCHAR(10) NOT NULL DEFAULT 'GET', -- GET, POST, HEAD
    interval_seconds INTEGER NOT NULL DEFAULT 300, -- Check interval in seconds (5 minutes)
    timeout_seconds INTEGER NOT NULL DEFAULT 30, -- Timeout in seconds
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Expected response validation
    expected_status INTEGER DEFAULT 200,
    expected_body TEXT,
    
    -- HTTP headers (stored as JSON)
    headers JSONB
);

-- Service checks table - stores individual check results
CREATE TABLE IF NOT EXISTS service_checks (
    id VARCHAR(36) PRIMARY KEY,
    service_id VARCHAR(36) NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL, -- up, down, timeout, error
    response_time_ms INTEGER NOT NULL DEFAULT 0, -- Response time in milliseconds
    status_code INTEGER,
    response_body TEXT,
    error_message TEXT,
    checked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- SSL Certificate information (for HTTPS services)
    ssl_expiry TIMESTAMP,
    ssl_issuer VARCHAR(255),
    ssl_days_left INTEGER
);

-- Uptime statistics table - aggregated stats for different time periods
CREATE TABLE IF NOT EXISTS uptime_stats (
    id VARCHAR(36) PRIMARY KEY,
    service_id VARCHAR(36) NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    period VARCHAR(10) NOT NULL, -- 1h, 24h, 7d, 30d
    uptime_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    total_checks INTEGER NOT NULL DEFAULT 0,
    successful_checks INTEGER NOT NULL DEFAULT 0,
    failed_checks INTEGER NOT NULL DEFAULT 0,
    avg_response_time_ms DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    min_response_time_ms INTEGER NOT NULL DEFAULT 0,
    max_response_time_ms INTEGER NOT NULL DEFAULT 0,
    last_updated TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(service_id, period)
);

-- Service incidents table - tracks downtime incidents
CREATE TABLE IF NOT EXISTS service_incidents (
    id VARCHAR(36) PRIMARY KEY,
    service_id VARCHAR(36) NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL, -- downtime, slow_response, ssl_expiry
    status VARCHAR(20) NOT NULL DEFAULT 'ongoing', -- ongoing, resolved
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP,
    duration_seconds INTEGER, -- Duration in seconds (calculated when resolved)
    description TEXT NOT NULL,
    alert_id VARCHAR(36) -- Reference to alerts table (optional)
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_service_checks_service_id ON service_checks(service_id);
CREATE INDEX IF NOT EXISTS idx_service_checks_checked_at ON service_checks(checked_at);
CREATE INDEX IF NOT EXISTS idx_service_checks_status ON service_checks(status);
CREATE INDEX IF NOT EXISTS idx_uptime_stats_service_period ON uptime_stats(service_id, period);
CREATE INDEX IF NOT EXISTS idx_service_incidents_service_id ON service_incidents(service_id);
CREATE INDEX IF NOT EXISTS idx_service_incidents_status ON service_incidents(status);

-- Insert sample service for testing
INSERT INTO services (id, name, url, type, method, interval_seconds, timeout_seconds, expected_status) 
VALUES (
    'service-001', 
    'Website', 
    'https://soundworks-ai.com', 
    'https', 
    'GET', 
    300, -- Check every 5 minutes
    30,  -- 30 second timeout
    200  -- Expect HTTP 200
) ON CONFLICT (id) DO NOTHING;

-- Initialize stats records for the sample service
INSERT INTO uptime_stats (id, service_id, period, uptime_percentage, total_checks, successful_checks, failed_checks)
VALUES 
    ('stats-001-1h', 'service-001', '1h', 100.00, 0, 0, 0),
    ('stats-001-24h', 'service-001', '24h', 100.00, 0, 0, 0),
    ('stats-001-7d', 'service-001', '7d', 100.00, 0, 0, 0),
    ('stats-001-30d', 'service-001', '30d', 100.00, 0, 0, 0)
ON CONFLICT (service_id, period) DO NOTHING; 