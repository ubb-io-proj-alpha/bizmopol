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

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"

	"github.com/nentgroup/slog-prettylogger"
)


func initLogger(cfg *config.Config) {
    var handler slog.Handler
    
    if cfg.AppEnvironment == "production" {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
    } else {
        handler = prettylogger.NewHandler(os.Stdout, prettylogger.HandlerOptions{
            SlogOpts: slog.HandlerOptions{
                AddSource: true,
                Level:     slog.LevelDebug,
            },
            TimeFormat: time.TimeOnly,
        })
    }


    slog.SetDefault(slog.New(handler))
}

func main() {
    cfg := config.LoadConfig()

    initLogger(cfg)

    db, err := config.ConnectDB(cfg.DBType, cfg.DBURL)
    if err != nil {
        slog.Error("Failed to connect to database", "error", err)
	}

    // migrations, should be good enough for now
    // later we can use other migration tools
    if err := db.AutoMigrate(&model.Test{}); err != nil {
		slog.Error("Failed to migrate database", "error", err)
	}

    testRepo := repository.NewTestRepository(db)
    testService := service.NewTestService(testRepo)
	testHandler := handler.NewTestHandler(testService)

    // use logger from gin
    r := gin.Default()

    // Enable CORS for frontend
    if cfg.AppEnvironment != "production" {
        r.Use(cors.Default())
    }

    api := r.Group("/api/v1")
    {
        tests := api.Group("/tests")
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

	// Run server in a goroutine so it doesn't block shutdown logic
    go func() {
        slog.Info("Server started", "port", cfg.Port)

        if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            slog.Error("Server listen failed", "error", err)
            os.Exit(1)
        }
    }()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	// Context with 5-second timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown:", "error", err)
        os.Exit(1)

	}

	slog.Info("Server exiting")
}