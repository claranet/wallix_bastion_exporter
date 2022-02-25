package exporter

import (
	"log"
	"net/http"
	"sync"

	"github.com/claranet/wallix_bastion_exporter/wallix"
	"github.com/prometheus/client_golang/prometheus"
)

func (e *Exporter) gatherMetricsUsers(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	users, err := wallix.GetUsers(client, e.Config.ScrapeURI)
	if err != nil {
		log.Printf("cannot get users: %v", err)
	} else {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricUsers, prometheus.GaugeValue, float64(len(users)),
		)
	}

	gatherGroup.Done()
}

func (e *Exporter) gatherMetricsGroups(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	groups, err := wallix.GetGroups(client, e.Config.ScrapeURI)
	if err != nil {
		log.Printf("cannot get groups: %v", err)
	} else {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricGroups, prometheus.GaugeValue, float64(len(groups)),
		)
	}

	gatherGroup.Done()
}

func (e *Exporter) gatherMetricsDevices(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	devices, err := wallix.GetDevices(client, e.Config.ScrapeURI)
	if err != nil {
		log.Printf("cannot get devices: %v", err)
	} else {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricDevices, prometheus.GaugeValue, float64(len(devices)),
		)
	}

	gatherGroup.Done()
}

func (e *Exporter) gatherMetricsTargetsSessionAccounts(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	targetType := "session_accounts"
	targetsSessionAccounts, err := wallix.GetTargets(client, e.Config.ScrapeURI, targetType)
	if err != nil {
		log.Printf("cannot get session accounts targets: %v", err)
	} else {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricTargets, prometheus.GaugeValue, float64(len(targetsSessionAccounts)), targetType,
		)
	}

	gatherGroup.Done()
}

func (e *Exporter) gatherMetricsTargetsSessionAccountMappings(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	targetType := "session_account_mappings"
	targetsSessionAccountMappings, err := wallix.GetTargets(client, e.Config.ScrapeURI, targetType)
	if err != nil {
		log.Printf("cannot get session account mappings targets: %v", err)
	} else {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricTargets, prometheus.GaugeValue, float64(len(targetsSessionAccountMappings)), targetType,
		)
	}

	gatherGroup.Done()
}

func (e *Exporter) gatherMetricsTargetsSessionInteractiveLogins(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	targetType := "session_interactive_logins"
	targetsSessionInteractiveLogins, err := wallix.GetTargets(client, e.Config.ScrapeURI, targetType)
	if err != nil {
		log.Printf("cannot get session interactive logins targets: %v", err)
	} else {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricTargets, prometheus.GaugeValue, float64(len(targetsSessionInteractiveLogins)), targetType,
		)
	}

	gatherGroup.Done()
}

func (e *Exporter) gatherMetricsTargetsSessionScenarioAccounts(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	targetType := "session_scenario_accounts"
	targetsSessionsScenarioAccounts, err := wallix.GetTargets(client, e.Config.ScrapeURI, targetType)
	if err != nil {
		log.Printf("cannot get session scenario accounts targets: %v", err)
	} else {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricTargets, prometheus.GaugeValue, float64(len(targetsSessionsScenarioAccounts)), targetType,
		)
	}

	gatherGroup.Done()
}

func (e *Exporter) gatherMetricsTargetsPasswordRetrievalAccounts(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	targetType := "password_retrieval_accounts"
	targetsPasswordRetrievalAccounts, err := wallix.GetTargets(client, e.Config.ScrapeURI, targetType)
	if err != nil {
		log.Printf("cannot get password retrieval accounts targets: %v", err)
	} else {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricTargets, prometheus.GaugeValue, float64(len(targetsPasswordRetrievalAccounts)), targetType,
		)
	}

	gatherGroup.Done()
}

func (e *Exporter) gatherMetricsEncryption(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
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
		log.Printf("cannot get encryption information: %v", err)
	} else {
		encryptionStatus, ok := encryptionInfo["encryption"].(string)
		if ok {
			encryptionSecurityLevel, ok := encryptionInfo["security_level"].(string)
			if ok {
				metricsChannel <- prometheus.MustNewConstMetric(
					metricEncryptionStatus,
					prometheus.GaugeValue,
					float64(encryptionMap[encryptionStatus]),
					encryptionStatus, encryptionSecurityLevel,
				)
				metricsChannel <- prometheus.MustNewConstMetric(
					metricEncryptionSecurityLevel,
					prometheus.GaugeValue,
					float64(encryptionMap[encryptionSecurityLevel]),
					encryptionSecurityLevel, encryptionStatus,
				)
			}
		}
	}

	gatherGroup.Done()
}

