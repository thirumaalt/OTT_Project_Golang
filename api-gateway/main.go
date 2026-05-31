package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/myflix/api-gateway/proxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var jwtSecret []byte

var httpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests",
	},
	[]string{"method", "path", "status"},
)

var httpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "path"},
)

func main() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET env var is required")
	}
	jwtSecret = []byte(secret)

	r := gin.Default()
	r.Use(MetricsMiddleware())
	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Use(corsMiddleware())

	// Health
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/system/health", proxy.SystemHealth)

	// Public — no auth required
	r.Any("/api/auth/*path", proxy.To(envURL("AUTH_BASE_URL", "http://auth-service:8083")+"/api/auth"))

	// Protected routes
	protected := r.Group("/", authMiddleware())
	protected.Any("/api/user/*path", proxy.To(envURL("USER_BASE_URL", "http://user-service:8085")+"/api/user"))
	protected.Any("/api/interaction/*path", proxy.To(envURL("INTERACTION_BASE_URL", "http://interaction-service:8089")+"/api/interaction"))
	protected.Any("/api/payment/*path", proxy.To(envURL("PAYMENT_BASE_URL", "http://payment-service:8091")+"/api/payment"))
	protected.Any("/api/subscription/*path", proxy.To(envURL("SUBSCRIPTION_BASE_URL", "http://subscription-service:8093")+"/api/subscription"))
	protected.Any("/api/analytics/*path", proxy.To(envURL("ANALYTICS_BASE_URL", "http://analytics-service:8086")+"/api/analytics"))
	protected.Any("/api/recommendation/*path", proxy.To(envURL("RECOMMENDATION_BASE_URL", "http://recommendation-service:8087")+"/api/recommendation"))
	protected.Any("/api/watchhistory/*path", proxy.To(envURL("WATCHHISTORY_BASE_URL", "http://watchhistory-service:8090")+"/api/watchhistory"))
	protected.Any("/api/metadata/*path", proxy.To(envURL("METADATA_BASE_URL", "http://metadata-service:8088")+"/api/metadata"))
	protected.Any("/api/media/*path", proxy.MediaProxy(envURL("MEDIA_BASE_URL", "http://media-library-service:8001")+"/api/media"))
	protected.Any("/api/transcoding/status", proxy.To(envURL("TRANSCODING_BASE_URL", "http://transcoding-service:8092")+"/queue/status"))
	protected.Any("/api/transcoding/transcode", proxy.To(envURL("TRANSCODING_BASE_URL", "http://transcoding-service:8092")+"/transcode"))

	port := envURL("PORT", "8094")
	log.Printf("API Gateway listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()

		c.Next()

		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(time.Since(start).Seconds())
	}
}
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""

		// 1. Try Authorization header (standard API calls)
		auth := c.GetHeader("Authorization")
		if len(auth) > 7 && auth[:7] == "Bearer " {
			tokenStr = auth[7:]
		}

		// 2. Fall back to ?token= query param (needed for <video> / HLS streaming
		//    where the browser cannot set custom headers)
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			return
		}
		_, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			return
		}
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsStr == "" {
		allowedOriginsStr = "http://localhost:5173"
	}
	allowedOrigins := strings.Split(allowedOriginsStr, ",")
	for i, o := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(o)
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			for _, allowed := range allowedOrigins {
				if allowed == "*" || allowed == origin {
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}
		} else {
			if len(allowedOrigins) > 0 {
				c.Header("Access-Control-Allow-Origin", allowedOrigins[0])
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func envURL(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
