package selfupdate

import (
	"errors"
	"runtime"
	"testing"
)

func TestPlatform_ArchString(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     string
	}{
		{
			name:     "amd64",
			platform: Platform{OS: "linux", Arch: "amd64"},
			want:     "amd64",
		},
		{
			name:     "arm64",
			platform: Platform{OS: "darwin", Arch: "arm64"},
			want:     "arm64",
		},
		{
			name:     "armv6",
			platform: Platform{OS: "linux", Arch: "arm", ARMVersion: "6"},
			want:     "armv6",
		},
		{
			name:     "armv7",
			platform: Platform{OS: "linux", Arch: "arm", ARMVersion: "7"},
			want:     "armv7",
		},
		{
			name:     "arm without version",
			platform: Platform{OS: "linux", Arch: "arm"},
			want:     "arm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.platform.ArchString()
			if got != tt.want {
				t.Errorf("ArchString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlatform_AssetSuffix(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     string
	}{
		{
			name:     "linux amd64",
			platform: Platform{OS: "linux", Arch: "amd64"},
			want:     "linux-amd64",
		},
		{
			name:     "darwin arm64",
			platform: Platform{OS: "darwin", Arch: "arm64"},
			want:     "darwin-arm64",
		},
		{
			name:     "linux armv7",
			platform: Platform{OS: "linux", Arch: "arm", ARMVersion: "7"},
			want:     "linux-armv7",
		},
		{
			name:     "windows amd64",
			platform: Platform{OS: "windows", Arch: "amd64", FileExtension: ".exe"},
			want:     "windows-amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.platform.AssetSuffix()
			if got != tt.want {
				t.Errorf("AssetSuffix() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlatform_AssetName(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		toolName string
		want     string
	}{
		{
			name:     "ha-ws-client linux amd64",
			platform: Platform{OS: "linux", Arch: "amd64"},
			toolName: "ha-ws-client",
			want:     "ha-ws-client-linux-amd64",
		},
		{
			name:     "validate-blueprint darwin arm64",
			platform: Platform{OS: "darwin", Arch: "arm64"},
			toolName: "validate-blueprint",
			want:     "validate-blueprint-darwin-arm64",
		},
		{
			name:     "ha-ws-client linux armv7",
			platform: Platform{OS: "linux", Arch: "arm", ARMVersion: "7"},
			toolName: "ha-ws-client",
			want:     "ha-ws-client-linux-armv7",
		},
		{
			name:     "ha-ws-client windows amd64",
			platform: Platform{OS: "windows", Arch: "amd64", FileExtension: ".exe"},
			toolName: "ha-ws-client",
			want:     "ha-ws-client-windows-amd64.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.platform.AssetName(tt.toolName)
			if got != tt.want {
				t.Errorf("AssetName(%q) = %q, want %q", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestDetectPlatform(t *testing.T) {
	// Save and restore global ArmVersion
	originalArmVersion := ArmVersion
	defer func() { ArmVersion = originalArmVersion }()

	// Current platform should be detected correctly (assuming we're on a supported arch)
	// We can only really test the current platform in a unit test
	t.Run("current platform", func(t *testing.T) {
		// For ARM platforms, set the ARM version
		if runtime.GOARCH == "arm" {
			ArmVersion = "7" // Assume armv7 for testing
		}

		p, err := DetectPlatform()
		// Check if we're on a supported platform
		if err != nil {
			var archErr *ArchitectureError
			if errors.As(err, &archErr) {
				t.Skipf("Current platform %s-%s is not supported", runtime.GOOS, runtime.GOARCH)
			}
			t.Fatalf("DetectPlatform() error = %v", err)
		}

		if p.OS != runtime.GOOS {
			t.Errorf("OS = %q, want %q", p.OS, runtime.GOOS)
		}

		if p.Arch != runtime.GOARCH {
			t.Errorf("Arch = %q, want %q", p.Arch, runtime.GOARCH)
		}

		// Windows should have .exe extension
		if runtime.GOOS == "windows" && p.FileExtension != ".exe" {
			t.Errorf("FileExtension = %q, want %q", p.FileExtension, ".exe")
		}

		// Non-Windows should have no extension
		if runtime.GOOS != "windows" && p.FileExtension != "" {
			t.Errorf("FileExtension = %q, want empty", p.FileExtension)
		}
	})

	t.Run("arm without version returns error", func(t *testing.T) {
		if runtime.GOARCH != "arm" {
			t.Skip("Skipping ARM-specific test on non-ARM platform")
		}

		ArmVersion = "" // Clear ARM version

		_, err := DetectPlatform()
		if err == nil {
			t.Error("DetectPlatform() should return error for ARM without version")
		}

		if !errors.Is(err, ErrUnsupportedArchitecture) {
			t.Errorf("error should be ErrUnsupportedArchitecture, got %v", err)
		}
	})
}

func TestSupportedArchitectures(t *testing.T) {
	// Verify all expected architectures are in the list
	expected := []string{
		"linux-amd64",
		"linux-arm64",
		"linux-armv7",
		"linux-armv6",
		"darwin-amd64",
		"darwin-arm64",
		"windows-amd64",
	}

	if len(SupportedArchitectures) != len(expected) {
		t.Errorf("SupportedArchitectures has %d entries, want %d", len(SupportedArchitectures), len(expected))
	}

	for _, arch := range expected {
		found := false
		for _, supported := range SupportedArchitectures {
			if supported == arch {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected architecture %q not found in SupportedArchitectures", arch)
		}
	}
}

func TestPlatform_String(t *testing.T) {
	p := Platform{OS: "linux", Arch: "amd64"}
	if got := p.String(); got != "linux-amd64" {
		t.Errorf("String() = %q, want %q", got, "linux-amd64")
	}
}
