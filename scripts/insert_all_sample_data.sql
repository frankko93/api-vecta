-- ============================================
-- Complete sample data for all 12 months
-- PBR, Financial, OPEX, CAPEX, Dore, Production, Revenue
-- ============================================

-- Ensure company exists
INSERT INTO mining_companies (id, name, code, country, active) 
VALUES (1, 'Cerro Moro', 'CM', 'Argentina', true) 
ON CONFLICT (id) DO NOTHING;

-- Ensure user exists
INSERT INTO users (id, username, email, password_hash, role, active)
VALUES (1, 'admin', 'admin@cerromoro.com', '$2a$10$dummy', 'admin', true)
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- PBR DATA - 12 months Budget + Actual
-- ============================================
DELETE FROM pbr_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025;

DO $$
DECLARE
    m INTEGER;
    base_date DATE;
    vf DECIMAL; -- variance factor
    -- Base values
    open_pit_ore DECIMAL := 12880;
    underground_ore DECIMAL := 11979;
    total_ore DECIMAL := 24859;
    waste DECIMAL := 262591;
    stripping DECIMAL := 20.39;
    -- Grades
    mining_ag DECIMAL := 205.51;
    mining_au DECIMAL := 7.46;
    op_ag DECIMAL := 237.28;
    ug_ag DECIMAL := 171.35;
    op_au DECIMAL := 9.11;
    ug_au DECIMAL := 5.68;
    -- Developments
    primary_dev DECIMAL := 468;
    secondary_dev DECIMAL := 130;
    expansionary_dev DECIMAL := 0;
    total_dev DECIMAL := 598;
    -- Headcount
    fte INT := 833;
    contractors_count INT := 562;
    total_hc INT := 1395;
BEGIN
    FOR m IN 1..12 LOOP
        base_date := DATE '2025-01-01' + ((m-1) * INTERVAL '1 month');
        vf := CASE WHEN m IN (1,4,7,10) THEN 1.0 WHEN m IN (2,5,8,11) THEN 1.03 ELSE 0.97 END;
        
        -- Budget
        INSERT INTO pbr_data (
            company_id, date,
            open_pit_ore_t, underground_ore_t, ore_mined_t,
            waste_mined_t, stripping_ratio,
            mining_grade_silver_gpt, mining_grade_gold_gpt,
            open_pit_grade_silver_gpt, underground_grade_silver_gpt,
            open_pit_grade_gold_gpt, underground_grade_gold_gpt,
            primary_development_m, secondary_development_opex_m, expansionary_development_m, developments_m,
            total_tonnes_processed, feed_grade_silver_gpt, feed_grade_gold_gpt,
            recovery_rate_silver_pct, recovery_rate_gold_pct,
            full_time_employees, contractors, total_headcount,
            data_type, created_by
        ) VALUES (
            1, base_date,
            open_pit_ore, underground_ore, total_ore,
            waste, stripping,
            mining_ag, mining_au,
            op_ag, ug_ag, op_au, ug_au,
            primary_dev, secondary_dev, expansionary_dev, total_dev,
            42000, 180.5, 2.8, 88.5, 92.3,
            fte, contractors_count, total_hc,
            'budget', 1
        );
        
        -- Actual (with variance)
        INSERT INTO pbr_data (
            company_id, date,
            open_pit_ore_t, underground_ore_t, ore_mined_t,
            waste_mined_t, stripping_ratio,
            mining_grade_silver_gpt, mining_grade_gold_gpt,
            open_pit_grade_silver_gpt, underground_grade_silver_gpt,
            open_pit_grade_gold_gpt, underground_grade_gold_gpt,
            primary_development_m, secondary_development_opex_m, expansionary_development_m, developments_m,
            total_tonnes_processed, feed_grade_silver_gpt, feed_grade_gold_gpt,
            recovery_rate_silver_pct, recovery_rate_gold_pct,
            full_time_employees, contractors, total_headcount,
            data_type, created_by
        ) VALUES (
            1, base_date,
            open_pit_ore * vf, underground_ore * vf, total_ore * vf,
            waste * vf, stripping * vf,
            mining_ag * vf, mining_au * vf,
            op_ag * vf, ug_ag * vf, op_au * vf, ug_au * vf,
            primary_dev * vf, secondary_dev * vf, expansionary_dev * vf, total_dev * vf,
            42000 * vf, 180.5 * vf, 2.8 * vf, 88.5, 92.3,
            ROUND(fte * vf)::INT, ROUND(contractors_count * vf)::INT, ROUND(total_hc * vf)::INT,
            'actual', 1
        );
    END LOOP;
