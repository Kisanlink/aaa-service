package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kisanlink/aaa-service/database"
	"github.com/Kisanlink/aaa-service/grpc_server"
	"github.com/Kisanlink/aaa-service/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

func main() {
	database.ConnectDB()
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	grpcServer, err := grpc_server.StartGRPCServer(database.DB)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
	defer grpcServer.GracefulStop()

	r := gin.Default()
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://example.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization", "Accept","aaa-auth-token"},
		AllowCredentials: true,
		Debug:            false, // Set to true for development
	})
	corsMiddleware.Handler(r)
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Welcome to AAA Service",
			"time":    time.Now().UTC(),
		})
	})
	api := r.Group("/api")
	routes.Routes(api,database.DB)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	log.Printf("Server running on port %s", port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}