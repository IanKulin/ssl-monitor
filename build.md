## to build & run for testing
- `docker build -t ghcr.io/iankulin/ssl-monitor:latest .`
- `docker run --name ssl-monitor -p 80:8080 ghcr.io/iankulin/ssl-monitor:latest`
- http://localhost

## to build and push for production
- `docker build --platform linux/amd64 -t ghcr.io/iankulin/ssl-monitor:latest .`
- `docker push ghcr.io/iankulin/ssl-monitor:latest`