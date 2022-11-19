package svu

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
)

const (
	baseImage = "ghcr.io/caarlos0/svu"
)

// Output is svu command output.
type Output struct {
	// Version
	Version string
	// Version without the prefix
	VersionWithoutPrefix string
}

// Run runs the svu command with the given options.
func Run(ctx context.Context, client *dagger.Client, workdir *dagger.Directory, options ...Option) (*Output, error) {
	cfg := defaultConfig()
	for _, o := range options {
		cfg = o(cfg)
	}

	svuFlags := flagsFromConfig(&cfg)

	container := client.Container().
		From(fmt.Sprintf("%s:%s", baseImage, cfg.Version)).
		WithMountedDirectory("/src", workdir).
		WithWorkdir("/src")

	container = container.Exec(dagger.ContainerExecOpts{Args: append([]string{string(cfg.Command)}, svuFlags...)})
	// Run container and get Exit code
	_, err := container.ExitCode(ctx)
	if err != nil {
		return nil, err
	}

	version, err := container.Stdout().Contents(ctx)
	if err != nil {
		return nil, err
	}

	svuFlags = append(svuFlags, "--strip-prefix")
	container = container.Exec(dagger.ContainerExecOpts{Args: append([]string{string(cfg.Command)}, svuFlags...)})
	// Run container and get Exit code
	_, err = container.ExitCode(ctx)
	if err != nil {
		return nil, err
	}

	versionWithoutPrefix, err := container.Stdout().Contents(ctx)
	if err != nil {
		return nil, err
	}

	return &Output{
		Version:              strings.TrimSpace(version),
		VersionWithoutPrefix: strings.TrimSpace(versionWithoutPrefix),
	}, nil
}

func flagsFromConfig(cfg *config) []string {
	var flags []string

	if cfg.Pattern != "" {
		flags = append(flags, "--pattern", cfg.Pattern)
	}
	if cfg.Prefix != "" {
		flags = append(flags, "--prefix", cfg.Prefix)
	}
	if cfg.Suffix != "" {
		flags = append(flags, "--suffix", cfg.Suffix)
	}
	if cfg.TagMode != "" {
		flags = append(flags, "--tag-mode", string(cfg.TagMode))
	}
	if cfg.Metadata {
		flags = append(flags, "--metadata")
	} else {
		flags = append(flags, "--no-metadata")
	}
	if cfg.Prerelease {
		flags = append(flags, "--pre-release")
	} else {
		flags = append(flags, "--no-pre-release")
	}
	if cfg.Build {
		flags = append(flags, "--build")
	} else {
		flags = append(flags, "--no-build")
	}

	return flags
}
