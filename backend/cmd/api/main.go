package main

import (
    "log"
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
)

func main() {
    cfg := config.LoadConfig()

    db, err := config.ConnectDB(cfg.DBType, cfg.DBURL)
    if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

    // migrations, should be good enough for now
    // later we can use other migration tools
    if err := db.AutoMigrate(&model.Test{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

    testRepo := repository.NewTestRepository(db)
    testService := service.NewTestService(testRepo)
	testHandler := handler.NewTestHandler(testService)

    // use logger from gin
    r := gin.Default()

    // Enable CORS for frontend
    r.Use(cors.Default())

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
		log.Printf("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Context with 5-second timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}