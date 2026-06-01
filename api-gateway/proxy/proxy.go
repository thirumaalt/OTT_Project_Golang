package proxy

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// To returns a gin handler that reverse-proxies to the given base URL.
func To(baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the wildcard path from gin's *path param
		path := c.Param("path")
		target := strings.TrimRight(baseURL, "/") + path
		if q := c.Request.URL.RawQuery; q != "" {
			target += "?" + q
		}
		forward(c, target, false)
	}
}

// MediaProxy handles media routes with streaming support.
func MediaProxy(baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("path")
		target := strings.TrimRight(baseURL, "/") + path
		if q := c.Request.URL.RawQuery; q != "" {
			target += "?" + q
		}
		stream := strings.HasPrefix(path, "/stream")
		forward(c, target, stream)
	}
}

func forward(c *gin.Context, target string, stream bool) {
	log.Printf("Proxying %s %s", c.Request.Method, target)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	req, err := http.NewRequestWithContext(
		c.Request.Context(),
		c.Request.Method,
		target,
		strings.NewReader(string(body)),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to build upstream request",
		})
		return
	}

	otel.GetTextMapPropagator().Inject(
		c.Request.Context(),
		propagation.HeaderCarrier(req.Header),
	)

	// Copy headers, skip Host
	for k, vals := range c.Request.Header {
		if strings.ToLower(k) == "host" {
			continue
		}
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Upstream error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, vals := range resp.Header {
		for _, v := range vals {
			c.Header(k, v)
		}
	}
	c.Status(resp.StatusCode)

	if stream {
		// Stream chunks directly to client
		buf := make([]byte, 32*1024)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
					break
				}
				c.Writer.Flush()
			}
			if readErr != nil {
				break
			}
		}
		return
	}

	io.Copy(c.Writer, resp.Body)
}
