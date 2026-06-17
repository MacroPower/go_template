# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
task format    # Format and tidy code, run generators
task lint      # golangci-lint, go mod tidy check, prettier
task test      # Run unit tests
task check     # Local gate: lint + test (tools on the devbox PATH, no Dagger)
task check:all # Everything CI runs (adds security + GitHub config + releaser, via Dagger)
task build     # Cross-compile snapshot binaries + runtime images via Dagger
```

## Code Style

### Go Conventions

- Document all exported items with doc comments.
- Package documentation in `doc.go` files.
- Wrap errors with `fmt.Errorf("context: %w", err)`, or `fmt.Errorf("%w: %w", ErrSentinel, err)`.
- Avoid using "failed" or "error" in library error messages.
- Use global error variables for common errors.
- Use constructors with functional options.
- Accept interfaces, return concrete types.
- Prefer consistency over performance, avoid "fast paths" that could lead to unpredictable behavior.

### Documentation

- Use `[Name]` syntax for Go doc links. Use `[*Name]` for pointer types.
- Constructors should always begin: `// NewThing creates a new [Thing].`
- Types with constructors should always note: `// Create instances with [NewThing].`
- Interfaces should note: `// See [Thing] for an implementation.`
- Interfaces should have sensible names: `type Builder interface { Build() Thing } // Builder builds [Thing]s.`
- Functional option types should have a list linking to all functions of that type.
- Functional options should always have a link to their type.
- Package docs should explain concepts and usage patterns; **do not enumerate exports**.

### Testing

- Use `github.com/stretchr/testify/assert` and `require`.
- Table-driven tests with `map[string]struct{}` format.
- Field names: prefer `want` for expected output, `err` for expected errors.
- For inputs, use clear contextual names (e.g., `before`/`after` for diffs, `line`/`col` for positions).
- Always use `t.Parallel()` in all tests.
- Create test packages (`package foo_test`) testing public API.
- Use `require.ErrorIs` for sentinel error checking.
- Use `require.ErrorAs` for error type extraction.
