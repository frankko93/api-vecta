package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"

	authHandler "github.com/gmhafiz/go8/internal/domain/auth/handler"
	authRepo "github.com/gmhafiz/go8/internal/domain/auth/repository"
	authUseCase "github.com/gmhafiz/go8/internal/domain/auth/usecase"
	companiesHandler "github.com/gmhafiz/go8/internal/domain/config/companies"
	mineralsHandler "github.com/gmhafiz/go8/internal/domain/config/minerals"
	"github.com/gmhafiz/go8/internal/domain/data"
	"github.com/gmhafiz/go8/internal/domain/health"
	"github.com/gmhafiz/go8/internal/domain/reports"
	"github.com/gmhafiz/go8/internal/middleware"
	"github.com/gmhafiz/go8/internal/utility/respond"
)

func (s *Server) InitDomains() {
	s.initVersion()
	s.initSwagger()
	s.initHealth()
	s.initAuth()
	s.initConfig()
	s.initData()
	s.initReports()
}

func (s *Server) initVersion() {
	s.router.Route("/version", func(router chi.Router) {
		router.Use(middleware.JSON)

		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			respond.JSON(w, http.StatusOK, map[string]string{"version": s.Version})
		})
	})
}

func (s *Server) initHealth() {
	newHealthRepo := health.NewRepo(s.sqlx)
	newHealthUseCase := health.New(newHealthRepo)
	health.RegisterHTTPEndPoints(s.router, newHealthUseCase)
}

//go:embed docs/*
var swaggerDocsAssetPath embed.FS

func (s *Server) initSwagger() {
	if s.Config().API.RunSwagger {
		docsPath, err := fs.Sub(swaggerDocsAssetPath, "docs")
		if err != nil {
			panic(err)
		}

		fileServer := http.FileServer(http.FS(docsPath))

		s.router.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
		})
		s.router.Handle("/swagger/", http.StripPrefix("/swagger", middleware.ContentType(fileServer)))
		s.router.Handle("/swagger/*", http.StripPrefix("/swagger", middleware.ContentType(fileServer)))
	}
}

func (s *Server) initAuth() {
	repo := authRepo.New(s.sqlx)
	uc := authUseCase.New(repo)
	handler := authHandler.RegisterHTTPEndPoints(s.router, s.validator, uc, repo)

	// Store authRepo in server for RequirePermission middleware
	s.authRepo = repo

	// Authenticated user can change their own password
	s.router.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(uc))
		r.Put("/api/v1/auth/password", handler.ChangePassword)
	})

	// User management routes
	s.router.Route("/api/v1/admin/users", func(r chi.Router) {
		r.Use(middleware.RequireAuth(uc))

		// Super admin only: full user management (no company filter)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(repo, "super_admin"))

			r.Get("/", handler.ListUsers)         // List all users
			r.Post("/", handler.CreateUser)       // Create new user
			r.Get("/{id}", handler.GetUser)       // Get specific user
			r.Put("/{id}", handler.UpdateUser)    // Update user
			r.Delete("/{id}", handler.DeactivateUser) // Deactivate user

			// Password management (super admin can set any user's password)
			r.Put("/{id}/password", handler.SetPassword)

			// Company assignment (super admin can assign any company)
			r.Post("/companies", handler.AssignUserToCompany)
			r.Put("/{user_id}/companies/{company_id}", handler.UpdateUserCompanyRole)
			r.Delete("/{user_id}/companies/{company_id}", handler.RemoveUserFromCompany)

			// Permissions
			r.Post("/permissions", handler.AssignPermissions)
		})
	})

	// Company admin routes: can only manage users within their companies
	s.router.Route("/api/v1/company/{company_id}/users", func(r chi.Router) {
		r.Use(middleware.RequireAuth(uc))
		r.Use(middleware.ValidateCompanyAccess(repo))
		r.Use(middleware.RequireCompanyRole(middleware.RoleAdmin)) // Must be admin in this company

		r.Get("/", handler.ListUsers)         // List users in company (filtered by company_id)
		r.Post("/", handler.CreateUser)       // Create user (will be assigned to this company)
		r.Get("/{id}", handler.GetUser)       // Get specific user
		r.Put("/{id}", handler.UpdateUser)    // Update user
		r.Delete("/{id}", handler.DeactivateUser) // Deactivate user

		// Company role management within this company
		r.Put("/{user_id}/role", handler.UpdateUserCompanyRole)
		r.Delete("/{user_id}", handler.RemoveUserFromCompany)
	})
}

