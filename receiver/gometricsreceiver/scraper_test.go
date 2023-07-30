// TODO
package gometricsreceiver

import (
	md "jba/work/metrics/metricdata"
	"reflect"
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestPopulateAttrs(t *testing.T) {
	attrs := []md.Attr{
		{"i", int64(7)},
		{"b", true},
		{"s", "str"},
	}
	pmap := pcommon.NewMap()
	if err := populateAttrs(pmap, attrs); err != nil {
		t.Fatal(err)
	}
	got := map[string]any{}
	pmap.Range(func(k string, v pcommon.Value) bool {
		got[k] = v.AsRaw()
		return true
	})
	want := map[string]any{}
	for _, a := range attrs {
		want[a.Key] = a.Value
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:  %v\nwant: %v", got, want)
	}
}

func TestPopulateMetric(t *testing.T) {
	for _, test := range []struct {
		gm    md.Metric
		wantf func(m pmetric.Metric)
	}{
		{
			gm: md.Metric{
				Name: "n", Description: "d", Unit: "u", IsSum: true,
				DataPoints: []md.DataPoint{{TimeUnixNano: 6, Value: 1.5}},
			},
			wantf: func(m pmetric.Metric) {
				m.SetName("n")
				m.SetDescription("d")
				m.SetUnit("u")
			},
		},
	} {
		got := pmetric.NewMetric()
		if err := populateMetric(got, test.gm, pcommon.Timestamp(11)); err != nil {
			t.Fatal(err)
		}
		want := pmetric.NewMetric()
		test.wantf(want)
		compareMetrics(t, got, want)
	}
}

func compareMetrics(t *testing.T, got, want pmetric.Metric) {
	t.Helper()
	if g, w := got.Name(), want.Name(); g != w {
		t.Errorf("name: got %q, want %q", g, w)
	}
	if g, w := got.Description(), want.Description(); g != w {
		t.Errorf("description: got %q, want %q", g, w)
	}
	if g, w := got.Unit(), want.Unit(); g != w {
		t.Errorf("unit: got %q, want %q", g, w)
	}
	t.Log("and more")
}
