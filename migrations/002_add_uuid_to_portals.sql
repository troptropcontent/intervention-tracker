-- Add UUID column to existing portals table
-- This migration handles existing records by generating UUIDs for them

-- Add UUID column (nullable initially since we have existing records)
ALTER TABLE portals ADD COLUMN uuid UUID;

-- Generate UUIDs for existing records
UPDATE portals SET uuid = gen_random_uuid() WHERE uuid IS NULL;

-- Now make the column NOT NULL and add unique constraint
ALTER TABLE portals ALTER COLUMN uuid SET NOT NULL;
ALTER TABLE portals ADD CONSTRAINT portals_uuid_unique UNIQUE (uuid);

-- Create index for efficient UUID lookups
CREATE INDEX idx_portals_uuid ON portals(uuid);

-- Optional: Add a comment explaining the UUID purpose
COMMENT ON COLUMN portals.uuid IS 'Unique identifier for QR code generation and public access';