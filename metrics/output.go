package metrics

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
)

// DatadogConfig describes configuration of metrics output to datadog
type DatadogConfig struct {
	Enabled bool
}

// LoggingConfig describes configuration of metrics output to log
type LoggingConfig struct {
	Enabled      bool
	PrintTags    bool
	ReduceValues bool
	PrintEmpty   bool
}

// OutputterConfig describes configuration of metrics output
type OutputterConfig struct {
	Datadog              *DatadogConfig
	Logging              *LoggingConfig
	FlushIntervalSeconds uint
}

// Outputter outputs metrics
type Outputter struct {
	config    *OutputterConfig
	providers []Provider

	quit chan bool
	wg   sync.WaitGroup
}

// NewOutputter creates new Outputter
func NewOutputter(config *OutputterConfig) *Outputter {
	return &Outputter{
		config:    config,
		providers: []Provider{},
	}
}

// AddProviders adds metric providers to metric group
func (outputter *Outputter) AddProviders(providers ...Provider) {
	outputter.providers = append(outputter.providers, providers...)
}

func (outputter *Outputter) sendToDatadog(metrics []*MetricUpdate) {
	body := datadog.MetricsPayload{Series: []datadog.Series{}}
	for _, metric := range metrics {
		if len(metric.Values) == 0 {
			continue
		}
		series := datadog.Series{
			Metric: metric.Name,
			Type:   datadog.PtrString(metricTypeToString(metric.Type)),
			Tags:   &metric.Tags,
			Points: [][]*float64{},
		}
		for _, value := range metric.Values {
			series.Points = append(series.Points, []*float64{
				datadog.PtrFloat64(float64(value.Timestamp.Unix())),
				datadog.PtrFloat64(value.Value),
			})
		}
		body.Series = append(body.Series, series)
	}
	_, r, err := datadog.NewAPIClient(datadog.NewConfiguration()).MetricsApi.SubmitMetrics(
		datadog.NewDefaultContext(context.Background()),
		body,
	)
	if err != nil {
		log.Printf("metrics: cant's submit to datadog [error: %v] [response: %v]", err, r)
	}
}

func (outputter *Outputter) logMetrics(metrics []*MetricUpdate) {
	var b bytes.Buffer

	fmt.Fprint(&b, "Metrics:\n")
	for _, metric := range metrics {
		if len(metric.Values) == 0 && !outputter.config.Logging.PrintEmpty {
			continue
		}
		fmt.Fprint(&b, metric.Name)
		if outputter.config.Logging.PrintTags {
			fmt.Fprintf(&b, "[%v]", strings.Join(metric.Tags, ","))
		}
		if outputter.config.Logging.ReduceValues {
			reducer := NewDefaultReducer(metric.Type)
			for _, value := range metric.Values {
				reducer.AddValue(value.Value)
			}
			value, _ := reducer.Reduce()
			fmt.Fprintf(&b, ": %v\n", value)
		} else {
			values := []string{}
			for _, value := range metric.Values {
				values = append(values, fmt.Sprint(value.Value))
			}
			fmt.Fprintf(&b, ": %v\n", strings.Join(values, "->"))
		}
	}

	log.Print(b.String())
}

// Flush outputs accumulated metrics
func (outputter *Outputter) Flush() {
	metrics := []*MetricUpdate{}
	for _, provider := range outputter.providers {
		metrics = append(metrics, provider.FlushMetrics()...)
	}

	if outputter.config.Logging.Enabled {
		outputter.logMetrics(metrics)
	}
	if outputter.config.Datadog.Enabled {
		outputter.sendToDatadog(metrics)
	}
}

// Start runs outputter in a separate goroutine
func (outputter *Outputter) Start() {
	outputter.quit = make(chan bool)
	go outputter.run()
}

func (outputter *Outputter) run() {
	outputter.wg.Add(1)
	defer outputter.wg.Done()
	for {
		select {
		case <-outputter.quit:
			return
		case <-time.After(time.Second * time.Duration(outputter.config.FlushIntervalSeconds)):
			outputter.Flush()
		}
	}
}

// Stop stops outputter goroutine
func (outputter *Outputter) Stop() {
	outputter.quit <- true
	outputter.wg.Wait()
}

func metricTypeToString(mtype MetricType) string {
	switch mtype {
	case GaugeMetric:
		return "gauge"
	case CountMetric:
		return "count"
	case RateMetric:
		return "rate"
	}
	return ""
}
