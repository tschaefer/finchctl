# The Minimal Observability Infrastructure [![Go Report Card](https://goreportcard.com/badge/github.com/tschaefer/finchctl)](https://goreportcard.com/report/github.com/tschaefer/finchctl)

`finchctl` is used to deploy an observability stack and agents.

The stack is based on Docker and consists of the following services:

- **Grafana** – Visualization and dashboards
- **Loki** – Log aggregation system
- **Mimir** – Metrics backend
- **Pyroscope** – Profiling data aggregation and visualization
- **Alloy** – Client-side agent for logs, metrics, and profiling data
- **Traefik** – Reverse proxy
- **Finch** – Agent manager

See the [Blog post](https://blog.tschaefer.org/posts/2025/08/17/finch-a-minimal-logging-stack/)
for background, motivation, and a walkthrough before you get started.

## Getting Started

Download the latest release from the
[releases page](https://github.com/tschaefer/finchctl/releases) or build it
from source.

## Install the Observability Stack

You need a blank Linux machine with SSH access and superuser privileges.

To install the stack with default configuration, run:

```bash
finchctl service deploy root@10.19.80.100
```

This deploys the stack and exposes services at `https://10.19.80.100`.
Admin credentials are in `~/.finch/config.json`.
Visualization and dashboards are available under `/grafana`.
TLS uses Traefik's default self-signed certificate.

For a public machine with DNS and Let's Encrypt certificate:

```bash
finchctl service deploy \
    --service.letsencrypt --service.letsencrypt.email acme@example.com \
    --service.user admin --service.password secret \
    --service.host finch.example.com root@cloud.machine
```

Services are exposed at `https://finch.example.com` with a Let's Encrypt
certificate. Credentials for Grafana and Finch: user `admin`, password `secret`.

To use a custom TLS certificate:

```bash
finchctl service deploy \
    --service.customtls --service.customtls.cert ~/.tls/cert.pem \
    --service.customtls.key ~/.tls/key.pem finch.example.com
```

## Enrolling an Observability Agent

```bash
finchctl agent register \
    --agent.hostname sparrow.example.com \
    --agent.log.journal finch.example.com
```

This registers a new agent for the specified Finch service.  
The agent config file is saved as `finch-agent.cfg`, ready to deploy, including
all endpoints and credentials. By default, it sends systemd journal records.

You can also collect logs from Docker containers (`--agent.log.docker`) and
files (`--agent.log.file /var/log/*.log`). Metrics can be included via
`--agent.metrics`.

To deploy the agent:

```bash
finchctl agent deploy --config finch-agent.cfg root@app.machine
```

Alloy will be enrolled and started with the provided configuration.

## Metrics and Profiling Data Collection

Applications can forward log, metrics and profiling data to Alloy:

- **Logs Listen** `http://localhost:3100`
- **Metrics Listen** `http://localhost:9091`
- **Profiling Listen** `http://localhost:4040`

Alloy is pre-configured to accept and forward this data to Loki, Mimir and
Pyroscope.

## Further Controller Commands

Both `service` and `agent` commands have several subcommands, including:

- `teardown` – Remove the deployed stack or agent
- `update` – Upgrade stack services or agent to the latest version

## Contributing

Contributions are welcome!
Fork the repository and submit a pull request. For major changes, open an issue
first to discuss your proposal.

Please ensure your code follows the project's style and includes appropriate
tests.

## License

This project is licensed under the [MIT License](LICENSE).
