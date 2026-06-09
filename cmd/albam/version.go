package main

import (
	"flag"
	"fmt"
	"runtime"
	"runtime/debug"
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
	return versionInfoFromBuildInfo(buildInfo, ok, versionInfo{
		Version: version,
		Commit:  commit,
		BuiltAt: builtAt,
	})
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
