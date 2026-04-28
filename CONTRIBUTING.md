# Contributing to LHM Exporter

Thank you for your interest in contributing to LHM Exporter!

## Development Setup

1. Clone the repository
2. Install dependencies: `go mod download`
3. Run tests: `make test`
4. Build: `make build`

## Code Style

- Follow Go idioms and `gofmt` formatting (`make fmt`)
- Pass `go vet ./...` without warnings
- Add tests for new features
- Update documentation as needed

## Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests and linting (`make test && make lint`)
5. Submit a pull request

## Commit Message Format

- Use English for commit messages
- Use the imperative mood (e.g., "Add feature" not "Added feature")
- Keep the first line under 72 characters

## Reporting Issues

Please include:
- Go version
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs

## Testing

- All new code should have unit tests
- Run benchmarks with `make bench` for performance-sensitive changes
- Ensure `go test ./...` passes before submitting a PR
