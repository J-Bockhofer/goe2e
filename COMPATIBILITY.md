# Compatibility and releases

## Current status

`goe2e` is pre-v1. The next tagged release should be `v0.2.0`; until `v1.0.0`, the project follows the Go module convention for pre-v1 APIs.

## Go version support

The `go` directive in `go.mod` is the minimum supported Go version. It currently requires Go 1.27 because the in-memory `httptest.NewTestServer` integration depends on that release.

CI runs the Go version declared in `go.mod`. A change to that directive is a compatibility change and must be called out in the release notes.

## Public API policy

The exported API consists of the types and functions in the module root, including `RequestConfig`, `TestConfig`, `SessionConfig`, `Session`, request/response modifiers, lifecycle statement factories, and diagnostics types.

- Additive API changes are suitable for a patch or minor release when they do not change existing behavior.
- Before `v1.0.0`, incompatible changes require a new minor version and a migration note.
- From `v1.0.0`, incompatible changes require a new major version and the corresponding Go module import path.
- Deprecated exported APIs remain available until the next planned incompatible release. Deprecations must name the replacement and include a migration note.

## Release checklist

Before a release:

1. Run `go test ./... -count=1` and `go vet ./...`.
2. Update the changelog/release notes with user-visible changes, compatibility notes, and migrations.
3. Tag the release with a semantic version such as `v0.2.0`.
4. Verify the tagged module and README links from a clean checkout.

This policy deliberately keeps the initial commitment small: stable APIs, clear migrations, and explicit Go-version requirements.
