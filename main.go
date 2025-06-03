package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"essay-stateless/internal/config"
	"essay-stateless/internal/handler"
	"essay-stateless/internal/repository"
	"essay-stateless/internal/service"
	"essay-stateless/pkg/database"
	"essay-stateless/pkg/logger"
	"essay-stateless/pkg/trace"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.Log)

	// 初始化追踪
	shutdown, err := trace.Init(cfg.Trace)
	if err != nil {
		log.Fatal("Failed to initialize tracing:", err)
	}
	defer shutdown()

	db, err := database.NewMongoDB(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Disconnect()

	rawLogsRepo := repository.NewRawLogsRepository(db.Database())

	ocrService := service.NewOcrService(&cfg.OCR)
	evaluateService := service.NewEvaluateService(&cfg.Beta, ocrService)

	evaluateHandler := handler.NewEvaluateHandler(evaluateService, rawLogsRepo)
	ocrHandler := handler.NewOcrHandler(ocrService, rawLogsRepo)

	router := setupRouter(evaluateHandler, ocrHandler)

	server := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRouter(evaluateHandler *handler.EvaluateHandler, ocrHandler *handler.OcrHandler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware("essay-stateless"))
	router.Use(trace.TraceIDMiddleware())

	v1 := router.Group("/evaluate")
	{
		v1.POST("", evaluateHandler.BetaEvaluate)
		v1.POST("/beta/ocr", evaluateHandler.BetaOcrEvaluate)
	}

	sts := router.Group("/sts")
	{
		sts.POST("/ocr/:provider/:imgType", ocrHandler.DefaultOcr)
		sts.POST("/ocr/title/:provider/:imgType", ocrHandler.TitleOcr)
	}

	return router
}