func (e *Exporter) gatherMetricsLicense(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	licenseInfo, err := wallix.GetLicense(client, e.Config.ScrapeURI)
	if err != nil {
		log.Printf("cannot get license information: %v", err)
		gatherGroup.Done()

		return
	}
	if licenseIsExpired, ok := licenseInfo["is_expired"].(bool); ok {
		var licenseIsExpiredGauge int8
		if licenseIsExpired {
			licenseIsExpiredGauge = 1
		}
		metricsChannel <- prometheus.MustNewConstMetric(
			metricLicenseIsExpired, prometheus.GaugeValue, float64(licenseIsExpiredGauge),
		)
	}
	if licenseIsValid, ok := licenseInfo["is_valid"].(bool); ok {
		var licenseIsExpiredGauge int8
		if !licenseIsValid {
			licenseIsExpiredGauge = 1
		}
		metricsChannel <- prometheus.MustNewConstMetric(
			metricLicenseIsExpired, prometheus.GaugeValue, float64(licenseIsExpiredGauge),
		)
	}
	licensePrimary, ok := licenseInfo["primary"].(float64) //nolint:varnamelen
	if !ok {
		licensePrimary = 0
	}
	if licensePrimaryMax, ok := licenseInfo["primary_max"].(float64); ok {
		licensePrimaryPct := licensePrimary / licensePrimaryMax
		metricsChannel <- prometheus.MustNewConstMetric(
			metricLicensePrimaryPct, prometheus.GaugeValue, licensePrimaryPct,
		)
	}
	licenseSecondary, ok := licenseInfo["secondary"].(float64)
	if !ok {
		licenseSecondary = 0
	}
	if licenseSecondaryMax, ok := licenseInfo["secondary_max"].(float64); ok {
		licenseSecondaryPct := licenseSecondary / licenseSecondaryMax
		metricsChannel <- prometheus.MustNewConstMetric(
			metricLicenseSecondaryPct, prometheus.GaugeValue, licenseSecondaryPct,
		)
	}
	licenseNamedUser, ok := licenseInfo["named_user"].(float64)
	if !ok {
		licenseNamedUser = 0
	}
	if licenseNamedUserMax, ok := licenseInfo["named_user_max"].(float64); ok {
		licenseNamedUserPct := licenseNamedUser / licenseNamedUserMax
		metricsChannel <- prometheus.MustNewConstMetric(
			metricLicenseNameUserPct, prometheus.GaugeValue, licenseNamedUserPct,
		)
	}
	licenseResource, ok := licenseInfo["resource"].(float64)
	if !ok {
		licenseResource = 0
	}
	if licenseResourceMax, ok := licenseInfo["resource_max"].(float64); ok {
		licenseResourcePct := licenseResource / licenseResourceMax
		metricsChannel <- prometheus.MustNewConstMetric(
			metricLicenseResourcePct, prometheus.GaugeValue, licenseResourcePct,
		)
	}
	licenseWaapm, ok := licenseInfo["waapm"].(float64)
	if !ok {
		licenseWaapm = 0
	}
	if licenseWaapmMax, ok := licenseInfo["waapm_max"].(float64); ok {
		licenseWaapmPct := licenseWaapm / licenseWaapmMax
		metricsChannel <- prometheus.MustNewConstMetric(
			metricLicenseWaapmPct, prometheus.GaugeValue, licenseWaapmPct,
		)
	}
	licensePmTarget, ok := licenseInfo["pm_target"].(float64)
	if !ok {
		licensePmTarget = 0
	}
	if licensePmTargetMax, ok := licenseInfo["pm_target_max"].(float64); ok {
		licensePmTargetPct := licensePmTarget / licensePmTargetMax
		metricsChannel <- prometheus.MustNewConstMetric(
			metricLicensePmTargetPct, prometheus.GaugeValue, licensePmTargetPct,
		)
	}
	licenseSmTarget, ok := licenseInfo["sm_target"].(float64)
	if !ok {
		licenseSmTarget = 0
	}
	if licenseSmTargetMax, ok := licenseInfo["sm_target_max"].(float64); ok {
		licenseSmTargetPct := licenseSmTarget / licenseSmTargetMax
		metricsChannel <- prometheus.MustNewConstMetric(
			metricLicenseSmTargetPct, prometheus.GaugeValue, licenseSmTargetPct,
		)
	}

	gatherGroup.Done()
}

func (e *Exporter) gatherMetricsSessions(
	gatherGroup *sync.WaitGroup, metricsChannel chan<- prometheus.Metric, client *http.Client,
) {
	sessionsCurrent, err := wallix.GetCurrentSessions(client, e.Config.ScrapeURI)
	if err != nil {
		log.Printf("cannot get current sessions: %v", err)
	} else {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricSessions, prometheus.GaugeValue, float64(len(sessionsCurrent)), "current",
		)
	}

	sessionsClosed, err := wallix.GetClosedSessions(client, e.Config.ScrapeURI, sessionsClosedMinutes)
	if err != nil {
		log.Printf("cannot get closed sessions: %v", err)
	} else {
		metricsChannel <- prometheus.MustNewConstMetric(
			metricSessions, prometheus.GaugeValue, float64(len(sessionsClosed)), "closed",
		)
	}

	// ch <- prometheus.MustNewConstMetric(
	// 	sessions, prometheus.GaugeValue, float64(
	// 		len(sessionsClosedResults)+len(sessionsCurrentResults),
	// 	),
	// )

	gatherGroup.Done()
}
