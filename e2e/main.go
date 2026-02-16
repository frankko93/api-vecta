package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gmhafiz/go8/config"
	"github.com/gmhafiz/go8/internal/domain/auth"
	configDomain "github.com/gmhafiz/go8/internal/domain/config"
	"github.com/gmhafiz/go8/internal/domain/data"
	"github.com/gmhafiz/go8/internal/domain/reports"
)

const Version = "v0.1.0"

var (
	url   = ""
	token = ""
)

func main() {
	log.Printf("Starting e2e tests - API version: %s\n", Version)
	cfg := config.New()

	url = fmt.Sprintf("http://%s:%s", cfg.API.Host, cfg.API.Port)

	waitForAPI(fmt.Sprintf("%s/api/health/readiness", url))

	run()
}

func run() {
	log.Println("=== Starting E2E Tests ===")

	// Critical flows
	testHealth()
	testAuthFlow()
	testConfigFlow()
	testImportFlow()
	testReportsFlow()
	testCompleteWorkflow()

	log.Println("=== All E2E tests passed! ===")
}

// Test 1: Health check
func testHealth() {
	log.Println("→ Testing health check...")

	resp, err := http.Get(fmt.Sprintf("%s/api/health/readiness", url))
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Health check failed: got %v want %v", resp.StatusCode, http.StatusOK)
	}

	log.Println("✓ Health check passed")
}

// Test 2: Authentication flow
func testAuthFlow() {
	log.Println("→ Testing auth flow...")

	// Login
	loginReq := auth.LoginRequest{
		DNI:      "99999999",
		Password: "admin123",
	}

	resp, err := doRequest(http.MethodPost, "/api/v1/auth/login", loginReq, "")
	if err != nil {
		log.Fatalf("Login request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Login failed: got %v want %v, body: %s", resp.StatusCode, http.StatusOK, string(body))
	}

	var loginResp auth.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		log.Fatalf("Failed to decode login response: %v", err)
	}
	resp.Body.Close()

	// Validate login response
	if loginResp.Token == "" {
		log.Fatal("Token is empty")
	}
	if len(loginResp.Token) != 64 {
		log.Fatalf("Token length incorrect: got %d want 64", len(loginResp.Token))
	}
	if loginResp.User.DNI != "99999999" {
		log.Fatalf("User DNI incorrect: got %s want 99999999", loginResp.User.DNI)
	}
	if len(loginResp.User.Permissions) == 0 {
		log.Fatal("User has no permissions")
	}

	token = loginResp.Token
	log.Printf("✓ Login successful - Token: %s...%s", token[:8], token[len(token)-8:])

	// Test /me endpoint
	meResp, err := doRequest(http.MethodGet, "/api/v1/auth/me", nil, token)
	if err != nil {
		log.Fatalf("Me request failed: %v", err)
	}
	if meResp.StatusCode != http.StatusOK {
		log.Fatalf("Me endpoint failed: got %v want %v", meResp.StatusCode, http.StatusOK)
	}

	var user auth.UserWithPermissions
	json.NewDecoder(meResp.Body).Decode(&user)
	meResp.Body.Close()

	if user.DNI != "99999999" {
		log.Fatalf("Me endpoint returned wrong user: got %s", user.DNI)
	}

	log.Println("✓ Auth flow passed")
}

// Test 3: Config flow
func testConfigFlow() {
	log.Println("→ Testing config flow...")

	// List minerals (should have 7 seeded)
	mineralsResp, err := doRequest(http.MethodGet, "/api/v1/config/minerals", nil, token)
	if err != nil {
		log.Fatalf("List minerals failed: %v", err)
	}
	if mineralsResp.StatusCode != http.StatusOK {
		log.Fatalf("List minerals failed: got %v want %v", mineralsResp.StatusCode, http.StatusOK)
	}

	var minerals []*configDomain.Mineral
	json.NewDecoder(mineralsResp.Body).Decode(&minerals)
	mineralsResp.Body.Close()

	if len(minerals) < 7 {
		log.Fatalf("Expected at least 7 minerals, got %d", len(minerals))
	}

	// Verify Au, Ag exist
	hasGold, hasSilver := false, false
	for _, m := range minerals {
		if m.Code == "AU" {
			hasGold = true
		}
		if m.Code == "AG" {
			hasSilver = true
		}
	}
	if !hasGold || !hasSilver {
		log.Fatal("Missing Au or Ag in minerals")
	}

	log.Println("✓ Config flow passed")
}

