-- Add missing columns to alerts table if they don't exist

-- Add author column
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'alerts' AND column_name = 'author') THEN
        ALTER TABLE alerts ADD COLUMN author TEXT;
    END IF;
END $$;

-- Add assigned_to column with foreign key reference
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'alerts' AND column_name = 'assigned_to') THEN
        ALTER TABLE alerts ADD COLUMN assigned_to TEXT REFERENCES users(id);
    END IF;
END $$;

-- Add assigned_at column
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'alerts' AND column_name = 'assigned_at') THEN
        ALTER TABLE alerts ADD COLUMN assigned_at TIMESTAMP;
    END IF;
END $$;

-- Create indexes for better performance (only if they don't exist)
CREATE INDEX IF NOT EXISTS idx_alerts_assigned_to ON alerts(assigned_to);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_author ON alerts(author); 