END $$;

-- ============================================
-- FINANCIAL DATA - 12 months Budget + Actual
-- ============================================
DELETE FROM financial_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025;

DO $$
DECLARE
    m INTEGER;
    base_date DATE;
    vf DECIMAL;
BEGIN
    FOR m IN 1..12 LOOP
        base_date := DATE '2025-01-01' + ((m-1) * INTERVAL '1 month');
        vf := CASE WHEN m IN (1,4,7,10) THEN 1.0 WHEN m IN (2,5,8,11) THEN 1.02 ELSE 0.98 END;
        
        -- Budget
        INSERT INTO financial_data (company_id, date, shipping_selling, sales_taxes_royalties, 
            other_adjustments, currency, data_type, created_by)
        VALUES (1, base_date, 150000, 280000, 25000, 'USD', 'budget', 1);
        
        -- Actual
        INSERT INTO financial_data (company_id, date, shipping_selling, sales_taxes_royalties, 
            other_adjustments, currency, data_type, created_by)
        VALUES (1, base_date, 150000*vf, 280000*vf, 25000*vf, 'USD', 'actual', 1);
    END LOOP;
END $$;

-- ============================================
-- OPEX DATA - 12 months Budget + Actual
-- ============================================
DELETE FROM opex_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025;

DO $$
DECLARE
    m INTEGER;
    base_date DATE;
    vf DECIMAL;
