package cmd

import (
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	prom "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	Exporters            []string
	ExportersHTTPTimeout int
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"RequestURI": r.RequestURI,
		"UserAgent":  r.UserAgent(),
	}).Debug("handling new request")
	h.Merge(w)
}

func (h Handler) Merge(w io.Writer) {
	mfs := map[string]*prom.MetricFamily{}

	responses := make([]map[string]*prom.MetricFamily, 1024)
	responsesMu := sync.Mutex{}
	httpClientTimeout := time.Second * time.Duration(h.ExportersHTTPTimeout)

	wg := sync.WaitGroup{}
	for _, url := range h.Exporters {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			log.WithField("url", u).Debug("getting remote metrics")
			httpClient := http.Client{Timeout: httpClientTimeout}
			resp, err := httpClient.Get(u)
			if err != nil {
				log.WithField("url", u).Errorf("HTTP connection failed: %v", err)
				return
			}
			defer resp.Body.Close()

			tp := new(expfmt.TextParser)
			part, err := tp.TextToMetricFamilies(resp.Body)
			if err != nil {
				log.WithField("url", u).Errorf("Parse response body to metrics: %v", err)
				return
			}
			responsesMu.Lock()
			responses = append(responses, part)
			responsesMu.Unlock()
		}(url)
	}
	wg.Wait()

	for _, part := range responses {
		for n, mf := range part {
			mfo, ok := mfs[n]
			if ok {
				mfo.Metric = append(mfo.Metric, mf.Metric...)
			} else {
				mfs[n] = mf
			}

		}
	}

	// Logic to merge metrics based on labels and types.
	for _, mf := range mfs {
		// Go through metrics of a metric family and merge them someplace else.
		merged_metrics_tmp := map[string]*prom.Metric{}

		for _, metric := range mf.GetMetric() {
			var labels strings.Builder

			// Construct label identifier.
			for _, label := range metric.GetLabel() {
				labels.WriteString(label.String())
			}

			key := labels.String()

			_, ok := merged_metrics_tmp[key]
			if !ok {
				// Simple case: metric doesn't exist yet, create it.
				merged_metrics_tmp[key] = metric
			} else {
				// Hard case: if the key already exists, update it.
				mergeMetricValues(merged_metrics_tmp[key], *metric)
			}
		}

		// Create a list of of those metrics.
		merged_metrics := []*prom.Metric{}
		for _, metric := range merged_metrics_tmp {
			merged_metrics = append(merged_metrics, metric)
		}

		// Replace metrics with merged variant.
		mf.Metric = merged_metrics
	}

	names := []string{}
	for name := range mfs {
		names = append(names, name)
	}
	sort.Strings(names)

	enc := expfmt.NewEncoder(w, expfmt.FmtText)
	for _, n := range names {
		err := enc.Encode(mfs[n])
		if err != nil {
			log.Error(err)
			return
		}
	}
}

func mergeMetricValues(dst *prom.Metric, src prom.Metric) {
	if dst.GetGauge() != nil && src.GetGauge() != nil {
		*(dst.GetGauge().Value) += src.GetGauge().GetValue()
	}
	if dst.GetCounter() != nil && src.GetCounter() != nil {
		*(dst.GetCounter().Value) += src.GetCounter().GetValue()
	}
	if dst.GetHistogram() != nil && src.GetHistogram() != nil {
		dst_buckets := dst.GetHistogram().GetBucket()
		src_buckets := src.GetHistogram().GetBucket()
		for i := range dst_buckets {
			*(dst_buckets[i].CumulativeCount) += src_buckets[i].GetCumulativeCount()
		}
	}
	if dst.GetSummary() != nil && src.GetSummary() != nil {
		*(dst.GetSummary().SampleCount) += src.GetSummary().GetSampleCount()
		*(dst.GetSummary().SampleSum) += src.GetSummary().GetSampleSum()
	}
}
