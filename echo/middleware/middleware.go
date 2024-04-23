package middleware

import (
	"fmt"
	"time"

	"github.com/labstack/echo"
	"github.com/topfreegames/extensions/v9/middleware"
)

const metricName = "response_time_milliseconds"

// ResponseTimeMetricsMiddleware struct encapsulating DDStatsD
type ResponseTimeMetricsMiddleware struct {
	DDStatsD      *middleware.DogStatsD
	AddCustomTags func(echo.Context) []string
}

//ResponseTimeMetricsMiddleware is a middleware to measure the response time
//of a route and send it do StatsD
func (responseTimeMiddleware ResponseTimeMetricsMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		startTime := time.Now()
		gameID := c.Param("gameID")
		result := next(c)
		status := c.Response().Status()
		route := c.Path()
		method := c.Request().Method()
		timeUsed := time.Since(startTime)

		tags := []string{
			fmt.Sprintf("route:%s", route),
			fmt.Sprintf("method:%s", method),
			fmt.Sprintf("status:%d", status),
		}
		if gameID != "" {
			tags = append(tags, fmt.Sprintf("game:%s", gameID))
		}
		if responseTimeMiddleware.AddCustomTags != nil {
			customizedTags := responseTimeMiddleware.AddCustomTags(c)
			tags = append(tags, customizedTags...)
		}

		responseTimeMiddleware.DDStatsD.Distribution(metricName, float64(timeUsed), tags...)

		return result
	}
}

//ResponseTimeMetricsMiddleware returns a new ResponseTimeMetricsMiddleware
func NewResponseTimeMetricsMiddleware(ddStatsD *middleware.DogStatsD) *ResponseTimeMetricsMiddleware {
	return &ResponseTimeMetricsMiddleware{
		DDStatsD: ddStatsD,
	}
}
