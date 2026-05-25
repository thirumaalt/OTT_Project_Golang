package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Subscription struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null"     json:"userId"`
	Plan      string    `gorm:"not null"                 json:"plan"`
	UpdatedAt time.Time `json:"updatedAt"`
}

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	db.AutoMigrate(&Subscription{})

	r := gin.Default()
	r.GET("/actuator/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })

	s := r.Group("/api/subscription")
	s.GET("/user/:userId", getSubscription)
	s.POST("/upgrade", upgradeSubscription)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8093"
	}
	log.Fatal(r.Run(":" + port))
}

func getSubscription(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	var sub Subscription
	if err := db.Where("user_id = ?", userID).First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

func upgradeSubscription(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
		return
	}
	plan := c.Query("plan")
	if plan == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plan required"})
		return
	}

	var sub Subscription
	result := db.Where("user_id = ?", userID).First(&sub)
	if result.Error != nil {
		// Create new subscription
		sub = Subscription{UserID: uint(userID), Plan: plan, UpdatedAt: time.Now()}
		db.Create(&sub)
	} else {
		sub.Plan = plan
		sub.UpdatedAt = time.Now()
		db.Save(&sub)
	}
	c.JSON(http.StatusOK, sub)
}
