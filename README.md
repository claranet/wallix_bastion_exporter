# Wallix Bastion exporter for Prometheus

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
- flags
- environment variables
- yaml configuration file

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

## Development

```bash
go build
```

## License

Mozilla Public License 2.0, see [LICENSE](LICENSE).
