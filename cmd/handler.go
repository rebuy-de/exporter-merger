package cmd

import (
	"io"
	"net/http"
	"sort"

	prom "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	Config Config
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"RequestURI": r.RequestURI,
		"UserAgent":  r.UserAgent(),
	}).Debug("handling new request")
	err := h.Merge(w)
	if err != nil {
		log.Error(err)
		w.WriteHeader(500)
	}
}

func (h Handler) Merge(w io.Writer) error {
	mfs := map[string]*prom.MetricFamily{}
	tp := new(expfmt.TextParser)

	for _, e := range h.Config.Exporters {
		resp, err := http.Get(e.URL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		part, err := tp.TextToMetricFamilies(resp.Body)
		if err != nil {
			return err
		}

		for n, mf := range part {
			mfo, ok := mfs[n]
			if ok {
				mfo.Metric = append(mfo.Metric, mf.Metric...)
			} else {
				mfs[n] = mf
			}

		}
	}

	names := []string{}
	for n := range mfs {
		names = append(names, n)
	}
	sort.Strings(names)

	enc := expfmt.NewEncoder(w, expfmt.FmtText)
	for _, n := range names {
		err := enc.Encode(mfs[n])
		if err != nil {
			return err
		}
	}

	return nil

}
