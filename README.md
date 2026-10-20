# alpinium-go

Go monorepo with Alpinium libraries. Every library is a separate Go module
with its own `go.mod` and its own version.

| Directory | Module | Description |
|---|---|---|
| `Cast` | `github.com/alpinium/alpinium-go/Cast` | safe conversions between bool, numbers and string with range checking |

## Installation

```bash
go get github.com/alpinium/alpinium-go/Cast@v0.1.0
```

## Local development

```bash
cd Cast
go vet ./...
go test -count=1 ./...
```

`go.work` is listed in `.gitignore`. CI and the release workflow run with `GOWORK=off`.

## Versioning and releases

The version of a module is a git tag prefixed with its directory, for example
`Cast/v0.2.0`. A release is started by pushing the tag:

```bash
git tag Cast/v0.2.0
git push origin Cast/v0.2.0
```

The `Release` workflow (`.github/workflows/release.yml`) runs for every tag
`<directory>/vX.Y.Z`. It checks the version format and that the tag is on
`main`, runs `go vet`, `go test -race` and `go mod tidy -diff`, creates a
GitHub Release (`v0.x` and `-rc` versions are marked as pre-releases) and
warms up the Go module proxy.

Rules ([semver](https://semver.org/)):

- Use `v0.x.y` while the API may still change. A new feature or a behaviour change bumps `x`, a bug fix bumps `y`.
- From `v1.0.0`: a bug fix bumps `y`, a backward compatible feature bumps `x`, a breaking change bumps the major version.
- `v2` and above require the module path `.../Cast/v2` in `go.mod` and a tag `Cast/v2.0.0`.
- A tag is immutable. A faulty version cannot be overwritten, only superseded by a new one (optionally with `retract` in `go.mod`).

## License

Apache License 2.0, see [LICENSE](LICENSE).