func (s *Server) initConfig() {
	// Companies
	companiesRepo := companiesHandler.NewRepository(s.sqlx)
	companiesUC := companiesHandler.NewUseCase(companiesRepo)
	companiesH := companiesHandler.NewHandler(companiesUC, s.validator)

	// Minerals
	mineralsRepo := mineralsHandler.NewRepository(s.sqlx)
	mineralsUC := mineralsHandler.NewUseCase(mineralsRepo)
	mineralsH := mineralsHandler.NewHandler(mineralsUC, s.validator)

	// Register routes
	s.router.Route("/api/v1/config", func(r chi.Router) {
		// Public endpoints (only authenticated)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(authUseCase.New(s.authRepo)))

			// Companies - Read
			r.Get("/companies", companiesH.List)
			r.Get("/companies/{id}", companiesH.GetByID)

			// Minerals - Read
			r.Get("/minerals", mineralsH.List)
			r.Get("/minerals/{id}", mineralsH.GetByID)

			// Units - Read
			r.Get("/units", companiesH.GetAvailableUnits)
		})

		// Admin only endpoints
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(authUseCase.New(s.authRepo)))
			r.Use(middleware.RequirePermission(s.authRepo, "admin"))

			// Companies - Write
			r.Post("/companies", companiesH.Create)
			r.Put("/companies/{id}", companiesH.Update)
			r.Delete("/companies/{id}", companiesH.Delete)
			r.Put("/companies/{id}/minerals", companiesH.AssignMinerals)
			r.Put("/companies/{id}/settings", companiesH.UpdateSettings)

			// Minerals - Write
			r.Post("/minerals", mineralsH.Create)
			r.Put("/minerals/{id}", mineralsH.Update)
			r.Delete("/minerals/{id}", mineralsH.Delete)
		})
	})
}

func (s *Server) initData() {
	repo := data.NewRepository(s.sqlx)
	uc := data.NewUseCase(repo)
	h := data.NewHandler(uc, s.validator)

	authUC := authUseCase.New(s.authRepo)

	s.router.Route("/api/v1/data", func(r chi.Router) {
		// All data endpoints require authentication + company access validation
		r.Use(middleware.RequireAuth(authUC))
		r.Use(middleware.ValidateCompanyAccess(s.authRepo))

		// Viewer role: can list/view data
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireCompanyRole(middleware.RoleViewer))
			r.Get("/{type}/list", h.List)
		})

		// Editor role: can import data
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireCompanyRole(middleware.RoleEditor))
			r.Post("/import", h.Import)
		})

		// Admin role: can delete data
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireCompanyRole(middleware.RoleAdmin))
			r.Delete("/{type}/{id}", h.Delete)
		})
	})
}

func (s *Server) initReports() {
	repo := reports.NewRepository(s.sqlx)
	uc := reports.NewUseCase(repo)
	detailUC := reports.NewDetailUseCase(repo)
	h := reports.NewHandler(uc, s.validator, s.authRepo)
	detailH := reports.NewDetailHandler(detailUC, s.validator, s.authRepo)

	authUC := authUseCase.New(s.authRepo)

	s.router.Route("/api/v1/reports", func(r chi.Router) {
		// All reports endpoints require authentication + company access validation
		r.Use(middleware.RequireAuth(authUC))
		r.Use(middleware.ValidateCompanyAccess(s.authRepo))

		// Viewer role: can view reports (read-only)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireCompanyRole(middleware.RoleViewer))

			// Summary and detailed reports
			r.Get("/summary", h.GetSummary)
			r.Get("/saved", h.ListSavedReports)
			r.Get("/pbr", detailH.GetPBRDetail)
			r.Get("/dore", detailH.GetDoreDetail)
			r.Get("/opex", detailH.GetOPEXDetail)
			r.Get("/capex", detailH.GetCAPEXDetail)
		})

		// Editor role: can save reports and compare
		// Note: SaveReport and CompareReports validate roles internally because company_id comes from JSON body
		r.Group(func(r chi.Router) {
			// No role middleware here - handlers validate internally
			r.Post("/save", h.SaveReport)
			r.Post("/compare", h.CompareReports)
		})
	})
}
