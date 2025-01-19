# Server Configuration

## Global

```yaml
interval: 5m0s # duration each process to fetch requesters and obtain certificate. default: 5m
lock_duration: 25m0s # max duration to lock process to obtain or renew certificate to prevent concurrency. default: 25m
unused_retention: 336h0m0s # time to keep in store unused certificate. default: 14 days
http:
    listen: 0.0.0.0:8080 # server listen address. default: 0.0.0.0:8080
    metrics_enable: false # enable metrics on path `/metrics`. default: false
    tls:
        enable: false
        cert_path: "/ssl/certificate.crt" # mandatory only when enable is true
        key_path: "/ssl/private.key" # mandatory only when enable is true
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

## State

State is used to save ACME account and all certificates.

### Filesystem (default)

```yaml
state:
    type: fs
    config:
        path: /var/lib/lets-go-tls/state.json # mandatory
```

## Requesters

A Requester defines the method by which certificate requests are retrieved.

### Static

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

### Agent

```yaml
requesters:
    - id: agents
      type: agent
      config:
          addresses: # mandatory
              - 127.0.0.1:8080
```

## Cache

Cache is used to prevent concurrency process.

### Memory (default)

```yaml
cache:
    type: memory
```

### Redis

```yaml
cache:
    type: redis
    config:
        address: 127.0.0.1:6379 # mandatory
        db: my_db
        username: user
        password: password
```

### Redis cluster

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

## ACME resolvers

By default, http ACME challenge is used.
The server must receive all URI `/.well-known/acme-challenge` for ACME Challenge. 

### DNS Challenges

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
