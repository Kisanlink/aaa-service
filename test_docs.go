package main

import (
	"fmt"
	"net/http"
	"os"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Test the swagger.json file exists
	if _, err := os.Stat("docs/swagger.json"); os.IsNotExist(err) {
		fmt.Println("ERROR: docs/swagger.json does not exist!")
		return
	}
	fmt.Println("✓ docs/swagger.json exists")

	// Serve the swagger files
	router.StaticFile("/docs/swagger.json", "docs/swagger.json")
	router.StaticFile("/docs/swagger.yaml", "docs/swagger.yaml")

	// Test Scalar API reference
	router.GET("/docs", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		specURL := scheme + "://" + c.Request.Host + "/docs/swagger.json"

		fmt.Printf("Generating docs with specURL: %s\n", specURL)

		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL:  specURL,
			DarkMode: true,
			CustomOptions: scalar.CustomOptions{
				PageTitle: "AAA Service API Reference - Test",
			},
		})
		if err != nil {
			fmt.Printf("ERROR generating HTML: %v\n", err)
			c.String(http.StatusInternalServerError, "failed to render API docs: %v", err)
			return
		}
		fmt.Printf("✓ HTML generated successfully, length: %d\n", len(htmlContent))
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	})

	// Redirect root to docs
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs")
	})

	fmt.Println("Starting test server on :8081")
	fmt.Println("Visit: http://localhost:8081/docs")
	router.Run(":8081")
}
