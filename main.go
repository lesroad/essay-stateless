package main

import (
	"context"
	appService "essay-stateless/internal/application/service"
	"essay-stateless/internal/config"
	"essay-stateless/internal/handler"
	"essay-stateless/internal/middleware"
	"essay-stateless/internal/repository"
	"essay-stateless/pkg/database"
	"essay-stateless/pkg/logger"
	"essay-stateless/pkg/trace"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// 初始化新版服务（基于DDD架构）
	evaluateServiceV2 := appService.NewEvaluateServiceV2(&cfg.Evaluate)
	ocrServiceV2 := appService.NewOcrServiceV2(&cfg.OCR)
	statisticsServiceV2 := appService.NewStatisticsServiceV2()

	// 初始化Handler（使用新版服务）
	evaluateHandler := handler.NewEvaluateHandler(evaluateServiceV2, rawLogsRepo)
	ocrHandler := handler.NewOcrHandler(ocrServiceV2, rawLogsRepo)
	statisticsHandler := handler.NewStatisticsHandler(statisticsServiceV2, rawLogsRepo)

	router := setupRouter(evaluateHandler, ocrHandler, statisticsHandler)

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

func setupRouter(evaluateHandler *handler.EvaluateHandler, ocrHandler *handler.OcrHandler, statisticsHandler *handler.StatisticsHandler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware("essay-stateless"))
	router.Use(trace.TraceIDMiddleware())
	router.Use(middleware.RequestLoggerMiddleware())

	v1 := router.Group("/evaluate")
	{
		v1.POST("/stream", evaluateHandler.EvaluateStream)
	}

	sts := router.Group("/sts")
	{
		sts.POST("/ocr/title/:provider/:imgType", ocrHandler.TitleOcr)
	}

	statistics := router.Group("/statistics")
	{
		statistics.POST("/class", statisticsHandler.AnalyzeClassStatistics)
	}

	return router
}
