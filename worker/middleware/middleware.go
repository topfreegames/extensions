package middleware

import (
	"fmt"
	"time"

	workers "github.com/jrallison/go-workers"

	"github.com/topfreegames/extensions/middleware"
)

const metricName = "worker_response_time_milliseconds"

// ResponseTimeMetricsMiddleware struct encapsulating DDStatsD
type ResponseTimeMetricsMiddleware struct {
	DDStatsD *middleware.DogStatsD
}

// NewResponseTimeMetricsMiddleware returns a new ResponseTimeMetricsMiddleware
func NewResponseTimeMetricsMiddleware(ddStatsD *middleware.DogStatsD) *ResponseTimeMetricsMiddleware {
	return &ResponseTimeMetricsMiddleware{
		DDStatsD: ddStatsD,
	}
}

// Call intercepts a worker call
func (m *ResponseTimeMetricsMiddleware) Call(queue string, message *workers.Msg, next func() bool) (acknowledge bool) {
	startTime := time.Now()
	tags := []string{fmt.Sprintf("queue:%s", queue)}

	defer func() {
		timeElapsed := time.Since(startTime)

		if r := recover(); r != nil {
			tags = append(tags, "status:success")
			m.DDStatsD.Timing(metricName, timeElapsed, tags...)

			panic(r)
		} else {
			tags = append(tags, "status:failure")
			m.DDStatsD.Timing(metricName, timeElapsed, tags...)
		}
	}()

	acknowledge = next()
	return acknowledge
}
