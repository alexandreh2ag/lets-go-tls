# Let's Go TLS

**Let's Go TLS** is a Golang-based tool designed to centralize the management of TLS certificates. This project simplifies the process of obtaining, distributing, and managing TLS certificates across multiple services. The tool is composed of two main components: the Server and the Agent.

## Features

- Centralized TLS certificate management
- Automated certificate issuance via Let's Encrypt
- Secure storage and distribution of certificates
- Support for various certificate request methods (HTTP calls, files, etc.)

## Components

### Server
The Server component handles the core certificate management tasks:

- Receives certificate requests from Agents.
- Obtains TLS certificates from Let's Encrypt.
- Stores certificates securely.
- Distributes certificates to requesting Agents.
- Cleanup unused certificates
- Expose metrics

### Agent
The Agent component is responsible for interacting with the Server and managing certificates locally:

- Expose certificate requests to the Server.
- Handles different types of certificate request methods (e.g., HTTP calls, files).
- Stores the retrieved certificates from server in specified locations.
- Makes certificates available to other services.
- Cleanup unused certificates
- Expose metrics

## Configuration

### Server Configuration
The Server can be configured via a `server.yaml` file, example can be found [here](examples/server.cfg.yml).

You can find details [here](docs/server_config.md).

#### Cleanup

When server process detect that a certificate is unused, it will mark as unused and delete it after `unused_retention` will be reach.

### Agent Configuration
The Agent can be configured via a `agent.yaml` file, example can be found [here](examples/agent.cfg.yml).

You can find details [here](docs/agent_config.md).

#### Cleanup

When agent process detect that a certificate is unused, it will to remove files on all storages.

## Installation

* Run with docker:

```bash
docker run -it -v ${PWD}:/config alexandreh2ag/lets-go-tls-server:${VERSION} start -c /config/server.yml
docker run -it -v ${PWD}:/config alexandreh2ag/lets-go-tls-agent:${VERSION} start -c /config/agent.yml
```

* Install binary to custom location:

```bash
export DESTINATION=/usr/local/bin
# latest
curl -L "https://github.com/alexandreh2ag/lets-go-tls/releases/latest/download/lets-go-tls_server_$(uname -s)_$(uname -m)" -o ${DESTINATION}/lets-go-tls_server
# specific version
curl -L "https://github.com/alexandreh2ag/lets-go-tls/releases/download/${VERSION}/lets-go-tls_server_$(uname -s)_$(uname -m)" -o ${DESTINATION}/lets-go-tls_server
chmod +x ${DESTINATION}/lets-go-tls_server

# latest
curl -L "https://github.com/alexandreh2ag/lets-go-tls/releases/latest/download/lets-go-tls_agent_$(uname -s)_$(uname -m)" -o ${DESTINATION}/lets-go-tls_server
# specific version
curl -L "https://github.com/alexandreh2ag/lets-go-tls/releases/download/${VERSION}/lets-go-tls_agent_$(uname -s)_$(uname -m)" -o ${DESTINATION}/lets-go-tls_agent
chmod +x ${DESTINATION}/lets-go-tls_agent
```

## Usage

1. Start the Server:
```bash
./lets-go-tls_server start -c ./server.yml
   ```
2. Configure and start the Agent:
```bash
./lets-go-tls_agent start -c ./agent.yml
```

### Server

## Build

### Prerequisites
- Go 1.23 or later
- Docker (optional, for containerized deployment)

### Clone the Repository
```bash
git clone https://github.com/alexandreh2ag/lets-go-tls.git
cd lets-go-tls
```

### Build and Run

#### Server
```bash
go build -o dist/lets-go-tls_server apps/server/main.go

# Run the server
./dist/lets-go-tls_server -c ./server.yml
```

#### Agent
```bash
go build -o dist/lets-go-tls_agent apps/agent/main.go

# Run the agent
./dist/lets-go-tls_agent -c ./agent.yml
```

### Tests

```bash
./bin/mock.sh
go test ./...
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request with any enhancements, bug fixes, or ideas.

## License

This project is licensed under the MIT License. See the `LICENSE.md` file for details.

