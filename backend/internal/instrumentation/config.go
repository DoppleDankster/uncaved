package instrumentation

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
)

type Config struct {
	ServiceName    string `mapstructure:"service-name"`
	ServiceVersion string `mapstructure:"service-version"`
	Environment    string `mapstructure:"environment"`

	OTLPEndpoint string `mapstructure:"otlp-endpoint"`

	LogLevel       string        `mapstructure:"log-level"`
	TraceSampling  float64       `mapstructure:"trace-sampling"`
	MetricInterval time.Duration `mapstructure:"metric-interval"`
}

func (c Config) Validate() error {
	if c.OTLPEndpoint != "" {
		u, err := url.Parse(c.OTLPEndpoint)
		if err != nil {
			return fmt.Errorf("instrumentation: otlp-endpoint %q: %w", c.OTLPEndpoint, err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf(
				"instrumentation: otlp-endpoint %q must be an http or https url",
				c.OTLPEndpoint,
			)
		}
		if u.Host == "" {
			return fmt.Errorf("instrumentation: otlp-endpoint %q has no host", c.OTLPEndpoint)
		}
	}
	if !strings.EqualFold(c.LogLevel, "off") {
		if _, err := parseLevel(c.LogLevel); err != nil {
			return fmt.Errorf("instrumentation: log-level %q: %w", c.LogLevel, err)
		}
	}
	if c.TraceSampling < 0 || c.TraceSampling > 1 {
		return fmt.Errorf("instrumentation: trace-sampling %v must be in [0,1]", c.TraceSampling)
	}
	return nil
}

// otlpSignalURL joins the configured OTLP base endpoint with a signal path
// ("v1/traces", "v1/metrics", "v1/logs").
func (c Config) otlpSignalURL(signal string) string {
	u, _ := url.Parse(c.OTLPEndpoint)
	u.Path = path.Join("/", u.Path, signal)
	return u.String()
}
