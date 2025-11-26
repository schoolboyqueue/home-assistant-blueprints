#!/usr/bin/env python3
"""
Validate Home Assistant Blueprint YAML files.

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

import sys
import yaml
import re
from pathlib import Path
from typing import Dict, List, Any, Optional


class HomeAssistantLoader(yaml.SafeLoader):
    """Custom YAML loader that handles Home Assistant's !input tag."""
    pass


def input_constructor(loader, node):
    """Handle !input tags in YAML."""
    return f"!input {loader.construct_scalar(node)}"


HomeAssistantLoader.add_constructor('!input', input_constructor)


class BlueprintValidator:
    """Validates Home Assistant Blueprint YAML files."""

    REQUIRED_BLUEPRINT_KEYS = ['name', 'description', 'domain', 'input']
    REQUIRED_ROOT_KEYS = ['blueprint', 'trigger', 'action']
    VALID_SELECTOR_TYPES = [
        'action', 'addon', 'area', 'attribute', 'boolean', 'color_rgb',
        'color_temp', 'condition', 'conversation_agent', 'country', 'date',
        'datetime', 'device', 'duration', 'entity', 'file', 'floor', 'icon',
        'label', 'language', 'location', 'media', 'navigation', 'number',
        'object', 'select', 'state', 'target', 'template', 'text', 'theme',
        'time', 'trigger', 'ui_action', 'ui_color'
    ]

    def __init__(self, file_path: str):
        self.file_path = Path(file_path)
        self.errors: List[str] = []
        self.warnings: List[str] = []
        self.data: Optional[Dict] = None

    def validate(self) -> bool:
        """Run all validation checks. Returns True if valid."""
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

        return self._report_results()

    def _load_yaml(self) -> bool:
        """Load and parse YAML file."""
        try:
            with open(self.file_path, 'r') as f:
                self.data = yaml.load(f, Loader=HomeAssistantLoader)
            return True
        except yaml.YAMLError as e:
            self.errors.append(f"YAML syntax error: {e}")
            return False
        except Exception as e:
            self.errors.append(f"Failed to load file: {e}")
            return False

    def _validate_structure(self):
        """Validate root-level structure."""
        if not isinstance(self.data, dict):
            self.errors.append("Root must be a dictionary")
            return

        # Check required root keys
        for key in self.REQUIRED_ROOT_KEYS:
            if key not in self.data:
                self.errors.append(f"Missing required root key: '{key}'")

        # Warn about variables not at root level
        if 'blueprint' in self.data and isinstance(self.data['blueprint'], dict):
            if 'variables' in self.data['blueprint']:
                self.errors.append(
                    "'variables' must be at root level, not nested under 'blueprint'"
                )

        # Check for variables at root (recommended)
        if 'variables' in self.data:
            if not isinstance(self.data['variables'], dict):
                self.errors.append("'variables' must be a dictionary")

    def _validate_blueprint_section(self):
        """Validate blueprint metadata section."""
        if 'blueprint' not in self.data:
            return

        blueprint = self.data['blueprint']
        if not isinstance(blueprint, dict):
            self.errors.append("'blueprint' must be a dictionary")
            return

        # Check required blueprint keys
        for key in self.REQUIRED_BLUEPRINT_KEYS:
            if key not in blueprint:
                self.errors.append(f"Missing required blueprint key: '{key}'")

        # Validate domain
        if 'domain' in blueprint:
            valid_domains = ['automation', 'script']
            if blueprint['domain'] not in valid_domains:
                self.errors.append(
                    f"Invalid domain '{blueprint['domain']}', must be one of: {valid_domains}"
                )

    def _validate_inputs(self):
        """Validate input definitions."""
        if 'blueprint' not in self.data or 'input' not in self.data['blueprint']:
            return

        inputs = self.data['blueprint']['input']
        if not isinstance(inputs, dict):
            self.errors.append("'blueprint.input' must be a dictionary")
            return

        self._validate_input_dict(inputs, path="blueprint.input")

    def _validate_input_dict(self, inputs: Dict, path: str):
        """Recursively validate input definitions."""
        for key, value in inputs.items():
            current_path = f"{path}.{key}"

            if not isinstance(value, dict):
                self.errors.append(f"{current_path}: Input must be a dictionary")
                continue

            # Check if this is an input group or actual input
            if 'input' in value:
                # This is a group
                if not isinstance(value['input'], dict):
                    self.errors.append(f"{current_path}.input: Must be a dictionary")
                else:
                    self._validate_input_dict(value['input'], current_path)
            else:
                # This is an actual input definition
                self._validate_single_input(value, current_path)

    def _validate_single_input(self, input_def: Dict, path: str):
        """Validate a single input definition."""
        # Check for selector
        if 'selector' not in input_def:
            self.warnings.append(f"{path}: No selector defined (inputs should have selectors)")
            return

        selector = input_def['selector']
        if not isinstance(selector, dict):
            self.errors.append(f"{path}.selector: Must be a dictionary")
            return

        # Validate selector type
        for selector_type in selector.keys():
            if selector_type not in self.VALID_SELECTOR_TYPES:
                self.warnings.append(
                    f"{path}.selector: Unknown selector type '{selector_type}'"
                )

    def _validate_variables(self):
        """Validate variables section."""
        if 'variables' not in self.data:
            self.warnings.append("No variables section defined")
            return

        variables = self.data['variables']
        if not isinstance(variables, dict):
            self.errors.append("'variables' must be a dictionary")
            return

        # Check for blueprint_version
        if 'blueprint_version' not in variables:
            self.warnings.append("No 'blueprint_version' variable defined")

    def _validate_triggers(self):
        """Validate trigger definitions."""
        if 'trigger' not in self.data:
            return

        triggers = self.data['trigger']
        if not isinstance(triggers, list):
            self.errors.append("'trigger' must be a list")
            return

        for i, trigger in enumerate(triggers):
            if not isinstance(trigger, dict):
                self.errors.append(f"trigger[{i}]: Must be a dictionary")
                continue

            # Check for platform or trigger type
            if 'platform' not in trigger and 'trigger' not in trigger:
                self.errors.append(f"trigger[{i}]: Missing 'platform' or 'trigger' key")

    def _validate_actions(self):
        """Validate action definitions."""
        if 'action' not in self.data:
            return

        actions = self.data['action']
        if not isinstance(actions, list):
            self.errors.append("'action' must be a list")
            return

        for i, action in enumerate(actions):
            self._validate_action_item(action, f"action[{i}]")

    def _validate_action_item(self, action: Any, path: str):
        """Validate a single action item."""
        if not isinstance(action, dict):
            self.errors.append(f"{path}: Must be a dictionary")
            return

        # Check for service calls
        if 'service' in action:
            # Must have either target or entity_id (or neither for some services)
            if 'data' in action:
                data = action['data']
                if data is None:
                    self.errors.append(f"{path}.data: Cannot be None/empty")
                elif not isinstance(data, dict):
                    self.errors.append(f"{path}.data: Must be a dictionary")

        # Check if/then/else structures
        if 'if' in action:
            if 'then' in action:
                then_actions = action['then']
                if isinstance(then_actions, list):
                    for j, then_action in enumerate(then_actions):
                        self._validate_action_item(then_action, f"{path}.then[{j}]")

            if 'else' in action:
                else_actions = action['else']
                if isinstance(else_actions, list):
                    for j, else_action in enumerate(else_actions):
                        self._validate_action_item(else_action, f"{path}.else[{j}]")

        # Check repeat structures
        if 'repeat' in action:
            if 'sequence' in action['repeat']:
                seq = action['repeat']['sequence']
                if isinstance(seq, list):
                    for j, seq_action in enumerate(seq):
                        self._validate_action_item(seq_action, f"{path}.repeat.sequence[{j}]")

    def _validate_templates(self):
        """Validate Jinja2 template syntax."""
        content = self.file_path.read_text()

        # Check for !input inside {{ }}
        if re.search(r'\{\{[^}]*!input', content):
            self.errors.append("Found !input tag inside {{ }} template - bind to variable first")

        # Check for unbalanced braces
        open_count = content.count('{{')
        close_count = content.count('}}')
        if open_count != close_count:
            self.errors.append(
                f"Unbalanced template braces: {{ appears {open_count} times, "
                f"}} appears {close_count} times"
            )

        # Check for common unsupported filters/functions
        unsupported = [
            r'import\s+math',
            r'math\.',
            r'pow\(',
            r'\|\\s*hypot',
            r'\|\\s*clip(?!\w)',  # clip but not clipboard
        ]

        for pattern in unsupported:
            if re.search(pattern, content):
                self.warnings.append(
                    f"Possible use of unsupported function/filter matching pattern: {pattern}"
                )

    def _report_results(self) -> bool:
        """Print validation results and return success status."""
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
        elif not self.errors:
            print(f"✅ Blueprint is valid (with {len(self.warnings)} warnings)")
            return True
        else:
            print(f"❌ Blueprint validation failed with {len(self.errors)} errors")
            return False


