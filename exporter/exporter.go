package exporter

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/claranet/wallix_bastion_exporter/config"
	"github.com/claranet/wallix_bastion_exporter/httpclient"
	"github.com/claranet/wallix_bastion_exporter/wallix"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// ony used for metrics based on past timeframe like the closed sessions.
	sessionsClosedMinutes = 5 // TODO expose as config parameter?
	// prometheus exporter Namespace.
	Namespace = "wallix_bastion"
)

var (
	metricUp = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "up"),
		"Was able to request and authenticate to Wallix Bastion API successfully.",
		nil, nil,
	)
	metricUsers = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "users"),
		"Current number of users.",
		nil, nil,
	)
	metricGroups = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "groups"),
		"Current number of groups.",
		nil, nil,
	)
	metricDevices = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "devices"),
		"Current number of devices.",
		nil, nil,
	)
	metricSessions = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "sessions"),
		fmt.Sprintf("Number of sessions for the last %dm.", sessionsClosedMinutes),
		[]string{"status"}, nil,
	)
	metricTargets = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "targets"),
		"Current number of targets.",
		[]string{"type"}, nil,
	)
	metricEncryptionStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "encryption_status"),
		"Encryption status (need_setup=0, ready=1, need_passphrase=2).",
		[]string{"status", "security_level"}, nil,
	)
	metricEncryptionSecurityLevel = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "encryption_security_level"),
		"Encryption security level (need_setup=0, passphrase_defined=1, passphrase_not_used=2, [hidden]=-1).",
		[]string{"security_level", "status"}, nil,
	)
	metricLicenseIsExpired = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "license_is_expired"),
		"Is the Wallix is expired (0=false, 1=true).",
		nil, nil,
	)
	metricLicensePrimaryPct = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "license_primary_ratio"),
		"License usage percentage of primary.",
		nil, nil,
	)
	metricLicenseSecondaryPct = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "license_secondary_ratio"),
		"License usage percentage of secondary.",
		nil, nil,
	)
	metricLicenseNameUserPct = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "license_named_user_ratio"),
		"License usage percentage of named user.",
		nil, nil,
	)
	metricLicenseResourcePct = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "license_resource_ratio"),
		"License usage percentage of resource.",
		nil, nil,
	)
	metricLicenseWaapmPct = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "license_waapm_ratio"),
		"License usage percentage of waapm.",
		nil, nil,
	)
	metricLicensePmTargetPct = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "license_sm_target_ratio"),
		"License usage percentage of pm target.",
		nil, nil,
	)
	metricLicenseSmTargetPct = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "license_pm_target_ratio"),
		"License usage percentage of sm target.",
		nil, nil,
	)
)

type Exporter struct {
	Config config.Config
}

func NewExporter(config config.Config) *Exporter {
	return &Exporter{
		Config: config,
	}
}

func (e *Exporter) Describe(metricsChannel chan<- *prometheus.Desc) {
	metricsChannel <- metricUp
	metricsChannel <- metricUsers
	metricsChannel <- metricGroups
	metricsChannel <- metricDevices
	metricsChannel <- metricSessions
	metricsChannel <- metricTargets
	metricsChannel <- metricEncryptionStatus
	metricsChannel <- metricEncryptionSecurityLevel
}

func (e *Exporter) Collect(metricsChannel chan<- prometheus.Metric) {
	httpConfig := httpclient.HTTPConfig{
		SkipVerify: e.Config.SkipVerify,
		Timeout:    e.Config.Timeout,
		Headers: map[string]string{
			"User-Agent": "prometheus_exporter_" + Namespace,
		},
		// Using a cookie speed up metrics fetch by avoiding basic auth on every requests
		CookieManager: true,
	}
	client, err := httpConfig.Build()
	if err != nil {
		log.Println(fmt.Errorf("init exporter failed: %w", err))

		return
	}

	err = e.AuthenticateWallixAPI(metricsChannel, client)
	if err != nil {
		log.Println(fmt.Errorf("determine up metric failed: %w", err))

		return
	}

	e.FetchWallixMetrics(metricsChannel, client)
}

// The first request done to wallix API. It allows to:
// - determine "up" metric for the exporter
// - prevent trying to fetch other metrics if down
// - retrieve the cookie to not have to authenticate subsequent requests
// Notice it uses "POST" methode in contrast to all other requests.
func (e *Exporter) AuthenticateWallixAPI(metricsChannel chan<- prometheus.Metric, client *http.Client) (err error) {
	err = wallix.Authenticate(
		client,
		e.Config.ScrapeURI,
		e.Config.WallixUsername,
		e.Config.WallixPassword,
	)
	if err != nil {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricUp, prometheus.GaugeValue, 0,
		)

		return fmt.Errorf("wallix authentication failed: %w", err)
	}

	metricsChannel <- prometheus.MustNewConstMetric(
		metricUp, prometheus.GaugeValue, 1,
	)

	return nil
}

// All other metrics fetched from the API essentially
// by counting the number of elements of list returned
// by different routes.
func (e *Exporter) FetchWallixMetrics(
	metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	var wg sync.WaitGroup

	wg.Add(1)
	go e.gatherMetricsUsers(&wg, metricsChannel, client)
	wg.Add(1)
	go e.gatherMetricsGroups(&wg, metricsChannel, client)
	wg.Add(1)
	go e.gatherMetricsDevices(&wg, metricsChannel, client)
	wg.Add(1)
	go e.gatherMetricsTargetsSessionAccounts(&wg, metricsChannel, client)
	wg.Add(1)
	go e.gatherMetricsTargetsSessionAccountMappings(&wg, metricsChannel, client)
	wg.Add(1)
	go e.gatherMetricsTargetsSessionInteractiveLogins(&wg, metricsChannel, client)
	wg.Add(1)
	go e.gatherMetricsTargetsSessionScenarioAccounts(&wg, metricsChannel, client)
	wg.Add(1)
	go e.gatherMetricsTargetsPasswordRetrievalAccounts(&wg, metricsChannel, client)
	wg.Add(1)
	go e.gatherMetricsEncryption(&wg, metricsChannel, client)
	wg.Add(1)
	go e.gatherMetricsLicense(&wg, metricsChannel, client)
	wg.Add(1)
	go e.gatherMetricsSessions(&wg, metricsChannel, client)

	wg.Wait()
}
