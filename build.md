## run tests
- `go test ./src/...`

## run locally
- `LOG_LEVEL=INFO ggo run $(find src -name '*.go' ! -name '*_test.go')`

## to build & run for testing
- `docker build -t ghcr.io/iankulin/ssl-monitor:latest .`
- `docker run --name ssl-monitor -p 80:8080 ghcr.io/iankulin/ssl-monitor:latest`
- http://localhost

## to build and push for production
- `docker build --platform linux/amd64 -t ghcr.io/iankulin/ssl-monitor:latest .`
- `docker push ghcr.io/iankulin/ssl-monitor:latest`

## to build binaries on Mac
- `go build -o release/ssl-monitor ./src`
- `GOOS=linux GOARCH=amd64 go build -o release/ssl-monitor-linux ./src`
- `GOOS=windows GOARCH=amd64 go build -o release/ssl-monitor.exe ./src`

## run binary locally
- `LOG_LEVEL=INFO release/ssl-monitor`
