# Changelog

All notable changes to validate-blueprint-go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
