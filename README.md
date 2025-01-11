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

#### Global

```yaml
interval: 5m0s # duration each process to fetch requesters and obtain certificate. default: 5m
lock_duration: 25m0s # max duration to lock process to obtain or renew certificate to prevent concurrency. default: 25m
unused_retention: 336h0m0s # time to keep in store unused certificate. default: 14 days
http:
    listen: 0.0.0.0:8080 # server listen address. default: 0.0.0.0:8080
    metrics_enable: false # enable metrics on path `/metrics`. default: false
jwt:
    key: superSecret # secret to sign JWT token
    method: HS256 # method used to sign JWT token. default: HS256
acme:
    ca_server: https://acme-v02.api.letsencrypt.org/directory # CA server address. default: https://acme-v02.api.letsencrypt.org/directory
    email: acme@example.com # email used for ACME registration
    renew_period: 240h0m0s # period before the end of a certificate. default: 10 days
    delay_failed: 24h0m0s # delay when a certificate reach max fail attempt to obtain or renew. default: 24h 
    max_attempt: 3 # max attempt when a certificate fail to obtain or renew. default: 3
```

##### State

State is used to save ACME account and all certificates.

###### Filesystem (default)

```yaml
state:
    type: fs
    config:
        path: /var/lib/lets-go-tls/state.json # mandatory
```

##### Requesters

A Requester defines the method by which certificate requests are retrieved.

###### Static

```yaml
requesters:
    - id: static
      type: static
      config:
          domains: # mandatory
            - - example.com
            - - foo.com
              - bar.com
```

###### Agent

```yaml
requesters:
    - id: agents
      type: agent
      config:
          addresses: # mandatory
              - 127.0.0.1:8080
```

##### Cache

Cache is used to prevent concurrency process. 

###### Memory (default)

```yaml
cache:
    type: memory
```

###### Redis

```yaml
cache:
    type: redis
    config:
        address: 127.0.0.1:6379 # mandatory
        db: my_db
        username: user
        password: password
```

###### Redis cluster

```yaml
cache:
    type: redis-cluster
    config:
        address: # mandatory
          - 127.0.0.1:6379 
          - 127.0.0.2:6379
        username: user
        password: password
```

##### ACME resolvers

By default, http ACME challenge is used.

###### DNS Challenges

The key `filters` define domains who must use specific resolver.

Examples:
* Whole domain:
  * filters: example.com
  * match: example and all subdomain
* Only subdomain:
  * filters: sub.example.com
  * match: sub.example.com and *.sub.example.com but don't match foo.example.com or example.com

```yaml
acme:
  resolvers:
      gandiv5:
          type: gandiv5
          config:
              api_key: ""
              http_timeout: 3m0s
              polling_interval: 2s
              propagation_timeout: 1m0s
          filters:
              - foo.com
      httpreq:
          type: httpreq
          config:
              endpoint: ""
              http_timeout: 3m0s
              mode: ""
              password: ""
              polling_interval: 2s
              propagation_timeout: 1m0s
              username: ""
          filters:
              - foo.com
      ovh:
          type: ovh
          config:
              access_token: ""
              application_key: ""
              application_secret: ""
              client_id: ""
              client_secret: ""
              consumer_key: ""
              endpoint: ovh-eu
              http_timeout: 3m0s
              polling_interval: 2s
              propagation_timeout: 1m0s
          filters:
              - foo.com
```

If you need other resolver you can open an issue or a pull request.

### Agent Configuration
The Agent can be configured via a `agent.yaml` file, example can be found [here](examples/agent.cfg.yml).

#### Global

```yaml
interval: 5m0s # duration each process to fetch certificates. default: 5m
http:
    listen: 0.0.0.0:8080 # server listen address. default: 0.0.0.0:8080
    metrics_enable: false # enable metrics on path `/metrics`. default: false
manager:
    address: 127.0.0.1:8080 # server address
    token: tokenJwt # JWT token used to authenticate on server
```

##### State

State is used to save all certificates.

###### Filesystem (default)

```yaml
state:
    type: fs
    config:
        path: /var/lib/lets-go-tls/state.json # mandatory
```

##### Requesters

A Requester defines the method by which certificate requests are retrieved.

###### Static

```yaml
requesters:
    - id: static
      type: static
      config:
          domains: # mandatory
            - - example.com
            - - foo.com
              - bar.com
```

###### Traefik (V2 & V3)

```yaml
requesters:
    - id: traefik
      type: traefikV2 # value can be traefikV2 or traefikV3
      config:
          addresses: # mandatory
              - http://127.0.0.1
              - http://127.0.0.1:8080
              - https://127.0.0.1/api
```

##### Storages

A storage defines where certificate must be stored.

###### Filesystem

```yaml
storages:
    - id: fs
      type: fs
      config:
          path: /var/lib/lets-go-tls/ssl # mandatory
          prefix_filename: "ssl."
```

tree structure example:

```
/var/lib/lets-go-tls
└── ssl/
    ├── ssl.example.com-0.key
    ├── ssl.example.com-0.crt
    ├── ssl.foo.com-0.key
    └── ssl.foo.com-0.crt
```

###### Traefik (V2 & V3)

Traefik configurations:
* [static configuration](./examples/traefik/config.yml)
* [dynamic configuration](./examples/traefik/router.yml)

```yaml
storages:
    - id: traefik
      type: traefikV2 # value can be traefikV2 or traefikV3
      config:
          path: /var/lib/traefik/ssl # mandatory
          prefix_filename: "ssl."
```

tree structure example:
```
/var/lib/traefik
└── ssl/
    ├── ssl.example.com-0.yml
    └── ssl.foo.com-0.yml
```

###### Cleanup

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
./lets-go-tls_server -c ./server.yml
   ```
2. Configure and start the Agent:
```bash
./lets-go-tls_agent -c ./agent.yml
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

