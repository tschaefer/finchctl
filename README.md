# The Minimal Logging Infrastructure

finchtcl can be used to deploy a logging stack and agents.

The logging stack bases on Docker and consists of following services:

- **grafana** - The visualization tool
- **loki** - The log aggregation system
- **alloy** - The log shipping agent
- **traefik** - The reverse proxy
- **finch** - The log agent manager

Consider to read the [Blog post](https://blog.tschaefer.org) for motivation
and a walkthrough before using this tool.

## Getting Started

Download the latest release from the
[releases page](https://github.com/tschaefer/finchctl/releases) or build it
from source.

## Install the logging stack

Prerequisites is a blank Linux machine with SSH access and superuser
privileges.

To install a logging stack the easiest way without further configuration
run the following command:

```bash
finchctl service deploy root@10.19.80.100
```

This will deploy the logging stack to the remote machine and exposes the
services under the base URL `https://10.19.80.100`. The admin credentials
can be found in the configuration file ~/.finch/config.json. The path to the
main services are `/grafana` and `/finch`. The used TLS certificate is a
traefik default self-signed certificate.

Assumed you have a public reachable machine and a DNS record for it, you
can deploy the logging stack with a custom URL and Let's Encrypt certificate.

```bash
finchctl service deploy \
    --service.letsencrypt --service.letsencrypt.email acme@example.com \
    --service.user admin --service.password secret \
    --service.host finch.example.com root@cloud.machine
```

This will deploy the logging stack to the remote machine and exposes the
services under the base URL `https://finch.example.com` and uses a
Let's Encrypt certificate. The credentials for Grafana and Finch are set
to user `admin` and password `secret`.

Beside Let's Encrypt you can also use a custom TLS certificate by specifying
the paths to the certificate and key files.

```bash
finchctl service deploy \
    --service.customtls --service.customtls.cert ~/.tls/cert.pem
    --service.customtls.key ~/.tls/key.pem finch.example.com
```

##  Enrolling a logging agent

```bash
finchctl agent register \
    --agent.hostname sparrow.example.com \
    --agent.log.journal finch.example.com
```

This will register a new agent with the given hostname to the given finch
service. The prepared agent configuration file will be stored as
`finch-agent.cfg`. It is set up ready to deploy, including all loki endpoint
configuration and authentication credentials. As requested the agent will send
systemd journal records to the logging stack.

Beside systemd journal records, docker `--agent.log.docker` and file records
`--agent.log.file /var/log/*.log` can be used as log sources.

```bash
finchctl agent deploy --config finch-agent.cfg root@app.machine
```
This will enroll the agent, alloy, on the remote machine and start it with the
configuration from the specified file.

## Further controller commands

Both commands, `service` and `agent`, have severals subcommands. Among others
`teardown` to remove the deployed logging stack or agent from the remote
machine and `update` to update the logging stack services or agent to the
latest version.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.
For major changes, open an issue first to discuss what you would like to change.

Ensure that your code adheres to the existing style and includes appropriate tests.

## License

This project is licensed under the [MIT License](LICENSE).
