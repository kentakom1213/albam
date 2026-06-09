package main

import (
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
)

var (
	version = "dev"
	commit  = "unknown"
	builtAt = "unknown"
)

type versionInfo struct {
	Version string
	Commit  string
	BuiltAt string
}

func currentVersionInfo() versionInfo {
	buildInfo, ok := debug.ReadBuildInfo()
	info := versionInfoFromBuildInfo(buildInfo, ok, versionInfo{
		Version: version,
		Commit:  commit,
		BuiltAt: builtAt,
	})
	if info.Version == "dev" && !buildInfoVCSModified(buildInfo, ok) {
		if tag, err := currentGitTag(); err == nil && tag != "" {
			info.Version = tag
		}
	}
	return info
}

func versionInfoFromBuildInfo(buildInfo *debug.BuildInfo, ok bool, fallback versionInfo) versionInfo {
	info := fallback
	if info.Version == "" {
		info.Version = "dev"
	}
	if info.Commit == "" {
		info.Commit = "unknown"
	}
	if info.BuiltAt == "" {
		info.BuiltAt = "unknown"
	}

	if !ok || buildInfo == nil {
		return info
	}

	if (info.Version == "" || info.Version == "dev") && buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
		info.Version = buildInfo.Main.Version
	}

	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			if info.Commit == "unknown" && setting.Value != "" {
				info.Commit = setting.Value
			}
		case "vcs.time":
			if info.BuiltAt == "unknown" && setting.Value != "" {
				info.BuiltAt = setting.Value
			}
		}
	}

	return info
}

func buildInfoVCSModified(buildInfo *debug.BuildInfo, ok bool) bool {
	if !ok || buildInfo == nil {
		return false
	}

	for _, setting := range buildInfo.Settings {
		if setting.Key == "vcs.modified" {
			return setting.Value == "true"
		}
	}

	return false
}

func currentGitTag() (string, error) {
	output, err := exec.Command("git", "describe", "--tags", "--exact-match", "HEAD").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func runVersion(args []string) error {
	fs := newFlagSet("version", "usage: albam version [--verbose]")

	var verbose bool
	fs.BoolVar(&verbose, "verbose", false, "print detailed version information")

	if err := parseFlags(fs, args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: albam version [--verbose]")
	}

	info := currentVersionInfo()
	fmt.Printf("albam %s\n", info.Version)
	if verbose {
		fmt.Printf("commit: %s\n", info.Commit)
		fmt.Printf("built: %s\n", info.BuiltAt)
		fmt.Printf("go: %s\n", runtime.Version())
	}

	return nil
}
