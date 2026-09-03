# Changelog

All notable changes to this project are documented here.

## [0.2.0] - Unreleased

### Breaking changes

- The public package moved from `github.com/J-Bockhofer/goe2e/pkg` to `github.com/J-Bockhofer/goe2e`.
- Request behavior now belongs to `RequestConfig`. Pass it through `TestConfig.Request` for package-level `TestRequest` calls.

### Added

- Stateful `Session` support for multi-request, cookie-backed workflows, including defaults and response JSON-pointer extraction.
- Exact request and response body assertion factories: `AssertRequestBodyEquals` and `AssertResponseBodyEquals`.

### Changed

- `TestRequest` returns the executed `*RequestHandler`, allowing response data to feed a subsequent request.
