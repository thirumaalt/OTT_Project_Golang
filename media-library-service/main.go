package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/myflix/media-library-service/handler"
	"github.com/myflix/media-library-service/scanner"
)

func main() {
	// Configuration
	mediaDir := os.Getenv("MEDIA_DATA_DIR")
	if mediaDir == "" {
		mediaDir = "/media"
	}

	// Initialize scanner
	s := scanner.New(mediaDir)

	// Initialize handler
	h := handler.New(s)

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "media_dir": mediaDir})
	})

	// Media routes
	media := r.Group("/api/media")
	media.GET("/library", h.GetLibrary)
	media.GET("/search", h.Search)
	media.GET("/stream", h.Stream)
	media.GET("/hls/:file_id/:filename", h.GetHLSFile)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	log.Printf("Media Library Service listening on :%s", port)
	log.Printf("Media directory: %s", mediaDir)
	log.Fatal(r.Run(":" + port))
}
