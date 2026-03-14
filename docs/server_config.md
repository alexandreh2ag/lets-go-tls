# Server Configuration

## Global

```yaml
interval: 5m0s # duration each process to fetch requesters and obtain certificate. default: 5m
lock_duration: 25m0s # max duration to lock process to obtain or renew certificate to prevent concurrency. default: 25m
unused_retention: 336h0m0s # time to keep in store unused certificate. default: 14 days
http:
    listen: 0.0.0.0:8080 # http server listen address. default: 0.0.0.0:8080
    metrics_enable: false # enable metrics on path `/metrics`. default: false
    tls:
        enable: false
        listen: 0.0.0.0:443 # https server listen address.
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
    http_challenge:
        enable_document_root: false # enable document root for http challenge.
        document_root: "" # document root for http challenge.
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

### HTTP Challenge manual

It is possible to manually add keyAuth files to validate HTTP challenges for external certificates (e.g., behind a CDN).

When `enable_document_root` is enabled, the server looks for a file named after the token directly in the `document_root` directory.
The path `/.well-known/acme-challenge` is **not** included in the file lookup, so the token file must be placed directly in the document root.

For example, with `document_root: "/var/www"` and a token `abc123`, the server will read the keyAuth value from `/var/www/abc123`.

```yaml
acme:
  http_challenge:
    enable_document_root: true
    document_root: "/var/www"
```

**Usage:**

1. Create the token file in the document root with the keyAuth value as content:
    ```bash
    echo -n "keyAuth_value" > /var/www/<token>
    ```
2. The ACME server will request `http://<domain>/.well-known/acme-challenge/<token>`, and the server will respond with the content of `/var/www/<token>`.
3. Once the challenge is validated, you can remove the token file.

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
