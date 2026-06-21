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
    "backend/internal/worker"
    "backend/internal/ws"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"

    "github.com/nentgroup/slog-prettylogger"
)

func initLogger(cfg *config.Config) {
    var h slog.Handler

    if cfg.AppEnvironment == "production" {
        level := slog.LevelInfo
        if cfg.DebugLog {
            level = slog.LevelDebug
        }
        h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
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
        &model.ContactStage{},
        &model.EmailThread{},
        &model.EmailMessage{},
        &model.EmailTemplate{},
        &model.BulkEmailJob{},
        &model.EmailSignature{},
        &model.EmailAccount{},
        &model.CalendarEvent{},
        &model.ZoomAccount{},
        &model.Document{},
        &model.Signature{},
        &model.SignatureLog{},
        &model.Funnel{},
        &model.Page{},
        &model.FunnelVisit{},
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
    pipelineService := service.NewPipelineService(pipelineRepo, contactRepo)
    pipelineHandler := handler.NewPipelineHandler(pipelineService)

    commRepo := repository.NewCommunicationRepository(db)

    calRepo := repository.NewCalendarRepository(db)
    wsHub := ws.NewHub()

    emailWorker := worker.NewEmailWorker(commRepo, wsHub)
    commService := service.NewCommunicationService(commRepo, emailWorker)
    commHandler := handler.NewCommunicationHandler(commService)

    zoomWorker := worker.NewZoomWorker(calRepo, wsHub)
    calService := service.NewCalendarService(calRepo, zoomWorker)
    calHandler := handler.NewCalendarHandler(calService)

    docRepo := repository.NewDocumentRepository(db)
    docService := service.NewDocumentService(docRepo, contactRepo, wsHub)
    docHandler := handler.NewDocumentHandler(docService)

    funnelRepo := repository.NewFunnelRepository(db)
    funnelService := service.NewFunnelService(funnelRepo, contactService)
    funnelHandler := handler.NewFunnelHandler(funnelService)

    r := gin.Default()

    r.Use(cors.Default())

    if cfg.DebugLog {
        slog.Info("Debug logging enabled (DEBUG_LOG=true)")
        r.Use(middleware.DebugLogger())
    }

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
            contacts.GET("/:id/threads", commHandler.GetContactThreads)
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
            pipelines.GET("/", pipelineHandler.List)
            pipelines.POST("/", pipelineHandler.Create)
            pipelines.GET("/:id", pipelineHandler.GetByID)
            pipelines.PUT("/:id", pipelineHandler.Update)
            pipelines.DELETE("/:id", pipelineHandler.Delete)
            pipelines.GET("/:id/kanban", pipelineHandler.GetKanban)
            pipelines.POST("/:id/move", pipelineHandler.MoveContact)
            pipelines.POST("/:id/stages", pipelineHandler.CreateStage)
            pipelines.PUT("/:id/stages/:stage_id", pipelineHandler.UpdateStage)
            pipelines.DELETE("/:id/stages/:stage_id", pipelineHandler.DeleteStage)
		}

        comm := api.Group("/communication")
        comm.Use(middleware.JWTAuth(cfg.JWTSecret))
        {
            comm.GET("/stats", commHandler.GetStats)
            comm.GET("/queue-status", commHandler.GetQueueStatus)
            comm.POST("/sync", commHandler.SyncInbox)

            threads := comm.Group("/threads")
            {
                threads.GET("/", commHandler.ListThreads)
                threads.POST("/", commHandler.SendEmail)
                threads.GET("/:id", commHandler.GetThread)
                threads.DELETE("/:id", commHandler.DeleteThread)
                threads.POST("/:id/reply", commHandler.ReplyToThread)
                threads.PUT("/:id/read", commHandler.MarkRead)
                threads.PUT("/:id/archive", commHandler.ArchiveThread)
                threads.PUT("/:id/status", commHandler.UpdateThreadStatus)
            }

            messages := comm.Group("/messages")
            {
                messages.PUT("/:id/star", commHandler.StarMessage)
            }

            bulk := comm.Group("/bulk")
            {
                bulk.GET("/", commHandler.ListBulkJobs)
                bulk.POST("/", commHandler.SendBulkEmail)
                bulk.GET("/:id", commHandler.GetBulkJob)
            }

            templates := comm.Group("/templates")
            {
                templates.GET("/", commHandler.ListTemplates)
                templates.POST("/", commHandler.CreateTemplate)
                templates.PUT("/:id", commHandler.UpdateTemplate)
                templates.DELETE("/:id", commHandler.DeleteTemplate)
            }

            signatures := comm.Group("/signatures")
            {
                signatures.GET("/", commHandler.ListSignatures)
                signatures.POST("/", commHandler.CreateSignature)
                signatures.PUT("/:id", commHandler.UpdateSignature)
                signatures.DELETE("/:id", commHandler.DeleteSignature)
            }

            account := comm.Group("/account")
            {
                account.GET("/", commHandler.GetEmailAccount)
                account.POST("/", commHandler.CreateEmailAccount)
                account.PUT("/", commHandler.UpdateEmailAccount)
                account.DELETE("/", commHandler.DeleteEmailAccount)
                account.POST("/test", commHandler.TestConnection)
            }
        }

        cal := api.Group("/calendar")
        cal.Use(middleware.JWTAuth(cfg.JWTSecret))
        {
            cal.GET("/stats", calHandler.GetStats)
            cal.GET("/upcoming", calHandler.UpcomingEvents)

            events := cal.Group("/events")
            {
                events.GET("/", calHandler.ListEvents)
                events.POST("/", calHandler.CreateEvent)
                events.GET("/:id", calHandler.GetEvent)
                events.PUT("/:id", calHandler.UpdateEvent)
                events.DELETE("/:id", calHandler.DeleteEvent)
            }

            zoom := cal.Group("/zoom")
            {
                zoom.GET("/", calHandler.GetZoomAccount)
                zoom.POST("/", calHandler.CreateZoomAccount)
                zoom.PUT("/", calHandler.UpdateZoomAccount)
                zoom.DELETE("/", calHandler.DeleteZoomAccount)
                zoom.POST("/test", calHandler.TestZoomConnection)
            }
        }

        docs := api.Group("/documents")
        docs.Use(middleware.JWTAuth(cfg.JWTSecret))
        {
            docs.GET("/", docHandler.List)
            docs.POST("/", docHandler.Upload)
            docs.GET("/:id", docHandler.Get)
            docs.DELETE("/:id", docHandler.Delete)
            docs.GET("/:id/download", docHandler.Download)
            docs.GET("/:id/view", docHandler.View)
            docs.POST("/:id/sign", docHandler.Sign)
            docs.GET("/:id/audit-log", docHandler.AuditLog)
            docs.GET("/:id/signatures/:sig_id/image", docHandler.SignatureImage)
        }

        api.GET("/ws", middleware.JWTAuth(cfg.JWTSecret), wsHub.HandleWS)

        funnels := api.Group("/funnels")
        funnels.Use(middleware.JWTAuth(cfg.JWTSecret))
        {
            funnels.GET("/", funnelHandler.ListFunnels)
            funnels.POST("/", funnelHandler.CreateFunnel)
            funnels.GET("/:id", funnelHandler.GetFunnel)
            funnels.PUT("/:id", funnelHandler.UpdateFunnel)
            funnels.DELETE("/:id", funnelHandler.DeleteFunnel)
            funnels.POST("/:id/pages", funnelHandler.CreatePage)
            funnels.PUT("/pages/:pageId", funnelHandler.UpdatePage)
            funnels.DELETE("/pages/:pageId", funnelHandler.DeletePage)
        }
    }

    public := r.Group("/public/v1")
    {
        public.POST("/funnels/submit", funnelHandler.Submit)
    }

    workerCtx, workerCancel := context.WithCancel(context.Background())
    zoomWorker.Start(workerCtx)
    emailWorker.Start(workerCtx)

    r.NoRoute(funnelHandler.ServeLivePage)

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

    workerCancel()
    zoomWorker.Wait()
    emailWorker.Wait()

    slog.Info("Server exiting")
}
