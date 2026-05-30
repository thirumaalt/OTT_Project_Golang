package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/myflix/metadata-service/handler"
	"github.com/myflix/metadata-service/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	db := store.Connect(os.Getenv("DATABASE_URL"))
	store.Migrate(db)

	h := handler.New(db)

	r := gin.Default()
	r.GET("/actuator/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	meta := r.Group("/api/metadata")
	meta.GET("", h.ListMedia)            // list / filter
	meta.POST("", h.CreateMedia)         // create
	meta.GET("/search", h.SearchMedia)   // full-text search
	meta.GET("/by-path", h.GetByPath)    // lookup by file path
	meta.GET("/title/:title", h.GetByTitle) // TMDB-proxy lookup by title (frontend compat)
	meta.POST("/bulk", h.BulkUpsert)     // bulk upsert
	meta.GET("/:id", h.GetMedia)         // get by id
	meta.PUT("/:id", h.UpdateMedia)      // update
	meta.DELETE("/:id", h.DeleteMedia)   // delete

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}
	log.Printf("Metadata service listening on :%s", port)
	log.Fatal(r.Run(":" + port))
}
