# API Test Server

A lightweight HTTP server for testing systems that generate API calls. It logs every incoming request to stdout in full detail, and returns configurable responses based on URI patterns defined in a JSON config file.

## Features

- Logs every request to stdout — method, path, query string, headers, and body
- Routes requests to configured responses using wildcard URI patterns
- Configurable status codes, response headers, and response bodies per route
- A `default` catch-all in the config handles anything not matched by a route; if its body is empty, the HTTP response is empty (request is still logged to stdout)
- Config reload at runtime via `SIGHUP` — no restart needed
- Single static binary, no dependencies beyond the Go standard library

## Building

```bash
go build -o testsvr .
```

## Running

```bash
./testsvr                           # uses config.json on port 8080
./testsvr -config myconfig.json     # custom config file
./testsvr -port 9000                # override port from config
./testsvr -config myconfig.json -port 9000
```

Reload config without restarting:

```bash
kill -HUP <pid>
```

## Config File

```json
{
  "port": 8080,
  "default": {
    "status": 200,
    "headers": { "Content-Type": "text/plain" },
    "body": ""
  },
  "routes": [
    {
      "uri": "/api/sites/*/create",
      "method": "POST",
      "status": 200,
      "headers": { "Content-Type": "application/json" },
      "body": "{\"success\": true, \"id\": 42}\n"
    },
    {
      "uri": "/health",
      "status": 200,
      "body": "OK\n"
    },
    {
      "uri": "/api/*/delete",
      "method": "DELETE",
      "status": 403,
      "body": "Forbidden\n"
    }
  ]
}
```

### Top-level fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | integer | `8080` | Port to listen on. Overridden by `-port` flag. |
| `default` | object | required | Response used when no route matches. |
| `routes` | array | `[]` | Ordered list of route rules. |

### Route fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `uri` | string | required | URI pattern to match. Supports `*` as a single-segment wildcard. |
| `method` | string | `"*"` | HTTP method to match (`GET`, `POST`, etc.). `"*"` matches any method. |
| `status` | integer | `200` | HTTP status code to return. |
| `headers` | object | `{}` | Response headers to set as key/value pairs. |
| `body` | string | `""` | Response body. |

### Default fields

Same as a route but without `uri` or `method`. If `body` is `""` or omitted, the HTTP response body is empty — the request is still printed to stdout.

## URI Wildcards

`*` matches exactly one non-empty path segment — it will not cross a `/`.

```
Pattern                  Request path                  Match?
/api/sites/*/create      /api/sites/123/create         yes
/api/sites/*/create      /api/sites/abc-xyz/create     yes
/api/sites/*/create      /api/sites/a/b/create         no  (two segments vs one *)
/api/sites/*/create      /api/sites/create             no  (missing segment)
```

Routes are evaluated in order; the first match wins.

## Stdout Output Format

Every request is printed to stdout regardless of whether it matches a route:

```
--- Request ---
Method: POST
Path:   /api/sites/42/create
Query:  foo=bar

Headers:
  Accept: */*
  Content-Type: application/json
  User-Agent: curl/8.14.1

Body:
{"name": "example"}
```

`Query:` is omitted when there is no query string. `Body:` is omitted when the request has no body.

## Using nginx as a Frontend

API Test Server only speaks plain HTTP and has no TLS support, so put
nginx in front of it if you need HTTPS, a stable hostname, or to share
port 80/443 with other services on the same host.

A sample config is provided at [nginx/testsvr.conf.example](nginx/testsvr.conf.example).
It reverse-proxies all paths and methods straight through to the server
on `127.0.0.1:8080`, with notes on what each directive is for:

- `client_max_body_size 50m` — test payloads can be large or malformed; don't reject them at the proxy
- `proxy_set_header X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto` — preserve original client info even though the server will only see nginx's headers on the upstream connection
- `proxy_pass_request_headers on` / `proxy_pass_request_body on` — forward everything unmodified so the stdout dump reflects the real client request
- Commented-out `server` block for TLS termination via Let's Encrypt or your own certs

To use it:

```bash
sudo cp nginx/testsvr.conf.example /etc/nginx/sites-available/testsvr
# edit server_name (and TLS paths if enabling HTTPS)
sudo ln -s /etc/nginx/sites-available/testsvr /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

Note that any header nginx adds or rewrites (e.g. `Host`) will appear in
the request dump as nginx sent it, not as the original client sent it —
`X-Forwarded-*` headers are how you recover the original values.

## Project Layout

```
API-Test-Server/
├── main.go                       entry point, flag parsing, SIGHUP handler, server start
├── config.go                     Config structs and JSON loading
├── router.go                     wildcard URI matching and route lookup
├── handler.go                    request logging, route dispatch, response writing
├── config.json                   starter config
└── nginx/testsvr.conf.example    sample reverse-proxy config
```
