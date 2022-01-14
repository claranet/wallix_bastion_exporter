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

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- metricUp
	ch <- metricUsers
	ch <- metricGroups
	ch <- metricDevices
	ch <- metricSessions
	ch <- metricTargets
	ch <- metricEncryptionStatus
	ch <- metricEncryptionSecurityLevel
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
	err = wallix.Authenticate(
		client,
		e.Config.ScrapeURI,
		e.Config.WallixUsername,
		e.Config.WallixPassword,
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

	encryptionMap := map[string]int{
		"ready":               1,
		"need_setup":          0,
		"need_passphrase":     2, // nolint:gomnd
		"passphrase_not_used": 2, // nolint:gomnd
		"passphrase_defined":  1,
		"[hidden]":            -1,
	}
	encryptionInfo, err := wallix.GetEncryption(client, e.Config.ScrapeURI)
	if err != nil {
		return fmt.Errorf("cannot get encryption information: %w", err)
	}
	encryptionStatus, ok := encryptionInfo["encryption"].(string)
	if ok {
		encryptionSecurityLevel, ok := encryptionInfo["security_level"].(string)
		if ok {
			ch <- prometheus.MustNewConstMetric(
				metricEncryptionStatus,
				prometheus.GaugeValue,
				float64(encryptionMap[encryptionStatus]),
				encryptionStatus, encryptionSecurityLevel,
			)
			ch <- prometheus.MustNewConstMetric(
				metricEncryptionSecurityLevel,
				prometheus.GaugeValue,
				float64(encryptionMap[encryptionSecurityLevel]),
				encryptionSecurityLevel, encryptionStatus,
			)
		}
	}

	licenseInfo, err := wallix.GetLicense(client, e.Config.ScrapeURI)
	if err != nil {
		return fmt.Errorf("cannot get license information: %w", err)
	}
	licenseIsExpired, ok := licenseInfo["is_expired"].(bool)
	if ok {
		var licenseIsExpiredGauge int8
		if licenseIsExpired {
			licenseIsExpiredGauge = 1
		}
		ch <- prometheus.MustNewConstMetric(
			metricLicenseIsExpired, prometheus.GaugeValue, float64(licenseIsExpiredGauge),
		)
	}
	licensePrimary, ok := licenseInfo["primary"].(float64)
	if !ok {
		licensePrimary = 0
	}
	licensePrimaryMax, ok := licenseInfo["primary_max"].(float64)
	if ok {
		licensePrimaryPct := licensePrimary / licensePrimaryMax
		ch <- prometheus.MustNewConstMetric(
			metricLicensePrimaryPct, prometheus.GaugeValue, licensePrimaryPct,
		)
	}
	licenseSecondary, ok := licenseInfo["secondary"].(float64)
	if !ok {
		licenseSecondary = 0
	}
	licenseSecondaryMax, ok := licenseInfo["secondary_max"].(float64)
	if ok {
		licenseSecondaryPct := licenseSecondary / licenseSecondaryMax
		ch <- prometheus.MustNewConstMetric(
			metricLicenseSecondaryPct, prometheus.GaugeValue, licenseSecondaryPct,
		)
	}
	licenseNamedUser, ok := licenseInfo["named_user"].(float64)
	if !ok {
		licenseNamedUser = 0
	}
	licenseNamedUserMax, ok := licenseInfo["named_user_max"].(float64)
	if ok {
		licenseNamedUserPct := licenseNamedUser / licenseNamedUserMax
		ch <- prometheus.MustNewConstMetric(
			metricLicenseNameUserPct, prometheus.GaugeValue, licenseNamedUserPct,
		)
	}
	licenseResource, ok := licenseInfo["resource"].(float64)
	if !ok {
		licenseResource = 0
	}
	licenseResourceMax, ok := licenseInfo["resource_max"].(float64)
	if ok {
		licenseResourcePct := licenseResource / licenseResourceMax
		ch <- prometheus.MustNewConstMetric(
			metricLicenseResourcePct, prometheus.GaugeValue, licenseResourcePct,
		)
	}
	licenseWaapm, ok := licenseInfo["waapm"].(float64)
	if !ok {
		licenseWaapm = 0
	}
	licenseWaapmMax, ok := licenseInfo["waapm_max"].(float64)
	if ok {
		licenseWaapmPct := licenseWaapm / licenseWaapmMax
		ch <- prometheus.MustNewConstMetric(
			metricLicenseWaapmPct, prometheus.GaugeValue, licenseWaapmPct,
		)
	}
	licensePmTarget, ok := licenseInfo["pm_target"].(float64)
	if !ok {
		licensePmTarget = 0
	}
	licensePmTargetMax, ok := licenseInfo["pm_target_max"].(float64)
	if ok {
		licensePmTargetPct := licensePmTarget / licensePmTargetMax
		ch <- prometheus.MustNewConstMetric(
			metricLicensePmTargetPct, prometheus.GaugeValue, licensePmTargetPct,
		)
	}
	licenseSmTarget, ok := licenseInfo["sm_target"].(float64)
	if !ok {
		licenseSmTarget = 0
	}
	licenseSmTargetMax, ok := licenseInfo["sm_target_max"].(float64)
	if ok {
		licenseSmTargetPct := licenseSmTarget / licenseSmTargetMax
		ch <- prometheus.MustNewConstMetric(
			metricLicenseSmTargetPct, prometheus.GaugeValue, licenseSmTargetPct,
		)
	}

	// /!\ Keep sessions relative metrics fetch to the end because it depends on a active wallix license
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
