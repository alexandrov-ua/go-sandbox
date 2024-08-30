package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {

	cleanup := initTracer()
	defer cleanup(context.Background())

	opt, err := redis.ParseURL(os.Getenv("REDIS_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(opt)
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		panic(err)
	}

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(StructuredLogger(logger))
	r.Use(otelgin.Middleware(serviceName))
	r.GET("/weather", func(ctx *gin.Context) {
		controller := WeatherController{
			redisClient: redisClient,
			httpClient:  &httpClient,
			context:     ctx.Request.Context(),
			logger:      CreateLoggerFromGinContext(logger, ctx),
		}
		controller.HandleGetWeather(ctx)
	})
	r.Run(os.Getenv("API_URL"))
}

func CreateLoggerFromGinContext(parent *slog.Logger, ctx *gin.Context) *slog.Logger {
	return parent.With(
		slog.String("marhod", ctx.Request.RequestURI),
		slog.String(CorrelationIdKey, ctx.GetString(CorrelationIdKey)),
	)
}
