-- ============================================
-- Authentication & Authorization Schema
-- ============================================

-- Drop tables if exist (for clean setup)
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS user_companies CASCADE;
DROP TABLE IF EXISTS user_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    dni VARCHAR(20) UNIQUE NOT NULL,
    birth_date DATE NOT NULL,
    work_area VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Permissions catalog
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

-- User permissions (many-to-many)
CREATE TABLE user_permissions (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

-- Sessions table
-- company_roles stores the user's roles per company at login time (cached for performance)
-- Format: {"company_id": "role", ...} e.g. {"1": "admin", "2": "viewer"}
CREATE TABLE sessions (
    token VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_roles JSONB DEFAULT '{}' NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- User companies (many-to-many relationship with mining_companies)
-- Note: mining_companies table is defined in config_schema.sql
-- This table links users to the companies they have access to
CREATE TABLE user_companies (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id BIGINT NOT NULL, -- References mining_companies(id) from config_schema
    role VARCHAR(50) DEFAULT 'member' NOT NULL, -- e.g., 'admin', 'member', 'viewer'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, company_id)
);

-- Indexes for performance
CREATE INDEX idx_users_dni ON users(dni);
CREATE INDEX idx_users_active ON users(active);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX idx_user_companies_user_id ON user_companies(user_id);
CREATE INDEX idx_user_companies_company_id ON user_companies(company_id);

-- Seed permissions
INSERT INTO permissions (name, description) VALUES
('super_admin', 'Global admin - can manage users and companies across the entire system'),
('admin', 'Company admin - can manage users and data within assigned companies'),
('editor', 'Can create and edit data'),
('viewer', 'Read-only access to data');

-- Create test super admin user
-- DNI: 99999999, Password: admin123
INSERT INTO users (first_name, last_name, dni, birth_date, work_area, password_hash, active)
VALUES ('Super', 'Admin', '99999999', '1990-01-01', 'IT', '$argon2id$v=19$m=65536,t=1,p=11$26wRAe/3D66n2EZzzR0QNw$FLiJupf5T0vQCFLryzB2gWdrR4jLMX8sFVAfq2UbnwE', true);

-- Assign super_admin permission to test user
INSERT INTO user_permissions (user_id, permission_id)
SELECT u.id, p.id 
FROM users u, permissions p 
WHERE u.dni = '99999999' AND p.name = 'super_admin';

