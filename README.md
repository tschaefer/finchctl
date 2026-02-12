# The Minimal Observability Infrastructure

[![Tag](https://img.shields.io/github/tag/tschaefer/finchctl.svg)](https://github.com/tschaefer/finchctl/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/tschaefer/finchctl)](https://goreportcard.com/report/github.com/tschaefer/finchctl)
[![Contributors](https://img.shields.io/github/contributors/tschaefer/finchctl)](https://github.com/tschaefer/finchctl/graphs/contributors)
[![License](https://img.shields.io/github/license/tschaefer/finchctl)](./LICENSE)

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
Visualization and dashboards are available under `/grafana`. TLS uses Traefik's
default self-signed certificate. Credentials for Grafana are user `admin` and
password `admin`. mTLS secures communication between the Finch client and
service. The client certificate and key are generated during deployment and
stored in `~/.config/finch.json`.

For a public machine with DNS and Let's Encrypt certificate:

```bash
finchctl service deploy \
    --service.letsencrypt --service.letsencrypt.email acme@example.com \
    --service.host finch.example.com root@cloud.machine
```

Services are exposed at `https://finch.example.com` with a Let's Encrypt
certificate.

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
    --agent.logs.journal finch.example.com
```

This registers a new agent for the specified Finch service.  
The agent config file is saved as `finch-agent.cfg`, ready to deploy, including
all endpoints and credentials. By default, it sends systemd journal records.

You can also collect logs from Docker containers (`--agents.log.docker`) and
files (`--agent.logs.file /var/log/*.log`). Node metrics can be included via
`--agent.metrics` and metrics targets to scrape can be added with
`--agent.metrics.targets <target>`. Profiling data collection can be enabled
with `--agent.profiles`.

Alternatively, you can register an agent with a [configuration
file](contrib/agent-file.yml):

```bash
finchctl agent register --agent.file agent-file.yml finch.example.com
```

To deploy the agent:

```bash
finchctl agent deploy --agent.config finch-agent.cfg root@app.machine
```

Alloy will be enrolled and started with the provided configuration. Alloy
authenticates with the Finch service using a JWT token with a **365-day
expiration** included in the config file. Request a new config at least after
one year or on compromise.

```bash
finchctl agent config --agent.rid rid:finch... finch.example.com
```

## Metrics and Profiling Data Collection

Applications can forward log, metrics and profiling data to Alloy:

- **Logs Listen** `http://localhost:3100`
- **Metrics Listen** `http://localhost:9091`
- **Profiling Listen** `http://localhost:4040`

Alloy is pre-configured to accept and forward this data to Loki, Mimir and
Pyroscope.

## Access Web Dashboard

Finch provides a lightweight dashboard for visualizing agents with real-time
updates. The dashboard is protected by token-based authentication.
Retrieve token and open in browser:

```bash
finchctl service dashboard --web=true --session-timeout=1800 finch.example.com
```

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
