package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var redisClient *redis.Client

func main() {

	cleanup := initTracer()
	defer cleanup(context.Background())

	opt, err := redis.ParseURL(os.Getenv("REDIS_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}

	redisClient = redis.NewClient(opt)
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(otelgin.Middleware(serviceName))
	r.GET("/weather", HandleGetWeather)
	r.Run(os.Getenv("API_URL"))
}
