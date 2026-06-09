package main

import (
	"runtime/debug"
	"testing"
)

func TestVersionInfoFromBuildInfoUsesModuleVersion(t *testing.T) {
	info := versionInfoFromBuildInfo(&debug.BuildInfo{
		Main: debug.Module{
			Version: "v0.2.3",
		},
	}, true, versionInfo{
		Version: "dev",
		Commit:  "unknown",
		BuiltAt: "unknown",
	})

	if info.Version != "v0.2.3" {
		t.Fatalf("Version = %q, want %q", info.Version, "v0.2.3")
	}
}

func TestVersionInfoFromBuildInfoKeepsLdflagsVersion(t *testing.T) {
	info := versionInfoFromBuildInfo(&debug.BuildInfo{
		Main: debug.Module{
			Version: "v0.2.3",
		},
	}, true, versionInfo{
		Version: "custom",
		Commit:  "unknown",
		BuiltAt: "unknown",
	})

	if info.Version != "custom" {
		t.Fatalf("Version = %q, want %q", info.Version, "custom")
	}
}

func TestVersionInfoFromBuildInfoUsesVCSSettings(t *testing.T) {
	info := versionInfoFromBuildInfo(&debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef123456"},
			{Key: "vcs.time", Value: "2026-06-09T12:00:00Z"},
		},
	}, true, versionInfo{
		Version: "dev",
		Commit:  "unknown",
		BuiltAt: "unknown",
	})

	if info.Commit != "abcdef123456" {
		t.Fatalf("Commit = %q, want %q", info.Commit, "abcdef123456")
	}
	if info.BuiltAt != "2026-06-09T12:00:00Z" {
		t.Fatalf("BuiltAt = %q, want %q", info.BuiltAt, "2026-06-09T12:00:00Z")
	}
}

func TestBuildInfoVCSModified(t *testing.T) {
	modified := buildInfoVCSModified(&debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: "vcs.modified", Value: "true"},
		},
	}, true)

	if !modified {
		t.Fatal("buildInfoVCSModified = false, want true")
	}
}
