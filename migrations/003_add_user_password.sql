-- Migration: Add password field to users table and create admin user
-- Created: $(date)

-- Add password column to users table
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'users' AND column_name = 'password_hash') THEN
        ALTER TABLE users ADD COLUMN password_hash TEXT;
    END IF;
END $$;

-- Create admin user with password
-- Password: admin123 (hashed with bcrypt)
INSERT INTO users (
    id, 
    name, 
    email, 
    phone, 
    role, 
    team, 
    fcm_token,
    password_hash,
    is_active, 
    created_at, 
    updated_at
) VALUES (
    'admin-user-id-001',
    'Admin User',
    'admin@slar.com',
    '+1234567890',
    'admin',
    'System Admin',
    '',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: admin123
    true,
    NOW(),
    NOW()
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    phone = EXCLUDED.phone,
    fcm_token = EXCLUDED.fcm_token,
    updated_at = NOW();

-- Create index for password lookups
CREATE INDEX IF NOT EXISTS idx_users_email_password ON users(email, password_hash); 