// TODO
package gometricsreceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	md "jba/work/metrics/metricdata"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
)

type goMetrics struct {
	Version string
	Metrics []md.Metric
}

type goMetricsScraper struct {
	cfg       *Config
	settings  *receiver.CreateSettings
	client    *http.Client
	startTime pcommon.Timestamp // start time that will be applied to all recorded data points.
}

func newGoMetricsScraper(cfg *Config, settings receiver.CreateSettings) *goMetricsScraper {
	return &goMetricsScraper{
		cfg:       cfg,
		settings:  &settings,
		startTime: pcommon.NewTimestampFromTime(time.Now()),
	}
}

func (e *goMetricsScraper) start(_ context.Context, host component.Host) error {
	client, err := e.cfg.HTTPClientSettings.ToClient(host, e.settings.TelemetrySettings)
	if err != nil {
		return err
	}
	e.client = client
	return nil
}

func (e *goMetricsScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	emptyMetrics := pmetric.NewMetrics()
	req, err := http.NewRequestWithContext(ctx, "GET", e.cfg.Endpoint, nil)
	if err != nil {
		return emptyMetrics, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return emptyMetrics, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return emptyMetrics, fmt.Errorf("expected 200 but received %d status code", resp.StatusCode)
	}

	var gms goMetrics
	if err := json.NewDecoder(resp.Body).Decode(&gms); err != nil {
		return emptyMetrics, fmt.Errorf("could not decode response body to JSON: %w", err)
	}
	wantVersion := "1.0.0"
	if gms.Version != wantVersion {
		return emptyMetrics, fmt.Errorf("got version %q, want %q", gms.Version, wantVersion)
	}

	return e.processMetrics(&gms)
}

func (s *goMetricsScraper) processMetrics(gms *goMetrics) (pmetric.Metrics, error) {
	result := pmetric.NewMetrics() // TODO: why do docs say I should only do this in tests?
	rm := pmetric.NewResourceMetrics()
	ils := rm.ScopeMetrics().AppendEmpty()
	ils.Scope().SetName("otelcol/gometricsreceiver")
	ils.Scope().SetVersion(s.settings.BuildInfo.Version)

	for _, gm := range gms.Metrics {
		if len(gm.DataPoints) == 0 {
			continue
		}
		pm := ils.Metrics().AppendEmpty()
		if err := populateMetric(pm, gm, s.startTime); err != nil {
			return result, err
		}
	}

	if ils.Metrics().Len() > 0 {
		rm.MoveTo(result.ResourceMetrics().AppendEmpty())
	}
	return result, nil
}

func populateMetric(dest pmetric.Metric, src md.Metric, startTime pcommon.Timestamp) error {
	dest.SetName(src.Name)
	dest.SetDescription(src.Description)
	dest.SetUnit(src.Unit)
	if src.DataPoints[0].Bounds != nil {
		// histogram
		h := dest.SetEmptyHistogram()
		h.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		dps := h.DataPoints()
		dps.EnsureCapacity(len(src.DataPoints))
		for _, sdp := range src.DataPoints {
			pdp := dps.AppendEmpty()
			pdp.SetStartTimestamp(startTime)
			pdp.SetTimestamp(pcommon.Timestamp(sdp.TimeUnixNano))
			pdp.BucketCounts().FromRaw(sdp.Counts)
			pdp.ExplicitBounds().FromRaw(sdp.Bounds)
			if err := populateAttrs(pdp.Attributes(), sdp.Attrs); err != nil {
				return err
			}
		}
	} else {
		// scalar
		var dps pmetric.NumberDataPointSlice
		if src.IsSum {
			sum := dest.SetEmptySum()
			sum.SetIsMonotonic(src.IsMonotonic)
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			dps = sum.DataPoints()
		} else {
			g := dest.SetEmptyGauge()
			dps = g.DataPoints()
		}
		dps.EnsureCapacity(len(src.DataPoints))
		for _, gdp := range src.DataPoints {
			pdp := dps.AppendEmpty()
			pdp.SetStartTimestamp(startTime)
			pdp.SetTimestamp(pcommon.Timestamp(gdp.TimeUnixNano))
			if src.IsInt {
				pdp.SetIntValue(int64(gdp.Value))
			} else {
				pdp.SetDoubleValue(gdp.Value)
			}
			if err := populateAttrs(pdp.Attributes(), gdp.Attrs); err != nil {
				return err
			}
		}
	}
	return nil
}

func populateAttrs(pmap pcommon.Map, attrs []md.Attr) error {
	pmap.EnsureCapacity(len(attrs))
	for _, a := range attrs {
		mv := pmap.PutEmpty(a.Key)
		if err := mv.FromRaw(a.Value); err != nil {
			return fmt.Errorf("attribute %q: %v", a.Key, err)
		}
	}
	return nil
}
