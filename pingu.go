package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ping/ping"
)

type PingAddress struct {
	Address string `json:"address" binding:"required"`
	Timeout int    `json:"timeout"`
}

type PingAddresses struct {
	Addresses []PingAddress `json:"addresses" binding:"required"`
}

type PingResult struct {
	Address string
	Status  bool
}

type PingResults struct {
	Addresses []PingResult
}

func pingString(address string) (error, bool) {
	pinger, err := ping.NewPinger(address)
	if err != nil {
		return err, false
	}
	pinger.Count = 3
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		panic(err)
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

		for _, jsonItem := range json.Addresses {
			var result PingResult

			result.Address = jsonItem.Address
			_, result.Status = pingString(jsonItem.Address)

			results.Addresses = append(results.Addresses, result)

		}

		c.JSON(http.StatusOK, results)
	})
	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
