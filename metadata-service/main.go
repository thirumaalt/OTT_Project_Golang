package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/myflix/metadata-service/handler"
	"github.com/myflix/metadata-service/store"
	"github.com/myflix/metadata-service/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	db := store.Connect(os.Getenv("DATABASE_URL"))
	store.Migrate(db)

	h := handler.New(db)

	shutdown := telemetry.InitTracer("metadata-service")
	defer shutdown()

	r := gin.Default()
	r.Use(TracingMiddleware())
	r.GET("/actuator/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	meta := r.Group("/api/metadata")
	meta.GET("", h.ListMedia)               // list / filter
	meta.POST("", h.CreateMedia)            // create
	meta.GET("/search", h.SearchMedia)      // full-text search
	meta.GET("/by-path", h.GetByPath)       // lookup by file path
	meta.GET("/title/:title", h.GetByTitle) // TMDB-proxy lookup by title (frontend compat)
	meta.POST("/bulk", h.BulkUpsert)        // bulk upsert
	meta.GET("/:id", h.GetMedia)            // get by id
	meta.PUT("/:id", h.UpdateMedia)         // update
	meta.DELETE("/:id", h.DeleteMedia)      // delete

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}
	log.Printf("Metadata service listening on :%s", port)
	log.Fatal(r.Run(":" + port))
}

func TracingMiddleware() gin.HandlerFunc {
	tracer := otel.Tracer("api-gateway")

	return func(c *gin.Context) {

		ctx := otel.GetTextMapPropagator().Extract(
			c.Request.Context(),
			propagation.HeaderCarrier(c.Request.Header),
		)

		ctx, span := tracer.Start(
			ctx,
			c.Request.Method+" "+c.FullPath(),
		)

		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
