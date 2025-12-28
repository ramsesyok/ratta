# ratta

## About

ratta is a desktop application for managing issues related to a single project shared between the Contractor
and the Vendor. It replaces spreadsheet-based tracking by keeping issue comments and status changes recorded
in a clear work log, without relying on cloud services.

## Development

Prerequisites: Go, Node.js, and Wails CLI.

Common commands:

- `wails dev` (or `make dev`) to start the desktop app in development mode
- `go test ./...` to run backend unit tests
- `cd frontend && npm test` to run frontend unit tests
- `make fmt` to format Go and frontend sources

## Building

To build a redistributable, production mode package, use `wails build`.
