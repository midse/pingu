package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ping/ping"
	"github.com/spf13/viper"
)

const (
	TTL        = 128
	TIMEOUT    = 1000 // milliseconds
	INTERVAL   = 1000 // milliseconds
	COUNT      = 1
	ENV_PREFIX = "PINGU"
)

var config Config

type Config struct {
	Address    string
	User       string
	Password   string
	Privileged bool
}

func (*Config) init() {
	viper.SetEnvPrefix(ENV_PREFIX)

	viper.BindEnv("address")
	viper.SetDefault("address", "0.0.0.0:8080")

	viper.BindEnv("user")
	viper.SetDefault("user", "pingu")

	viper.BindEnv("password")

	viper.BindEnv("privileged")
	viper.SetDefault("privileged", true)

	viper.AutomaticEnv()

	viper.Unmarshal(&config)

	if config.Password == "" {
		panic(ENV_PREFIX + "_PASSWORD is required!")
	}
}

type PingAddresses struct {
	Addresses []string `json:"addresses" binding:"required,min=1,lte=10,dive,ipv4"`
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

func pingAddress(address string, count int, interval int, timeout int, ttl int) (bool, error) {
	pinger, err := ping.NewPinger(address)

	if err != nil {
		return false, err
	}

	pinger.Count = count
	pinger.Interval = time.Millisecond * time.Duration(interval)
	pinger.Timeout = time.Millisecond * time.Duration(timeout)
	pinger.TTL = ttl

	pinger.SetPrivileged(config.Privileged)

	if err := pinger.Run(); err != nil {
		return false, err
	}

	stats := pinger.Statistics()
	return stats.PacketsRecv > 0, nil
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

			status, err = pingAddress(address, json.Count, json.Interval, json.Timeout, json.TTL)

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

		// Read data from channel
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

	// Load configuration from env vars
	config.init()

	status, err := pingAddress("127.0.0.1", COUNT, INTERVAL, TIMEOUT, TTL)
	if !status {
		panic("Pingu is unable to ping. " + err.Error())
	}

	router := gin.Default()

	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{config.User: config.Password}))

	authorized.POST("/ping", func(c *gin.Context) {
		// Set default values before binding from json data
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

	router.Run(config.Address)
}
