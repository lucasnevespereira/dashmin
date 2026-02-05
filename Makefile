.PHONY: help build run lint clean release

APP_NAME = dashmin

help:
	@echo "📊 $(APP_NAME) development commands"
	@echo ""
	@echo "  build      - Build binary to bin/$(APP_NAME)"
	@echo "  run        - Run $(APP_NAME) locally"
	@echo "  lint       - Run code quality checks"
	@echo "  clean      - Remove build artifacts"
	@echo "  release    - Create a new release"
	@echo ""

build:
	go build -o bin/$(APP_NAME) .

run:
	go run .

lint:
	golangci-lint run

clean:
	rm -rf $(APP_NAME)
	rm -rf bin/

release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=0.3.0"; \
		exit 1; \
	fi
	@echo "Releasing v$(VERSION)..."
	@sed -i '' 's/const Version = .*/const Version = "$(VERSION)"/' cmd/version.go
	@git add -A
	@git commit -m "chore: bump version to $(VERSION)"
	@git tag -a "v$(VERSION)" -m "v$(VERSION)"
	@git push origin main
	@git push origin "v$(VERSION)"
	@echo ""
	@echo "Tag pushed. Now create the GitHub release:"
	@echo "  gh release create v$(VERSION) --generate-notes"
