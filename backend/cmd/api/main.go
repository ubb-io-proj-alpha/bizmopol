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
	"backend/internal/middleware"
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
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Contact{},
		&model.ContactHistory{},
		&model.Tag{},
		&model.CustomField{},
		&model.CustomFieldValue{},
		&model.Pipeline{},
		&model.Stage{},
		&model.Lead{},
	); err != nil {
		slog.Error("Failed to migrate database", "error", err)
	}

	if cfg.AppEnvironment == "development" {
		config.SeedDatabase(db)
	}

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	contactRepo := repository.NewContactRepository(db)
	tagRepo := repository.NewTagRepository(db)
	cfRepo := repository.NewCustomFieldRepository(db)
	contactService := service.NewContactService(contactRepo, tagRepo, cfRepo)
	contactHandler := handler.NewContactHandler(contactService)

	tagService := service.NewTagService(tagRepo)
	tagHandler := handler.NewTagHandler(tagService)

	cfService := service.NewCustomFieldService(cfRepo)
	cfHandler := handler.NewCustomFieldHandler(cfService)

	pipelineRepo := repository.NewPipelineRepository(db)
	pipelineService := service.NewPipelineService(pipelineRepo)
	pipelineHandler := handler.NewPipelineHandler(pipelineService)

	r := gin.Default()
	r.Use(cors.Default())

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		contacts := api.Group("/contacts")
		contacts.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			contacts.GET("/", contactHandler.List)
			contacts.POST("/", contactHandler.Create)
			contacts.POST("/merge", contactHandler.Merge)
			contacts.GET("/:id", contactHandler.GetByID)
			contacts.PUT("/:id", contactHandler.Update)
			contacts.DELETE("/:id", contactHandler.Delete)
			contacts.GET("/:id/history", contactHandler.ListHistory)
			contacts.POST("/:id/history", contactHandler.AddHistory)
			contacts.GET("/:id/members", contactHandler.GetGroupMembers)
			contacts.GET("/:id/group-history", contactHandler.GetGroupHistory)
			contacts.PUT("/:id/dnd", contactHandler.UpdateDnd)
		}

		tags := api.Group("/tags")
		tags.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			tags.GET("/", tagHandler.List)
			tags.POST("/", tagHandler.Create)
			tags.PUT("/:id", tagHandler.Update)
			tags.DELETE("/:id", tagHandler.Delete)
		}

		customFields := api.Group("/custom-fields")
		customFields.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			customFields.GET("/", cfHandler.List)
			customFields.POST("/", cfHandler.Create)
			customFields.PUT("/:id", cfHandler.Update)
			customFields.DELETE("/:id", cfHandler.Delete)
		}

		pipelines := api.Group("/pipelines")
		pipelines.Use(middleware.JWTAuth(cfg.JWTSecret))
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
