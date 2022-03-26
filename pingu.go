package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ping/ping"
)

type PingAddresses struct {
	Addresses []string `json:"addresses" binding:"required"`
	Timeout   int      `json:"timeout"`
	Count     int      `json:"count"`
	TTL       int      `json:"ttl"`
}

type PingResult struct {
	Address string `json:"address"`
	Status  bool   `json:"status"`
}

type PingResults struct {
	Addresses []PingResult `json:"addresses"`
}

func pingString(address string, count int, ttl int, timeout int) (error, bool) {
	pinger, err := ping.NewPinger(address)

	if err != nil {
		return err, false
	}

	pinger.Count = count
	pinger.Timeout = time.Millisecond * time.Duration(timeout)
	pinger.TTL = ttl

	if err := pinger.Run(); err != nil {
		return err, false
	}

	stats := pinger.Statistics()
	return nil, stats.PacketsRecv > 0
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	//router.SetTrustedProxies([]string{"10.1.2.3"})

	router.POST("/ping", func(c *gin.Context) {
		var json PingAddresses

		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var results PingResults
		results.Addresses = []PingResult{}

		if json.TTL == 0 {
			json.TTL = 128
		}

		if json.Timeout == 0 {
			json.Timeout = 500
		}

		if json.Count == 0 {
			json.Count = 1
		}

		for _, address := range json.Addresses {
			var result PingResult

			result.Address = address
			_, result.Status = pingString(address, json.Count, json.TTL, json.Timeout)

			results.Addresses = append(results.Addresses, result)

		}

		c.JSON(http.StatusOK, results)
	})
	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
