# pingu

Ping multiple IPv4 addresses using HTTP requests.

## Configuration

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| PINGU_ADDRESS | No | 0.0.0.0:8080 | Listen address & port |
| PINGU_USER | No | pingu | Username used for basic auth |
| PINGU_PASSWORD | Yes | - | Password used for basic auth |
| PINGU_PRIVILEGED | No | true | Privileged mode uses raw sockets on Linux ([more details](https://github.com/prometheus-community/pro-bing#supported-operating-systems)) |

If you want to run Pingu with a normal user (not root), you will need to allow it to bind to raw sockets :

```bash
setcap cap_net_raw=+ep /path/to/pingu
```

## Run Pingu

```bash
PINGU_PASSWORD="your_strong_password" ./pingu
```

## Interact with Pingu

### Simple request

```bash
curl -u user:password -d '{"addresses": ["127.0.0.1", "8.8.8.8", "1.1.1.1"]}' http://127.0.0.1:8080/ping

{
  "addresses": [
    {
      "address": "127.0.0.1",
      "status": true
    },
    {
      "address": "8.8.8.8",
      "status": true
    },
    {
      "address": "1.1.1.1",
      "status": true
    }
  ]
}
```

### Customize ping parameters

| Parameter | Default | Constraints |
| --- | --- | --- |
| addresses | - | Max 10 IPv4 addresses |
| count | 1 | min=1 max=10 |
| ttl | 128 | min=1 max=128
| timeout | 1000 | Milliseconds. min=1 max=10000 |
| interval | 1000 | Milliseconds. min=1 max=10000 |

```bash
curl -u user:password -d '{"addresses": ["127.0.0.1"], "count": 5, "ttl": 128, "interval": 5000, "timeout": 5000}' http://127.0.0.1:8080/ping
```
