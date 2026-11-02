# Contributing to go-csupa

Thank you for considering contributing to `go-csupa`! We welcome contributions from everyone.

## Getting Started

1. **Fork the repository** on GitHub.
2. **Clone your fork**:
   ```bash
   git clone https://github.com/<your-username>/go-csupa.git
   cd go-csupa
   ```
3. **Create a branch**:
   ```bash
   git checkout -b feature/my-new-feature
   # or
   git checkout -b fix/my-bug-fix
   ```

## Development Guidelines

- **Standard Library First**: Keep external dependencies to zero unless strictly necessary and discussed beforehand.
- **Code Formatting**: Ensure code adheres to standard Go conventions (`go fmt ./...`).
- **Testing**: Add or update unit tests for any new logic or bug fixes (`go test -v ./...`).
- **Linting**: Ensure `go vet ./...` passes without errors or warnings.

## Submitting Pull Requests

1. Commit your changes with clear, descriptive commit messages.
2. Push your branch to your GitHub fork:
   ```bash
   git push origin feature/my-new-feature
   ```
3. Open a Pull Request against the `main` branch of `kevadamar/go-csupa`.
4. Provide a clear description of the problem solved or feature added in the PR description template.