BEGIN
    FOR m IN 1..12 LOOP
        base_date := DATE '2025-01-01' + ((m-1) * INTERVAL '1 month');
        vf := CASE WHEN m IN (1,4,7,10) THEN 1.0 WHEN m IN (2,5,8,11) THEN 1.04 ELSE 0.96 END;
        
        -- Budget - Mine cost center
        INSERT INTO opex_data (company_id, date, cost_center, subcategory, expense_type, amount, currency, data_type, created_by) VALUES
        (1, base_date, 'Mine', 'Drilling', 'Labour', 1536353, 'USD', 'budget', 1),
        (1, base_date, 'Mine', 'Blasting', 'Labour', 1027840, 'USD', 'budget', 1),
        (1, base_date, 'Mine', 'Loading', 'Labour', 620478, 'USD', 'budget', 1),
        (1, base_date, 'Mine', 'Hauling', 'Labour', 999752, 'USD', 'budget', 1),
        (1, base_date, 'Mine', 'Ground Support', 'Labour', 826390, 'USD', 'budget', 1),
        (1, base_date, 'Mine', 'Mine Services', 'Materials', 405349, 'USD', 'budget', 1),
        (1, base_date, 'Mine', 'Mine Geology', 'Labour', 373433, 'USD', 'budget', 1),
        (1, base_date, 'Mine', 'Mine Engineering', 'Labour', 396466, 'USD', 'budget', 1),
        (1, base_date, 'Mine', 'Mine Maintenance', 'Materials', 968061, 'USD', 'budget', 1),
        (1, base_date, 'Mine', 'General Operating', 'Other', 1383876, 'USD', 'budget', 1),
        (1, base_date, 'Mine', 'Stockpile/WIP', 'Other', 1740162, 'USD', 'budget', 1),
        -- Budget - Processing cost center
        (1, base_date, 'Processing', 'CO General Operating', 'Labour', 338318, 'USD', 'budget', 1),
        (1, base_date, 'Processing', 'CO Primary Crushing', 'Labour', 338119, 'USD', 'budget', 1),
        (1, base_date, 'Processing', 'CO Grinding and Classifying', 'Labour', 386865, 'USD', 'budget', 1),
        (1, base_date, 'Processing', 'CO Regrinding and Flotation', 'Labour', 256661, 'USD', 'budget', 1),
        (1, base_date, 'Processing', 'CO Thickening and Filtering', 'Labour', 116097, 'USD', 'budget', 1),
        (1, base_date, 'Processing', 'CO Tailing Disposal', 'Materials', 179333, 'USD', 'budget', 1),
        (1, base_date, 'Processing', 'CO Sampling and Assaying', 'Materials', 456318, 'USD', 'budget', 1),
        (1, base_date, 'Processing', 'CO Plant Maintenance', 'Materials', 744675, 'USD', 'budget', 1),
        (1, base_date, 'Processing', 'PR Leaching', 'Materials', 445046, 'USD', 'budget', 1),
        (1, base_date, 'Processing', 'PR Refining', 'Materials', 352245, 'USD', 'budget', 1),
        -- Budget - G&A cost center
        (1, base_date, 'G&A', 'General Administration', 'Third Party', 568217, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Maintenance Shops (Overhead)', 'Labour', 241787, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Warehouse', 'Labour', 374376, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Purchasing', 'Labour', 207287, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Safety', 'Labour', 492813, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Security', 'Third Party', 333262, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Legal', 'Third Party', 1638, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Environmental Services', 'Labour', 150525, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Camp', 'Third Party', 1460458, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Public Community Relations', 'Other', 346944, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Human Resources', 'Labour', 106843, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'New Business / Project Development', 'Third Party', 797554, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Financings & Cost', 'Other', 110638, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Information Systems', 'Materials', 105692, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Contract Administration', 'Labour', 23862, 'USD', 'budget', 1),
        (1, base_date, 'G&A', 'Other Indirect', 'Other', 149824, 'USD', 'budget', 1);

        -- Actual - Mine cost center (with variance)
        INSERT INTO opex_data (company_id, date, cost_center, subcategory, expense_type, amount, currency, data_type, created_by) VALUES
        (1, base_date, 'Mine', 'Drilling', 'Labour', 1536353*vf, 'USD', 'actual', 1),
        (1, base_date, 'Mine', 'Blasting', 'Labour', 1027840*vf, 'USD', 'actual', 1),
        (1, base_date, 'Mine', 'Loading', 'Labour', 620478*vf, 'USD', 'actual', 1),
        (1, base_date, 'Mine', 'Hauling', 'Labour', 999752*vf, 'USD', 'actual', 1),
        (1, base_date, 'Mine', 'Ground Support', 'Labour', 826390*vf, 'USD', 'actual', 1),
        (1, base_date, 'Mine', 'Mine Services', 'Materials', 405349*vf, 'USD', 'actual', 1),
        (1, base_date, 'Mine', 'Mine Geology', 'Labour', 373433*vf, 'USD', 'actual', 1),
        (1, base_date, 'Mine', 'Mine Engineering', 'Labour', 396466*vf, 'USD', 'actual', 1),
        (1, base_date, 'Mine', 'Mine Maintenance', 'Materials', 968061*vf, 'USD', 'actual', 1),
        (1, base_date, 'Mine', 'General Operating', 'Other', 1383876*vf, 'USD', 'actual', 1),
        (1, base_date, 'Mine', 'Stockpile/WIP', 'Other', 1740162*vf, 'USD', 'actual', 1),
        -- Actual - Processing cost center
        (1, base_date, 'Processing', 'CO General Operating', 'Labour', 338318*vf, 'USD', 'actual', 1),
        (1, base_date, 'Processing', 'CO Primary Crushing', 'Labour', 338119*vf, 'USD', 'actual', 1),
        (1, base_date, 'Processing', 'CO Grinding and Classifying', 'Labour', 386865*vf, 'USD', 'actual', 1),
        (1, base_date, 'Processing', 'CO Regrinding and Flotation', 'Labour', 256661*vf, 'USD', 'actual', 1),
        (1, base_date, 'Processing', 'CO Thickening and Filtering', 'Labour', 116097*vf, 'USD', 'actual', 1),
        (1, base_date, 'Processing', 'CO Tailing Disposal', 'Materials', 179333*vf, 'USD', 'actual', 1),
        (1, base_date, 'Processing', 'CO Sampling and Assaying', 'Materials', 456318*vf, 'USD', 'actual', 1),
        (1, base_date, 'Processing', 'CO Plant Maintenance', 'Materials', 744675*vf, 'USD', 'actual', 1),
        (1, base_date, 'Processing', 'PR Leaching', 'Materials', 445046*vf, 'USD', 'actual', 1),
        (1, base_date, 'Processing', 'PR Refining', 'Materials', 352245*vf, 'USD', 'actual', 1),
        -- Actual - G&A cost center
        (1, base_date, 'G&A', 'General Administration', 'Third Party', 568217*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Maintenance Shops (Overhead)', 'Labour', 241787*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Warehouse', 'Labour', 374376*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Purchasing', 'Labour', 207287*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Safety', 'Labour', 492813*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Security', 'Third Party', 333262*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Legal', 'Third Party', 1638*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Environmental Services', 'Labour', 150525*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Camp', 'Third Party', 1460458*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Public Community Relations', 'Other', 346944*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Human Resources', 'Labour', 106843*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'New Business / Project Development', 'Third Party', 797554*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Financings & Cost', 'Other', 110638*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Information Systems', 'Materials', 105692*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Contract Administration', 'Labour', 23862*vf, 'USD', 'actual', 1),
        (1, base_date, 'G&A', 'Other Indirect', 'Other', 149824*vf, 'USD', 'actual', 1);
    END LOOP;
END $$;

-- ============================================
-- DORE DATA - 12 months Budget + Actual
-- ============================================
DELETE FROM dore_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025;

DO $$
DECLARE
    m INTEGER;
    base_date DATE;
    vf DECIMAL;
BEGIN
    FOR m IN 1..12 LOOP
        base_date := DATE '2025-01-01' + ((m-1) * INTERVAL '1 month');
        vf := CASE WHEN m IN (1,4,7,10) THEN 1.0 WHEN m IN (2,5,8,11) THEN 1.02 ELSE 0.98 END;
        
        -- Budget
        INSERT INTO dore_data (company_id, date, dore_produced_oz, silver_grade_pct, gold_grade_pct,
            pbr_price_silver, pbr_price_gold, realized_price_silver, realized_price_gold,
            silver_adjustment_oz, gold_adjustment_oz, ag_deductions_pct, au_deductions_pct,
            treatment_charge, refining_deductions_au, streaming, data_type, created_by)
        VALUES (1, base_date, 250000, 96.5, 3.5, 23.50, 1950.00, 24.00, 1980.00,
            1500, 55, 2.0, 1.5, 85000, 42000, -150000, 'budget', 1);
        
        -- Actual
        INSERT INTO dore_data (company_id, date, dore_produced_oz, silver_grade_pct, gold_grade_pct,
            pbr_price_silver, pbr_price_gold, realized_price_silver, realized_price_gold,
            silver_adjustment_oz, gold_adjustment_oz, ag_deductions_pct, au_deductions_pct,
            treatment_charge, refining_deductions_au, streaming, data_type, created_by)
        VALUES (1, base_date, 250000*vf, 96.5, 3.5, 23.50*vf, 1950.00*vf, 24.00*vf, 1980.00*vf,
            1500*vf, 55*vf, 2.0, 1.5, 85000*vf, 42000*vf, -150000*vf, 'actual', 1);
    END LOOP;
END $$;

-- ============================================
-- PRODUCTION DATA - 12 months Budget + Actual
-- ============================================
DELETE FROM production_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025;

-- Ensure minerals exist
INSERT INTO minerals (id, code, name, unit, active) VALUES 
(1, 'AG', 'Silver', 'oz', true),
(2, 'AU', 'Gold', 'oz', true)
ON CONFLICT (id) DO NOTHING;

DO $$
DECLARE
    m INTEGER;
    base_date DATE;
    vf DECIMAL;
BEGIN
    FOR m IN 1..12 LOOP
        base_date := DATE '2025-01-01' + ((m-1) * INTERVAL '1 month');
        vf := CASE WHEN m IN (1,4,7,10) THEN 1.0 WHEN m IN (2,5,8,11) THEN 1.03 ELSE 0.97 END;
        
        -- Budget - Silver
        INSERT INTO production_data (company_id, date, mineral_id, quantity, unit, data_type, created_by)
        VALUES (1, base_date, 1, 240000, 'oz', 'budget', 1);
        -- Budget - Gold
        INSERT INTO production_data (company_id, date, mineral_id, quantity, unit, data_type, created_by)
        VALUES (1, base_date, 2, 8750, 'oz', 'budget', 1);
        
        -- Actual - Silver
        INSERT INTO production_data (company_id, date, mineral_id, quantity, unit, data_type, created_by)
        VALUES (1, base_date, 1, 240000*vf, 'oz', 'actual', 1);
        -- Actual - Gold
        INSERT INTO production_data (company_id, date, mineral_id, quantity, unit, data_type, created_by)
        VALUES (1, base_date, 2, 8750*vf, 'oz', 'actual', 1);
    END LOOP;
END $$;

-- ============================================
-- REVENUE DATA - 12 months Budget + Actual
-- ============================================
DELETE FROM revenue_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025;

DO $$
DECLARE
    m INTEGER;
    base_date DATE;
    vf DECIMAL;
BEGIN
    FOR m IN 1..12 LOOP
        base_date := DATE '2025-01-01' + ((m-1) * INTERVAL '1 month');
        vf := CASE WHEN m IN (1,4,7,10) THEN 1.0 WHEN m IN (2,5,8,11) THEN 1.02 ELSE 0.98 END;
        
        -- Budget - Silver
        INSERT INTO revenue_data (company_id, date, mineral_id, quantity_sold, unit_price, currency, data_type, created_by)
        VALUES (1, base_date, 1, 235000, 24.00, 'USD', 'budget', 1);
        -- Budget - Gold
        INSERT INTO revenue_data (company_id, date, mineral_id, quantity_sold, unit_price, currency, data_type, created_by)
        VALUES (1, base_date, 2, 8500, 1980.00, 'USD', 'budget', 1);
        
        -- Actual - Silver
        INSERT INTO revenue_data (company_id, date, mineral_id, quantity_sold, unit_price, currency, data_type, created_by)
        VALUES (1, base_date, 1, 235000*vf, 24.00*vf, 'USD', 'actual', 1);
        -- Actual - Gold
        INSERT INTO revenue_data (company_id, date, mineral_id, quantity_sold, unit_price, currency, data_type, created_by)
        VALUES (1, base_date, 2, 8500*vf, 1980.00*vf, 'USD', 'actual', 1);
    END LOOP;
END $$;

-- ============================================
-- CAPEX DATA - 12 months Budget + Actual
-- ============================================
DELETE FROM capex_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025;

DO $$
DECLARE
    m INTEGER;
    base_date DATE;
    vf DECIMAL;
BEGIN
    FOR m IN 1..12 LOOP
        base_date := DATE '2025-01-01' + ((m-1) * INTERVAL '1 month');
        vf := CASE WHEN m IN (1,4,7,10) THEN 1.0 WHEN m IN (2,5,8,11) THEN 1.05 ELSE 0.95 END;
        
        -- Budget
        INSERT INTO capex_data (company_id, date, category, car_number, project_name, type, amount, accretion_of_mine_closure_liability, currency, data_type, created_by) VALUES
        (1, base_date, 'Exploration/Mine Geology', 'C487EY21001', 'CAPEX EXPLORACIONES', 'sustaining', 700000 + (m * 5000), 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25001', '', 'sustaining', 5000 + (m * 500), 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25002', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25003', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25004', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25005', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25006', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25007', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25008', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25009', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25010', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Plant Upgrades', 'C487PY25001', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Administration Projects', 'C487AY25001', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Administration Projects', 'C487AY25002', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Administration Projects', 'C487AY25003', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Site Infrastructure', 'C487AY24001', '', 'sustaining', -3261, 0, 'USD', 'budget', 1),
        (1, base_date, 'Site Infrastructure', 'C487AY24005', '', 'sustaining', 3915, 0, 'USD', 'budget', 1),
        (1, base_date, 'Site Infrastructure', 'C487AY24003', '', 'sustaining', 3261, 0, 'USD', 'budget', 1),
        (1, base_date, 'Administration Projects', 'C48703300', '', 'sustaining', 2970, 0, 'USD', 'budget', 1),
        (1, base_date, 'Pre-Stripping and Capital Developments', 'PRE-001', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Mine Infrastructure', 'INF-001', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Tailings Dams and Leach Pads', 'TAIL-001', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Community Projects', 'COM-001', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Right-of-Use Asset (IFRS16)', 'IFRS16-001', '', 'leasing', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Total MPPE Additions', 'MPPE-TOTAL', '', 'sustaining', 700000 + (m * 5000) + 5000 + (m * 500) + 6885, 0, 'USD', 'budget', 1),
        (1, base_date, 'Project Capital', 'PROJ-CAP', '', 'project', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Leasing Addition - Project Capital', 'LEASE-PROJ', '', 'leasing', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Other', 'OTHER-001', '', 'sustaining', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Sustaining MPPE Additions', 'MPPE-SUST', '', 'sustaining', 700000 + (m * 5000) + 5000 + (m * 500) + 6885, 0, 'USD', 'budget', 1),
        (1, base_date, 'Leasing Addition - Sustaining Capital', 'LEASE-SUST', '', 'leasing', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'Sustaining Capital Lease Cash Outflows', 'LEASE-CASH', '', 'leasing', 0, 0, 'USD', 'budget', 1),
        (1, base_date, 'IFRS16', 'IFRS16-SUM', '', 'leasing', 0, 0, 'USD', 'budget', 1);

        -- Actual (with variance)
        INSERT INTO capex_data (company_id, date, category, car_number, project_name, type, amount, accretion_of_mine_closure_liability, currency, data_type, created_by) VALUES
        (1, base_date, 'Exploration/Mine Geology', 'C487EY21001', 'CAPEX EXPLORACIONES', 'sustaining', (700000 + (m * 5000)) * vf, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25001', '', 'sustaining', (5000 + (m * 500)) * vf, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25002', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25003', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25004', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25005', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25006', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25007', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25008', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25009', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25010', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Plant Upgrades', 'C487PY25001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Administration Projects', 'C487AY25001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Administration Projects', 'C487AY25002', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Administration Projects', 'C487AY25003', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Site Infrastructure', 'C487AY24001', '', 'sustaining', -3261 * vf, 0, 'USD', 'actual', 1),
        (1, base_date, 'Site Infrastructure', 'C487AY24005', '', 'sustaining', 3915 * vf, 0, 'USD', 'actual', 1),
        (1, base_date, 'Site Infrastructure', 'C487AY24003', '', 'sustaining', 3261 * vf, 0, 'USD', 'actual', 1),
        (1, base_date, 'Administration Projects', 'C48703300', '', 'sustaining', 2970 * vf, 0, 'USD', 'actual', 1),
        (1, base_date, 'Pre-Stripping and Capital Developments', 'PRE-001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Infrastructure', 'INF-001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Tailings Dams and Leach Pads', 'TAIL-001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Community Projects', 'COM-001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Right-of-Use Asset (IFRS16)', 'IFRS16-001', '', 'leasing', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Total MPPE Additions', 'MPPE-TOTAL', '', 'sustaining', ((700000 + (m * 5000)) + (5000 + (m * 500)) + 6885) * vf, 0, 'USD', 'actual', 1),
        (1, base_date, 'Project Capital', 'PROJ-CAP', '', 'project', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Leasing Addition - Project Capital', 'LEASE-PROJ', '', 'leasing', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Other', 'OTHER-001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Sustaining MPPE Additions', 'MPPE-SUST', '', 'sustaining', ((700000 + (m * 5000)) + (5000 + (m * 500)) + 6885) * vf, 0, 'USD', 'actual', 1),
        (1, base_date, 'Leasing Addition - Sustaining Capital', 'LEASE-SUST', '', 'leasing', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Sustaining Capital Lease Cash Outflows', 'LEASE-CASH', '', 'leasing', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'IFRS16', 'IFRS16-SUM', '', 'leasing', 0, 0, 'USD', 'actual', 1);
    END LOOP;
END $$;

-- ============================================
-- VERIFICATION
-- ============================================
SELECT 'PBR' as table_name, data_type, COUNT(DISTINCT EXTRACT(MONTH FROM date)) as months 
FROM pbr_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025 GROUP BY data_type
UNION ALL
SELECT 'Financial', data_type, COUNT(DISTINCT EXTRACT(MONTH FROM date)) 
FROM financial_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025 GROUP BY data_type
UNION ALL
SELECT 'OPEX', data_type, COUNT(DISTINCT EXTRACT(MONTH FROM date)) 
FROM opex_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025 GROUP BY data_type
UNION ALL
SELECT 'CAPEX', data_type, COUNT(DISTINCT EXTRACT(MONTH FROM date)) 
FROM capex_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025 GROUP BY data_type
UNION ALL
SELECT 'Dore', data_type, COUNT(DISTINCT EXTRACT(MONTH FROM date)) 
FROM dore_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025 GROUP BY data_type
UNION ALL
SELECT 'Production', data_type, COUNT(DISTINCT EXTRACT(MONTH FROM date)) 
FROM production_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025 GROUP BY data_type
UNION ALL
SELECT 'Revenue', data_type, COUNT(DISTINCT EXTRACT(MONTH FROM date)) 
FROM revenue_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025 GROUP BY data_type
ORDER BY table_name, data_type;
