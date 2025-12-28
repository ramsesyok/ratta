.PHONY: fmt test test-go test-frontend dev

fmt:
	gofmt -w .
	cd frontend && npm run format

test: test-go test-frontend

test-go:
	go test ./...

test-frontend:
	cd frontend && npm test

dev:
	wails dev
