package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nentgroup/slog-prettylogger"
)

func initLogger(cfg *config.Config) {
	var h slog.Handler

	if cfg.AppEnvironment == "production" {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		h = prettylogger.NewHandler(os.Stdout, prettylogger.HandlerOptions{
			SlogOpts: slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
			TimeFormat: time.TimeOnly,
		})
	}

	slog.SetDefault(slog.New(h))
}

func main() {
	cfg := config.LoadConfig()

	initLogger(cfg)

	db, err := config.ConnectDB(cfg.DBType, cfg.DBURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(&model.Test{}, &model.User{}, &model.Pipeline{}, &model.Stage{}, &model.Lead{}); err != nil {
		slog.Error("Failed to migrate database", "error", err)
		os.Exit(1)
	}

	// Ensure seeded accounts (admin/coach/user with password123) exist and are updated on each startup.
	config.SeedDatabase(db)

	testRepo := repository.NewTestRepository(db)
	testService := service.NewTestService(testRepo)
	testHandler := handler.NewTestHandler(testService)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	pipelineRepo := repository.NewPipelineRepository(db)
	pipelineService := service.NewPipelineService(pipelineRepo)
	pipelineHandler := handler.NewPipelineHandler(pipelineService)

	r := gin.Default()

	if cfg.AppEnvironment != "production" {
		r.Use(cors.Default())
	}

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		tests := api.Group("/tests")
		{
			tests.GET("/", testHandler.GetAll)
			tests.POST("/add", testHandler.Create)
		}

		pipelines := api.Group("/pipelines")
		{
			pipelines.POST("/", pipelineHandler.CreatePipeline)
			pipelines.GET("/", pipelineHandler.GetUserPipelines)
			pipelines.GET("/:id", pipelineHandler.GetPipeline)
			pipelines.PUT("/:id", pipelineHandler.UpdatePipeline)
			pipelines.DELETE("/:id", pipelineHandler.DeletePipeline)

			stages := pipelines.Group("/:id/stages")
			{
				stages.POST("/", pipelineHandler.CreateStage)
				stages.PUT("/:stageId", pipelineHandler.UpdateStage)
				stages.PUT("/:stageId/position", pipelineHandler.UpdateStagePosition)
				stages.DELETE("/:stageId", pipelineHandler.DeleteStage)
			}

			leads := pipelines.Group("/:id/leads")
			{
				leads.POST("/", pipelineHandler.CreateLead)
				leads.PUT("/:leadId", pipelineHandler.UpdateLead)
				leads.PUT("/:leadId/move", pipelineHandler.MoveLeadToStage)
				leads.DELETE("/:leadId", pipelineHandler.DeleteLead)
			}
		}
	}

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("Server started", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server listen failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown:", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exiting")
}
