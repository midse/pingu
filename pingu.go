package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ping/ping"
)

const (
	TTL      = 128
	TIMEOUT  = 500 // milliseconds
	INTERVAL = 500 // milliseconds
	COUNT    = 1
)

type PingAddresses struct {
	Addresses []string `json:"addresses" binding:"required,lte=10,dive,ipv4"`
	Count     int      `json:"count" binding:"omitempty,min=1,lte=10"`
	Interval  int      `json:"interval" binding:"omitempty,min=1,lte=10000"`
	Timeout   int      `json:"timeout" binding:"omitempty,min=1,lte=10000"`
	TTL       int      `json:"ttl" binding:"omitempty,min=1,lte=128"`
}

type PingResult struct {
	Address string `json:"address"`
	Status  bool   `json:"status"`
}

type PingResults struct {
	Addresses []PingResult `json:"addresses"`
}

func pingAddress(address string, count int, interval int, timeout int, ttl int) (error, bool) {
	pinger, err := ping.NewPinger(address)

	if err != nil {
		return err, false
	}

	pinger.Count = count
	pinger.Interval = time.Millisecond * time.Duration(interval)
	pinger.Timeout = time.Millisecond * time.Duration(timeout)
	pinger.TTL = ttl

	if err := pinger.Run(); err != nil {
		return err, false
	}

	stats := pinger.Statistics()
	return nil, stats.PacketsRecv > 0
}

func pingAddresses(data PingAddresses) PingResults {
	channel := make(chan struct {
		string
		bool
	})

	var results PingResults
	results.Addresses = []PingResult{}

	for _, address := range data.Addresses {
		go func(address string, json PingAddresses) {
			var err error
			var status bool

			err, status = pingAddress(address, json.Count, json.Interval, json.Timeout, json.TTL)

			if err != nil {
				fmt.Println(err)
			}

			channel <- struct {
				string
				bool
			}{address, status}

		}(address, data)

	}

	for range data.Addresses {
		var result PingResult
		data := <-channel

		result.Address = data.string
		result.Status = data.bool
		results.Addresses = append(results.Addresses, result)
	}

	close(channel)

	return results
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	//router.SetTrustedProxies([]string{"10.1.2.3"})

	router.POST("/ping", func(c *gin.Context) {
		data := PingAddresses{
			TTL:      TTL,
			Timeout:  TIMEOUT,
			Interval: INTERVAL,
			Count:    COUNT,
		}

		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, pingAddresses(data))
	})
	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
