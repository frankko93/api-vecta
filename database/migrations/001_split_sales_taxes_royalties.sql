-- Migration: Split sales_taxes_royalties into separate columns
-- Date: 2026-02-15
-- Description: Separates the combined sales_taxes_royalties column into
--   sales_taxes, royalties, and adds other_sales_deductions.
--   Existing data is migrated to sales_taxes (royalties defaults to 0).

-- Step 1: Add new columns
ALTER TABLE financial_data ADD COLUMN IF NOT EXISTS sales_taxes DECIMAL(15,2) DEFAULT 0;
ALTER TABLE financial_data ADD COLUMN IF NOT EXISTS royalties DECIMAL(15,2) DEFAULT 0;
ALTER TABLE financial_data ADD COLUMN IF NOT EXISTS other_sales_deductions DECIMAL(15,2) DEFAULT 0;

-- Step 2: Migrate existing data (sales_taxes_royalties -> sales_taxes, royalties stays 0)
UPDATE financial_data
SET sales_taxes = COALESCE(sales_taxes_royalties, 0),
    royalties = 0,
    other_sales_deductions = 0
WHERE sales_taxes_royalties IS NOT NULL OR sales_taxes = 0;

-- Step 3: Drop the old combined column
ALTER TABLE financial_data DROP COLUMN IF EXISTS sales_taxes_royalties;
