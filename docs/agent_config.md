# Agent Configuration

## Global

```yaml
interval: 5m0s # duration each process to fetch certificates. default: 5m
http:
    listen: 0.0.0.0:8080 # http server listen address. default: 0.0.0.0:8080
    metrics_enable: false # enable metrics on path `/metrics`. default: false
    tls:
        enable: false
        listen: 0.0.0.0:443 # https server listen address.
        cert_path: "/ssl/certificate.crt" # mandatory only when enable is true
        key_path: "/ssl/private.key" # mandatory only when enable is true
manager:
    address: 127.0.0.1:8080 # server address
    token: tokenJwt # JWT token used to authenticate on server
```

## State

State is used to save all certificates.

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

### Traefik (V2 & V3)

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

## Storages

A storage defines where certificate must be stored.

### Filesystem

```yaml
storages:
    - id: fs
      type: fs
      config:
          path: /var/lib/lets-go-tls/ssl # mandatory
          prefix_filename: "ssl."
          owner: "root"
          group: "root"
          only_matched_domains: false # when true, only store certificate specified in specific_domains
          specific_domains:
            - identifier: custom # mandatory
              path: "custom" # optional, can be absolut or relative
              domains: # mandatory
                - bar.com
          post_hook:
            cmd: "bash -c 'nginx -t && systemctl reload nginx'" # mandatory, run command when certificates have changed
            timeout: 1m0s
```

tree structure example:

```
/var/lib/lets-go-tls
└── ssl/
    ├── ssl.example.com-0.key
    ├── ssl.example.com-0.crt
    ├── ssl.foo.com-0.key
    └── ssl.foo.com-0.crt
    ├── ssl.custom.key
    └── ssl.custom.crt
```

### Traefik (V2 & V3)

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
          owner: "root"
          group: "root"
```

tree structure example:
```
/var/lib/traefik
└── ssl/
    ├── ssl.example.com-0.yml
    └── ssl.foo.com-0.yml
```
