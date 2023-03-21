package main

import (
	"database/sql"
	"log"
	"net"
	u "net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/PolicyPuma4/gort/internal/db"
	"github.com/PolicyPuma4/gort/internal/generate"

	"github.com/gin-gonic/gin"
)

var (
	baseUrl u.URL
	token   string
)

func authMiddleware(c *gin.Context) {
	t := c.Request.Header.Get("TOKEN")
	if t != token {
		c.AbortWithStatus(401)
		return
	}

	c.Next()
}

func createEndpoint(c *gin.Context) {
	longUrl := c.Request.Header.Get("LONG_URL")
	if longUrl == "" {
		c.AbortWithStatus(400)
		return
	}

	code, err := generate.NewCode()
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(500)
		return
	}

	statement, err := db.DB.Prepare("INSERT INTO shorts (code, url, timestamp, ip) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(500)
		return
	}
	defer func(statement *sql.Stmt) {
		err := statement.Close()
		if err != nil {
			log.Println(err)
		}
	}(statement)

	short := db.Short{
		Code:      code,
		Url:       longUrl,
		Timestamp: time.Now().UTC(),
		Ip: func() (ip string) {
			ip = c.Request.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = c.ClientIP()
			}

			return
		}(),
	}
	_, err = statement.Exec(short.Code, short.Url, short.Timestamp, short.Ip)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(500)
		return
	}

	baseUrl.Path = code
	c.Header("SHORT_URL", baseUrl.String())
	c.Writer.WriteHeader(200)
}

func deleteEndpoint(c *gin.Context) {
	shortUrl := c.Request.Header.Get("SHORT_URL")
	if shortUrl == "" {
		c.AbortWithStatus(400)
		return
	}

	url, err := u.Parse(shortUrl)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(400)
		return
	}

	code := strings.TrimPrefix(url.Path, "/")

	statement, err := db.DB.Prepare("DELETE FROM shorts WHERE code = ?")
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(500)
		return
	}
	defer func(statement *sql.Stmt) {
		err := statement.Close()
		if err != nil {
			log.Println(err)
		}
	}(statement)

	_, err = statement.Exec(code)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(500)
		return
	}

	c.Writer.WriteHeader(200)
}

func rootEndpoint(c *gin.Context) {
	userAgent := c.Request.Header.Get("User-Agent")
	if userAgent == "" {
		c.AbortWithStatus(400)
		return
	}
	ip := c.Request.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.ClientIP()
	}

	code := c.Param("code")

	var longUrl string
	err := db.DB.QueryRow("SELECT url FROM shorts WHERE code = ?", code).Scan(&longUrl)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Redirect(301, "https://youtu.be/dQw4w9WgXcQ")
			return
		}

		log.Println(err)
		c.AbortWithStatus(500)
		return
	}

	statement, err := db.DB.Prepare("INSERT INTO visits (code, timestamp, ip, user_agent) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(500)
		return
	}
	defer func(statement *sql.Stmt) {
		err := statement.Close()
		if err != nil {
			log.Println(err)
		}
	}(statement)

	visit := db.Visit{
		Code:      code,
		Timestamp: time.Now().UTC(),
		Ip:        ip,
		UserAgent: userAgent,
	}
	_, err = statement.Exec(visit.Code, visit.Timestamp, visit.Ip, visit.UserAgent)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(500)
		return
	}

	c.Redirect(301, longUrl)
}

func main() {
	baseUrlScheme := os.Getenv("BASE_URL_SCHEME")
	baseUrlHost := os.Getenv("BASE_URL_HOST")
	baseUrlPort := os.Getenv("BASE_URL_PORT")
	baseUrl = u.URL{
		Scheme: baseUrlScheme,
		Host: func() (host string) {
			host = baseUrlHost
			if baseUrlPort == "" {
				return
			}

			host = net.JoinHostPort(baseUrlHost, baseUrlPort)
			return
		}(),
	}
	token = os.Getenv("TOKEN")
	path := os.Getenv("DATABASE_PATH")
	if path == "" {
		path = "./database.sqlite"
	}

	err := db.Connect(path)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := db.DB.Close()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		os.Exit(0)
	}()

	r := gin.Default()

	r.GET("/:code", rootEndpoint)

	v1 := r.Group("/api/v1")

	v1.Use(authMiddleware)
	{
		v1.POST("", createEndpoint)
		v1.DELETE("", deleteEndpoint)
	}

	_ = r.Run(":3000")
}
