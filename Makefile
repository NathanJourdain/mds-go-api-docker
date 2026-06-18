# Lint avec golangci-lint
lint:
	docker run --rm \
		-v "./:/app" \
		-v "golangci-cache:/root/.cache" \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run ./...
