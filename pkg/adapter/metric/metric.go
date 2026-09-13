package metric

import "go.opentelemetry.io/otel/metric"

var (
	// Request metrics
	RequestCounter  metric.Int64Counter
	RequestDuration metric.Float64Histogram
	RequestInFlight metric.Int64UpDownCounter

	// Business metrics
	LetterCounter   metric.Int64Counter
	WordLengthGauge metric.Int64Gauge
)

func Initialize(meter metric.Meter) error {
	var err error

	RequestCounter, err = meter.Int64Counter(
		"request_total",
		metric.WithDescription("Total number of requests processed"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return err
	}

	RequestDuration, err = meter.Float64Histogram(
		"request_duration_seconds",
		metric.WithDescription("Duration of requests"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	RequestInFlight, err = meter.Int64UpDownCounter(
		"requests_in_flight",
		metric.WithDescription("Current number of requests being processed"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return err
	}

	LetterCounter, err = meter.Int64Counter(
		"letter_concatenations_total",
		metric.WithDescription("Total number of letter concatenations"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return err
	}

	WordLengthGauge, err = meter.Int64Gauge(
		"word_length",
		metric.WithDescription("Current length of the word"),
		metric.WithUnit("chars"),
	)
	if err != nil {
		return err
	}

	return nil
}
