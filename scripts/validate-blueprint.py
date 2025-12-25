#!/usr/bin/env python3
"""Validate Home Assistant Blueprint YAML files.

This script performs comprehensive validation of Home Assistant blueprint files:
1. YAML syntax validation
2. Blueprint schema validation (required keys, structure)
3. Input/selector validation
4. Template syntax checking
5. Service call structure validation

Usage:
    python3 validate-blueprint.py <blueprint.yaml>
    python3 validate-blueprint.py --all  # validate all blueprints in repo
"""

from __future__ import annotations

import re
import sys
from pathlib import Path
from typing import Any

import yaml


class HomeAssistantLoader(yaml.SafeLoader):
    """Custom YAML loader that handles Home Assistant's !input tag."""


def _input_constructor(loader: yaml.SafeLoader, node: yaml.Node) -> str:
    """Handle !input tags in YAML."""
    return f"!input {loader.construct_scalar(node)}"  # type: ignore[arg-type]


HomeAssistantLoader.add_constructor("!input", _input_constructor)


class BlueprintValidator:
    """Validates Home Assistant Blueprint YAML files."""

    REQUIRED_BLUEPRINT_KEYS = ["name", "description", "domain", "input"]
    REQUIRED_ROOT_KEYS = ["blueprint", "trigger", "action"]
    VALID_SELECTOR_TYPES = [
        "action",
        "addon",
        "area",
        "attribute",
        "boolean",
        "color_rgb",
        "color_temp",
        "condition",
        "conversation_agent",
        "country",
        "date",
        "datetime",
        "device",
        "duration",
        "entity",
        "file",
        "floor",
        "icon",
        "label",
        "language",
        "location",
        "media",
        "navigation",
        "number",
        "object",
        "select",
        "state",
        "target",
        "template",
        "text",
        "theme",
        "time",
        "trigger",
        "ui_action",
        "ui_color",
    ]

    def __init__(self, file_path: Path) -> None:
        """Initialize the validator.

        Args:
            file_path: Path to the blueprint YAML file.
        """
        self.file_path = file_path
        self.errors: list[str] = []
        self.warnings: list[str] = []
        self.data: dict[str, Any] = {}
        self.join_variables: set[str] = set()

    def validate(self) -> bool:
        """Run all validation checks.

        Returns:
            True if the blueprint is valid (no errors), False otherwise.
        """
        print(f"Validating: {self.file_path}")

        if not self._load_yaml():
            return False

        self._validate_structure()
        self._validate_blueprint_section()
        self._validate_inputs()
        self._validate_variables()
        self._validate_triggers()
        self._validate_actions()
        self._validate_templates()
        self._check_readme_exists()

        return self._report_results()

    def _load_yaml(self) -> bool:
        """Load and parse YAML file.

        Returns:
            True if the file was loaded successfully, False otherwise.
        """
        try:
            with self.file_path.open("r") as f:
                loaded = yaml.load(f, Loader=HomeAssistantLoader)
                if isinstance(loaded, dict):
                    self.data = loaded
                else:
                    self.errors.append("Root must be a dictionary")
                    return False
            return True
        except yaml.YAMLError as e:
            self.errors.append(f"YAML syntax error: {e}")
            return False
        except OSError as e:
            self.errors.append(f"Failed to load file: {e}")
            return False

    def _validate_structure(self) -> None:
        """Validate root-level structure."""
        # Check required root keys
        for key in self.REQUIRED_ROOT_KEYS:
            if key not in self.data:
                self.errors.append(f"Missing required root key: '{key}'")

        # Warn about variables not at root level
        blueprint = self.data.get("blueprint")
        if isinstance(blueprint, dict) and "variables" in blueprint:
            self.errors.append(
                "'variables' must be at root level, not nested under 'blueprint'"
            )

        # Check for variables at root (recommended)
        variables = self.data.get("variables")
        if variables is not None and not isinstance(variables, dict):
            self.errors.append("'variables' must be a dictionary")

    def _validate_blueprint_section(self) -> None:
        """Validate blueprint metadata section."""
        blueprint = self.data.get("blueprint")
        if blueprint is None:
            return

        if not isinstance(blueprint, dict):
            self.errors.append("'blueprint' must be a dictionary")
            return

        # Check required blueprint keys
        for key in self.REQUIRED_BLUEPRINT_KEYS:
            if key not in blueprint:
                self.errors.append(f"Missing required blueprint key: '{key}'")

        # Validate domain
        domain = blueprint.get("domain")
        if domain is not None:
            valid_domains = ["automation", "script"]
            if domain not in valid_domains:
                self.errors.append(
                    f"Invalid domain '{domain}', must be one of: {valid_domains}"
                )

    def _validate_inputs(self) -> None:
        """Validate input definitions."""
        blueprint = self.data.get("blueprint")
        if not isinstance(blueprint, dict):
            return

        inputs = blueprint.get("input")
        if inputs is None:
            return

        if not isinstance(inputs, dict):
            self.errors.append("'blueprint.input' must be a dictionary")
            return

        self._validate_input_dict(inputs, path="blueprint.input")

    def _validate_input_dict(self, inputs: dict[str, Any], path: str) -> None:
        """Recursively validate input definitions.

        Args:
            inputs: Dictionary of input definitions.
            path: Current path for error messages.
        """
        for key, value in inputs.items():
            current_path = f"{path}.{key}"

            if not isinstance(value, dict):
                self.errors.append(f"{current_path}: Input must be a dictionary")
                continue

            # Check if this is an input group or actual input
            nested_input = value.get("input")
            if nested_input is not None:
                # This is a group
                if not isinstance(nested_input, dict):
                    self.errors.append(f"{current_path}.input: Must be a dictionary")
                else:
                    self._validate_input_dict(nested_input, current_path)
            else:
                # This is an actual input definition
                self._validate_single_input(value, current_path)

    def _validate_single_input(self, input_def: dict[str, Any], path: str) -> None:
        """Validate a single input definition.

        Args:
            input_def: The input definition dictionary.
            path: Current path for error messages.
        """
        # Check for selector
        selector = input_def.get("selector")
        if selector is None:
            self.warnings.append(
                f"{path}: No selector defined (inputs should have selectors)"
            )
            return

        if not isinstance(selector, dict):
            self.errors.append(f"{path}.selector: Must be a dictionary")
            return

        # Validate selector type
        for selector_type in selector:
            if selector_type not in self.VALID_SELECTOR_TYPES:
                self.warnings.append(
                    f"{path}.selector: Unknown selector type '{selector_type}'"
                )

    def _validate_variables(self) -> None:
        """Validate variables section."""
        variables = self.data.get("variables")
        if variables is None:
            self.warnings.append("No variables section defined")
            return

        if not isinstance(variables, dict):
            self.errors.append("'variables' must be a dictionary")
            return

        # Record variables that appear to build comma-joined strings
        for name, value in variables.items():
            if isinstance(value, str):
                if re.search(r"\|\s*join\b", value) or re.search(r"\bjoin\s*\(", value):
                    self.join_variables.add(name)

        # Check for blueprint_version
        if "blueprint_version" not in variables:
            self.warnings.append("No 'blueprint_version' variable defined")

    def _validate_triggers(self) -> None:
        """Validate trigger definitions."""
        triggers = self.data.get("trigger")
        if triggers is None:
            return

        if not isinstance(triggers, list):
            self.errors.append("'trigger' must be a list")
            return

        # Get list of variable names for template trigger validation
        variable_names: set[str] = set()
        variables = self.data.get("variables")
        if isinstance(variables, dict):
            variable_names = set(variables.keys())

        for i, trigger in enumerate(triggers):
            if not isinstance(trigger, dict):
                self.errors.append(f"trigger[{i}]: Must be a dictionary")
                continue

            # Check for platform or trigger type
            if "platform" not in trigger and "trigger" not in trigger:
                self.errors.append(f"trigger[{i}]: Missing 'platform' or 'trigger' key")

            entity_id = trigger.get("entity_id")
            if entity_id is not None:
                self._check_trigger_entity_id(entity_id, f"trigger[{i}].entity_id")

            # Check template triggers for automation variable references
            if trigger.get("platform") == "template":
                value_template = trigger.get("value_template", "")
                if isinstance(value_template, str):
                    self._check_template_variable_refs(
                        value_template, variable_names, f"trigger[{i}]"
                    )

    def _check_template_variable_refs(
        self, template: str, variable_names: set[str], path: str
    ) -> None:
        """Check if a template references automation variables.

        Args:
            template: The template string to check.
            variable_names: Set of defined variable names.
            path: Current path for error messages.
        """
        for var_name in variable_names:
            # Look for variable references: {{ var_name or {{var_name
            pattern = rf"\{{\{{\s*{re.escape(var_name)}(?:\s|[\|\(\)\[\]\.,]|$)"
            if re.search(pattern, template):
                self.errors.append(
                    f"{path}: Template trigger references automation variable "
                    f"'{var_name}'. Template triggers cannot access automation "
                    "variables - use state triggers instead."
                )

    def _validate_service_format(self, service: str, path: str) -> None:
        """Validate service format (domain.service).

        Args:
            service: The service string to validate.
            path: Current path for error messages.
        """
        # Allow !input references
        if service.startswith("!input "):
            return

        # Allow templates
        if "{{" in service and "}}" in service:
            return

        # Check format: domain.service_name
        if not re.match(r"^[a-z_][a-z0-9_]*\.[a-z_][a-z0-9_]*$", service):
            self.warnings.append(
                f"{path}: Service '{service}' should be in format 'domain.service_name' "
                "(lowercase letters, numbers, underscores only)"
            )

    def _validate_actions(self) -> None:
        """Validate action definitions."""
        actions = self.data.get("action")
        if actions is None:
            return

        if not isinstance(actions, list):
            self.errors.append("'action' must be a list")
            return

        for i, action in enumerate(actions):
            self._validate_action_item(action, f"action[{i}]")

    def _validate_action_item(self, action: Any, path: str) -> None:
        """Validate a single action item.

        Args:
            action: The action to validate.
            path: Current path for error messages.
        """
        if not isinstance(action, dict):
            self.errors.append(f"{path}: Must be a dictionary")
            return

        # Check for service calls
        service = action.get("service")
        if isinstance(service, str):
            self._validate_service_format(service, f"{path}.service")

            # Must have either target or entity_id (or neither for some services)
            data = action.get("data")
            if data is None and "data" in action:
                self.errors.append(f"{path}.data: Cannot be None/empty")
            elif data is not None and not isinstance(data, dict):
                self.errors.append(f"{path}.data: Must be a dictionary")

            target = action.get("target")
            if isinstance(target, dict):
                entity_id = target.get("entity_id")
                if entity_id is not None:
                    self._check_entity_id_value(entity_id, f"{path}.target.entity_id")

            entity_id = action.get("entity_id")
            if entity_id is not None:
                self._check_entity_id_value(entity_id, f"{path}.entity_id")

        # Check if/then/else structures
        if "if" in action:
            then_actions = action.get("then")
            if isinstance(then_actions, list):
                for j, then_action in enumerate(then_actions):
                    self._validate_action_item(then_action, f"{path}.then[{j}]")

            else_actions = action.get("else")
            if isinstance(else_actions, list):
                for j, else_action in enumerate(else_actions):
                    self._validate_action_item(else_action, f"{path}.else[{j}]")

        # Check repeat structures
        repeat = action.get("repeat")
        if isinstance(repeat, dict):
            seq = repeat.get("sequence")
            if isinstance(seq, list):
                for j, seq_action in enumerate(seq):
                    self._validate_action_item(
                        seq_action, f"{path}.repeat.sequence[{j}]"
                    )

            for_each = repeat.get("for_each")
            if isinstance(for_each, str) and "join" in for_each:
                self.warnings.append(
                    f"{path}.repeat.for_each: uses 'join' which may not produce "
                    "a valid list; ensure it returns a sequence"
                )

        # Check choose structures
        choose = action.get("choose")
        if isinstance(choose, list):
            for j, choice in enumerate(choose):
                if isinstance(choice, dict):
                    choice_seq = choice.get("sequence")
                    if isinstance(choice_seq, list):
                        for k, choice_action in enumerate(choice_seq):
                            self._validate_action_item(
                                choice_action, f"{path}.choose[{j}].sequence[{k}]"
                            )

    def _validate_entity_id_format(self, entity_id: str, path: str) -> None:
        """Validate entity_id format (domain.entity_name).

        Args:
            entity_id: The entity ID to validate.
            path: Current path for error messages.
        """
        # Allow !input references
        if entity_id.startswith("!input "):
            return

        # Allow templates
        if "{{" in entity_id and "}}" in entity_id:
            return

        stripped = entity_id.strip()
        if not stripped:
            return

        # Check format: domain.entity_name (lowercase, numbers, underscores)
        if not re.match(r"^[a-z_][a-z0-9_]*\.[a-z0-9_]+$", stripped):
            self.warnings.append(
                f"{path}: Entity ID '{stripped}' should be in format "
                "'domain.entity_name' (lowercase letters, numbers, underscores only)"
            )

    def _check_trigger_entity_id(self, value: Any, path: str) -> None:
        """Ensure trigger entity_id fields are static strings.

        Args:
            value: The entity_id value to check.
            path: Current path for error messages.
        """
        if value is None:
            return

        if isinstance(value, list):
            for idx, item in enumerate(value):
                self._check_trigger_entity_id(item, f"{path}[{idx}]")
            return

        if not isinstance(value, str):
            return

        stripped = value.strip()
        if not stripped:
            self.errors.append(f"{path}: entity_id cannot be empty")
            return

        if "{{" in stripped or "}}" in stripped:
            self.errors.append(
                f"{path}: entity_id cannot use templates; provide a concrete "
                "entity reference or !input value"
            )
        else:
            # Validate format if it's a static string
            self._validate_entity_id_format(stripped, path)

    def _validate_templates(self) -> None:
        """Validate Jinja2 template syntax."""
        content = self.file_path.read_text()

        # Check for !input inside {{ }}
        if re.search(r"\{\{[^}]*!input", content):
            self.errors.append(
                "Found !input tag inside {{ }} template - bind to variable first"
            )

        # Check for balanced Jinja2 delimiters
        jinja_patterns = [
            ("{{", "}}", "Jinja expressions"),
            ("{%", "%}", "Jinja control blocks"),
            ("{#", "#}", "Jinja comments"),
        ]

        for open_tag, close_tag, name in jinja_patterns:
            open_count = content.count(open_tag)
            close_count = content.count(close_tag)
            if open_count != close_count:
                self.errors.append(
                    f"Unbalanced {name}: {open_tag} appears {open_count} times, "
                    f"{close_tag} appears {close_count} times"
                )

        # Check for common unsupported filters/functions
        unsupported = [
            r"import\s+math",
            r"math\.",
            r"pow\(",
            r"\|\\s*hypot",
            r"\|\\s*clip(?!\w)",  # clip but not clipboard
        ]

        for pattern in unsupported:
            if re.search(pattern, content):
                self.warnings.append(
                    f"Possible use of unsupported function/filter matching "
                    f"pattern: {pattern}"
                )

    def _check_entity_id_value(self, value: Any, path: str) -> None:
        """Validate entity_id fields for multi-entity pitfalls.

        Args:
            value: The entity_id value to check.
            path: Current path for error messages.
        """
        if value is None:
            return

        if isinstance(value, list):
            for idx, item in enumerate(value):
                if not isinstance(item, str):
                    self.errors.append(
                        f"{path}[{idx}]: entity_id entries must be strings"
                    )
            return

        if not isinstance(value, str):
            return

        stripped = value.strip()
        if not stripped:
            return

        if "," in stripped and "{{" not in stripped and not stripped.startswith("["):
            self.errors.append(
                f"{path}: Multiple entity IDs must be provided as a YAML list "
                "or loop, not a comma-separated string"
            )

        if "{{" in stripped and "}}" in stripped:
            if re.search(r"\bjoin\b", stripped):
                self.errors.append(
                    f"{path}: entity_id template uses 'join', which produces an "
                    "invalid comma-separated string; iterate over entities instead"
                )

            match = re.match(r"^\{\{\s*([\w\.]+)", stripped)
            if match:
                var_name = match.group(1)
                if var_name in self.join_variables:
                    self.errors.append(
                        f"{path}: References variable '{var_name}' built with join(); "
                        "entity_id templates cannot combine multiple entities"
                    )

    def _check_readme_exists(self) -> None:
        """Check if README.md exists in the blueprint directory."""
        readme_path = self.file_path.parent / "README.md"
        if not readme_path.exists():
            self.warnings.append(
                f"No README.md found in {self.file_path.parent.name}/ directory"
            )

    def _report_results(self) -> bool:
        """Print validation results and return success status.

        Returns:
            True if no errors were found, False otherwise.
        """
        print()

        if self.errors:
            print("❌ ERRORS:")
            for error in self.errors:
                print(f"  • {error}")
            print()

        if self.warnings:
            print("⚠️  WARNINGS:")
            for warning in self.warnings:
                print(f"  • {warning}")
            print()

        if not self.errors and not self.warnings:
            print("✅ Blueprint is valid!")
            return True

        if not self.errors:
            print(f"✅ Blueprint is valid (with {len(self.warnings)} warnings)")
            return True

        print(f"❌ Blueprint validation failed with {len(self.errors)} errors")
        return False


