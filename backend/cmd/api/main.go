package main

import (
    "log/slog"
    "net/http"
    "time"
    "errors"
    "context"
    "os"
    "os/signal"
    "syscall"

    "backend/internal/repository"
    "backend/internal/service"
    "backend/internal/handler"
    "backend/internal/config"
    "backend/internal/model"
    "backend/internal/middleware"

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

    if err := db.AutoMigrate(&model.Test{}, &model.User{}); err != nil {
        slog.Error("Failed to migrate database", "error", err)
    }

    if cfg.AppEnvironment == "development") {
        config.SeedDatabase(db)
    }

    testRepo := repository.NewTestRepository(db)
    testService := service.NewTestService(testRepo)
    testHandler := handler.NewTestHandler(testService)

    userRepo := repository.NewUserRepository(db)
    authService := service.NewAuthService(userRepo, cfg.JWTSecret)
    authHandler := handler.NewAuthHandler(authService)

    r := gin.Default()

    // tutaj usunalem warunek
    // ze ENV musi byc prod
    r.Use(cors.Default())

    api := r.Group("/api/v1")
    {
        auth := api.Group("/auth")
        {
            auth.POST("/register", authHandler.Register)
            auth.POST("/login", authHandler.Login)
        }

        tests := api.Group("/tests")
        tests.Use(middleware.JWTAuth(cfg.JWTSecret))
        {
            tests.GET("/", testHandler.GetAll)
            tests.POST("/add", testHandler.Create)
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

