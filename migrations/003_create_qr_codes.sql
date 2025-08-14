-- Create qr_codes table for tracking pre-printed QR codes
CREATE TABLE IF NOT EXISTS qr_codes (
    id SERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    portal_id INTEGER REFERENCES portals(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'associated', 'damaged', 'lost')),
    associated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for efficient lookups
CREATE INDEX idx_qr_codes_uuid ON qr_codes(uuid);
CREATE INDEX idx_qr_codes_portal_id ON qr_codes(portal_id);
CREATE INDEX idx_qr_codes_status ON qr_codes(status);

-- Add comments
COMMENT ON TABLE qr_codes IS 'Tracks pre-printed QR codes and their association with portals';
COMMENT ON COLUMN qr_codes.uuid IS 'Unique identifier printed on the physical QR code';
COMMENT ON COLUMN qr_codes.portal_id IS 'Associated portal ID, NULL if not associated';
COMMENT ON COLUMN qr_codes.status IS 'Current status: available, associated, damaged, or lost';
COMMENT ON COLUMN qr_codes.associated_at IS 'Timestamp when QR code was associated with a portal';