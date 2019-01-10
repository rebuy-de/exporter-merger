package cmd_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rebuy-de/exporter-merger/cmd"
	log "github.com/sirupsen/logrus"
)

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

	want := `# TYPE bar untyped
bar 4
# TYPE conflict untyped
conflict 2
conflict 5
# TYPE foo untyped
foo 1
# TYPE shared untyped
shared{meh="a"} 3
shared{meh="b"} 6
`
	have, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if want != string(have) {
		t.Error("Got wrong response.")
		t.Error("Want:")
		t.Error(want)
		t.Error("Have:")
		t.Error(string(have))
	}
}
