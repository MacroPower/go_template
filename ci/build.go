package main

import (
	"context"
	"strings"
	"time"

	"dagger/ci/internal/dagger"
)

// Build runs GoReleaser in snapshot mode, producing binaries for all
// platforms. Returns the dist/ directory. Docker, signing, and SBOM stages
// are skipped here -- images are built natively via Dagger in [Ci.BuildImages]
// and signing requires OIDC credentials only available during a real release.
func (m *Ci) Build(ctx context.Context) (*dagger.Directory, error) {
	ctr, err := m.releaserBase(ctx)
	if err != nil {
		return nil, err
	}
	return ctr.
		WithExec([]string{
			"goreleaser", "release", "--snapshot", "--clean",
			"--skip=docker,sign,sbom",
			"--parallelism=0",
		}).
		Directory("/src/dist"), nil
}

// BinarySnapshot builds the go_template binary for a single platform via GoReleaser
// in snapshot mode.
func (m *Ci) BinarySnapshot(
	ctx context.Context,
	// Target build platform (e.g. "linux/arm64").
	platform dagger.Platform,
) (*dagger.File, error) {
	ctr, err := m.releaserBase(ctx)
	if err != nil {
		return nil, err
	}
	goos, goarch, _ := strings.Cut(string(platform), "/")
	return ctr.
		WithEnvVariable("GOOS", goos).
		WithEnvVariable("GOARCH", goarch).
		// GoReleaser does not create the --output parent directory.
		WithDirectory("/out", dag.Directory()).
		WithExec([]string{
			"goreleaser", "build", "--snapshot", "--clean",
			"--single-target", "--output", "/out/go_template",
		}).
		File("/out/go_template"), nil
}

// BuildImages builds multi-arch runtime container images from a GoReleaser
// dist directory. If no dist is provided, a snapshot build is run.
func (m *Ci) BuildImages(
	ctx context.Context,
	// Version label for OCI metadata.
	// +default="snapshot"
	version string,
	// Pre-built GoReleaser dist directory. If not provided, runs a snapshot build.
	// +optional
	dist *dagger.Directory,
) ([]*dagger.Container, error) {
	if dist == nil {
		var err error
		dist, err = m.Build(ctx)
		if err != nil {
			return nil, err
		}
	}

	return runtimeImages(dist, version, ociCreated())
}

// ociCreated renders the current time for the
// org.opencontainers.image.created annotation.
func ociCreated() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// platformDistDir maps a target platform to the GoReleaser dist
// subdirectory holding its go_template binary.
type platformDistDir struct {
	platform dagger.Platform
	distDir  string
}

// platformDistDirs is the release platform matrix. Both runtimeImages
// and [Ci.ReleaseDryRun] range over it so the images that ship
// and the dry-run that verifies them cover the same platforms.
var platformDistDirs = []platformDistDir{
	{platform: "linux/amd64", distDir: "go_template_linux_amd64_v1"},
	{platform: "linux/arm64", distDir: "go_template_linux_arm64_v8.0"},
}

// runtimeImages builds a multi-arch set of runtime container images from a
// pre-built GoReleaser dist/ directory.
func runtimeImages(dist *dagger.Directory, version, created string) ([]*dagger.Container, error) {
	containers := make([]*dagger.Container, len(platformDistDirs))

	for i, p := range platformDistDirs {
		containers[i] = withOCILabels(dag.Container(dagger.ContainerOpts{Platform: p.platform})).
			WithLabel("org.opencontainers.image.version", version).
			WithLabel("org.opencontainers.image.created", created).
			WithAnnotation("org.opencontainers.image.version", version).
			WithAnnotation("org.opencontainers.image.created", created).
			WithFile("/usr/local/bin/go_template", dist.File(p.distDir+"/go_template")).
			WithEntrypoint([]string{"go_template"})
	}

	return containers, nil
}

// withOCILabels applies the static OCI labels and annotations.
func withOCILabels(ctr *dagger.Container) *dagger.Container {
	return ctr.
		WithLabel("org.opencontainers.image.title", "go_template").
		WithLabel("org.opencontainers.image.source", "https://github.com/MacroPower/go_template").
		WithLabel("org.opencontainers.image.url", "https://github.com/MacroPower/go_template").
		WithLabel("org.opencontainers.image.licenses", "Apache-2.0").
		WithAnnotation("org.opencontainers.image.title", "go_template").
		WithAnnotation("org.opencontainers.image.source", "https://github.com/MacroPower/go_template")
}

// releaserBase builds the full release toolset: the shared GoReleaser base
// (the Go build base plus the goreleaser binary, from the [Goreleaser]
// toolchain) extended with cosign, syft, project source, and a bootstrapped
// git repo -- everything goreleaser release needs for signing and SBOMs.
// cosign and syft are folded into the goreleaser toolchain, so its
// WithCosign/WithSyft install those binaries for GoReleaser's sign and sbom
// steps. Config-only validation goes through the [Goreleaser] toolchain
// directly -- see [Ci.LintReleaser].
func (m *Ci) releaserBase(_ context.Context) (*dagger.Container, error) {
	// WithCosign/WithSyft take and return a container, so they are applied as
	// statements rather than chained.
	ctr := m.Goreleaser.GoreleaserBase()
	ctr = m.Goreleaser.WithCosign(ctr)
	ctr = m.Goreleaser.WithSyft(ctr)
	ctr = ctr.
		// Env vars used by GoReleaser ldflags and templates.
		WithEnvVariable("HOSTNAME", "dagger").
		WithEnvVariable("USER", "dagger").
		// Mount source after all tools so that source changes only invalidate
		// layers from here onward, preserving the tool installation layers above.
		WithMountedDirectory("/src", m.Source)
	return m.Goreleaser.EnsureGitRepo(ctr, dagger.GoreleaserEnsureGitRepoOpts{RemoteURL: cloneURL}), nil
}
