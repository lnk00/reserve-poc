# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Run server: `go run main.go`
- Build binary: `go build`
- Format code: `gofmt -w .`
- Run all tests: `go test ./...`
- Run specific test: `go test -run TestName ./path/to/package`

## Code Style Guidelines
- **Formatting**: Use `gofmt` for consistent formatting
- **Imports**: Group imports by standard lib, then third-party
- **Error Handling**: Always check and handle errors explicitly
- **Naming**: 
  - Use camelCase for variables, PascalCase for exported functions/types
  - Prefer meaningful names over abbreviations
- **Documentation**: Add comments for exported functions and types
- **HTTP Handlers**: Follow the standard http.Handler interface pattern
- **Config**: Use JSON for configuration files
- **Types**: Favor strong typing and explicit type declarations

## Project Structure
This project is a Go-based API proxy/reserve proxy defined in the reserve-config.json file.