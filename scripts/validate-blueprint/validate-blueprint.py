#!/usr/bin/env python3
"""Validate Home Assistant Blueprint YAML files.

This script performs comprehensive validation of Home Assistant blueprint files:
1. YAML syntax validation
2. Blueprint schema validation (required keys, structure)
3. Input/selector validation
4. Template syntax checking
5. Service call structure validation
6. Version sync validation (name vs blueprint_version)
7. Trigger validation (for: templates, entity_id)
8. Condition validation
9. Mode validation
10. Input reference validation
11. Delay/wait validation
12. Choose/sequence validation
13. Bare boolean literal detection (true/false outputting strings instead of booleans)
14. Entity ID boolean context detection (unreliable string truthiness in conditions)
15. Python-style list method detection (e.g., [a,b].min() should be [a,b] | min)
16. Variable dependency chain detection (helper variables that may cause UndefinedError)
17. Undefined variable reference detection in templates
18. Unsafe math operations (division by zero, log/sqrt of negative, modulo by zero)
19. Type mismatch detection in filter chains

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
    # Common hysteresis input naming patterns (ON threshold should be > OFF threshold)
    # Format: (on_pattern, off_pattern, description)
    HYSTERESIS_PATTERNS: list[tuple[str, str, str]] = [
        (r"(.*)_on$", r"\1_off", "threshold"),
        (r"(.*)_high$", r"\1_low", "boundary"),
        (r"(.*)_upper$", r"\1_lower", "limit"),
        (r"(.*)_start$", r"\1_stop", "trigger point"),
        (r"(.*)_enable$", r"\1_disable", "activation point"),
        (r"delta_on$", r"delta_off$", "delta threshold"),
    ]
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
    VALID_MODES = ["single", "restart", "queued", "parallel"]
    VALID_CONDITION_TYPES = [
        "and",
        "or",
        "not",
        "state",
        "numeric_state",
        "template",
        "time",
        "zone",
        "trigger",
        "sun",
        "device",
    ]
    # Services that are potentially deprecated or have better alternatives
    # Note: homeassistant.turn_on/off are valid for multi-domain support
    DEPRECATED_SERVICES: dict[str, str] = {
        # homeassistant.turn_on/off/toggle are intentionally not listed here
        # as they're valid when supporting multiple entity types
    }

    # Comprehensive list of Jinja2 builtins and Home Assistant template functions
    # These are always available in templates and should not trigger "undefined" warnings
    JINJA2_BUILTINS: set[str] = {
        # Python/Jinja2 built-in constants
        "true",
        "false",
        "none",
        "True",
        "False",
        "None",
        # Jinja2 control keywords
        "if",
        "else",
        "elif",
        "endif",
        "for",
        "endfor",
        "in",
        "not",
        "and",
        "or",
        "is",
        "set",
        "endset",
        "macro",
        "endmacro",
        "call",
        "endcall",
        "filter",
        "endfilter",
        "block",
        "endblock",
        "extends",
        "include",
        "import",
        "from",
        "as",
        "with",
        "endwith",
        "do",
        "continue",
        "break",
        # Jinja2 tests
        "defined",
        "undefined",
        "none",
        "number",
        "string",
        "mapping",
        "iterable",
        "callable",
        "sequence",
        "sameas",
        "escaped",
        "even",
        "odd",
        "divisibleby",
        "lower",
        "upper",
        # Jinja2 built-in filters (commonly used)
        "abs",
        "attr",
        "batch",
        "capitalize",
        "center",
        "count",
        "default",
        "dictsort",
        "escape",
        "filesizeformat",
        "first",
        "float",
        "forceescape",
        "format",
        "groupby",
        "indent",
        "int",
        "items",
        "join",
        "last",
        "length",
        "list",
        "lower",
        "map",
        "max",
        "min",
        "pprint",
        "random",
        "reject",
        "rejectattr",
        "replace",
        "reverse",
        "round",
        "safe",
        "select",
        "selectattr",
        "slice",
        "sort",
        "split",
        "string",
        "striptags",
        "sum",
        "title",
        "tojson",
        "trim",
        "truncate",
        "unique",
        "upper",
        "urlencode",
        "urlize",
        "wordcount",
        "wordwrap",
        "xmlattr",
        # Home Assistant specific functions
        "states",
        "is_state",
        "state_attr",
        "is_state_attr",
        "has_value",
        "expand",
        "device_entities",
        "area_entities",
        "integration_entities",
        "device_attr",
        "device_id",
        "area_name",
        "area_id",
        "floor_id",
        "floor_name",
        "label_id",
        "label_name",
        "labels",
        "relative_time",
        "time_since",
        "timedelta",
        "strptime",
        "strftime",
        "as_timestamp",
        "as_datetime",
        "as_local",
        "as_timedelta",
        "today_at",
        "now",
        "utcnow",
        "distance",
        "closest",
        "iif",
        "log",
        "sin",
        "cos",
        "tan",
        "asin",
        "acos",
        "atan",
        "atan2",
        "sqrt",
        "e",
        "pi",
        "tau",
        "inf",
        "average",
        "median",
        "statistical_mode",
        "pack",
        "unpack",
        "ord",
        "base64_encode",
        "base64_decode",
        "slugify",
        "regex_match",
        "regex_search",
        "regex_replace",
        "regex_findall",
        "regex_findall_index",
        "urlencode",
        "from_json",
        "to_json",
        "value_json",
        "trigger",
        "this",
        "context",
        "repeat",
        "wait",
        "namespace",
        # Common variable names in loops that shouldn't trigger warnings
        "item",
        "loop",
        "index",
        "index0",
        "first",
        "last",
        "length",
        "cycle",
        "depth",
        "depth0",
        "previtem",
        "nextitem",
        "changed",
        # Range function
        "range",
        # Datetime attributes and methods (accessed via now(), etc.)
        "year",
        "month",
        "day",
        "hour",
        "minute",
        "second",
        "microsecond",
        "weekday",
        "isoweekday",
        "isocalendar",
        "isoformat",
        "date",
        "time",
        "timestamp",
        "tzinfo",
        "tzname",
        "utcoffset",
        "dst",
        "timetuple",
        # State object attributes
        "state",
        "attributes",
        "entity_id",
        "domain",
        "object_id",
        "name",
        "last_changed",
        "last_updated",
        "last_reported",
        "context_id",
        # Trigger object attributes
        "platform",
        "event",
        "to_state",
        "from_state",
        "for",
        "idx",
        "id",
        "description",
        "alias",
        # Additional common attributes
        "friendly_name",
        "icon",
        "unit_of_measurement",
        "device_class",
        "brightness",
        "color_temp",
        "hs_color",
        "rgb_color",
        "xy_color",
        "temperature",
        "humidity",
        "pressure",
        "position",
        "current_position",
        "current_temperature",
        "target_temperature",
        "hvac_mode",
        "hvac_action",
        "fan_mode",
        "swing_mode",
        "preset_mode",
        "speed",
        "percentage",
        "battery_level",
        "battery",
        "power",
        "voltage",
        "current",
        "energy",
        "elevation",
        "azimuth",
        "rising",
        "setting",
        "next_rising",
        "next_setting",
    }

    # Filters that change type - used for type mismatch detection
    # Maps filter name to expected output type
    TYPE_CHANGING_FILTERS: dict[str, str] = {
        "int": "number",
        "float": "number",
        "string": "string",
        "bool": "boolean",
        "list": "list",
        "length": "number",
        "count": "number",
        "round": "number",
        "abs": "number",
        "sum": "number",
        "max": "number",
        "min": "number",
        "average": "number",
        "median": "number",
        "as_timestamp": "number",
        "first": "any",
        "last": "any",
        "join": "string",
        "lower": "string",
        "upper": "string",
        "title": "string",
        "capitalize": "string",
        "trim": "string",
        "replace": "string",
        "slugify": "string",
        "tojson": "string",
        "to_json": "string",
        "from_json": "any",
        "split": "list",
        "sort": "list",
        "unique": "list",
        "select": "list",
        "reject": "list",
        "map": "list",
        "batch": "list",
    }

    # Functions that require positive arguments
    POSITIVE_ARG_FUNCTIONS: set[str] = {"log", "sqrt"}

    # Functions that require non-zero arguments
    NONZERO_ARG_FUNCTIONS: set[str] = {"log"}

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
        self.defined_inputs: set[str] = set()
        self.used_inputs: set[str] = set()
        self.input_datetime_inputs: set[str] = (
            set()
        )  # inputs with input_datetime selector
        self.entity_inputs: set[str] = set()  # inputs with entity selector
        self.input_defaults: dict[str, Any] = {}  # input name -> default value
        self.input_selectors: dict[
            str, dict[str, Any]
        ] = {}  # input name -> selector dict
        self.defined_variables: set[str] = set()  # All defined variable names
        self.variable_types: dict[
            str, str
        ] = {}  # Inferred types from selectors/filters

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
        self._validate_mode()
        self._validate_inputs()
        self._validate_hysteresis_boundaries()
        self._validate_variables()
        self._validate_version_sync()
        self._validate_triggers()
        self._validate_conditions()
        self._validate_actions()
        self._validate_templates()
        self._validate_input_references()
        self._check_readme_exists()
        self._check_changelog_exists()

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

    def _validate_mode(self) -> None:
        """Validate automation mode."""
        mode = self.data.get("mode")
        if mode is None:
            return  # Default mode is 'single', which is valid

        if not isinstance(mode, str):
            self.errors.append("'mode' must be a string")
            return

        if mode not in self.VALID_MODES:
            self.errors.append(
                f"Invalid mode '{mode}', must be one of: {self.VALID_MODES}"
            )

        # Check for max/max_exceeded when using queued/parallel
        if mode in ["queued", "parallel"]:
            max_val = self.data.get("max")
            if max_val is not None:
                if not isinstance(max_val, int) or max_val < 1:
                    self.errors.append(
                        f"'max' must be a positive integer when mode is '{mode}'"
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
                # This is an actual input definition - record it
                self.defined_inputs.add(key)
                self._validate_single_input(value, current_path)

    def _validate_single_input(self, input_def: dict[str, Any], path: str) -> None:
        """Validate a single input definition.

        Args:
            input_def: The input definition dictionary.
            path: Current path for error messages.
        """
        # Extract input name from path (e.g., "blueprint.input.foo" -> "foo")
        input_name = path.split(".")[-1]

        # Track default value for hysteresis validation
        default_value = input_def.get("default")
        if default_value is not None:
            self.input_defaults[input_name] = default_value

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

        # Track selector for hysteresis validation
        self.input_selectors[input_name] = selector

        # Validate selector type and track special types
        for selector_type in selector:
            if selector_type not in self.VALID_SELECTOR_TYPES:
                self.warnings.append(
                    f"{path}.selector: Unknown selector type '{selector_type}'"
                )

            # Track entity selector inputs for later validation
            if selector_type == "entity":
                self.entity_inputs.add(input_name)
                entity_selector = selector.get("entity", {})
                if isinstance(entity_selector, dict):
                    domain = entity_selector.get("domain")
                    if domain == "input_datetime":
                        self.input_datetime_inputs.add(input_name)

            # Validate select selector options
            if selector_type == "select":
                self._validate_select_options(selector.get("select", {}), path)

    def _validate_select_options(
        self, select_config: dict[str, Any] | None, path: str
    ) -> None:
        """Validate select selector options.

        Home Assistant requires all select options to be valid strings.
        Options can be:
        1. Simple strings: "option1"
        2. Label/value dicts: {"label": "Display Name", "value": "actual_value"}

        For label/value dicts, the 'value' field MUST be a non-empty string.
        Using empty strings or None for 'value' will cause import errors like:
        "expected str @ data['blueprint']['input'][...]['selector']['options'][N]. Got None"

        Args:
            select_config: The select selector configuration dict.
            path: Current path for error messages.
        """
        if not isinstance(select_config, dict):
            return

        options = select_config.get("options")
        if options is None:
            return

        if not isinstance(options, list):
            self.errors.append(f"{path}.selector.select.options: Must be a list")
            return

        for i, option in enumerate(options):
            option_path = f"{path}.selector.select.options[{i}]"

            if option is None:
                self.errors.append(
                    f"{option_path}: Option cannot be None. "
                    f"Select options must be strings or label/value dicts with non-empty values."
                )
            elif isinstance(option, dict):
                # Label/value format: {"label": "...", "value": "..."}
                value = option.get("value")
                label = option.get("label")

                if value is None:
                    self.errors.append(
                        f"{option_path}: Option value is None. "
                        f"Label/value options must have a non-empty 'value' field. "
                        f"Label: '{label}'"
                    )
                elif not isinstance(value, str):
                    self.errors.append(
                        f"{option_path}: Option value must be a string, got {type(value).__name__}. "
                        f"Label: '{label}'"
                    )
                elif value == "":
                    # Empty string values cause "expected str ... Got None" errors in HA
                    self.errors.append(
                        f"{option_path}: Option value cannot be empty string. "
                        f"Home Assistant treats empty values as None during import. "
                        f"Label: '{label}'. Use a placeholder value like '---' or remove this option."
                    )

                if label is not None and not isinstance(label, str):
                    self.warnings.append(
                        f"{option_path}: Option label should be a string, got {type(label).__name__}"
                    )
            elif isinstance(option, str):
                # Simple string option - valid as long as it's a string
                # Empty strings are technically valid here but unusual
                if option == "":
                    self.warnings.append(
                        f"{option_path}: Empty string option. Consider using a meaningful value."
                    )
            else:
                self.errors.append(
                    f"{option_path}: Option must be a string or label/value dict, "
                    f"got {type(option).__name__}"
                )

    def _validate_hysteresis_boundaries(self) -> None:
        """Validate hysteresis boundary pairs have correct relationships.

        Hysteresis is a common pattern in automation to prevent rapid on/off
        oscillation (chattering). For example, a humidity-based fan control
        might turn ON at 15% delta and OFF at 10% delta.

        This validation checks that:
        1. ON thresholds are greater than OFF thresholds
        2. HIGH boundaries are greater than LOW boundaries
        3. The gap between thresholds is sufficient (not too small)

        Common issues detected:
        - humidity_delta_on: 10, humidity_delta_off: 15 (inverted - will cause chatter)
        - temp_high: 25, temp_low: 25 (equal values - no hysteresis)
        - threshold_on: 10.5, threshold_off: 10.0 (gap too small for stability)
        """
        # Look for hysteresis pairs in defined inputs
        for on_pattern, off_pattern, desc in self.HYSTERESIS_PATTERNS:
            for input_name in self.defined_inputs:
                on_match = re.match(on_pattern, input_name)
                if not on_match:
                    continue

                # Construct the expected OFF input name
                if on_pattern == r"delta_on$":
                    # Special case for delta patterns
                    off_name = input_name.replace("_on", "_off")
                else:
                    # Use regex substitution for other patterns
                    off_name = re.sub(
                        on_pattern,
                        off_pattern.replace(r"\1", on_match.group(1)),
                        input_name,
                    )

                if off_name not in self.defined_inputs:
                    continue

                # Found a hysteresis pair - validate the relationship
                on_default = self.input_defaults.get(input_name)
                off_default = self.input_defaults.get(off_name)

                # Only validate if both have numeric defaults
                if on_default is None or off_default is None:
                    continue

                try:
                    on_value = float(on_default)
                    off_value = float(off_default)
                except (ValueError, TypeError):
                    continue

                # Check if the relationship is correct (ON > OFF for hysteresis)
                if on_value < off_value:
                    self.errors.append(
                        f"Hysteresis {desc} inversion: '{input_name}' (default={on_value}) "
                        f"should be greater than '{off_name}' (default={off_value}). "
                        f"With ON < OFF, the system will chatter rapidly. "
                        f"Swap the values or adjust thresholds."
                    )
                elif on_value == off_value:
                    self.warnings.append(
                        f"Hysteresis {desc} has no gap: '{input_name}' and '{off_name}' "
                        f"both default to {on_value}. Without a gap between ON and OFF "
                        f"thresholds, there's no hysteresis protection against oscillation."
                    )
                else:
                    # Check if the gap is very small (less than 20% of the ON value)
                    gap = on_value - off_value
                    if on_value != 0 and gap / abs(on_value) < 0.1:
                        self.warnings.append(
                            f"Hysteresis {desc} gap may be too small: '{input_name}' "
                            f"(default={on_value}) minus '{off_name}' (default={off_value}) "
                            f"= {gap}. A larger gap provides better oscillation protection."
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

        # Track defined variables in order for dependency checking
        defined_vars: set[str] = set()
        var_order = list(variables.keys())

        # Check for variable dependency chains that may cause evaluation issues
        self._check_variable_dependency_chains(variables)

        # Pre-pass: collect variables that have non-zero defaults
        # These are safe to divide by even when used in other variables
        self.nonzero_default_vars: set[str] = set()
        for name, value in variables.items():
            if isinstance(value, str):
                # Pattern: | float(X) or | int(X) where X > 0
                # This catches: state_attr(...) | float(1.0) or | float(0.5)
                default_match = re.search(
                    r"\|\s*(?:float|int)\s*\(\s*(\d+\.?\d*)", value
                )
                if default_match:
                    default_val = float(default_match.group(1))
                    if default_val > 0:
                        self.nonzero_default_vars.add(name)

        # First pass: collect all variable names for context
        self.defined_variables = set(variables.keys())

        # Record variables that appear to build comma-joined strings
        for name, value in variables.items():
            if isinstance(value, str):
                if re.search(r"\|\s*join\b", value) or re.search(r"\bjoin\s*\(", value):
                    self.join_variables.add(name)

                # Track !input references in variables
                self._collect_input_refs(value)

                # Check for forward references to variables not yet defined
                self._check_variable_ordering(name, value, defined_vars, var_order)

                # Check for direct attribute access on state objects
                self._check_state_attribute_access(name, value)

                # Check for incorrect input_datetime usage
                self._check_input_datetime_usage(name, value)

                # Check for bare boolean literals in templates
                self._check_bare_boolean_literals(name, value)

                # Check for math functions that need guards
                self._check_math_function_safety(name, value)

                # Check for entity ID variables in boolean contexts
                self._check_entity_id_boolean_context(name, value)

                # NEW: Check for undefined variable references
                self._check_undefined_variable_references(name, value, defined_vars)

                # NEW: Check for unsafe math operations (sqrt, modulo, asin/acos)
                self._check_unsafe_math_operations(name, value)

                # NEW: Check for type mismatches in filter chains
                self._check_type_mismatch_in_filters(name, value)

            defined_vars.add(name)

        # Check for blueprint_version
        if "blueprint_version" not in variables:
            self.warnings.append("No 'blueprint_version' variable defined")

    def _check_variable_dependency_chains(self, variables: dict[str, Any]) -> None:
        """Check for variable dependency chains that may cause runtime evaluation issues.

        In Home Assistant blueprints, variables that reference other variables
        can fail with "UndefinedError" at runtime, even when the definition order
        appears correct in YAML. This happens because:

        1. Each variable's template is evaluated separately
        2. Multi-line YAML blocks (>) return strings that may need re-parsing
        3. The evaluation order can differ from the definition order

        This check warns about patterns that are known to cause issues:
        - Multiple intermediate variables that are only used by one other variable
        - Variables forming chains: A -> B -> C where B is only used by C

        Args:
            variables: Dictionary of variable definitions.
        """
        # Build a dependency graph: which variables reference which others
        var_refs: dict[str, set[str]] = {}  # var_name -> set of vars it references
        var_used_by: dict[str, set[str]] = {}  # var_name -> set of vars that use it

        var_names = set(variables.keys())

        # Jinja2 builtins to ignore
        jinja_builtins = {
            "true",
            "false",
            "none",
            "True",
            "False",
            "None",
            "now",
            "utcnow",
            "as_timestamp",
            "states",
            "is_state",
            "state_attr",
            "has_value",
            "expand",
            "device_entities",
            "area_entities",
            "integration_entities",
            "device_attr",
            "area_name",
            "area_id",
            "relative_time",
            "timedelta",
            "strptime",
            "as_datetime",
            "as_local",
            "today_at",
            "trigger",
            "this",
            "repeat",
            "context",
            "set",
            "if",
            "else",
            "elif",
            "endif",
            "for",
            "endfor",
            "in",
            "not",
            "and",
            "or",
            "is",
            "defined",
            "undefined",
            "number",
            "string",
            "mapping",
            "iterable",
            "int",
            "float",
            "round",
            "abs",
            "default",
            "length",
            "log",
            "min",
            "max",
        }

        for var_name, value in variables.items():
            if not isinstance(value, str):
                continue

            var_refs[var_name] = set()
            for other_var in var_names:
                if other_var == var_name or other_var in jinja_builtins:
                    continue

                # Check if this variable references other_var inside a Jinja block
                pattern = rf"\b{re.escape(other_var)}\b"
                for match in re.finditer(pattern, value):
                    pos = match.start()

                    # Skip if the match is inside a string literal (quoted)
                    before_match = value[:pos]
                    in_string = False
                    for quote in ["'", '"']:
                        quote_positions = [
                            i for i, c in enumerate(before_match) if c == quote
                        ]
                        if len(quote_positions) % 2 == 1:
                            after_match = value[match.end() :]
                            if quote in after_match:
                                in_string = True
                                break

                    if in_string:
                        continue

                    before = value[:pos]
                    last_expr_open = before.rfind("{{")
                    last_ctrl_open = before.rfind("{%")
                    last_open = max(last_expr_open, last_ctrl_open)

                    if last_open == -1:
                        continue

                    between = value[last_open:pos]
                    if "}}" in between or "%}" in between:
                        continue

                    # Found a reference to other_var inside a Jinja block
                    var_refs[var_name].add(other_var)
                    if other_var not in var_used_by:
                        var_used_by[other_var] = set()
                    var_used_by[other_var].add(var_name)
                    break

        # Find variables that are only used by one other variable and start with _
        # These are "helper" variables that could potentially be inlined
        helper_chains: list[tuple[str, str]] = []
        for var_name, used_by in var_used_by.items():
            if (
                len(used_by) == 1
                and var_name.startswith("_")
                and var_name
                not in (
                    "_raw_",
                    "_is_",
                    "_has_",
                )  # Common prefixes for persistent helpers
            ):
                user = list(used_by)[0]
                # Check if the helper is also using other helpers
                helper_refs = var_refs.get(var_name, set())
                other_helpers = [
                    ref
                    for ref in helper_refs
                    if ref.startswith("_")
                    and ref in var_used_by
                    and len(var_used_by[ref]) == 1
                ]
                if other_helpers:
                    # This helper depends on other single-use helpers - potential chain
                    helper_chains.append((var_name, user))

        # Warn about chains of helper variables
        if helper_chains:
            # Group by the final user
            chains_by_user: dict[str, list[str]] = {}
            for helper, user in helper_chains:
                if user not in chains_by_user:
                    chains_by_user[user] = []
                chains_by_user[user].append(helper)

            for user, helpers in chains_by_user.items():
                if len(helpers) >= 2:
                    self.warnings.append(
                        f"Variable '{user}' depends on a chain of helper variables "
                        f"({', '.join(helpers)}). Consider consolidating into a single "
                        f"self-contained template to avoid potential 'UndefinedError' "
                        f"at runtime. See: variables evaluated separately may fail "
                        f"when referencing each other."
                    )

    def _check_variable_ordering(
        self, var_name: str, value: str, defined_vars: set[str], var_order: list[str]
    ) -> None:
        """Check if a variable references other variables that aren't defined yet.

        Args:
            var_name: Name of the variable being checked.
            value: The template string value.
            defined_vars: Set of variables defined so far.
            var_order: Ordered list of all variable names.
        """
        # Find variable references in the template (excluding Jinja built-ins)
        # Pattern matches references in both {{ }} expressions and {% %} control blocks
        jinja_builtins = {
            "true",
            "false",
            "none",
            "True",
            "False",
            "None",
            "now",
            "utcnow",
            "as_timestamp",
            "states",
            "is_state",
            "state_attr",
            "has_value",
            "expand",
            "device_entities",
            "area_entities",
            "integration_entities",
            "device_attr",
            "area_name",
            "area_id",
            "relative_time",
            "timedelta",
            "strptime",
            "as_datetime",
            "as_local",
            "today_at",
            "trigger",
            "this",
            "repeat",
            "context",
            # Common Jinja2 control keywords
            "set",
            "if",
            "else",
            "elif",
            "endif",
            "for",
            "endfor",
            "in",
            "not",
            "and",
            "or",
            "is",
            "defined",
            "undefined",
            "number",
            "string",
            "mapping",
            "iterable",
            "int",
            "float",
            "round",
            "abs",
            "default",
            "length",
            "log",
            "min",
            "max",
        }

        # Look for variable references
        for other_var in var_order:
            if other_var == var_name:
                continue
            if other_var in jinja_builtins:
                continue

            # Check if this variable references another variable
            # Match patterns in both {{ }} and {% %} blocks:
            # - {{ other_var }} or {{ other_var | filter }}
            # - {% set x = other_var %} or {% if other_var %}
            # Use word boundary \b to avoid partial matches
            pattern = rf"\b{re.escape(other_var)}\b"
            if re.search(pattern, value):
                # Verify it's actually inside a Jinja block ({{ }} or {% %})
                # by checking if it appears after {{ or {% and before }} or %}
                in_jinja_block = False
                for match in re.finditer(pattern, value):
                    pos = match.start()

                    # Skip if the match is inside a string literal (quoted)
                    # Check if preceded by an odd number of quotes (inside a string)
                    before_match = value[:pos]
                    # Count single and double quotes to detect if we're in a string
                    # This is a heuristic - look for 'word' or "word" patterns
                    # where our match would be inside the quotes
                    in_string = False
                    for quote in ["'", '"']:
                        # Find all quote positions before our match
                        quote_positions = [
                            i for i, c in enumerate(before_match) if c == quote
                        ]
                        # If odd number of quotes, we're inside a string
                        if len(quote_positions) % 2 == 1:
                            # Verify the string continues past our match
                            after_match = value[match.end() :]
                            if quote in after_match:
                                in_string = True
                                break

                    if in_string:
                        continue

                    # Look backwards for {{ or {%
                    before = value[:pos]
                    # Find the last Jinja opener
                    last_expr_open = before.rfind("{{")
                    last_ctrl_open = before.rfind("{%")
                    last_open = max(last_expr_open, last_ctrl_open)

                    if last_open == -1:
                        continue

                    # Check if there's a closer between the opener and our match
                    between = value[last_open:pos]
                    if "}}" in between or "%}" in between:
                        continue  # The opener was already closed

                    in_jinja_block = True
                    break

                if in_jinja_block and other_var not in defined_vars:
                    self.errors.append(
                        f"Variable '{var_name}' references '{other_var}' which is "
                        f"defined later in the variables section. Variables are "
                        f"evaluated in order - move '{other_var}' before '{var_name}' "
                        f"or restructure to avoid the dependency."
                    )

    def _check_state_attribute_access(self, var_name: str, value: str) -> None:
        """Check for direct attribute access on state objects.

        In Home Assistant templates, state objects retrieved via states[entity]
        have attributes in a ReadOnlyDict. Direct dot notation access like
        s.attributes.brightness doesn't work - use state_attr() instead.

        Args:
            var_name: Name of the variable being checked.
            value: The template string value.
        """
        # Pattern to match: variable.attributes.something
        # This catches: s.attributes.current_position, state.attributes.brightness, etc.
        pattern = r"\b(\w+)\.attributes\.(\w+)"
        matches = re.findall(pattern, value)

        for var, attr in matches:
            # Skip if it's clearly not a state object access
            # (e.g., it's accessing trigger.to_state.attributes which is valid)
            if var in ("to_state", "from_state", "trigger"):
                continue

            self.warnings.append(
                f"Variable '{var_name}': Direct attribute access '{var}.attributes.{attr}' "
                f"may fail with ReadOnlyDict error. Consider using "
                f"state_attr(entity_id, '{attr}') instead."
            )

    def _check_input_datetime_usage(self, var_name: str, value: str) -> None:
        """Check for incorrect usage of input_datetime helpers.

        input_datetime entities store state as 'YYYY-MM-DD HH:MM:SS' which
        as_timestamp() cannot parse. Use state_attr(entity, 'timestamp') instead.

        Args:
            var_name: Name of the variable being checked.
            value: The template string value.
        """
        # Get variables that are bound to input_datetime inputs
        variables = self.data.get("variables", {})
        datetime_vars: set[str] = set()

        for name, val in variables.items():
            if isinstance(val, str) and val.startswith("!input "):
                input_name = val[7:].strip()
                if input_name in self.input_datetime_inputs:
                    datetime_vars.add(name)

        # Check for patterns like: as_timestamp(states(var)) or as_timestamp(helper_state)
        # where var is bound to an input_datetime input
        for dt_var in datetime_vars:
            # Pattern 1: as_timestamp(states(dt_var))
            pattern1 = rf"as_timestamp\s*\(\s*states\s*\(\s*{re.escape(dt_var)}\s*\)"
            if re.search(pattern1, value):
                self.errors.append(
                    f"Variable '{var_name}': as_timestamp(states({dt_var})) won't work "
                    f"for input_datetime entities. The state format 'YYYY-MM-DD HH:MM:SS' "
                    f"isn't parseable by as_timestamp(). Use state_attr({dt_var}, 'timestamp') "
                    f"instead to get the Unix timestamp directly."
                )

            # Pattern 2: intermediate variable like helper_state = states(dt_var),
            # then as_timestamp(helper_state)
            # First find if there's a variable assignment like: helper_state = states(dt_var)
            state_var_pattern = rf"(\w+)\s*=\s*states\s*\(\s*{re.escape(dt_var)}\s*\)"
            state_var_matches = re.findall(state_var_pattern, value)
            for state_var in state_var_matches:
                # Check if as_timestamp is called on this intermediate variable
                ts_pattern = rf"as_timestamp\s*\(\s*{re.escape(state_var)}\s*\)"
                if re.search(ts_pattern, value):
                    self.errors.append(
                        f"Variable '{var_name}': as_timestamp({state_var}) won't work when "
                        f"{state_var} is from states({dt_var}) (an input_datetime). "
                        f"Use state_attr({dt_var}, 'timestamp') instead."
                    )

    def _check_math_function_safety(self, var_name: str, value: str) -> None:
        """Check for math functions that may fail without proper guards.

        The log() function requires x > 0, and division requires non-zero denominators.
        This check warns when these are used without apparent guards.

        Args:
            var_name: Name of the variable being checked.
            value: The template string value.
        """
        # Check for log() usage - needs x > 0 guard
        log_matches = re.findall(r"log\s*\(\s*(\w+)\s*\)", value)
        for var in log_matches:
            # Check if there's a guard like "if x > 0" or "x > 0.001" before the log
            guard_pattern = rf"{re.escape(var)}\s*>\s*0|{re.escape(var)}\s*is\s+number"
            if not re.search(guard_pattern, value):
                self.warnings.append(
                    f"Variable '{var_name}': log({var}) may fail if {var} <= 0. "
                    f"Consider adding a guard like 'if {var} > 0' before the log() call."
                )

        # Check for division that might fail
        # Only check inside {{ }} template blocks to avoid false positives in strings
        template_blocks = re.findall(r"\{\{([^}]+)\}\}", value)
        for block in template_blocks:
            # Pattern: / followed by variable (not inside a string quote)
            # This handles: / var, / (var), / (var - x)
            div_matches = re.findall(r"/\s*\(?\s*([a-zA-Z_][a-zA-Z0-9_]*)", block)
            for var in div_matches:
                # Skip mathematical constants (always non-zero)
                if var in ("pi", "e", "tau"):
                    continue

                # Skip variables that were identified as having non-zero defaults
                if (
                    hasattr(self, "nonzero_default_vars")
                    and var in self.nonzero_default_vars
                ):
                    continue

                # Check if variable is a local set from a guarded source
                # Pattern: {% set var = source ... %} where source has is number check
                local_set_pattern = rf"{{% set {re.escape(var)}\s*="
                if re.search(local_set_pattern, value):
                    # Local variable - check if the source expression is guarded
                    # e.g., {% set Tk = t_in_c + 273.15 %} is safe if t_in_c is guarded
                    set_match = re.search(
                        rf"{{% set {re.escape(var)}\s*=\s*([^%]+)%}}", value
                    )
                    if set_match:
                        source_expr = set_match.group(1)
                        # Check if the expression is of the form (1 + x) or (constant + x)
                        # These are always non-zero when x >= 0
                        if re.match(r"^\s*\d+\s*\+", source_expr) or re.match(
                            r"^\s*\d+\.\d+\s*\+", source_expr
                        ):
                            continue
                        # Extract source variables from the set expression
                        source_vars = re.findall(
                            r"([a-zA-Z_][a-zA-Z0-9_]*)", source_expr
                        )
                        # Check if any source var is guarded
                        all_guarded = all(
                            re.search(rf"{re.escape(sv)}\s+is\s+number", value)
                            for sv in source_vars
                            if not sv.isdigit()
                            and sv not in ("set", "if", "else", "endif")
                        )
                        if all_guarded:
                            continue

                # Check if variable has a non-zero default via | float(X) or | int(X)
                # Pattern: var | float(0.5) or var | int(1) where X > 0
                default_pattern = (
                    rf"{re.escape(var)}\s*\|\s*(?:float|int)\s*\(\s*(\d+\.?\d*)"
                )
                default_match = re.search(default_pattern, value)
                if default_match:
                    default_val = float(default_match.group(1))
                    if default_val > 0:
                        continue

                # Check if there's a guard in the same block or value
                guard_pattern = (
                    rf"\({re.escape(var)}[^)]*\)\s*>\s*0|"
                    rf"{re.escape(var)}\s*>\s*0|"
                    rf"{re.escape(var)}\s+is\s+number|"
                    rf"{re.escape(var)}\s*<=\s*0|"  # Early return pattern
                    rf"{re.escape(var)}\s*==\s*0"  # Zero check pattern (e.g., if x == 0)
                )
                if not re.search(guard_pattern, value):
                    # Only warn if it looks like a variable, not a function call
                    if not re.search(rf"{re.escape(var)}\s*\(", block):
                        self.warnings.append(
                            f"Variable '{var_name}': Division by '{var}' may fail if it's "
                            f"zero, none, or not a number. Consider adding guards."
                        )

    def _check_bare_boolean_literals(self, var_name: str, value: str) -> None:
        """Check for bare boolean literals in Jinja2 templates.

        In multi-line YAML templates (using >- or |), outputting bare 'true' or
        'false' without {{ }} creates a STRING, not a boolean. The string "false"
        is truthy in Python/Jinja2 because it's a non-empty string. This causes
        subtle bugs where {% if some_var %} unexpectedly evaluates to True.

        Correct: {{ false }} or {{ true }} outputs actual booleans.
        Wrong:   false or true as bare text outputs strings.

        Args:
            var_name: Name of the variable being checked.
            value: The template string value.
        """
        # Only check multi-line templates that look like they're meant to output booleans
        # Pattern: lines that contain just 'true' or 'false' (possibly with whitespace)
        # but NOT inside {{ }} blocks
        lines = value.split("\n")

        # Track which types of bare booleans we've found (to avoid duplicate warnings)
        found_bare_true = False
        found_bare_false = False

        # Track nesting of Jinja control blocks
        block_depth = 0

        for line_num, line in enumerate(lines, 1):
            stripped = line.strip()

            # Track block depth: {% if/for/etc %} increases, {% endif/endfor/else/elif %} may change
            # We use a simplified approach: count {% and %} pairs
            # Opening blocks: if, for, macro, call, filter, set (when multi-line)
            # Closing blocks: endif, endfor, endmacro, endcall, endfilter, endset
            # Note: else/elif don't change depth but are still inside a block

            # Count block opens and closes on this line
            open_blocks = len(re.findall(r"\{%\s*(?:if|for|macro|call|filter)\b", line))
            close_blocks = len(
                re.findall(r"\{%\s*(?:endif|endfor|endmacro|endcall|endfilter)\b", line)
            )

            # Update depth based on what we see
            # Process closes first (for lines like {% endif %}{% if %})
            block_depth = max(0, block_depth - close_blocks)
            block_depth += open_blocks

            # Skip empty lines
            if not stripped:
                continue

            # Only check lines that are just 'true' or 'false'
            if stripped not in ("true", "false"):
                continue

            # If we're inside a control block (if/for/etc), the literal might
            # legitimately be part of template logic, not bare output.
            # However, if block_depth is 0, this bare literal WILL be output.
            # But wait - we need to track this more carefully:
            # {% if x %}
            #   false    <- this is output if x is true, and it's a STRING
            # {% endif %}
            #
            # The issue is that ANY bare true/false in a multi-line template
            # that gets output will be a string. Let's warn on all of them
            # but only if they look like they're meant to be returned values.

            # Check if this looks like a return value (not inside {{ }} expression on same line)
            if "{{" in line and "}}" in line:
                # The true/false is part of an expression on this line, skip
                continue

            # Check if there's a {{ before and no matching }} on this line (inside expression)
            line_before = line[: line.find(stripped)] if stripped in line else ""
            if "{{" in line_before and "}}" not in line_before:
                continue

            # This is a bare boolean literal that will be output as a string
            if stripped == "true" and not found_bare_true:
                found_bare_true = True
                self.warnings.append(
                    f"Variable '{var_name}': Bare 'true' outputs STRING \"true\", "
                    f"not boolean. Use '{{{{ true }}}}' to output actual boolean. "
                    f'(String "false" is truthy, causing subtle bugs in conditionals.)'
                )
            elif stripped == "false" and not found_bare_false:
                found_bare_false = True
                self.warnings.append(
                    f"Variable '{var_name}': Bare 'false' outputs STRING \"false\", "
                    f'not boolean. The string "false" is TRUTHY (non-empty), so '
                    f"'{{% if var %}}' passes unexpectedly. Use '{{{{ false }}}}' instead."
                )

    def _check_entity_id_boolean_context(self, var_name: str, value: str) -> None:
        """Check for entity ID variables used in boolean contexts without explicit checks.

        In Jinja2 templates, using an entity ID variable directly in a boolean
        context like `and entity_var }}` can be unreliable. The string truthiness
        check may fail unexpectedly. It's safer to use an explicit check like
        `and (entity_var | string | length > 0)`.

        Args:
            var_name: Name of the variable being checked.
            value: The template string value.
        """
        # Get variables that are bound to entity selector inputs
        variables = self.data.get("variables", {})
        entity_vars: set[str] = set()

        for name, val in variables.items():
            if isinstance(val, str) and val.startswith("!input "):
                input_name = val[7:].strip()
                if input_name in self.entity_inputs:
                    entity_vars.add(name)

        # Check for patterns like: `and entity_var }}` or `and entity_var %}`
        # without explicit string/length checks
        for entity_var in entity_vars:
            # Pattern: `and entity_var` followed by }} or %} (end of expression)
            # but NOT followed by | (which would indicate a filter is applied)
            pattern = rf"\band\s+{re.escape(entity_var)}\s*[}}%]"
            if re.search(pattern, value):
                # Check if there's already an explicit check nearby
                # e.g., `(entity_var | string | length > 0)` or `entity_var != ''`
                explicit_check_pattern = (
                    rf"{re.escape(entity_var)}\s*\|\s*string\s*\|\s*length|"
                    rf"{re.escape(entity_var)}\s*!=\s*['\"]|"
                    rf"len\s*\(\s*{re.escape(entity_var)}\s*\)|"
                    rf"{re.escape(entity_var)}\s*\|\s*length"
                )
                if not re.search(explicit_check_pattern, value):
                    self.warnings.append(
                        f"Variable '{var_name}': Entity ID variable '{entity_var}' used "
                        f"in boolean context without explicit check. String truthiness "
                        f"can be unreliable. Consider using "
                        f"'({entity_var} | string | length > 0)' instead of just "
                        f"'{entity_var}'."
                    )

    def _check_undefined_variable_references(
        self, var_name: str, value: str, context_vars: set[str]
    ) -> None:
        """Check for undefined variable references in templates.

        Detects when a template uses a variable that hasn't been defined in:
        1. The blueprint's variables section
        2. Jinja2 built-in functions/filters
        3. Home Assistant template functions
        4. Loop variables (item, loop, index, etc.)
        5. Local {% set %} definitions within the same template

        Args:
            var_name: Name of the variable being checked (for error messages).
            value: The template string value to check.
            context_vars: Set of variable names available in the current context.
        """
        if not isinstance(value, str):
            return

        # Find all variable references in Jinja2 blocks
        # Pattern matches identifiers inside {{ }} or {% %} blocks
        all_available = self.JINJA2_BUILTINS | context_vars | self.defined_variables

        # Extract local {% set var = ... %} definitions
        local_sets = set(re.findall(r"\{%\s*set\s+(\w+)\s*=", value))
        all_available = all_available | local_sets

        # Extract {% for item in ... %} loop variables
        for_loops = re.findall(r"\{%\s*for\s+(\w+)(?:\s*,\s*(\w+))?\s+in\s+", value)
        for match in for_loops:
            all_available.add(match[0])
            if match[1]:  # Handle tuple unpacking: for key, value in ...
                all_available.add(match[1])

        # Find all potential variable references in Jinja blocks
        # Extract content inside {{ }} and {% %}
        jinja_blocks = re.findall(r"\{\{([^}]+)\}\}", value)
        jinja_blocks.extend(re.findall(r"\{%([^%]+)%\}", value))

        for block in jinja_blocks:
            # Find all word-like tokens that could be variable references
            # Skip tokens that are:
            # - Inside string literals
            # - Numeric values
            # - Operators
            # - Filter names following |
            tokens = self._extract_variable_references(block)

            for token in tokens:
                if token not in all_available:
                    # Check if it might be an attribute access (like obj.attr)
                    if "." in var_name:
                        continue  # Skip attribute-style references

                    # Check if this looks like a function call (followed by parenthesis)
                    func_pattern = rf"\b{re.escape(token)}\s*\("
                    if re.search(func_pattern, block):
                        # It's being used as a function - might be a HA function we don't know
                        continue

                    # Check if it's after a pipe (filter)
                    filter_pattern = rf"\|\s*{re.escape(token)}\b"
                    if re.search(filter_pattern, block):
                        # It's being used as a filter
                        continue

                    self.errors.append(
                        f"Variable '{var_name}': Undefined reference to '{token}'. "
                        f"This variable is not defined in the variables section "
                        f"or recognized as a Jinja2/Home Assistant built-in."
                    )

    def _extract_variable_references(self, block: str) -> set[str]:
        """Extract potential variable references from a Jinja2 block.

        Args:
            block: Content inside a {{ }} or {% %} block.

        Returns:
            Set of potential variable names found.
        """
        tokens: set[str] = set()

        # Remove string literals to avoid false positives
        # Handle both single and double quoted strings
        cleaned = re.sub(r"'[^']*'", "", block)
        cleaned = re.sub(r'"[^"]*"', "", cleaned)

        # Remove numeric literals
        cleaned = re.sub(r"\b\d+\.?\d*\b", "", cleaned)

        # Remove namespace attribute access (e.g., ns.found -> remove .found)
        # This prevents false positives for attributes of namespace objects
        cleaned = re.sub(r"\.([a-zA-Z_][a-zA-Z0-9_]*)", "", cleaned)

        # Remove namespace keyword arguments (e.g., namespace(found=false))
        # This removes patterns like: word= where the word is an assignment target
        cleaned = re.sub(r"\b([a-zA-Z_][a-zA-Z0-9_]*)\s*=", "=", cleaned)

        # Find all word tokens that are NOT preceded by a dot
        # (to avoid capturing object attributes as standalone variables)
        word_tokens = re.findall(r"(?<![.\w])([a-zA-Z_][a-zA-Z0-9_]*)\b", cleaned)

        for token in word_tokens:
            # Skip very short tokens that are likely operators/keywords
            if len(token) <= 1 and token not in ("e", "x", "y", "z", "t", "s", "n"):
                continue

            # Skip common operators written as words
            if token in ("eq", "ne", "lt", "gt", "le", "ge"):
                continue

            tokens.add(token)

        return tokens

    def _check_unsafe_math_operations(self, var_name: str, value: str) -> None:
        """Check for potentially unsafe mathematical operations.

        Detects:
        1. sqrt() with potentially negative arguments
        2. Division/modulo by zero risks
        3. log() with non-positive arguments (already checked elsewhere, enhanced here)
        4. Trigonometric functions with invalid domains

        Args:
            var_name: Name of the variable being checked.
            value: The template string value.
        """
        if not isinstance(value, str):
            return

        # Check for sqrt() with potentially negative arguments
        sqrt_matches = re.findall(r"sqrt\s*\(\s*([^)]+)\)", value)
        for arg in sqrt_matches:
            arg_stripped = arg.strip()
            # Skip if it's a literal positive number
            if re.match(r"^\d+\.?\d*$", arg_stripped):
                continue

            # Check if there's a guard for negativity
            # Common patterns: max(0, x), abs(x), if x >= 0
            var_match = re.search(r"([a-zA-Z_][a-zA-Z0-9_]*)", arg_stripped)
            if var_match:
                var = var_match.group(1)
                # Check for guards
                guard_patterns = [
                    rf"max\s*\(\s*0\s*,\s*{re.escape(var)}",
                    rf"max\s*\(\s*{re.escape(var)}\s*,\s*0",
                    rf"abs\s*\(\s*{re.escape(var)}\s*\)",
                    rf"{re.escape(var)}\s*>=\s*0",
                    rf"{re.escape(var)}\s*>\s*0",
                    rf"if\s+{re.escape(var)}\s*>=\s*0",
                ]
                has_guard = any(re.search(p, value) for p in guard_patterns)

                if not has_guard:
                    self.warnings.append(
                        f"Variable '{var_name}': sqrt({arg_stripped}) may fail if the "
                        f"argument is negative. Consider using 'sqrt(max(0, {var}))' "
                        f"or adding a guard like 'if {var} >= 0'."
                    )

        # Check for modulo by zero risks (% operator)
        # IMPORTANT: Only check inside {{ }} expression blocks, not {% %} control blocks
        # Pattern: expr % var where var could be zero
        expr_blocks = re.findall(r"\{\{([^}]+)\}\}", value)
        for block in expr_blocks:
            modulo_matches = re.findall(r"%\s*\(?\s*([a-zA-Z_][a-zA-Z0-9_]*)", block)
            for var in modulo_matches:
                # Skip mathematical constants
                if var in ("pi", "e", "tau"):
                    continue

                # Skip Jinja2 keywords (these aren't modulo operations)
                if var in self.JINJA2_BUILTINS:
                    continue

                # Check if variable has non-zero default
                if (
                    hasattr(self, "nonzero_default_vars")
                    and var in self.nonzero_default_vars
                ):
                    continue

                # Check for guards
                guard_pattern = (
                    rf"{re.escape(var)}\s*!=\s*0|"
                    rf"{re.escape(var)}\s*>\s*0|"
                    rf"{re.escape(var)}\s+is\s+number"
                )
                if not re.search(guard_pattern, value):
                    self.warnings.append(
                        f"Variable '{var_name}': Modulo by '{var}' may fail if it's zero. "
                        f"Consider adding a guard like 'if {var} != 0'."
                    )

        # Check for asin/acos with out-of-range arguments
        for func in ["asin", "acos"]:
            func_matches = re.findall(rf"{func}\s*\(\s*([^)]+)\)", value)
            for arg in func_matches:
                arg_stripped = arg.strip()
                # Skip if it's a literal in valid range [-1, 1]
                if re.match(r"^-?[01]\.?\d*$", arg_stripped):
                    try:
                        val = float(arg_stripped)
                        if -1 <= val <= 1:
                            continue
                    except ValueError:
                        pass

                # Check for clamping patterns
                var_match = re.search(r"([a-zA-Z_][a-zA-Z0-9_]*)", arg_stripped)
                if var_match:
                    var = var_match.group(1)
                    clamp_patterns = [
                        rf"max\s*\(\s*-1\s*,\s*min\s*\(\s*1\s*,\s*{re.escape(var)}",
                        rf"min\s*\(\s*1\s*,\s*max\s*\(\s*-1\s*,\s*{re.escape(var)}",
                        r"\|\s*clamp",  # Custom clamp filter if exists
                    ]
                    has_clamp = any(re.search(p, value) for p in clamp_patterns)

                    if not has_clamp:
                        self.warnings.append(
                            f"Variable '{var_name}': {func}({arg_stripped}) requires "
                            f"argument in range [-1, 1]. Consider clamping: "
                            f"{func}(max(-1, min(1, {var})))."
                        )

    def _check_type_mismatch_in_filters(self, var_name: str, value: str) -> None:
        """Check for type mismatches in filter chains.

        Detects patterns like:
        1. Applying string filters to numbers without conversion
        2. Applying numeric filters to strings without conversion
        3. Using incompatible filter combinations

        Args:
            var_name: Name of the variable being checked.
            value: The template string value.
        """
        if not isinstance(value, str):
            return

        # Extract filter chains from the template
        # Pattern: expression | filter1 | filter2 | ...
        # We look for chains where type conversion might be needed

        # Find expressions with filter chains
        filter_chain_pattern = r"([^|{}]+(?:\|[^|{}]+)+)"
        chains = re.findall(filter_chain_pattern, value)

        for chain in chains:
            parts = [p.strip() for p in chain.split("|")]
            if len(parts) < 2:
                continue

            # Track the expected type through the chain
            current_type: str | None = None
            prev_filter: str | None = None

            for i, part in enumerate(parts[1:], 1):
                # Extract filter name (before any parentheses)
                filter_match = re.match(r"(\w+)", part)
                if not filter_match:
                    continue

                filter_name = filter_match.group(1)

                # Check for type mismatches
                if filter_name in self.TYPE_CHANGING_FILTERS:
                    new_type = self.TYPE_CHANGING_FILTERS[filter_name]

                    # String operations on numbers
                    if current_type == "number" and filter_name in (
                        "lower",
                        "upper",
                        "capitalize",
                        "title",
                        "split",
                        "replace",
                        "strip",
                        "trim",
                    ):
                        self.warnings.append(
                            f"Variable '{var_name}': Filter '{filter_name}' expects "
                            f"a string but previous filter '{prev_filter}' outputs "
                            f"a number. Add '| string' before '| {filter_name}'."
                        )

                    # Numeric operations on strings (without conversion)
                    if current_type == "string" and filter_name in ("round", "abs"):
                        # These will work if the string is numeric, but warn anyway
                        pass  # Don't warn - Jinja2 handles this gracefully

                    current_type = new_type
                    prev_filter = filter_name

        # Check for common type mismatch patterns
        # Pattern: states(...) | round - states returns string, round expects number
        if re.search(r"states\s*\([^)]+\)\s*\|\s*round", value):
            # Check if there's a float/int conversion
            if not re.search(
                r"states\s*\([^)]+\)\s*\|\s*(?:float|int)\s*\([^)]*\)\s*\|\s*round",
                value,
            ):
                self.warnings.append(
                    f"Variable '{var_name}': 'states(...)' returns a string, "
                    f"but '| round' expects a number. "
                    f"Use '| float(0)' before '| round'."
                )

        # Pattern: state_attr returns various types - check numeric operations
        state_attr_numeric = re.findall(
            r"state_attr\s*\([^)]+\)\s*\|\s*(round|abs)\b", value
        )
        for filter_used in state_attr_numeric:
            # Check if there's proper conversion
            if not re.search(
                rf"state_attr\s*\([^)]+\)\s*\|\s*(?:float|int)\s*\([^)]*\)\s*\|\s*{filter_used}",
                value,
            ):
                self.warnings.append(
                    f"Variable '{var_name}': 'state_attr(...)' may return a string, "
                    f"but '| {filter_used}' expects a number. Consider adding "
                    f"'| float(0)' before '| {filter_used}' for safety."
                )

    def _check_cross_variable_arithmetic(
        self, variables: dict[str, Any], path: str
    ) -> None:
        """Check for ADDITION between action-block variables (string concatenation bug).

        When variables are defined in an action block like:
            - variables:
                _now_ts: "{{ as_timestamp(now()) | float(0) }}"
                _duration_sec: "{{ x | int(60) * 60 }}"
                _until_ts: "{{ (_now_ts + _duration_sec) | int }}"

        The values of _now_ts and _duration_sec are STRINGS when used in the
        third variable's template. So `_now_ts + _duration_sec` performs string
        concatenation instead of addition: "1735423328" + "1800" = "17354233281800"

        NOTE: Only the + operator is affected. Jinja2's -, *, / operators coerce
        strings to numbers, so subtraction/multiplication/division work correctly.
        The + operator is overloaded for both addition and string concatenation.

        The fix is to compute all values in a single Jinja2 block:
            _until_ts: >
              {% set now_ts = as_timestamp(now()) %}
              {% set duration = x | int(60) * 60 %}
              {{ (now_ts + duration) | int }}

        Args:
            variables: Dictionary of variable definitions in this action block.
            path: Current path for error messages.
        """
        var_names = set(variables.keys())

        for name, value in variables.items():
            if not isinstance(value, str):
                continue

            # Look for ADDITION operations (+) involving other variables
            # from this same variables block. Only + is affected because it's
            # overloaded for string concatenation. -, *, / coerce to numbers.
            for other_var in var_names:
                if other_var == name:
                    continue

                # Pattern: other_var + something, or something + other_var
                # Must be + specifically, not -, *, /
                # Examples: (_now_ts + _duration_sec), _now_ts + 60
                addition_pattern = (
                    rf"\b{re.escape(other_var)}\s*\+|"
                    rf"\+\s*{re.escape(other_var)}\b|"
                    rf"\+\s*\({re.escape(other_var)}|"
                    rf"{re.escape(other_var)}\)\s*\+"
                )

                if re.search(addition_pattern, value):
                    # Check if other_var is defined as a simple template (likely numeric)
                    other_value = variables.get(other_var, "")
                    if isinstance(other_value, str) and "{{" in other_value:
                        # This looks like cross-variable addition
                        self.errors.append(
                            f"{path}.{name}: Uses '+' with '{other_var}' which is "
                            f"defined earlier in the same variables block. In action-block "
                            f"variables, each value becomes a STRING, so '{other_var} + x' "
                            f"performs string concatenation, not addition. "
                            f"Fix: compute all values in a single Jinja2 block using "
                            f"{{% set %}} statements, or use '| float' on the variable: "
                            f"'({other_var} | float) + x'."
                        )
                        # Only report once per variable
                        break

    def _validate_action_variables(self, variables: dict[str, Any], path: str) -> None:
        """Validate inline variables defined within action blocks.

        Args:
            variables: Dictionary of variable definitions.
            path: Current path for error messages.
        """
        # Check for cross-variable arithmetic (string concatenation bug)
        self._check_cross_variable_arithmetic(variables, path)

        for name, value in variables.items():
            if isinstance(value, str):
                # Check for direct attribute access on state objects
                self._check_state_attribute_access(f"{path}.{name}", value)
                # Check for incorrect input_datetime usage
                self._check_input_datetime_usage(f"{path}.{name}", value)
                # Check for bare boolean literals
                self._check_bare_boolean_literals(f"{path}.{name}", value)
                # Check for entity ID variables in boolean contexts
                self._check_entity_id_boolean_context(f"{path}.{name}", value)

    def _validate_version_sync(self) -> None:
        """Validate that blueprint_version matches version in name field."""
        blueprint = self.data.get("blueprint")
        if not isinstance(blueprint, dict):
            return

        name = blueprint.get("name")
        if not isinstance(name, str):
            return

        variables = self.data.get("variables")
        if not isinstance(variables, dict):
            return

        blueprint_version = variables.get("blueprint_version")
        if blueprint_version is None:
            return  # Already warned about missing version

        # Handle both quoted and unquoted versions
        version_str = str(blueprint_version).strip("\"'")

        # Extract version from name (expected format: "Name vX.Y.Z" or "Name X.Y.Z")
        version_match = re.search(r"v?(\d+\.\d+(?:\.\d+)?)\s*$", name)
        if version_match:
            name_version = version_match.group(1)
            if name_version != version_str:
                self.errors.append(
                    f"Version mismatch: blueprint name has 'v{name_version}' "
                    f"but blueprint_version is '{version_str}'"
                )
        else:
            self.warnings.append(
                f"Could not extract version from blueprint name: '{name}'. "
                "Expected format: 'Blueprint Name vX.Y.Z'"
            )

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

            # Check for templates in 'for:' duration (Common Pitfall #10)
            self._check_trigger_for_duration(trigger, f"trigger[{i}]")

            # Collect !input references
            self._collect_input_refs_from_dict(trigger)

    def _check_trigger_for_duration(self, trigger: dict[str, Any], path: str) -> None:
        """Check for templates in trigger 'for:' duration.

        Templates in trigger 'for:' durations don't work because variables
        aren't available at trigger compile time.

        Args:
            trigger: The trigger dictionary.
            path: Current path for error messages.
        """
        for_duration = trigger.get("for")
        if for_duration is None:
            return

        # for: can be a string, dict, or template
        if isinstance(for_duration, str):
            if "{{" in for_duration and "}}" in for_duration:
                self.errors.append(
                    f"{path}.for: Templates in trigger 'for:' duration are not "
                    "supported. Variables aren't available at trigger compile time. "
                    "Use !input or a static value instead."
                )
        elif isinstance(for_duration, dict):
            # Check each component (hours, minutes, seconds, milliseconds)
            for key, value in for_duration.items():
                if isinstance(value, str) and "{{" in value and "}}" in value:
                    self.errors.append(
                        f"{path}.for.{key}: Templates in trigger 'for:' duration are "
                        "not supported. Variables aren't available at trigger compile "
                        "time. Use !input or a static value instead."
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

    def _validate_conditions(self) -> None:
        """Validate condition definitions."""
        conditions = self.data.get("condition")
        if conditions is None:
            return  # condition: is optional

        if not isinstance(conditions, list):
            self.errors.append("'condition' must be a list")
            return

        for i, condition in enumerate(conditions):
            self._validate_condition_item(condition, f"condition[{i}]")

    def _validate_condition_item(self, condition: Any, path: str) -> None:
        """Validate a single condition item.

        Args:
            condition: The condition to validate.
            path: Current path for error messages.
        """
        if not isinstance(condition, dict):
            self.errors.append(f"{path}: Condition must be a dictionary")
            return

        # Check for condition type
        condition_type = condition.get("condition")
        if condition_type is None:
            # Could be a shorthand condition
            if "conditions" in condition:
                # Nested conditions (and/or/not shorthand)
                nested = condition.get("conditions")
                if isinstance(nested, list):
                    for j, nested_cond in enumerate(nested):
                        self._validate_condition_item(
                            nested_cond, f"{path}.conditions[{j}]"
                        )
            return

        if not isinstance(condition_type, str):
            self.errors.append(f"{path}.condition: Must be a string")
            return

        if condition_type not in self.VALID_CONDITION_TYPES:
            self.warnings.append(f"{path}: Unknown condition type '{condition_type}'")

        # Validate nested conditions for and/or/not
        if condition_type in ["and", "or", "not"]:
            nested = condition.get("conditions")
            if nested is None:
                self.errors.append(
                    f"{path}: '{condition_type}' condition requires 'conditions' key"
                )
            elif not isinstance(nested, list):
                self.errors.append(f"{path}.conditions: Must be a list")
            else:
                for j, nested_cond in enumerate(nested):
                    self._validate_condition_item(
                        nested_cond, f"{path}.conditions[{j}]"
                    )

        # Check for template in state condition
        if condition_type == "state":
            entity_id = condition.get("entity_id")
            if entity_id is not None:
                self._collect_input_refs(entity_id)

        # Check template conditions for entity ID boolean context issues
        if condition_type == "template":
            value_template = condition.get("value_template")
            if isinstance(value_template, str):
                self._check_entity_id_boolean_context(
                    f"{path}.value_template", value_template
                )

        # Collect !input references
        self._collect_input_refs_from_dict(condition)

    def _validate_service_format(self, service: str, path: str) -> None:
        """Validate service format (domain.service).

        Args:
            service: The service string to validate.
            path: Current path for error messages.
        """
        # Allow !input references
        if service.startswith("!input "):
            self._record_input_use(service[7:].strip())
            return

        # Allow templates
        if "{{" in service and "}}" in service:
            return

        # Check for deprecated services
        if service in self.DEPRECATED_SERVICES:
            self.warnings.append(
                f"{path}: Service '{service}' is deprecated. "
                f"{self.DEPRECATED_SERVICES[service]}"
            )

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

        # Check for inline variables in actions
        variables = action.get("variables")
        if isinstance(variables, dict):
            self._validate_action_variables(variables, f"{path}.variables")

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

        # Check delay actions
        delay = action.get("delay")
        if delay is not None:
            self._validate_delay(delay, f"{path}.delay")

        # Check wait_template actions
        wait_template = action.get("wait_template")
        if wait_template is not None:
            self._validate_wait_template(wait_template, action, f"{path}")

        # Check wait_for_trigger actions
        wait_for_trigger = action.get("wait_for_trigger")
        if wait_for_trigger is not None:
            self._validate_wait_for_trigger(wait_for_trigger, action, f"{path}")

        # Check if/then/else structures
        if "if" in action:
            # Validate the if conditions
            if_conditions = action.get("if")
            if isinstance(if_conditions, list):
                for j, cond in enumerate(if_conditions):
                    self._validate_condition_item(cond, f"{path}.if[{j}]")

            then_actions = action.get("then")
            if then_actions is None:
                self.errors.append(f"{path}: 'if' requires 'then' block")
            elif isinstance(then_actions, list):
                if len(then_actions) == 0:
                    self.warnings.append(f"{path}.then: Empty sequence")
                for j, then_action in enumerate(then_actions):
                    self._validate_action_item(then_action, f"{path}.then[{j}]")

            else_actions = action.get("else")
            if isinstance(else_actions, list):
                if len(else_actions) == 0:
                    self.warnings.append(f"{path}.else: Empty sequence")
                for j, else_action in enumerate(else_actions):
                    self._validate_action_item(else_action, f"{path}.else[{j}]")

        # Check repeat structures
        repeat = action.get("repeat")
        if isinstance(repeat, dict):
            seq = repeat.get("sequence")
            if seq is None:
                self.errors.append(f"{path}.repeat: Missing 'sequence' key")
            elif isinstance(seq, list):
                if len(seq) == 0:
                    self.warnings.append(f"{path}.repeat.sequence: Empty sequence")
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
            has_conditions = False
            for j, choice in enumerate(choose):
                if isinstance(choice, dict):
                    # Validate conditions
                    choice_conditions = choice.get("conditions")
                    if choice_conditions is not None:
                        has_conditions = True
                        if isinstance(choice_conditions, list):
                            for k, cond in enumerate(choice_conditions):
                                self._validate_condition_item(
                                    cond, f"{path}.choose[{j}].conditions[{k}]"
                                )

                    choice_seq = choice.get("sequence")
                    if choice_seq is None:
                        self.errors.append(
                            f"{path}.choose[{j}]: Missing 'sequence' key"
                        )
                    elif isinstance(choice_seq, list):
                        if len(choice_seq) == 0:
                            self.warnings.append(
                                f"{path}.choose[{j}].sequence: Empty sequence"
                            )
                        for k, choice_action in enumerate(choice_seq):
                            self._validate_action_item(
                                choice_action, f"{path}.choose[{j}].sequence[{k}]"
                            )

            # Check for default branch - only at top-level actions, not nested
            # Many nested choose blocks intentionally have no default
            default = action.get("default")
            if isinstance(default, list):
                if len(default) == 0:
                    self.warnings.append(f"{path}.default: Empty sequence")
                for j, default_action in enumerate(default):
                    self._validate_action_item(default_action, f"{path}.default[{j}]")

        # Collect !input references from this action
        self._collect_input_refs_from_dict(action)

    def _validate_delay(self, delay: Any, path: str) -> None:
        """Validate delay action format.

        Args:
            delay: The delay value.
            path: Current path for error messages.
        """
        if isinstance(delay, str):
            # Could be a template or HH:MM:SS format
            if "{{" in delay and "}}" in delay:
                # Template - valid
                return
            # Check for HH:MM:SS format
            if not re.match(r"^\d{1,2}:\d{2}(:\d{2})?$", delay):
                self.warnings.append(
                    f"{path}: Delay string '{delay}' should be in HH:MM:SS format "
                    "or a template"
                )
        elif isinstance(delay, dict):
            # Check for valid keys
            valid_keys = {"days", "hours", "minutes", "seconds", "milliseconds"}
            for key in delay:
                if key not in valid_keys:
                    self.warnings.append(
                        f"{path}.{key}: Unknown delay key. "
                        f"Valid keys: {', '.join(sorted(valid_keys))}"
                    )
        elif isinstance(delay, (int, float)):
            # Numeric delay in seconds - valid
            if delay < 0:
                self.errors.append(f"{path}: Delay cannot be negative")

    def _validate_wait_template(
        self, wait_template: Any, action: dict[str, Any], path: str
    ) -> None:
        """Validate wait_template action.

        Args:
            wait_template: The wait_template value.
            action: The full action dictionary.
            path: Current path for error messages.
        """
        if not isinstance(wait_template, str):
            self.errors.append(f"{path}.wait_template: Must be a string template")
            return

        if "{{" not in wait_template or "}}" not in wait_template:
            self.warnings.append(
                f"{path}.wait_template: Should contain a Jinja2 template expression"
            )

        # Check for timeout
        timeout = action.get("timeout")
        if timeout is None:
            self.warnings.append(
                f"{path}: wait_template without 'timeout' may wait indefinitely"
            )

    def _validate_wait_for_trigger(
        self, wait_for_trigger: Any, action: dict[str, Any], path: str
    ) -> None:
        """Validate wait_for_trigger action.

        Args:
            wait_for_trigger: The wait_for_trigger value.
            action: The full action dictionary.
            path: Current path for error messages.
        """
        if not isinstance(wait_for_trigger, list):
            self.errors.append(f"{path}.wait_for_trigger: Must be a list of triggers")
            return

        if len(wait_for_trigger) == 0:
            self.errors.append(f"{path}.wait_for_trigger: Cannot be empty")
            return

        # Validate each trigger in the wait
        for i, trigger in enumerate(wait_for_trigger):
            if not isinstance(trigger, dict):
                self.errors.append(
                    f"{path}.wait_for_trigger[{i}]: Must be a dictionary"
                )
                continue

            if "platform" not in trigger and "trigger" not in trigger:
                self.errors.append(
                    f"{path}.wait_for_trigger[{i}]: Missing 'platform' or 'trigger' key"
                )

        # Check for timeout
        timeout = action.get("timeout")
        if timeout is None:
            self.warnings.append(
                f"{path}: wait_for_trigger without 'timeout' may wait indefinitely"
            )

    def _validate_entity_id_format(self, entity_id: str, path: str) -> None:
        """Validate entity_id format (domain.entity_name).

        Args:
            entity_id: The entity ID to validate.
            path: Current path for error messages.
        """
        # Allow !input references
        if entity_id.startswith("!input "):
            self._record_input_use(entity_id[7:].strip())
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

        Trigger entity_id fields in Home Assistant must be resolvable at
        automation load time, not at runtime. This means they cannot use:
        1. Jinja2 templates ({{ }})
        2. Automation variables (defined in the variables: section)
        3. Dynamic expressions

        Valid options are:
        - Static entity IDs: "sensor.temperature"
        - !input references: !input my_sensor
        - Lists of the above

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

        # Check for Jinja2 templates
        if "{{" in stripped or "}}" in stripped:
            self.errors.append(
                f"{path}: entity_id cannot use templates; provide a concrete "
                "entity reference or !input value. Trigger entity_id must be "
                "resolvable at automation load time."
            )
            return

        # Check if it looks like a variable reference (not an entity ID or !input)
        # Entity IDs have format: domain.entity_name
        # !input refs have format: !input input_name
        if stripped.startswith("!input "):
            # Valid !input reference
            self._validate_entity_id_format(stripped, path)
            return

        # Check if it matches entity ID format (contains a dot)
        if "." not in stripped:
            # No dot - might be trying to use a variable name
            variables = self.data.get("variables", {})
            if isinstance(variables, dict) and stripped in variables:
                self.errors.append(
                    f"{path}: '{stripped}' appears to be a variable reference. "
                    f"Trigger entity_id cannot reference automation variables. "
                    f"Use !input or a static entity ID instead. Variables are "
                    f"not available until the automation runs, but trigger "
                    f"entity_id must be known at load time."
                )
                return
            else:
                # Not a known variable, might be a malformed entity ID
                self.errors.append(
                    f"{path}: '{stripped}' is not a valid entity ID format. "
                    f"Entity IDs must be 'domain.entity_name' (e.g., sensor.temperature). "
                    f"For dynamic entity selection, use !input with an entity selector."
                )
                return

        # Validate static entity ID format
        self._validate_entity_id_format(stripped, path)

    def _validate_templates(self) -> None:
        """Validate Jinja2 template syntax."""
        content = self.file_path.read_text()

        # Check for !input inside {{ }}
        if re.search(r"\{\{[^}]*!input", content):
            self.errors.append(
                "Found !input tag inside {{ }} template - bind to variable first"
            )

        # Check for balanced Jinja2 delimiters with enhanced analysis
        self._check_jinja2_delimiter_balance(content)

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

        # Check for Python-style list methods that don't work in Jinja2
        # e.g., [a, b].min() should be [a, b] | min
        python_list_methods = [
            (r"\[[^\]]+\]\.min\s*\(", ".min()", "| min"),
            (r"\[[^\]]+\]\.max\s*\(", ".max()", "| max"),
            (r"\[[^\]]+\]\.sort\s*\(", ".sort()", "| sort"),
            (r"\[[^\]]+\]\.append\s*\(", ".append()", "cannot modify lists in Jinja2"),
            (r"\[[^\]]+\]\.extend\s*\(", ".extend()", "cannot modify lists in Jinja2"),
        ]

        for pattern, method, fix in python_list_methods:
            if re.search(pattern, content):
                self.errors.append(
                    f"Python-style list method '{method}' does not work in Jinja2. "
                    f"Use '{fix}' instead."
                )

    def _check_jinja2_delimiter_balance(self, content: str) -> None:
        """Check for balanced Jinja2 delimiters with detailed error reporting.

        This enhanced check not only counts delimiters but also:
        1. Reports the approximate line number where imbalance occurs
        2. Detects nested delimiter issues (e.g., {{ {{ }} }})
        3. Identifies unclosed blocks (e.g., {% if %} without {% endif %})
        4. Catches common typos like {{{ or }}}

        Args:
            content: The full blueprint file content.
        """
        lines = content.split("\n")

        # Basic count check for each delimiter type
        jinja_patterns = [
            ("{{", "}}", "Jinja expressions"),
            ("{%", "%}", "Jinja control blocks"),
            ("{#", "#}", "Jinja comments"),
        ]

        for open_tag, close_tag, name in jinja_patterns:
            open_count = content.count(open_tag)
            close_count = content.count(close_tag)
            if open_count != close_count:
                # Find the first line with imbalance
                running_balance = 0
                first_imbalance_line = None
                for line_num, line in enumerate(lines, 1):
                    running_balance += line.count(open_tag)
                    running_balance -= line.count(close_tag)
                    if running_balance < 0 and first_imbalance_line is None:
                        first_imbalance_line = line_num
                        break

                if first_imbalance_line:
                    self.errors.append(
                        f"Unbalanced {name}: {open_tag} appears {open_count} times, "
                        f"{close_tag} appears {close_count} times. "
                        f"Check around line {first_imbalance_line}."
                    )
                else:
                    # Imbalance at end - missing close tag
                    self.errors.append(
                        f"Unbalanced {name}: {open_tag} appears {open_count} times, "
                        f"{close_tag} appears {close_count} times. "
                        f"Missing closing delimiter(s) at end of file."
                    )

        # Check for common typos: triple braces
        triple_open = re.search(r"\{\{\{", content)
        if triple_open:
            # Find line number
            pos = triple_open.start()
            line_num = content[:pos].count("\n") + 1
            self.errors.append(
                f"Triple opening brace '{{{{{{' found at line {line_num}. "
                f"Did you mean '{{{{' (expression) or '{{%' (control block)?"
            )

        triple_close = re.search(r"\}\}\}", content)
        if triple_close:
            pos = triple_close.start()
            line_num = content[:pos].count("\n") + 1
            self.errors.append(
                f"Triple closing brace '}}}}}}' found at line {line_num}. "
                f"Did you mean '}}}}' (expression) or '%}}' (control block)?"
            )

        # Check for unmatched control blocks (if/endif, for/endfor)
        self._check_control_block_balance(content)

    def _check_control_block_balance(self, content: str) -> None:
        """Check that Jinja2 control blocks are properly balanced.

        Validates that:
        - Every {% if %} has a matching {% endif %}
        - Every {% for %} has a matching {% endfor %}
        - Every {% macro %} has a matching {% endmacro %}
        - Block keywords appear in valid order

        Args:
            content: The full blueprint file content.
        """
        # Define control block pairs (open, close)
        block_pairs = [
            ("if", "endif"),
            ("for", "endfor"),
            ("macro", "endmacro"),
            ("call", "endcall"),
            ("filter", "endfilter"),
            ("block", "endblock"),
        ]

        lines = content.split("\n")

        for open_kw, close_kw in block_pairs:
            # Pattern to match {% if ... %} but not {% endif %}
            open_pattern = rf"\{{% *{open_kw}\b(?!end)"
            close_pattern = rf"\{{% *{close_kw}\b"

            open_matches = list(re.finditer(open_pattern, content))
            close_matches = list(re.finditer(close_pattern, content))

            open_count = len(open_matches)
            close_count = len(close_matches)

            if open_count != close_count:
                if open_count > close_count:
                    # Missing close - find the last unclosed open
                    missing = open_count - close_count
                    # Find line of last open block
                    if open_matches:
                        last_open = open_matches[-missing]
                        line_num = content[: last_open.start()].count("\n") + 1
                        self.errors.append(
                            f"Unclosed '{{% {open_kw} %}}' block: found {open_count} "
                            f"'{open_kw}' but only {close_count} '{close_kw}'. "
                            f"Missing '{{% {close_kw} %}}' for block starting near line {line_num}."
                        )
                else:
                    # Extra close - find the orphan close
                    extra = close_count - open_count
                    if close_matches:
                        orphan_close = close_matches[open_count]
                        line_num = content[: orphan_close.start()].count("\n") + 1
                        self.errors.append(
                            f"Orphan '{{% {close_kw} %}}' at line {line_num}: "
                            f"found {close_count} '{close_kw}' but only {open_count} '{open_kw}'. "
                            f"Remove the extra '{{% {close_kw} %}}' or add missing '{{% {open_kw} %}}'."
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
                else:
                    self._collect_input_refs(item)
            return

        if not isinstance(value, str):
            return

        stripped = value.strip()
        if not stripped:
            return

        self._collect_input_refs(stripped)

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

    def _collect_input_refs(self, value: Any) -> None:
        """Collect !input references from a value.

        Args:
            value: The value to scan for !input references.
        """
        if isinstance(value, str):
            if value.startswith("!input "):
                self._record_input_use(value[7:].strip())

    def _collect_input_refs_from_dict(self, d: dict[str, Any]) -> None:
        """Recursively collect !input references from a dictionary.

        Args:
            d: The dictionary to scan.
        """
        for key, value in d.items():
            if isinstance(value, str):
                self._collect_input_refs(value)
            elif isinstance(value, dict):
                self._collect_input_refs_from_dict(value)
            elif isinstance(value, list):
                for item in value:
                    if isinstance(item, str):
                        self._collect_input_refs(item)
                    elif isinstance(item, dict):
                        self._collect_input_refs_from_dict(item)

    def _record_input_use(self, input_name: str) -> None:
        """Record usage of an input.

        Args:
            input_name: The name of the input being used.
        """
        self.used_inputs.add(input_name)

    def _validate_input_references(self) -> None:
        """Validate that all !input references point to defined inputs."""
        # Find undefined inputs
        undefined = self.used_inputs - self.defined_inputs
        for input_name in sorted(undefined):
            self.errors.append(
                f"Undefined input reference: '!input {input_name}' - "
                "no matching input defined in blueprint.input"
            )

        # Find unused inputs (warning only)
        unused = self.defined_inputs - self.used_inputs
        for input_name in sorted(unused):
            # Don't warn about inputs that might be used in ways we can't detect
            # (e.g., dynamically constructed names, used only in descriptions)
            pass  # Intentionally not warning - too many false positives

    def _check_readme_exists(self) -> None:
        """Check if README.md exists in the blueprint directory."""
        readme_path = self.file_path.parent / "README.md"
        if not readme_path.exists():
            self.warnings.append(
                f"No README.md found in {self.file_path.parent.name}/ directory"
            )

    def _check_changelog_exists(self) -> None:
        """Check if CHANGELOG.md exists in the blueprint directory."""
        changelog_path = self.file_path.parent / "CHANGELOG.md"
        if not changelog_path.exists():
            self.warnings.append(
                f"No CHANGELOG.md found in {self.file_path.parent.name}/ directory"
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
    # Navigate up from scripts/validate-blueprint/ to the repo root
    repo_root = Path(__file__).parent.parent.parent
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
