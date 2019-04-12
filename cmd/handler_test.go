package cmd_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/prometheus/common/expfmt"
	"github.com/rebuy-de/exporter-merger/cmd"
	log "github.com/sirupsen/logrus"
)

func Equal(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func testExporter(t testing.TB, content string) (string, func()) {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, content)
	}))

	return ts.URL, ts.Close
}

func TestHandler(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	te1, deferrer := testExporter(t,
		"foo{} 1\nconflict 2\nshared{meh=\"a\"} 3")
	defer deferrer()

	te2, deferrer := testExporter(t,
		"bar{} 4\nconflict 5\nshared{meh=\"b\"} 6")
	defer deferrer()

	exporters := []string{
		te1,
		te2,
	}

	server := httptest.NewServer(cmd.Handler{
		Exporters: exporters,
	})
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("Received non-200 response: %d\n", resp.StatusCode)
	}

	// 	want := `# TYPE bar untyped
	// bar 4
	// # TYPE conflict untyped
	// conflict 2
	// conflict 5
	// # TYPE foo untyped
	// foo 1
	// # TYPE shared untyped
	// shared{meh="a"} 3
	// shared{meh="b"} 6
	// `
	// have, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	eFmt := new(expfmt.TextParser)
	part, err := eFmt.TextToMetricFamilies(resp.Body)

	fooWanted := 1.0
	var foo float64

	barWanted := 4.0
	var bar float64

	var conflictWanted sort.Float64Slice = []float64{2.0, 5.0}
	var conflict sort.Float64Slice = make([]float64, 0)

	sharedWanted := map[string]float64{"a": 3.0, "b": 6.0}
	shared := make(map[string]float64)

	for n, mf := range part {
		if n == "bar" {
			bar = mf.GetMetric()[0].GetUntyped().GetValue()
		}

		if n == "foo" {
			foo = mf.GetMetric()[0].GetUntyped().GetValue()
		}

		if n == "conflict" {
			for _, metric := range mf.GetMetric() {
				conflict = append(conflict, metric.GetUntyped().GetValue())
			}
		}

		if n == "shared" {
			for _, metric := range mf.GetMetric() {
				label := metric.GetLabel()[0].GetValue()
				value := metric.GetUntyped().GetValue()
				shared[label] = value
			}
		}
	}

	if bar != barWanted {
		t.Errorf("bar is %f but wanted %f", bar, barWanted)
	}

	if foo != 1.0 {
		t.Errorf("foo is %f but wanted %f", foo, fooWanted)
	}

	conflictWanted.Sort()
	conflict.Sort()

	if !Equal(conflict, conflictWanted) {
		t.Errorf("conflict is %v but wanted %v", conflict, conflictWanted)
	}

	if !reflect.DeepEqual(shared, sharedWanted) {
		t.Errorf("shared is %v but wanted %v", shared, sharedWanted)
	}
}
