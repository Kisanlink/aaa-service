package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Kisanlink/aaa-service/docs"
	"github.com/Kisanlink/aaa-service/server"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title AAA Service API
// @version 2.0
// @description Authentication, Authorization, and Accounting Service
// @host localhost:8080
// @BasePath /api

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbManager := db.NewDatabaseManager()
	ctx := context.Background()
	if err := dbManager.Connect(ctx); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbManager.Close()

	httpServer := server.NewHTTPServer(dbManager, port, logger)
	router := httpServer.GetRouter()

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})

	go func() {
		logger.Info("Server starting", zap.String("port", port))
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Stop(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}
}
