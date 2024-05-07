# keyvaluestore

A naive key value store implemented in Go, serving over HTTP.

## Features

- Uses the Go 1.22 "[Enhanced routing patterns](https://tip.golang.org/doc/go1.22)".
- Best-attempt tries to implement a RESTful API.
- Uses OpenAPI 3 for API documentation.

## HTTP API

### Getting a value

```bash
curl -X GET 'localhost:8080/kv?key=simon'
```

### Setting a value

```bash
curl -X PUT 'localhost:8080/kv' -d '{"key": "yo", "value": "hello"}'
```

### Getting the history

```bash
curl -X PUT 'localhost:8080/kv/hist?yo'
```

### Deleting all history

```bash
curl -X PUT 'localhost:8080/kv/del_hist' -d '{"key": "yo"}'
```
