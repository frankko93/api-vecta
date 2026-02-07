-- Saved Reports / Scenarios for Comparison

-- Drop table if exists (for clean reset)
DROP TABLE IF EXISTS saved_reports CASCADE;

-- Saved Reports Table
CREATE TABLE IF NOT EXISTS saved_reports (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES mining_companies(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    year INT NOT NULL,
    budget_version INT DEFAULT 1,
    report_data JSONB NOT NULL,  -- Complete summary response (12 months)
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Indexes
CREATE INDEX idx_saved_reports_company ON saved_reports(company_id);
CREATE INDEX idx_saved_reports_year ON saved_reports(year);
CREATE INDEX idx_saved_reports_created_by ON saved_reports(created_by);
CREATE INDEX idx_saved_reports_company_year ON saved_reports(company_id, year);

COMMENT ON TABLE saved_reports IS 'Saved budget scenarios for comparison (what-if analysis)';
COMMENT ON COLUMN saved_reports.report_data IS 'Budget projections for the year (12 months)';
COMMENT ON COLUMN saved_reports.budget_version IS 'Which budget version this scenario represents (1=optimistic, 2=conservative, 3=realistic, etc)';