def find_all_blueprints(base_path: Path) -> List[Path]:
    """Find all blueprint YAML files in the repository."""
    blueprints = []

    # Look in standard blueprint directories
    for pattern in ['**/*_pro.yaml', '**/*_pro_blueprint.yaml', '**/blueprint.yaml']:
        blueprints.extend(base_path.glob(pattern))

    # Exclude certain directories
    exclude = {'.git', 'node_modules', 'venv', '.venv', '__pycache__'}
    blueprints = [
        bp for bp in blueprints
        if not any(ex in bp.parts for ex in exclude)
    ]

    return sorted(set(blueprints))


def main():
    """Main entry point."""
    if len(sys.argv) < 2:
        print("Usage: validate-blueprint.py <blueprint.yaml>")
        print("       validate-blueprint.py --all")
        sys.exit(1)

    if sys.argv[1] == '--all':
        # Validate all blueprints in repository
        repo_root = Path(__file__).parent.parent
        blueprints = find_all_blueprints(repo_root)

        if not blueprints:
            print("No blueprints found in repository")
            sys.exit(1)

        print(f"Found {len(blueprints)} blueprint(s) to validate\n")

        results = []
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

        sys.exit(0 if failed == 0 else 1)
    else:
        # Validate single blueprint
        blueprint_path = sys.argv[1]
        validator = BlueprintValidator(blueprint_path)
        success = validator.validate()
        sys.exit(0 if success else 1)


if __name__ == '__main__':
    main()
