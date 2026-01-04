# Changelog

All notable changes to validate-blueprint-go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.8.0] - 2026-01-04

### Added

- Add `update --list` flag to list all available versions from GitHub releases

## [1.7.0] - 2026-01-04

### Added

- Add `update` subcommand for self-updating the binary from GitHub releases
  - `update --check` to check for available updates
  - `update --version=X.Y.Z` to install a specific version
  - `update` to install the latest version
- Add shared `selfupdate` package in `scripts/go-tools/selfupdate` for update functionality

### Changed

- Update build system to embed ARM version for correct binary selection on ARM platforms

## [1.6.2] - 2026-01-03

### Changed

- Extract test fixtures to shared `scripts/testfixtures` package for code reuse across Go tools

## [1.6.1] - 2026-01-03

### Added

- Centralized error handling system with new `internal/errors` package
  - `Error` type with typed error categories (Syntax, Schema, Validation, Parsing, Reference, Template, Input, Trigger, Condition, Action, Documentation, Internal)
  - Error registry pattern for consistent error creation
  - Path-aware errors with `WithPath()` for precise error location
  - Warning support via `AsWarning()` for non-blocking issues
  - Helper functions for error type checking (`IsSyntax`, `IsSchema`, `IsTemplate`, etc.)

### Changed

- Refactor all validator modules to use centralized error types
- Update reporter to leverage typed error information for categorization

## [1.6.0] - 2026-01-03

### Added

- Graceful shutdown coordination with new `internal/shutdown` package
  - `Coordinator` type for managing shutdown lifecycle with configurable grace periods
  - Signal handling for SIGINT/SIGTERM with proper cleanup
  - Context-aware validation for interruptible batch operations
- Context propagation for cancellation support during batch validation
  - `runValidateAllWithContext` replaces `runValidateAll` for interrupt handling
- User feedback during batch validation ("Ctrl+C to interrupt" message)
- Exit code 130 for SIGINT interruption (standard Unix convention)

### Changed

- Refactor main.go to use shutdown coordinator for signal handling
- Update batch validation to check context for early termination

## [1.5.0] - 2026-01-03

### Added

- Strongly-typed struct definitions for blueprint data structures (`types.go`)
- `BlueprintData` struct as root type with typed fields for triggers, conditions, actions
- Comprehensive `Selector` type with all Home Assistant selector variants
- `RawData` and `AnyList` type aliases for improved code documentation
- Typed selector support in ValidationContext (`TypedInputSelectors`)

### Changed

- Replace `map[string]interface{}` with `RawData` type alias throughout codebase
- Update test fixtures to use common type aliases
- Add `TypedData` field to ValidationContext for gradual migration to type-safe code

## [1.4.0] - 2026-01-03

### Added

- ValidationContext struct for centralized state management during validation
- Context-based validation API with cleaner separation of concerns
- Comprehensive context tests (`context_test.go`)

### Changed

- Refactor validator to use ValidationContext pattern for cleaner state handling
- Simplify main.go CLI logic with switch statement refactoring
- Update urfave/cli dependency

### Fixed

- Fix unused parameter lint warning in runValidation
- Fix unchecked error returns for cli.ShowAppHelp
- Fix if-else chain converted to switch statement per linter recommendation

## [1.3.0] - 2026-01-03

### Added

- Error categorization system with category types (Syntax, Schema, References, Templates, Inputs, Triggers, Conditions, Actions, Documentation)
- CategorizedError and CategorizedWarning types for enhanced error reporting
- Category-based error grouping and filtering functions
- Enhanced reporter with category summary output

### Changed

- Refactored test files to use shared test fixtures and reduce duplication
- Enhanced reporter to support both flat and categorized error display

## [1.2.0] - 2026-01-02

### Changed

- Extract shared validation helpers into internal/common package
- Refactor validator files to use common.TraverseValue for consistent traversal
- Add MergeValidationResult helper for integrating common validators

## [1.1.0] - 2026-01-02

### Changed

- Refactor validation logic into internal/validator package for better organization
- Extract constants, helpers, and validators into focused files
- Update module path to match ha-ws-client-go pattern

## [1.0.0] - 2025-01-01

### Added

- Initial release of validate-blueprint-go
- Comprehensive Home Assistant Blueprint YAML validation
- YAML syntax validation with custom !input tag support
- Blueprint schema validation (required keys, structure)
- Input/selector validation with support for all Home Assistant selector types
- Template syntax checking (balanced delimiters, no !input inside {{ }})
- Service call structure validation
- Version sync validation (blueprint name vs blueprint_version variable)
- Trigger validation (platform types, entity_id format, template restrictions)
- Condition validation (type checking, nested conditions)
- Mode validation (single, restart, queued, parallel)
- Input reference validation (undefined input detection)
- Hysteresis boundary validation (on/off threshold checking)
- Variable definition validation with safety checks for math operations
- README.md and CHANGELOG.md existence checks
- Batch validation with --all flag
- Cross-platform builds for Linux (amd64, arm64, armv7, armv6), macOS (amd64, arm64), Windows (amd64)
- Version information with --version flag
