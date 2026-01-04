// Package selfupdate provides self-update functionality for CLI tools.
package selfupdate

import (
	"fmt"
	"runtime"
	"slices"
)

// Platform OS and architecture constants.
const (
	osWindows = "windows"
	archARM   = "arm"
)

// ArmVersion is set at build time via ldflags for ARM builds.
// For armv6: -ldflags "-X github.com/home-assistant-blueprints/selfupdate.ArmVersion=6"
// For armv7: -ldflags "-X github.com/home-assistant-blueprints/selfupdate.ArmVersion=7"
var ArmVersion string

// Platform represents a supported OS/architecture combination.
type Platform struct {
	OS            string // Operating system: linux, darwin, windows
	Arch          string // Architecture: amd64, arm64, arm
	ARMVersion    string // ARM variant: "6" or "7" for ARM, empty otherwise
	FileExtension string // Binary file extension: ".exe" for Windows, empty otherwise
}

// SupportedArchitectures lists all architectures we build binaries for.
var SupportedArchitectures = []string{
	"linux-amd64",
	"linux-arm64",
	"linux-armv7",
	"linux-armv6",
	"darwin-amd64",
	"darwin-arm64",
	"windows-amd64",
}

// DetectPlatform returns the current platform's OS, architecture, and other details.
// For ARM architectures, it uses the ArmVersion variable that must be set at build time.
func DetectPlatform() (Platform, error) {
	p := Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Set file extension for Windows
	if p.OS == osWindows {
		p.FileExtension = ".exe"
	}

	// Handle ARM version detection
	if p.Arch == archARM {
		if ArmVersion == "" {
			// No ARM version set at build time - this is a problem for ARM builds
			return Platform{}, &ArchitectureError{
				OS:                     p.OS,
				Arch:                   "arm (unknown version)",
				SupportedArchitectures: SupportedArchitectures,
			}
		}
		p.ARMVersion = ArmVersion
	}

	// Validate this is a supported architecture
	assetSuffix := p.AssetSuffix()
	supported := slices.Contains(SupportedArchitectures, assetSuffix)

	if !supported {
		return Platform{}, &ArchitectureError{
			OS:                     p.OS,
			Arch:                   p.ArchString(),
			SupportedArchitectures: SupportedArchitectures,
		}
	}

	return p, nil
}

// ArchString returns the architecture as a string, including ARM version if applicable.
// Examples: "amd64", "arm64", "armv6", "armv7"
func (p Platform) ArchString() string {
	if p.Arch == archARM && p.ARMVersion != "" {
		return fmt.Sprintf("armv%s", p.ARMVersion)
	}
	return p.Arch
}

// AssetSuffix returns the OS-architecture suffix used in asset names.
// Examples: "linux-amd64", "darwin-arm64", "windows-amd64", "linux-armv7"
func (p Platform) AssetSuffix() string {
	return fmt.Sprintf("%s-%s", p.OS, p.ArchString())
}

// AssetName returns the expected asset name for a given tool.
// Examples: "ha-ws-client-linux-arm64", "validate-blueprint-windows-amd64.exe"
func (p Platform) AssetName(toolName string) string {
	name := fmt.Sprintf("%s-%s", toolName, p.AssetSuffix())
	if p.FileExtension != "" {
		name += p.FileExtension
	}
	return name
}

// String returns a human-readable representation of the platform.
func (p Platform) String() string {
	return p.AssetSuffix()
}