// Test 4: Import flow
func testImportFlow() {
	log.Println("→ Testing import flow...")

	// Create test company first
	companyReq := configDomain.CreateCompanyRequest{
		Name:         "E2E Test Company",
		LegalName:    "E2E Test Company S.A.",
		TaxID:        "20-99999999-8",
		Address:      "Test Address",
		ContactEmail: "test@e2e.com",
	}

	companyResp, err := doRequest(http.MethodPost, "/api/v1/config/companies", companyReq, token)
	if err != nil {
		log.Fatalf("Create company failed: %v", err)
	}
	if companyResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(companyResp.Body)
		log.Fatalf("Create company failed: got %v want %v, body: %s", companyResp.StatusCode, http.StatusCreated, string(body))
	}

	var company configDomain.MiningCompany
	json.NewDecoder(companyResp.Body).Decode(&company)
	companyResp.Body.Close()

	companyID := company.ID
	log.Printf("✓ Company created - ID: %d", companyID)

	// Import PBR data
	pbrCSV := []byte(`date,ore_mined_t,waste_mined_t,developments_m,total_tonnes_processed,feed_grade_silver_gpt,feed_grade_gold_gpt,recovery_rate_silver_pct,recovery_rate_gold_pct
2025-01-15,24859,262591,598,35951,209.79,7.35,94.01,95.36`)

	importResp, err := doMultipartRequest("/api/v1/data/import", map[string]string{
		"type":       "pbr",
		"data_type":  "actual",
		"company_id": fmt.Sprintf("%d", companyID),
	}, "file", "pbr.csv", pbrCSV, token)

	if err != nil {
		log.Fatalf("Import PBR failed: %v", err)
	}
	if importResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(importResp.Body)
		log.Fatalf("Import PBR failed: got %v want %v, body: %s", importResp.StatusCode, http.StatusOK, string(body))
	}

	var importResult data.ImportResponse
	json.NewDecoder(importResp.Body).Decode(&importResult)
	importResp.Body.Close()

	if !importResult.Success {
		log.Fatalf("Import failed: %+v", importResult.Errors)
	}
	if importResult.RowsInserted != 1 {
		log.Fatalf("Expected 1 row inserted, got %d", importResult.RowsInserted)
	}

	log.Println("✓ Import flow passed")
}

// Test 5: Reports flow
func testReportsFlow() {
	log.Println("→ Testing reports flow...")

	// Get summary (should work even with partial data)
	resp, err := doRequest(http.MethodGet, "/api/v1/reports/summary?company_id=1&year=2025&budget_version=1&months=1", nil, token)
	if err != nil {
		log.Fatalf("Get summary failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Get summary failed: got %v want %v, body: %s", resp.StatusCode, http.StatusOK, string(body))
	}

	var summary reports.SummaryReport
	json.NewDecoder(resp.Body).Decode(&summary)
	resp.Body.Close()

	// Validate structure
	if summary.Year != 2025 {
		log.Fatalf("Wrong year: got %d want 2025", summary.Year)
	}
	if len(summary.Months) == 0 {
		log.Fatal("No months in summary")
	}
	if summary.Months[0].Month != "2025-01" {
		log.Fatalf("Wrong month: got %s want 2025-01", summary.Months[0].Month)
	}

	log.Println("✓ Reports flow passed")
}

// Test 6: Complete workflow (most critical)
func testCompleteWorkflow() {
	log.Println("→ Testing complete workflow...")

	// Create dedicated company
	companyReq := configDomain.CreateCompanyRequest{
		Name:      "Workflow Test Company",
		LegalName: "Workflow Test S.A.",
		TaxID:     "20-88888888-8",
	}

	companyResp, _ := doRequest(http.MethodPost, "/api/v1/config/companies", companyReq, token)
	var company configDomain.MiningCompany
	json.NewDecoder(companyResp.Body).Decode(&company)
	companyResp.Body.Close()

	// Assign minerals
	assignReq := configDomain.AssignMineralsRequest{
		MineralIDs: []int{1, 2}, // Au, Ag
	}
	assignResp, _ := doRequest(http.MethodPut, fmt.Sprintf("/api/v1/config/companies/%d/minerals", company.ID), assignReq, token)
	if assignResp.StatusCode != http.StatusOK {
		log.Fatalf("Assign minerals failed: got %v", assignResp.StatusCode)
	}
	assignResp.Body.Close()

	// Import complete dataset
	datasets := map[string][]byte{
		"pbr":       buildPBRCSV(),
		"dore":      buildDoreCSV(),
		"opex":      buildOPEXCSV(),
		"capex":     buildCAPEXCSV(),
		"financial": buildFinancialCSV(),
	}

	for dataType, csvContent := range datasets {
		importResp, _ := doMultipartRequest("/api/v1/data/import", map[string]string{
			"type":       dataType,
			"data_type":  "actual",
			"company_id": fmt.Sprintf("%d", company.ID),
		}, "file", fmt.Sprintf("%s.csv", dataType), csvContent, token)

		if importResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(importResp.Body)
			log.Fatalf("Import %s failed: %v, body: %s", dataType, importResp.StatusCode, string(body))
		}
		importResp.Body.Close()
	}

	// Get summary and validate calculations
	summaryResp, _ := doRequest(http.MethodGet, fmt.Sprintf("/api/v1/reports/summary?company_id=%d&year=2025&budget_version=1&months=1", company.ID), nil, token)
	if summaryResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(summaryResp.Body)
		log.Fatalf("Get summary failed: got %v, body: %s", summaryResp.StatusCode, string(body))
	}

	var summary reports.SummaryReport
	json.NewDecoder(summaryResp.Body).Decode(&summary)
	summaryResp.Body.Close()

	// Validate complete structure
	if len(summary.Months) == 0 {
		log.Fatal("Summary has no months")
	}

	month := summary.Months[0]
	actual := month.Actual

	// Validate all metrics have data
	if !actual.Mining.HasData {
		log.Fatal("Mining data missing")
	}
	if !actual.Processing.HasData {
		log.Fatal("Processing data missing")
	}
	if !actual.Production.HasData {
		log.Fatal("Production data missing")
	}
	if !actual.Costs.HasData {
		log.Fatal("Costs data missing")
	}
	if !actual.NSR.HasData {
		log.Fatal("NSR data missing")
	}
	if !actual.CAPEX.HasData {
		log.Fatal("CAPEX data missing")
	}

	// Validate calculations are correct
	if actual.Production.TotalProductionSilverOz <= 0 {
		log.Fatalf("Invalid production silver: %f", actual.Production.TotalProductionSilverOz)
	}
	if actual.NSR.NetSmelterReturn <= 0 {
		log.Fatalf("Invalid NSR: %f", actual.NSR.NetSmelterReturn)
	}
	if actual.CAPEX.ProductionBasedMargin <= 0 {
		log.Fatalf("Invalid margin: %f", actual.CAPEX.ProductionBasedMargin)
	}

	// Validate margin calculation: NSR - Costs
	expectedMargin := actual.NSR.NetSmelterReturn - actual.Costs.ProductionBasedCosts
	if actual.CAPEX.ProductionBasedMargin != expectedMargin {
		log.Fatalf("Margin calculation wrong: got %f want %f", actual.CAPEX.ProductionBasedMargin, expectedMargin)
	}

	// Validate cash flow: Margin - Sustaining CAPEX
	expectedCashFlow := expectedMargin - actual.CAPEX.Sustaining
	if actual.CAPEX.PBRNetCashFlow != expectedCashFlow {
		log.Fatalf("Cash flow calculation wrong: got %f want %f", actual.CAPEX.PBRNetCashFlow, expectedCashFlow)
	}

	log.Println("✓ Complete workflow passed with validated calculations")
}

