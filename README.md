# ProxyMini

ProxyMini is a lightweight reverse proxy server written in Go that provides HTTP request logging capabilities

## Configuration

### Environment variables

- `PROXYMINI_PORT`: The port on which the ProxyMini server will listen. Default is 14443.
- `PROXYMINI_CONFIG`: The path to the TOML configuration file. Default is "proxymini.conf.toml".
- `PROXYMINI_DB`: The path to the database file used for request logging. Default is "rl.db".

### Configuration file

ProxyMini uses a TOML file to define proxy routing rules. Example config:
```toml
[[proxy]]
prefix = "/api"
target = "http://api-server:8080"

[[proxy]]
prefix = "/auth"
target = "http://auth-service:9000"
```

More specific rules should be before more general ones.

## Web UI

ProxyMini includes a web interface for viewing request logs. Access it by navigating to `/app`

![web-ui](docs/images/web.png)
