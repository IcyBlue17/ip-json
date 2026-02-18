package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ipCache    *cache
	httpClient *http.Client
	providers  []provider
)

func main() {
	port := getEnv("PORT", "8080")
	ttl := getDuration("CACHE_TTL", 10*time.Minute)
	maxSize := getInt("CACHE_MAX_SIZE", 10000)
	timeout := getDuration("PROVIDER_TIMEOUT", 3*time.Second)

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ipCache = newCache(ttl, maxSize)
	defer ipCache.close()

	httpClient = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	providers = []provider{p1{}, p2{}, p3{}, p4{}, p5{}, p6{}}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	r.GET("/", handleIP)
	r.GET("/:ip", handleIP)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func handleIP(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" || ip == "self" {
		ip = getClientIP(c)
	}
	ip = strings.TrimSpace(strings.Trim(ip, "[]"))

	if !validIP(ip) {
		c.JSON(400, gin.H{"error": "invalid IP address", "ip": ip})
		return
	}
	if !publicIP(ip) {
		c.JSON(422, gin.H{"error": "private or reserved IP address", "ip": ip})
		return
	}

	if info, ok := ipCache.get(ip); ok {
		c.JSON(200, info)
		return
	}

	info, err := lookupAll(c.Request.Context(), ip, providers, httpClient)
	if err != nil {
		c.JSON(502, gin.H{"error": "all upstream providers failed", "ip": ip})
		ipCache.setTTL(ip, nil, 60*time.Second)
		return
	}

	ipCache.set(ip, info)
	c.JSON(200, info)
}

func getClientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		if s := strings.TrimSpace(strings.Split(xff, ",")[0]); s != "" {
			return s
		}
	}
	if v := c.GetHeader("X-Real-IP"); v != "" {
		return strings.TrimSpace(v)
	}
	if v := c.GetHeader("CF-Connecting-IP"); v != "" {
		return strings.TrimSpace(v)
	}
	return c.ClientIP()
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func validIP(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' || c == ':' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			continue
		}
		return false
	}
	return s != "" && (strings.Contains(s, ".") || strings.Contains(s, ":"))
}

func publicIP(ip string) bool {
	p := net.ParseIP(ip)
	if p == nil {
		return false
	}
	if p.IsLoopback() || p.IsPrivate() || p.IsUnspecified() ||
		p.IsMulticast() || p.IsLinkLocalUnicast() || p.IsLinkLocalMulticast() {
		return false
	}
	if v4 := p.To4(); v4 != nil && v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
		return false
	}
	return true
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return def
}
