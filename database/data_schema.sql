-- ============================================
-- Data Import Schema
-- Production, Dore, PBR, OPEX, CAPEX, Revenue
-- ============================================

-- Production Data
CREATE TABLE production_data (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES mining_companies(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    mineral_id INT NOT NULL REFERENCES minerals(id),
    quantity DECIMAL(15,3) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    data_type VARCHAR(20) NOT NULL DEFAULT 'actual',
    version INT NOT NULL DEFAULT 1,
    description TEXT DEFAULT '',
    deleted_at TIMESTAMP,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Dore Data
CREATE TABLE dore_data (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES mining_companies(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    -- Production
    dore_produced_oz DECIMAL(15,3) NOT NULL,
    silver_grade_pct DECIMAL(5,2) NOT NULL,
    gold_grade_pct DECIMAL(5,2) NOT NULL,
    -- Prices
    pbr_price_silver DECIMAL(10,2) NOT NULL,
    pbr_price_gold DECIMAL(10,2) NOT NULL,
    realized_price_silver DECIMAL(10,2) NOT NULL,
    realized_price_gold DECIMAL(10,2) NOT NULL,
    -- Adjustments
    silver_adjustment_oz DECIMAL(15,3) DEFAULT 0,
    gold_adjustment_oz DECIMAL(15,3) DEFAULT 0,
    -- Deductions
    ag_deductions_pct DECIMAL(5,2) NOT NULL,
    au_deductions_pct DECIMAL(5,2) NOT NULL,
    -- Charges
    treatment_charge DECIMAL(15,2) NOT NULL,
    refining_deductions_au DECIMAL(15,2) NOT NULL,
    -- Metadata
    data_type VARCHAR(20) NOT NULL DEFAULT 'actual',
    version INT NOT NULL DEFAULT 1,
    description TEXT DEFAULT '',
    deleted_at TIMESTAMP,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- PBR Data (Plan Beneficio Regional)
CREATE TABLE pbr_data (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES mining_companies(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    -- Mining
    ore_mined_t DECIMAL(15,3) NOT NULL,
    waste_mined_t DECIMAL(15,3) NOT NULL,
    developments_m DECIMAL(10,2) NOT NULL,
    -- Processing
    total_tonnes_processed DECIMAL(15,3) NOT NULL,
    feed_grade_silver_gpt DECIMAL(10,3) NOT NULL,
    feed_grade_gold_gpt DECIMAL(10,3) NOT NULL,
    recovery_rate_silver_pct DECIMAL(5,2) NOT NULL,
    recovery_rate_gold_pct DECIMAL(5,2) NOT NULL,
    -- Metadata
    data_type VARCHAR(20) NOT NULL DEFAULT 'actual',
    version INT NOT NULL DEFAULT 1,
    description TEXT DEFAULT '',
    deleted_at TIMESTAMP,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- OPEX Data
CREATE TABLE opex_data (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES mining_companies(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    cost_center VARCHAR(100) NOT NULL,
    subcategory VARCHAR(100) NOT NULL,
    expense_type VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    data_type VARCHAR(20) NOT NULL DEFAULT 'actual',
    version INT NOT NULL DEFAULT 1,
    description TEXT DEFAULT '',
    deleted_at TIMESTAMP,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- CAPEX Data
CREATE TABLE capex_data (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES mining_companies(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    category VARCHAR(100) NOT NULL,
    car_number VARCHAR(50),
    project_name TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    data_type VARCHAR(20) NOT NULL DEFAULT 'actual',
    version INT NOT NULL DEFAULT 1,
    description TEXT DEFAULT '',
    deleted_at TIMESTAMP,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Revenue Data
CREATE TABLE revenue_data (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES mining_companies(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    mineral_id INT NOT NULL REFERENCES minerals(id),
    quantity_sold DECIMAL(15,3) NOT NULL,
    unit_price DECIMAL(15,2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    data_type VARCHAR(20) NOT NULL DEFAULT 'actual',
    version INT NOT NULL DEFAULT 1,
    description TEXT DEFAULT '',
    deleted_at TIMESTAMP,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Financial Data (Shipping, Sales Taxes, Royalties)
CREATE TABLE financial_data (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES mining_companies(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    shipping_selling DECIMAL(15,2) DEFAULT 0,
    sales_taxes_royalties DECIMAL(15,2) DEFAULT 0,
    other_adjustments DECIMAL(15,2) DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    data_type VARCHAR(20) NOT NULL DEFAULT 'actual',
    version INT NOT NULL DEFAULT 1,
    description TEXT DEFAULT '',
    deleted_at TIMESTAMP,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_production_data_company ON production_data(company_id);
CREATE INDEX idx_production_data_date ON production_data(date);
CREATE INDEX idx_production_data_mineral ON production_data(mineral_id);
CREATE INDEX idx_production_data_type ON production_data(data_type);
CREATE INDEX idx_production_data_deleted ON production_data(deleted_at);
CREATE INDEX idx_production_data_company_date_type ON production_data(company_id, date, data_type) WHERE deleted_at IS NULL;

CREATE INDEX idx_dore_data_company ON dore_data(company_id);
CREATE INDEX idx_dore_data_date ON dore_data(date);
CREATE INDEX idx_dore_data_type ON dore_data(data_type);
CREATE INDEX idx_dore_data_deleted ON dore_data(deleted_at);
CREATE INDEX idx_dore_data_company_date_type ON dore_data(company_id, date, data_type) WHERE deleted_at IS NULL;

CREATE INDEX idx_pbr_data_company ON pbr_data(company_id);
CREATE INDEX idx_pbr_data_date ON pbr_data(date);
CREATE INDEX idx_pbr_data_type ON pbr_data(data_type);
CREATE INDEX idx_pbr_data_deleted ON pbr_data(deleted_at);
CREATE INDEX idx_pbr_data_company_date_type ON pbr_data(company_id, date, data_type) WHERE deleted_at IS NULL;

CREATE INDEX idx_opex_data_company ON opex_data(company_id);
CREATE INDEX idx_opex_data_date ON opex_data(date);
CREATE INDEX idx_opex_data_cost_center ON opex_data(cost_center);
CREATE INDEX idx_opex_data_type ON opex_data(data_type);
CREATE INDEX idx_opex_data_deleted ON opex_data(deleted_at);
CREATE INDEX idx_opex_data_company_date_type ON opex_data(company_id, date, data_type) WHERE deleted_at IS NULL;

CREATE INDEX idx_capex_data_company ON capex_data(company_id);
CREATE INDEX idx_capex_data_date ON capex_data(date);
CREATE INDEX idx_capex_data_category ON capex_data(category);
CREATE INDEX idx_capex_data_type ON capex_data(data_type);
CREATE INDEX idx_capex_data_deleted ON capex_data(deleted_at);
CREATE INDEX idx_capex_data_company_date_type ON capex_data(company_id, date, data_type) WHERE deleted_at IS NULL;

CREATE INDEX idx_revenue_data_company ON revenue_data(company_id);
CREATE INDEX idx_revenue_data_date ON revenue_data(date);
CREATE INDEX idx_revenue_data_mineral ON revenue_data(mineral_id);
CREATE INDEX idx_revenue_data_type ON revenue_data(data_type);
CREATE INDEX idx_revenue_data_deleted ON revenue_data(deleted_at);
CREATE INDEX idx_revenue_data_company_date_type ON revenue_data(company_id, date, data_type) WHERE deleted_at IS NULL;

CREATE INDEX idx_financial_data_company ON financial_data(company_id);
CREATE INDEX idx_financial_data_date ON financial_data(date);
CREATE INDEX idx_financial_data_type ON financial_data(data_type);
CREATE INDEX idx_financial_data_deleted ON financial_data(deleted_at);
CREATE INDEX idx_financial_data_company_date_type ON financial_data(company_id, date, data_type) WHERE deleted_at IS NULL;

