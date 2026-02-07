-- Configuration Schema
-- Drop existing tables
DROP TABLE IF EXISTS company_minerals CASCADE;
DROP TABLE IF EXISTS company_settings CASCADE;
DROP TABLE IF EXISTS mining_companies CASCADE;
DROP TABLE IF EXISTS minerals CASCADE;

-- Mining Companies
CREATE TABLE mining_companies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50) UNIQUE NOT NULL,
    address TEXT,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Minerals
CREATE TABLE minerals (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    code VARCHAR(10) UNIQUE NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Company Minerals (Many-to-Many)
CREATE TABLE company_minerals (
    company_id BIGINT REFERENCES mining_companies(id) ON DELETE CASCADE,
    mineral_id INT REFERENCES minerals(id) ON DELETE CASCADE,
    PRIMARY KEY (company_id, mineral_id)
);

-- Company Settings
CREATE TABLE company_settings (
    company_id BIGINT PRIMARY KEY REFERENCES mining_companies(id) ON DELETE CASCADE,
    mining_type VARCHAR(50) DEFAULT 'underground',
    country VARCHAR(100),
    royalty_percentage DECIMAL(5,2) DEFAULT 0.00,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Indexes
CREATE INDEX idx_mining_companies_active ON mining_companies(active);
CREATE INDEX idx_minerals_active ON minerals(active);
CREATE INDEX idx_company_minerals_company ON company_minerals(company_id);
CREATE INDEX idx_company_minerals_mineral ON company_minerals(mineral_id);

-- Seed Minerals (7 principales)
INSERT INTO minerals (name, code, description) VALUES
('Oro', 'AU', 'Oro metálico'),
('Plata', 'AG', 'Plata metálica'),
('Cobre', 'CU', 'Cobre'),
('Zinc', 'ZN', 'Zinc'),
('Plomo', 'PB', 'Plomo'),
('Litio', 'LI', 'Litio'),
('Hierro', 'FE', 'Hierro');

-- Seed Mining Company: Cerro Moro (para datos de ejemplo)
INSERT INTO mining_companies (name, legal_name, tax_id, address, contact_email, contact_phone, active)
VALUES 
('Cerro Moro', 'Cerro Moro S.A.', '30-12345678-9', 'Argentina', 'info@cerromoro.com', '+54 261 123-4567', true)
ON CONFLICT (name) DO UPDATE 
SET contact_phone = COALESCE(EXCLUDED.contact_phone, mining_companies.contact_phone, '+54 261 123-4567');

-- Assign minerals AU and AG to Cerro Moro
INSERT INTO company_minerals (company_id, mineral_id)
SELECT mc.id, m.id
FROM mining_companies mc, minerals m
WHERE mc.name = 'Cerro Moro' AND m.code IN ('AU', 'AG')
ON CONFLICT DO NOTHING;

-- Seed company settings for Cerro Moro
INSERT INTO company_settings (company_id, mining_type, country, royalty_percentage, notes)
SELECT mc.id, 'both', 'Argentina', 3.5, 'Open Pit and Underground operations'
FROM mining_companies mc
WHERE mc.name = 'Cerro Moro'
ON CONFLICT (company_id) DO NOTHING;

-- Assign test admin user to Cerro Moro company
-- Note: user_companies table is defined in schema.sql (must be run first)
INSERT INTO user_companies (user_id, company_id, role)
SELECT u.id, mc.id, 'admin'
FROM users u, mining_companies mc
WHERE u.dni = '99999999' AND mc.name = 'Cerro Moro'
ON CONFLICT (user_id, company_id) DO NOTHING;
