package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/myflix/auth-service/handler"
	"github.com/myflix/auth-service/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	db := store.Connect(os.Getenv("DATABASE_URL"))
	store.Migrate(db)

	h := handler.New(db, os.Getenv("JWT_SECRET"))

	r := gin.Default()
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/actuator/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })

	auth := r.Group("/api/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	log.Printf("Auth service listening on :%s", port)
	log.Fatal(r.Run(":" + port))
}
