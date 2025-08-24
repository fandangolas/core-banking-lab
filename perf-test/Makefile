.PHONY: build run-cli run-server clean deps

# Build the load test binary
build:
	go build -o bin/loadtest cmd/loadtest/main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run CLI mode with default scenario
run-cli:
	go run cmd/loadtest/main.go \
		-mode=cli \
		-workers=100 \
		-duration=60s \
		-ramp-up=10s

# Run CLI with high load scenario
run-high-load:
	go run cmd/loadtest/main.go \
		-mode=cli \
		-scenario=scenarios/high-load.json \
		-workers=200 \
		-duration=300s \
		-ramp-up=30s

# Run CLI with concurrent transfers scenario
run-concurrent:
	go run cmd/loadtest/main.go \
		-mode=cli \
		-scenario=scenarios/concurrent-transfers.json \
		-workers=500 \
		-duration=120s \
		-ramp-up=20s

# Run web server mode
run-server:
	go run cmd/loadtest/main.go \
		-mode=server \
		-server-port=9999

# Run with process isolation (requires API to be running separately)
run-isolated:
	go run cmd/loadtest/main.go \
		-mode=cli \
		-isolate=true \
		-workers=100 \
		-duration=60s

# Clean build artifacts and reports
clean:
	rm -rf bin/ reports/

# Create report directory
init:
	mkdir -p reports

# Full test suite
test-suite: init
	@echo "Running performance test suite..."
	@echo "1. Default scenario..."
	go run cmd/loadtest/main.go -mode=cli -workers=50 -duration=30s
	@echo "2. High concurrency..."
	go run cmd/loadtest/main.go -mode=cli -workers=200 -duration=30s
	@echo "3. Read heavy..."
	go run cmd/loadtest/main.go -mode=cli -workers=100 -duration=30s -scenario=scenarios/read-heavy.json

# Help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the load test binary"
	@echo "  make deps           - Install dependencies"
	@echo "  make run-cli        - Run in CLI mode with default settings"
	@echo "  make run-high-load  - Run high load scenario"
	@echo "  make run-concurrent - Run concurrent transfers scenario"
	@echo "  make run-server     - Run web UI server on port 9999"
	@echo "  make run-isolated   - Run with process isolation"
	@echo "  make test-suite     - Run full test suite"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make help           - Show this help message"