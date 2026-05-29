package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/myflix/user-service/handler"
	"github.com/myflix/user-service/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	db := store.Connect(os.Getenv("DATABASE_URL"))
	store.Migrate(db)

	h := handler.New(db)

	r := gin.Default()
	r.GET("/actuator/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	u := r.Group("/api/user")
	u.POST("", h.CreateUser)
	u.GET("/:id", h.GetUser)

	// Profiles
	u.POST("/profiles", h.CreateProfile)
	u.GET("/profiles", h.GetProfiles) // ?userId=
	u.GET("/profiles/:id", h.GetProfile)
	u.DELETE("/profiles/:id", h.DeleteProfile)

	// Watch history
	u.POST("/history", h.SaveHistory)
	u.GET("/history", h.GetHistory) // ?profileId=

	// Watchlist
	u.POST("/watchlist", h.AddToWatchlist)
	u.DELETE("/watchlist", h.RemoveFromWatchlist) // ?profileId=&mediaPath=
	u.GET("/watchlist", h.GetWatchlist)           // ?profileId=

	u.GET("/trending", h.GetTrending)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}
	log.Printf("User service listening on :%s", port)
	log.Fatal(r.Run(":" + port))
}
