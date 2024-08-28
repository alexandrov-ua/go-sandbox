package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func main() {

	opt, err := redis.ParseURL(os.Getenv("REDIS_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}

	redisClient = redis.NewClient(opt)

	r := gin.Default()
	r.GET("/weather", HandleGetWeather)
	r.Run(os.Getenv("API_URL"))
}
