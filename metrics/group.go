package metrics

// Group aggregates multiple metrics with common name prefix and tags
type Group struct {
	namePrefix string
	tags       []string
	providers  []Provider
}

// NewGroup creates new group of metrics
func NewGroup(consumer Consumer, namePrefix string, tags ...string) *Group {
	group := &Group{
		namePrefix: namePrefix,
		tags:       tags,
		providers:  []Provider{},
	}
	if consumer != nil {
		consumer.AddProviders(group)
	}
	return group
}

// AddProviders adds metric providers to metric group
func (group *Group) AddProviders(providers ...Provider) {
	group.providers = append(group.providers, providers...)
}

// FlushMetrics flushes and returns accumulated metrics
func (group *Group) FlushMetrics() []*MetricUpdate {
	metrics := []*MetricUpdate{}
	for _, provider := range group.providers {
		for _, metric := range provider.FlushMetrics() {
			metric.Name = group.namePrefix + metric.Name
			metric.Tags = append(append([]string{}, group.tags...), metric.Tags...)
			metrics = append(metrics, metric)
		}
	}
	return metrics
}
