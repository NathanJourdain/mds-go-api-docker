# Lint avec golangci-lint
lint:
	docker run --rm \
		-v "./:/app" \
		-v "golangci-cache:/root/.cache" \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run ./...


vulncheck:
	docker run --rm \
		-v "./:/app" \
		-w /app \
		golang:latest \
		sh -c "go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./..."