// HTTP helpers

func doRequest(method, path string, body interface{}, authToken string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	}

	return http.DefaultClient.Do(req)
}

func doMultipartRequest(path string, fields map[string]string, fileField, fileName string, fileContent []byte, authToken string) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, val := range fields {
		writer.WriteField(key, val)
	}

	part, err := writer.CreateFormFile(fileField, fileName)
	if err != nil {
		return nil, err
	}
	part.Write(fileContent)
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, url+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	}

	return http.DefaultClient.Do(req)
}

// CSV builders
func buildPBRCSV() []byte {
	return []byte(`date,ore_mined_t,waste_mined_t,developments_m,total_tonnes_processed,feed_grade_silver_gpt,feed_grade_gold_gpt,recovery_rate_silver_pct,recovery_rate_gold_pct
2025-01-15,24859,262591,598,35951,209.79,7.35,94.01,95.36`)
}

func buildDoreCSV() []byte {
	return []byte(`date,dore_produced_oz,silver_grade_pct,gold_grade_pct,pbr_price_silver,pbr_price_gold,realized_price_silver,realized_price_gold,silver_adjustment_oz,gold_adjustment_oz,ag_deductions_pct,au_deductions_pct,treatment_charge,refining_deductions_au
2025-01-15,236064,85.5,14.5,24.50,2000,24.30,1985,10,5,2.5,1.5,5000,1200`)
}

func buildOPEXCSV() []byte {
	return []byte(`date,cost_center,subcategory,expense_type,amount,currency
2025-01-15,Mine,Drilling,Labour,8537997,USD
2025-01-15,Processing,CO General,Labour,3613678,USD
2025-01-15,G&A,Admin,Third Party,5471220,USD
2025-01-15,Mine,Inventory Variations,Other,1740162,USD`)
}

func buildCAPEXCSV() []byte {
	return []byte(`date,category,car_number,project_name,type,amount,currency
2025-01-15,Mine Equipment,C001,Equipment,sustaining,500000,USD
2025-01-15,Infrastructure,,Infra,sustaining,211052,USD`)
}

func buildFinancialCSV() []byte {
	return []byte(`date,shipping_selling,sales_taxes,royalties,other_sales_deductions,other_adjustments
2025-01-15,-202,465867,0,0,0`)
}

func waitForAPI(readinessURL string) {
	log.Println("Waiting for API...")
	for {
		_, err := http.Get(readinessURL)
		if err == nil {
			log.Println("✓ API is ready")
			return
		}

		base, capacity := time.Second, time.Minute
		for backoff := base; err != nil; backoff <<= 1 {
			if backoff > capacity {
				backoff = capacity
			}
			jitter := rand.Int63n(int64(backoff * 3))
			sleep := base + time.Duration(jitter)
			time.Sleep(sleep)
			_, err := http.Get(readinessURL)
			if err == nil {
				log.Println("✓ API is ready")
				return
			}
		}
	}
}
