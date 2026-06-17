# ci

This repository's own CI module, registered as `ci` in the root `dagger.json`.
It is not designed for remote consumption: it orchestrates the repo's
`dagger -> devbox -> task` flow so CI reproduces exactly what `task check:all`
runs locally, and it owns go_template's release pipeline.

## Functions

### Checks (run via devbox)

- `lint`, `test` (both +check) run the matching Taskfile target inside the
  project's devbox environment via the `devbox` toolchain, with the Go
  module/build and golangci-lint caches mounted.
- `test-coverage` runs the coverage target the same way and returns the coverage
  profile file.
- `lint-renovate` (+check) validates the Renovate configuration with
  renovate-config-validator at a pinned version in a Node container — the one
  gate that runs through neither devbox nor a shared toolchain, so Renovate can
  bump its own validator.

### Lint actions & Security (compose sibling toolchains)

These gates compose a sibling toolchain directly rather than running through
devbox, because their tools are not on the devbox PATH — the same pattern the
release functions use for `goreleaser`.

- `lint-actions` (+check) lints the GitHub Actions workflows by composing the
  `zizmor` toolchain. It pins `.github/zizmor.yaml` as the config path.
- `security` (+check) scans source dependencies for known vulnerabilities by
  composing the `security` toolchain (Trivy). `security-source-sarif` and
  `security-image-sarif` are the non-gating counterparts that emit SARIF for
  GitHub Code Scanning; the image SARIF builds the runtime image the way a
  release publishes it and surfaces OS-layer CVEs the source scan cannot see.
- `lint-releaser` (+check) runs `goreleaser check` via the `goreleaser`
  toolchain.

### Release pipeline (composes the goreleaser toolchain; see `build.go` + `publish.go`)

The release pipeline builds the binaries with GoReleaser, then constructs and
publishes the multi-arch runtime images natively via Dagger — GoReleaser's
Docker support is skipped to avoid Docker-in-Docker. It composes the
`goreleaser` toolchain directly, which carries the folded-in cosign signing and
syft SBOM tooling (`with-cosign`/`with-syft`/`sign-keyless`).

- `releaserBase` (private, in `build.go`) builds the release container: the
  goreleaser Go base + cosign + syft, then mounts the source and bootstraps a
  git repo via `ensure-git-repo` (now on the goreleaser toolchain).
- `build` snapshot-cross-compiles the binaries (no publishing).
  `binary` / `binary-snapshot` produce a single-platform binary the same way.
- `release --tag=vX.Y.Z` runs GoReleaser for binaries/archives/SBOMs/signing,
  creates the GitHub release, builds the multi-arch runtime image from the dist
  binaries (scratch base + the go_template binary), publishes it natively via
  Dagger, and signs the published digests with cosign keyless signing. Signing
  is keyless (Sigstore Fulcio + Rekor): the workflow forwards the GitHub Actions
  OIDC token; with no token the release is unsigned.
- `publish-images` publishes pre-built runtime images under the supplied tags.
- `release-dry-run` (non-+check) builds the snapshot, verifies each binary's
  architecture matches its target platform, and constructs the runtime images.

The container image build (scratch + binary) is the preserved release artifact.

## Layout

- `main.go` defines the `Ci` module (Go module path `dagger/ci`), the check
  functions that run via devbox, the version constants, and `Binary`.
- `check.go` holds the remaining `+check` functions and the manual
  `release-dry-run`.
- `build.go` holds `Build`/`BinarySnapshot`/`BuildImages`, the runtime image
  builders, and `releaserBase`.
- `publish.go` holds `Release`/`PublishImages` and the image publish/sign
  helpers.
- Dependencies in `dagger.json`: the `devbox` toolchain (checks), the
  `goreleaser` toolchain (release, carrying cosign + syft), the `security`
  toolchain (the vulnerability scan), and the `zizmor` toolchain (the Actions
  workflow lint), all referenced remotely from `github.com/MacroPower/x`.
- The `tests/` submodule exercises the +check functions (build dist, image
  metadata, lint-releaser, binary, lint-actions).

The `engineVersion` in `dagger.json` is pinned in lockstep with the root
`dagger.json` and with the CLI version in `.github/workflows`; bump them together
via `task dagger:update VERSION=<tag>`.
