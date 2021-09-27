package exporter

import (
	"fmt"
	"log"
	"net/http"

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
)

type Exporter struct {
	Config config.Config
}

func NewExporter(config config.Config) *Exporter {
	return &Exporter{
		Config: config,
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- metricUp
	ch <- metricUsers
	ch <- metricGroups
	ch <- metricDevices
	ch <- metricSessions
	ch <- metricTargets
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
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

	err = e.AuthenticateWallixAPI(ch, client)
	if err != nil {
		log.Println(fmt.Errorf("determine up metric failed: %w", err))

		return
	}

	err = e.FetchWallixMetrics(ch, client)
	if err != nil {
		log.Println(fmt.Errorf("fetch wallix metrics failed: %w", err))

		return
	}
}

// The first request done to wallix API. It allows to:
// - determine "up" metric for the exporter
// - prevent trying to fetch other metrics if down
// - retrieve the cookie to not have to authenticate subsequent requests
// Notice it uses "POST" methode in contrast to all other requests.
func (e *Exporter) AuthenticateWallixAPI(ch chan<- prometheus.Metric, client *http.Client) (err error) {
	_, _, err = wallix.DoRequest(
		client,
		http.MethodPost,
		e.Config.ScrapeURI,
		nil,
		&wallix.BasicAuth{
			Username: e.Config.WallixUsername,
			Password: e.Config.WallixPassword,
		},
	)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			metricUp, prometheus.GaugeValue, 0,
		)

		return fmt.Errorf("wallix authentication failed: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		metricUp, prometheus.GaugeValue, 1,
	)

	return nil
}

// All other metrics fetched from the API essentially
// by counting the number of elements of list returned
// by different routes.
func (e *Exporter) FetchWallixMetrics(
	ch chan<- prometheus.Metric, client *http.Client,
) (err error) {
	users, err := wallix.GetUsers(client, e.Config.ScrapeURI)
	if err != nil {
		return fmt.Errorf("cannot get users: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		metricUsers, prometheus.GaugeValue, float64(len(users)),
	)

	groups, err := wallix.GetGroups(client, e.Config.ScrapeURI)
	if err != nil {
		return fmt.Errorf("cannot get groups: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		metricGroups, prometheus.GaugeValue, float64(len(groups)),
	)

	devices, err := wallix.GetDevices(client, e.Config.ScrapeURI)
	if err != nil {
		return fmt.Errorf("cannot get devices: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		metricDevices, prometheus.GaugeValue, float64(len(devices)),
	)

	targetsSessionAccounts, err := wallix.GetTargets(client, e.Config.ScrapeURI, "session_accounts")
	if err != nil {
		return fmt.Errorf("cannot get session accounts targets: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		metricTargets, prometheus.GaugeValue, float64(len(targetsSessionAccounts)), "session_accounts",
	)

	targetsSessionAccountMappings, err := wallix.GetTargets(client, e.Config.ScrapeURI, "session_account_mappings")
	if err != nil {
		return fmt.Errorf("cannot get session account mappings targets: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		metricTargets, prometheus.GaugeValue, float64(len(targetsSessionAccountMappings)), "session_account_mappings",
	)

	targetsSessionInteractiveLogins, err := wallix.GetTargets(client, e.Config.ScrapeURI, "session_interactive_logins")
	if err != nil {
		return fmt.Errorf("cannot get session interactive logins targets: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		metricTargets, prometheus.GaugeValue, float64(len(targetsSessionInteractiveLogins)), "session_interactive_logins",
	)

	targetsSessionsScenarioAccounts, err := wallix.GetTargets(client, e.Config.ScrapeURI, "session_scenario_accounts")
	if err != nil {
		return fmt.Errorf("cannot get session scenario accounts targets: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		metricTargets, prometheus.GaugeValue, float64(len(targetsSessionsScenarioAccounts)), "session_scenario_accounts",
	)

	targetsPasswordRetrievalAccounts, err := wallix.GetTargets(client, e.Config.ScrapeURI, "password_retrieval_accounts")
	if err != nil {
		return fmt.Errorf("cannot get password retrieval accounts targets: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		metricTargets, prometheus.GaugeValue, float64(len(targetsPasswordRetrievalAccounts)), "password_retrieval_accounts",
	)

	// Keep sessions relative metrics fetch to the end because it depends on a active wallix license
	sessionsCurrent, err := wallix.GetCurrentSessions(client, e.Config.ScrapeURI)
	if err != nil {
		return fmt.Errorf("cannot get current sessions: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		metricSessions, prometheus.GaugeValue, float64(len(sessionsCurrent)), "current",
	)

	sessionsClosed, err := wallix.GetClosedSessions(client, e.Config.ScrapeURI, sessionsClosedMinutes)
	if err != nil {
		return fmt.Errorf("cannot get closed sessions: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		metricSessions, prometheus.GaugeValue, float64(len(sessionsClosed)), "closed",
	)

	// ch <- prometheus.MustNewConstMetric(
	// 	sessions, prometheus.GaugeValue, float64(
	// 		len(sessionsClosedResults)+len(sessionsCurrentResults),
	// 	),
	// )

	return nil
}
