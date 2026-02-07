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

	// Register users endpoint (requires auth)
	s.router.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(uc))
		r.Get("/api/v1/users", handler.ListUsers)
	})

	// Store authRepo in server for RequirePermission middleware
	s.authRepo = repo
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

	// Data import endpoint (requires authentication)
	s.router.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(authUseCase.New(s.authRepo)))
		data.RegisterHTTPEndPoints(r.(*chi.Mux), s.validator, uc)
	})
}

func (s *Server) initReports() {
	repo := reports.NewRepository(s.sqlx)
	uc := reports.NewUseCase(repo)

	// Reports endpoint (requires authentication)
	s.router.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(authUseCase.New(s.authRepo)))
		detailUC := reports.NewDetailUseCase(repo)
		reports.RegisterHTTPEndPoints(r.(*chi.Mux), s.validator, uc, detailUC)
	})
}
