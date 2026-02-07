-- Insert sample CAPEX data for testing
-- Assumes company_id=1 and user_id=1 exist

-- First, clear existing CAPEX data for 2025
DELETE FROM capex_data WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025;

-- =====================================================
-- Helper: Generate monthly data for all 12 months
-- Budget and Actual for each month
-- =====================================================

-- BUDGET - All 12 months
DO $$
DECLARE
    m INTEGER;
    base_date DATE;
BEGIN
    FOR m IN 1..12 LOOP
        base_date := DATE '2025-01-01' + ((m-1) * INTERVAL '1 month');
        
        -- Projects with varying amounts per month
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
        -- Required categories
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
    END LOOP;
END $$;

-- ACTUAL - All 12 months (with slight variations from budget)
DO $$
DECLARE
    m INTEGER;
    base_date DATE;
    variance_factor DECIMAL;
BEGIN
    FOR m IN 1..12 LOOP
        base_date := DATE '2025-01-01' + ((m-1) * INTERVAL '1 month');
        -- Varying the actuals: some months over, some under budget
        variance_factor := CASE 
            WHEN m IN (1, 4, 7, 10) THEN 1.0  -- on budget
            WHEN m IN (2, 5, 8, 11) THEN 1.05 -- 5% over
            ELSE 0.95 -- 5% under
        END;
        
        INSERT INTO capex_data (company_id, date, category, car_number, project_name, type, amount, accretion_of_mine_closure_liability, currency, data_type, created_by) VALUES
        (1, base_date, 'Exploration/Mine Geology', 'C487EY21001', 'CAPEX EXPLORACIONES', 'sustaining', (700000 + (m * 5000)) * variance_factor, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Equipment', 'C487MY25001', '', 'sustaining', (5000 + (m * 500)) * variance_factor, 0, 'USD', 'actual', 1),
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
        (1, base_date, 'Site Infrastructure', 'C487AY24001', '', 'sustaining', -3261 * variance_factor, 0, 'USD', 'actual', 1),
        (1, base_date, 'Site Infrastructure', 'C487AY24005', '', 'sustaining', 3915 * variance_factor, 0, 'USD', 'actual', 1),
        (1, base_date, 'Site Infrastructure', 'C487AY24003', '', 'sustaining', 3261 * variance_factor, 0, 'USD', 'actual', 1),
        (1, base_date, 'Administration Projects', 'C48703300', '', 'sustaining', 2970 * variance_factor, 0, 'USD', 'actual', 1),
        -- Required categories
        (1, base_date, 'Pre-Stripping and Capital Developments', 'PRE-001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Mine Infrastructure', 'INF-001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Tailings Dams and Leach Pads', 'TAIL-001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Community Projects', 'COM-001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Right-of-Use Asset (IFRS16)', 'IFRS16-001', '', 'leasing', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Total MPPE Additions', 'MPPE-TOTAL', '', 'sustaining', ((700000 + (m * 5000)) + (5000 + (m * 500)) + 6885) * variance_factor, 0, 'USD', 'actual', 1),
        (1, base_date, 'Project Capital', 'PROJ-CAP', '', 'project', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Leasing Addition - Project Capital', 'LEASE-PROJ', '', 'leasing', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Other', 'OTHER-001', '', 'sustaining', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Sustaining MPPE Additions', 'MPPE-SUST', '', 'sustaining', ((700000 + (m * 5000)) + (5000 + (m * 500)) + 6885) * variance_factor, 0, 'USD', 'actual', 1),
        (1, base_date, 'Leasing Addition - Sustaining Capital', 'LEASE-SUST', '', 'leasing', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'Sustaining Capital Lease Cash Outflows', 'LEASE-CASH', '', 'leasing', 0, 0, 'USD', 'actual', 1),
        (1, base_date, 'IFRS16', 'IFRS16-SUM', '', 'leasing', 0, 0, 'USD', 'actual', 1);
    END LOOP;
END $$;

-- Verify: Should show 12 months for both budget and actual
SELECT 
    data_type,
    EXTRACT(MONTH FROM date) as month,
    COUNT(*) as rows,
    ROUND(SUM(amount)::numeric, 2) as total_amount
FROM capex_data 
WHERE company_id = 1 AND EXTRACT(YEAR FROM date) = 2025
GROUP BY data_type, EXTRACT(MONTH FROM date)
ORDER BY data_type, month;
