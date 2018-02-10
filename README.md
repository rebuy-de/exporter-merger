# exporter-merger

[![Build Status](https://travis-ci.org/rebuy-de/exporter-merger.svg?branch=master)](https://travis-ci.org/rebuy-de/exporter-merger)
[![license](https://img.shields.io/github/license/rebuy-de/exporter-merger.svg)]()
[![GitHub release](https://img.shields.io/github/release/rebuy-de/exporter-merger.svg)]()

Merges Prometheus metrics from multiple sources.

> **Development Status** *exporter-merger* is in an early development phase.
> Expect incompatible changes and abandoment at any time.

## But Why?!

> [prometheus/prometheus#3756](https://github.com/prometheus/prometheus/issues/3756)

## Usage

*exporter-merger* needs a configuration file. Currently, nothing but URLs are accepted:

```yaml
exporters:
- url: http://localhost:9100/metrics
- url: http://localhost:9101/metrics
```

To start the exporter:

```
exporter-merger --config-path merger.yaml --listen-port 8080
```

## Planned Features

* Allow transforming of metrics from backend exporters.
  * eg add a prefix to the metric names
  * eg add labels to the metrics
* Allow dynamic adding of exporters.
