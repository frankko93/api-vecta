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
