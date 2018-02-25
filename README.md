# visit.danwolf.net

Website source for visit.danwolf.net which shows people how to visit me.

# Roadmap

This project will begin as a simple Go static webserver and stick to a code and content layout that will make it relatively easy to migrate to Rhyvu once it is stable.

# Build

```
go build
```

# Run

```
.\visit
```

# Install

To install as a service (assuming Ubuntu 16.04 and Systemd) you will need to first run `chmod u+x /path/to/visit/visit` and create the file below:

**/etc/systemd/system/visit.service**:

```
[Unit]
Description=Visit service
StartLimitBurst=5

[Service]
WorkingDirectory=/path/to/visit/
ExecStart=/path/to/visit/visit
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

Start it: `sudo systemctl start visit`

Enable it to run at boot (should create a symlink in `/etc/systemd/system/multi-user.target.wants/`): `sudo systemctl enable visit`

Stop it: `sudo systemctl stop visit`

Soft reload Systemd dependencies: `sudo systemctl daemon-reload`

# License

This project is available under an MIT license. Dependencies might carry other licenses.