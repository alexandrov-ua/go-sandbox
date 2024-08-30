package main

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CorrelationIdKey = "correlation-id"

func StructuredLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		correlationId := uuid.New().String()
		c.Set(CorrelationIdKey, correlationId)
		c.Next()
		param := gin.LogFormatterParams{}

		param.TimeStamp = time.Now()
		param.Latency = param.TimeStamp.Sub(start)
		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		param.BodySize = c.Writer.Size()
		if raw != "" {
			path = path + "?" + raw
		}
		param.Path = path

		level := slog.LevelInfo
		msg := "ok"
		if c.Writer.Status() >= 500 {
			level = slog.LevelError
			msg = param.ErrorMessage
		}

		logger.Log(c.Request.Context(), level, msg,
			slog.Int("status_code", param.StatusCode),
			slog.String("client_ip", param.ClientIP),
			slog.String("path", param.Path),
			slog.String(CorrelationIdKey, correlationId),
			slog.String("latency", param.Latency.String()),
			slog.Int("body_size", param.BodySize))
	}
}
