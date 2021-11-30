package metrics

// SumReducer accumulates sum of metric values
type SumReducer struct {
	sum float64
}

// NewSumReducer creates new SumReducer
func NewSumReducer() *SumReducer {
	return &SumReducer{}
}

// AddValue adds value to reducer
func (r *SumReducer) AddValue(value float64) {
	r.sum += value
}

// Reduce returns reduced value of metric and resets the state
// If nothing to return - (0, false) will be returned
func (r *SumReducer) Reduce() (float64, bool) {
	res := r.sum
	r.sum = 0
	return res, true
}

// OverrideReducer keeps only last value of metric
type OverrideReducer struct {
	value    float64
	hasValue bool
}

// NewOverrideReducer creates new OverrideReducer
func NewOverrideReducer() *OverrideReducer {
	return &OverrideReducer{}
}

// AddValue adds value to reducer
func (r *OverrideReducer) AddValue(value float64) {
	r.value = value
	r.hasValue = true
}

// Reduce returns reduced value of metric and resets the state
// If nothing to return - (0, false) will be returned
func (r *OverrideReducer) Reduce() (float64, bool) {
	if !r.hasValue {
		return 0, false
	}
	r.hasValue = false
	return r.value, true
}

// AvgReducer will produce average value of metric
type AvgReducer struct {
	sum   float64
	count uint64
}

// NewAvgReducer creates new AvgReducer
func NewAvgReducer() *AvgReducer {
	return &AvgReducer{}
}

// AddValue adds value to reducer
func (r *AvgReducer) AddValue(value float64) {
	r.sum += value
	r.count++
}

// Reduce returns reduced value of metric and resets the state
// If nothing to return - (0, false) will be returned
func (r *AvgReducer) Reduce() (float64, bool) {
	if r.count == 0 {
		return 0, false
	}
	res := r.sum / float64(r.count)
	r.count = 0
	return res, true
}

// MinReducer will produce minimum value of metric
type MinReducer struct {
	min      float64
	hasValue bool
}

// NewMinReducer creates new MinReducer
func NewMinReducer() *MinReducer {
	return &MinReducer{}
}

// AddValue adds value to reducer
func (r *MinReducer) AddValue(value float64) {
	if !r.hasValue || value < r.min {
		r.min = value
	}
	r.hasValue = true
}

// Reduce returns reduced value of metric and resets the state
// If nothing to return - (0, false) will be returned
func (r *MinReducer) Reduce() (float64, bool) {
	if !r.hasValue {
		return 0, false
	}
	r.hasValue = false
	return r.min, true
}

// MaxReducer will produce maximum value of metric
type MaxReducer struct {
	max      float64
	hasValue bool
}

// NewMaxReducer creates new MaxReducer
func NewMaxReducer() *MaxReducer {
	return &MaxReducer{}
}

// AddValue adds value to reducer
func (r *MaxReducer) AddValue(value float64) {
	if !r.hasValue || value > r.max {
		r.max = value
	}
	r.hasValue = true
}

// Reduce returns reduced value of metric and resets the state
// If nothing to return - (0, false) will be returned
func (r *MaxReducer) Reduce() (float64, bool) {
	if !r.hasValue {
		return 0, false
	}
	r.hasValue = false
	return r.max, true
}

// NewDefaultReducer returns new Reducer which is default for given metric type
func NewDefaultReducer(mtype MetricType) Reducer {
	if mtype == CountMetric {
		return NewSumReducer()
	}
	return NewOverrideReducer()
}
