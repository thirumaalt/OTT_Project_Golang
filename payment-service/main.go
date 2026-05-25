package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PaymentOrder struct {
	ID      string `gorm:"primaryKey"   json:"id"`
	Amount  int    `gorm:"not null"     json:"amount"`
	Status  string `gorm:"not null"     json:"status"`
	Receipt string `json:"receipt"`
}

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	db.AutoMigrate(&PaymentOrder{})

	r := gin.Default()
	r.GET("/actuator/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })

	p := r.Group("/api/payment")
	p.POST("/create-order", createOrder)
	p.POST("/capture-payment", capturePayment)
	p.GET("/status/:orderId", getStatus)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8091"
	}
	log.Fatal(r.Run(":" + port))
}

type createOrderReq struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

func createOrder(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: integrate Razorpay SDK — generate real order ID via Razorpay API
	// For now, create a pending order record
	order := PaymentOrder{
		ID:      generateOrderID(),
		Amount:  req.Amount,
		Status:  "CREATED",
		Receipt: "receipt_" + generateOrderID(),
	}
	db.Create(&order)
	c.JSON(http.StatusCreated, order)
}

func capturePayment(c *gin.Context) {
	orderID := c.Query("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId required"})
		return
	}
	var order PaymentOrder
	if err := db.First(&order, "id = ?", orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	order.Status = "CAPTURED"
	db.Save(&order)
	c.JSON(http.StatusOK, order)
}

func getStatus(c *gin.Context) {
	var order PaymentOrder
	if err := db.First(&order, "id = ?", c.Param("orderId")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func generateOrderID() string {
	return "order_" + os.Getenv("HOSTNAME") + "_" + randomSuffix()
}

func randomSuffix() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[i%36]
	}
	return string(b)
}
