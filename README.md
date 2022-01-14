# Wallix Bastion exporter for Prometheus
[![Maintainer](https://img.shields.io/badge/maintained%20by-claranet-red?style=flat-square)](https://www.claranet.fr/)
[![License](https://img.shields.io/github/license/claranet/wallix_bastion_exporter?style=flat-square)](LICENSE)
[![Release](https://img.shields.io/github/v/release/claranet/wallix_bastion_exporter?style=flat-square)](https://github.com/claranet/wallix_bastion_exporter/releases)
[![Lint](https://img.shields.io/github/workflow/status/claranet/wallix_bastion_exporter/golangci-lint?style=flat-square&label=lint)](https://github.com/claranet/wallix_bastion_exporter/actions/workflows/lint.yml)
[![CodeQL](https://img.shields.io/github/workflow/status/claranet/wallix_bastion_exporter/codeql-analysis?style=flat-square&label=security)](https://github.com/claranet/wallix_bastion_exporter/actions/workflows/analyze.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/claranet/wallix_bastion_exporter?style=flat-square)](https://goreportcard.com/report/github.com/claranet/wallix_bastion_exporter)
[![Code Climate](https://img.shields.io/codeclimate/maintainability/claranet/wallix_bastion_exporter?style=flat-square)](https://codeclimate.com/github/claranet/wallix_bastion_exporter)
[![Go mod version](https://img.shields.io/github/go-mod/go-version/claranet/wallix_bastion_exporter?style=flat-square)](https://golang.org/)
[![Godoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/claranet/wallix_bastion_exporter)

This is a simple server that scrapes Wallix Bastion stats and exports them via HTTP for Prometheus consumption.

## Getting Started

Download and extract the lastest precompiled binary from [releases page](https://github.com/claranet/wallix_bastion_exporter/releases).

Then, run it:

```bash
./wallix_bastion_exporter [flags]
```

Help on flags:

```bash
./wallix_bastion_exporter --help
```

Check the [examples](examples) directory for common installations like Systemd or Opentelemetry Collector.

## Usage

First of all, it requires an available user on the target Wallix bastion with full read only permissions. Here is an example of the Terraform configuration:

```hcl
resource "wallix-bastion_profile" "full_ro" {
  profile_name  = "full_ro"
  description   = "Grant read only access to everything"
  target_access = false

  gui_features {
    wab_audit      = "view"
    approval       = "view"
    authorizations = "view"
    devices        = "view"
    system_audit   = "view"
    target_groups  = "view"
    user_groups    = "view"
    users          = "view"
    wab_settings   = "view"
  }

  gui_transmission {
    system_audit   = "view"
    approval       = "view"
    authorizations = "view"
    devices        = "view"
    target_groups  = "view"
    user_groups    = "view"
    users          = "view"
    wab_settings   = "view"
  }
}

resource "wallix-bastion_user" "monitoring" {
  user_name        = "monitoring"
  display_name     = "Monitoring"
  email            = "monitoring@localhost.localdomain"
  profile          = wallix-bastion_profile.full_ro.profile_name
  user_auths       = ["local_password"]
  password         = "password should be retrieved from secure place like vault_generic_secret datasource"
  force_change_pwd = false
  ip_source        = "127.0.0.1"
}
```

If the exporter is not installed on Wallix bastion host, so you must:
- change the `ip_source` restriction for the user in the configuration above for the address from where the exporter will query Wallix bastion API.
- specify the custom URL for the Wallix bastion API (e.g. `./wallix_bastion_exporter --scrape-uri https://10.42.13.37/api`)

Then, you must configure at least `wallix-username` and `wallix-password` corresponding to this user.
See [Configuration](#configuration) section below for more information about how to configure the exporter.


## Configuration

Configuration can be done, in precendence order, using:
1. flags
1. environment variables
1. yaml configuration file

For the last, you can copy [config.yaml.sample](config.yaml.sample) next to the exporter binary and modify depending on your setup.

Here is a matrix with all available configurations depending on their sources:


| Config option | Environment variable |  Flag | Description |
|---|---|---|---|
| `listen-address` | `LISTEN_ADDRESS` | `--listen-address` | Address to listen on for web interface and telemetry |
| `telemetry-path` | `TELEMETRY_PATH` | `--telemetry-path` | Path under which to expose metrics |
| `scrape-uri` | `SCRAPE_URI` | `--scrape-uri` | URI on which to scrape Wallix Bastion API |
| `skip-verify` | `SKIP_VERIFY` | `--skip-verify` | Flag that disables TLS certificate verification for the scrape URI |
| `timeout` | `TIMEOUT` | `--timeout` | Timeout in seconds for requests to Wallix Bastion API |
| `wallix-username` | `WALLIX_USERNAME` | `--wallix-username` | The username used for authentication to request Wallix Bastion API |
| `wallix-password` | `WALLIX_PASSWORD` | `--wallix-password` | The password used for authentication to request Wallix Bastion API |

You can mix the three sources as you wish like:

```bash
$ cat config.yaml
scrape-uri: "https://127.0.0.1/api"
listen: ":4242"

$ WALLIX_PASSWORD=$(gopass show -o wallix-bastion/password) ./wallix_bastion_exporter --wallix-username "monitoring" --scrape-uri "https://10.42.13.37/api"
```

In this example:
- `wallix-username` is defined by `--wallix-username` flag to `monitoring`
- `wallix-password` is defined by `WALLIX_PASSWORD` environment variable using `gopass` command
- `scrape-uri` is defined by both configuration file and flag but the last has the priority so the value is `https://10.42.13.37/api`
- `listen` is defined by `listen` configuration file directive to `:4242` to change the default port `9191`

## Metrics

The statistics retrieved from Wallix API are not very dynamic so __it is recommended to configure the scrape interval to `5m`__.
Below could cause undesired load on the server. Above will desynchronize closed sessions metric timeframe.

| Metric | Labels | Note |
|---|---|---|
| `wallix_bastion_up` | | `0` if the exporter cannot authenticate to Wallix API, `1` if request is successful |
| `wallix_bastion_users` | | Total number of local users as gauge |
| `wallix_bastion_groups` | | Total number of user groups as gauge |
| `wallix_bastion_devices` | | Total number of devices as gauge |
| `wallix_bastion_targets` | `type` | Number of targets per `type` |
| `wallix_bastion_sessions` | `status` | Number of sessions per `status`. `closed` status count is done __over the last `5m` independently of the scrape interval__ |
| `wallix_bastion_encryption_status` | `status`,`security_level` | Encryption status (need_setup=0, ready=1, need_passphrase=2) |
| `wallix_bastion_encryption_security_level` | `security_level`,`status` | Encryption security level (need_setup=0, passphrase_defined=1, passphrase_not_used=2, [hidden]=-1) |
| `wallix_bastion_license_is_expired` | | Is the Wallix is expired (0=false, 1=true) |
| `wallix_bastion_license_primary_ratio` | | License usage percentage of primary |
| `wallix_bastion_license_secondary_ratio` | | License usage percentage of secondary |
| `wallix_bastion_license_named_user_ratio` | | License usage percentage of named user |
| `wallix_bastion_license_resource_ratio` | | License usage percentage of resource |
| `wallix_bastion_license_waapm_ratio` | | License usage percentage of waapm |
| `wallix_bastion_license_sm_target_ratio` | | License usage percentage of sm target |
| `wallix_bastion_license_pm_target_ratio` | | License usage percentage of pm target |

## Development

```bash
go build
```

## License

Mozilla Public License 2.0, see [LICENSE](LICENSE).
