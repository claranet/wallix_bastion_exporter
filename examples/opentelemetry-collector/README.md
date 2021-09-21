# OpenTelemetry Collector

The [agent_config.yaml](agent_config.yaml) contains a minimalist configuration of the OpenTelemetry Collector
to work with the Wallix Bastion Prometheus Exporter.

It leverages the [prometheusexec receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusexecreceiver)
to automatically run the exporter and configure its corresponding scraper.

If you prefer, you can run yourself this exporter binary, for example thanks to systemd and simply configure the OpenTelemetry Collector
to scrape it only by using [prometheus receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver)
instead.

Notice you can pass configuration as flags or environment variables. We prefer the last for `WALLIX_PASSWORD`.
It is also possible to install the configuration file next to the binary (in this example: `/etc/otel/collector/scripts/wallix/wallix_bastion_exporter/config.yaml`).
