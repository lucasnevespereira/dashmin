.PHONY: help build run lint clean

APP_NAME = dashmin

help:
	@echo "📊 $(APP_NAME) development commands"
	@echo ""
	@echo "  build      - Build binary to bin/$(APP_NAME)"
	@echo "  run        - Run $(APP_NAME) locally"
	@echo "  lint       - Run code quality checks"
	@echo "  clean      - Remove build artifacts"
	@echo ""

build:
	go build -o bin/$(APP_NAME) .

run:
	go run .

lint:
	golangci-lint run

clean:
	rm -rf bin/