def find_all_blueprints(base_path: Path) -> list[Path]:
    """Find all blueprint YAML files in the repository.

    Args:
        base_path: The root path to search from.

    Returns:
        A sorted list of paths to blueprint files.
    """
    blueprints: list[Path] = []

    # Look in standard blueprint directories
    for pattern in ["**/*_pro.yaml", "**/*_pro_blueprint.yaml", "**/blueprint.yaml"]:
        blueprints.extend(base_path.glob(pattern))

    # Exclude certain directories
    exclude = {".git", "node_modules", "venv", ".venv", "__pycache__"}
    blueprints = [bp for bp in blueprints if not any(ex in bp.parts for ex in exclude)]

    return sorted(set(blueprints))


def validate_single(blueprint_path: str) -> bool:
    """Validate a single blueprint file.

    Args:
        blueprint_path: Path to the blueprint file.

    Returns:
        True if valid, False otherwise.
    """
    validator = BlueprintValidator(Path(blueprint_path))
    return validator.validate()


def validate_all() -> bool:
    """Validate all blueprints in the repository.

    Returns:
        True if all blueprints are valid, False otherwise.
    """
    repo_root = Path(__file__).parent.parent
    blueprints = find_all_blueprints(repo_root)

    if not blueprints:
        print("No blueprints found in repository")
        return False

    print(f"Found {len(blueprints)} blueprint(s) to validate\n")

    results: list[tuple[Path, bool]] = []
    for bp in blueprints:
        validator = BlueprintValidator(bp)
        success = validator.validate()
        results.append((bp, success))
        print("-" * 80)
        print()

    # Summary
    print("=" * 80)
    print("SUMMARY")
    print("=" * 80)
    passed = sum(1 for _, success in results if success)
    failed = len(results) - passed

    for bp, success in results:
        status = "✅" if success else "❌"
        print(f"{status} {bp.relative_to(repo_root)}")

    print()
    print(f"Total: {len(results)} | Passed: {passed} | Failed: {failed}")

    return failed == 0


def main() -> None:
    """Main entry point."""
    if len(sys.argv) < 2:
        print("Usage: validate-blueprint.py <blueprint.yaml>")
        print("       validate-blueprint.py --all")
        sys.exit(1)

    if sys.argv[1] == "--all":
        success = validate_all()
    else:
        success = validate_single(sys.argv[1])

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
