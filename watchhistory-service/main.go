package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type WatchHistory struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index"           json:"userId"`
	ContentID uint      `gorm:"not null"                 json:"contentId"`
	WatchedAt time.Time `json:"watchedAt"`
}

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	db.AutoMigrate(&WatchHistory{})

	shutdown := telemetry.InitTracer("watchhistory-service")
	defer shutdown()

	r := gin.Default()
	shutdown := telemetry.InitTracer("user-service")
	defer shutdown()
	r.GET("/actuator/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	wh := r.Group("/api/watch-history")
	wh.POST("/record", recordWatch)
	wh.GET("/all", getAll)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	log.Fatal(r.Run(":" + port))
}

func recordWatch(c *gin.Context) {
	userID, err1 := strconv.ParseUint(c.Query("userId"), 10, 64)
	contentID, err2 := strconv.ParseUint(c.Query("contentId"), 10, 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId and contentId required"})
		return
	}
	wh := WatchHistory{UserID: uint(userID), ContentID: uint(contentID), WatchedAt: time.Now()}
	db.Create(&wh)
	c.JSON(http.StatusCreated, wh)
}

func getAll(c *gin.Context) {
	var list []WatchHistory
	db.Order("watched_at desc").Find(&list)
	c.JSON(http.StatusOK, list)
}

func TracingMiddleware() gin.HandlerFunc {
	tracer := otel.Tracer("user-service")

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
