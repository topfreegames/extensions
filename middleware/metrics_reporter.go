package middleware

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/dogstatsd"
	"strconv"
)

// MetricTypes constants
var MetricTypes = struct {
	ResponseTimeMs string
}{
	ResponseTimeMs: "response_time_ms",
}

// MetricsReporter interface
type MetricsReporter interface {
	Distribution(metric string, value float64, tags ...string) error
	Gauge(metrics string, value float64, tags ...string) error
	Increment(metric string, tags ...string) error
}

// NewMetricsReporter ctor
func NewMetricsReporter(config *viper.Viper) (MetricsReporter, error) {
	return NewDogStatsD(config)
}

// DogStatsD metrics reporter struct
type DogStatsD struct {
	client     dogstatsd.Client
	rate       float64
	tagsPrefix string
}

func loadDefaultConfigsDogStatsD(config *viper.Viper) {
	config.SetDefault("extensions.dogstatsd.host", "localhost:8125")
	config.SetDefault("extensions.dogstatsd.prefix", "middleware_dev.")
	config.SetDefault("extensions.dogstatsd.tags_prefix", "")
	config.SetDefault("extensions.dogstatsd.rate", "1")
}

// NewDogStatsD ctor
func NewDogStatsD(config *viper.Viper) (*DogStatsD, error) {
	loadDefaultConfigsDogStatsD(config)
	host := config.GetString("extensions.dogstatsd.host")
	prefix := config.GetString("extensions.dogstatsd.prefix")
	tagsPrefix := config.GetString("extensions.dogstatsd.tags_prefix")
	rate, err := strconv.ParseFloat(config.GetString("extensions.dogstatsd.rate"), 64)
	if err != nil {
		return nil, err
	}
	c, err := dogstatsd.New(host, prefix)
	if err != nil {
		return nil, err
	}
	return &DogStatsD{
		client:     c,
		rate:       rate,
		tagsPrefix: tagsPrefix,
	}, nil
}

func prefixTags(prefix string, tags ...string) {
	if len(prefix) > 0 {
		for i, t := range tags {
			tags[i] = fmt.Sprintf("%s%s", prefix, t)
		}
	}
}

// Distribution reports distribution interval taken for something
func (d *DogStatsD) Distribution(
	metric string, value float64, tags ...string,
) error {
	prefixTags(d.tagsPrefix, tags...)
	return d.client.Distribution(metric, value, tags, d.rate)
}

// Gauge reports a numeric value that can go up or down
func (d *DogStatsD) Gauge(
	metric string, value float64, tags ...string,
) error {
	prefixTags(d.tagsPrefix, tags...)
	return d.client.Gauge(metric, value, tags, d.rate)
}

// Increment reports an increment to some metric
func (d *DogStatsD) Increment(metric string, tags ...string) error {
	prefixTags(d.tagsPrefix, tags...)
	return d.client.Incr(metric, tags, d.rate)
}
