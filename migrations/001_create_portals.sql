-- Create portals table
CREATE TABLE IF NOT EXISTS portals (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address_street VARCHAR(255) NOT NULL,
    address_zipcode VARCHAR(10) NOT NULL,
    address_city VARCHAR(100) NOT NULL,
    contractor_company VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(20) NOT NULL,
    contact_email VARCHAR(255),
    installation_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for searching by company and city
CREATE INDEX idx_portals_contractor ON portals(contractor_company);
CREATE INDEX idx_portals_city ON portals(address_city);