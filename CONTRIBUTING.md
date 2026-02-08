# Contributing to ynab-cli

Thanks for your interest in contributing! This document covers the basics.

## Development Setup

```bash
git clone https://github.com/joeyhipolito/ynab-cli.git
cd ynab-cli
make build
```

Requires Go 1.21+. No external dependencies.

## Making Changes

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Run `make lint` (formats code and runs vet)
4. Run `make test` to ensure tests pass
5. Submit a pull request

## Code Style

- Follow standard Go conventions (`gofmt`)
- Keep the zero-dependency constraint — standard library only
- Add tests for new commands or API methods
- Use the existing package structure (`internal/api`, `internal/cmd`, etc.)

## Reporting Issues

Open an issue with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Your Go version and OS

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
