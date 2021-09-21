# Systemd Unit

The [wallix_bastion_exporter.service](wallix_bastion_exporter.service) contains an example of [Systemd Service
Unit](https://www.freedesktop.org/software/systemd/man/systemd.service.html) to work with the Wallix Bastion
Prometheus Exporter which can be installed into `/etc/systemd/system`.

The [wallix_bastion_exporter.env](wallix_bastion_exporter.env) can be used to configure the exporter and
installed to `/etc/default/wallix_bastion_exporter.env`.

Then you can reload systemd, start and enable the service:

```bash
systemctl daemon-reload
systemctl enable --now wallix_bastion_exporter.service
```

