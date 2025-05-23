# ProxyMini

ProxyMini is a lightweight proxy server written in Go that provides HTTP request logging capabilities

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

## Deployment

Download binary from [release page](https://github.com/mishankov/proxymini/releases) or use [docker image](https://github.com/mishankov/proxymini/pkgs/container/proxymini):

```shell
docker pull ghcr.io/mishankov/proxymini:latest
```

### Docker compose example

```yml
services:
  proxy:
    image: ghcr.io/mishankov/proxymini:latest
    ports:
      - "14443:14443"
    volumes:
      # mapping config file to container
      - ./proxymini.conf.toml:/app/proxymini.conf.toml:ro
      # mapping folder with database
      - ./data:/app/data:rw
```
