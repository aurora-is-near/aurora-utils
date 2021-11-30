package metrics

import (
	"sync"
	"time"
)

// MetricType describes type of metric
type MetricType uint

// MetricType's
const (
	GaugeMetric MetricType = iota
	CountMetric
	RateMetric
)

// MetricValue contains value of metric at some time point
type MetricValue struct {
	Timestamp time.Time
	Value     float64
}

// MetricUpdate contains sequence of values of some metric
type MetricUpdate struct {
	Name   string
	Type   MetricType
	Tags   []string
	Values []*MetricValue
}

// Provider is an object which provides metrics
type Provider interface {
	FlushMetrics() []*MetricUpdate
}

// Consumer is an object which consumes metrics
type Consumer interface {
	AddProviders(providers ...Provider)
}

// Reducer reduces metric values to minimize data sizes
type Reducer interface {
	AddValue(value float64)
	Reduce() (float64, bool)
}

// Metric describes metric and contains its values
type Metric struct {
	name  string
	mtype MetricType
	tags  []string

	reducer                Reducer
	reduceIntervalDuration *time.Duration
	reduceIntervalStart    *time.Time

	mutex  sync.Mutex
	values []*MetricValue
}

// NewMetric creates new Metric
func NewMetric(consumer Consumer, name string, mtype MetricType, tags ...string) *Metric {
	metric := &Metric{
		name:   name,
		mtype:  mtype,
		tags:   tags,
		values: []*MetricValue{},
	}
	if consumer != nil {
		consumer.AddProviders(metric)
	}
	return metric
}

func (metric *Metric) flushReducer() {
	if value, ok := metric.reducer.Reduce(); ok {
		metric.values = append(metric.values, &MetricValue{time.Now(), value})
	}
	metric.reduceIntervalStart = nil
}

// SetReducer sets reducer for metric
func (metric *Metric) SetReducer(reducer Reducer, reduceIntervalDuration time.Duration) {
	metric.mutex.Lock()
	defer metric.mutex.Unlock()
	if metric.reduceIntervalStart != nil {
		metric.flushReducer()
	}
	metric.reducer = reducer
	metric.reduceIntervalDuration = &reduceIntervalDuration
}

// NewAutoreducableMetric creates new Metric and sets default reducer for that metric type
func NewAutoreducableMetric(consumer Consumer, name string, mtype MetricType, tags ...string) *Metric {
	metric := NewMetric(consumer, name, mtype, tags...)
	metric.SetReducer(NewDefaultReducer(mtype), time.Second)
	return metric
}

// AddValue adds value for metric
func (metric *Metric) AddValue(value float64) {
	metric.mutex.Lock()
	defer metric.mutex.Unlock()
	if metric.reduceIntervalDuration == nil {
		metric.values = append(metric.values, &MetricValue{time.Now(), value})
		return
	}
	if metric.reduceIntervalStart == nil {
		t := time.Now()
		metric.reduceIntervalStart = &t
	} else if time.Since(*metric.reduceIntervalStart) > *metric.reduceIntervalDuration {
		metric.flushReducer()
	}
	metric.reducer.AddValue(value)
}

// FlushMetrics flushes and returns accumulated metrics
func (metric *Metric) FlushMetrics() []*MetricUpdate {
	metric.mutex.Lock()
	defer metric.mutex.Unlock()
	if metric.reduceIntervalStart != nil {
		metric.flushReducer()
	}
	if len(metric.values) == 0 && metric.mtype == CountMetric {
		metric.values = append(metric.values, &MetricValue{time.Now(), 0})
	}
	update := []*MetricUpdate{{
		Name:   metric.name,
		Type:   metric.mtype,
		Tags:   append([]string{}, metric.tags...),
		Values: metric.values,
	}}
	metric.values = []*MetricValue{}
	return update
}
