package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/ha-ws-client-go/internal/types"
)

func TestWithConfig(t *testing.T) {
	t.Parallel()

	t.Run("creates config if nil", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{}
		config := WithConfig(ctx)

		require.NotNil(t, config)
		assert.Same(t, config, ctx.Config)
	})

	t.Run("returns existing config", func(t *testing.T) {
		t.Parallel()
		existing := &HandlerConfig{AutomationID: "test"}
		ctx := &Context{Config: existing}
		config := WithConfig(ctx)

		assert.Same(t, existing, config)
		assert.Equal(t, "test", config.AutomationID)
	})
}

func TestRequireArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		indices     []int
		usage       string
		expectError bool
		expectedVal []string
	}{
		{
			name:        "single required arg at index 1",
			args:        []string{"cmd", "arg1"},
			indices:     []int{1},
			usage:       "Usage: cmd <arg>",
			expectError: false,
			expectedVal: []string{"arg1"},
		},
		{
			name:        "multiple required args",
			args:        []string{"cmd", "arg1", "arg2"},
			indices:     []int{1, 2},
			usage:       "Usage: cmd <arg1> <arg2>",
			expectError: false,
			expectedVal: []string{"arg1", "arg2"},
		},
		{
			name:        "missing single arg",
			args:        []string{"cmd"},
			indices:     []int{1},
			usage:       "Usage: cmd <arg>",
			expectError: true,
		},
		{
			name:        "missing second arg",
			args:        []string{"cmd", "arg1"},
			indices:     []int{1, 2},
			usage:       "Usage: cmd <arg1> <arg2>",
			expectError: true,
		},
		{
			name:        "non-sequential indices",
			args:        []string{"cmd", "a", "b", "c", "d"},
			indices:     []int{1, 3},
			usage:       "test",
			expectError: false,
			expectedVal: []string{"a", "c"},
		},
		{
			name:        "empty args",
			args:        []string{},
			indices:     []int{0},
			usage:       "Usage: cmd",
			expectError: true,
		},
		{
			name:        "index 0 (command name)",
			args:        []string{"cmd", "arg1"},
			indices:     []int{0},
			usage:       "test",
			expectError: false,
			expectedVal: []string{"cmd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handlerCalled := false
			handler := RequireArgs(tt.usage, tt.indices...)(func(ctx *Context) error {
				handlerCalled = true
				if tt.expectedVal != nil {
					assert.Equal(t, tt.expectedVal, ctx.Config.Args)
				}
				return nil
			})

			ctx := &Context{Args: tt.args}
			err := handler(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "missing argument")
				assert.Contains(t, err.Error(), tt.usage)
				assert.False(t, handlerCalled)
			} else {
				assert.NoError(t, err)
				assert.True(t, handlerCalled)
			}
		})
	}
}

func TestRequireArg1(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		handler := RequireArg1("Usage: state <entity_id>")(func(ctx *Context) error {
			assert.Equal(t, []string{"light.kitchen"}, ctx.Config.Args)
			return nil
		})

		ctx := &Context{Args: []string{"state", "light.kitchen"}}
		err := handler(ctx)
		require.NoError(t, err)
	})

	t.Run("missing arg", func(t *testing.T) {
		t.Parallel()
		handler := RequireArg1("Usage: state <entity_id>")(func(_ *Context) error {
			return nil
		})

		ctx := &Context{Args: []string{"state"}}
		err := handler(ctx)
		require.Error(t, err)
	})
}

func TestRequireArg2(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		handler := RequireArg2("Usage: call <domain> <service>")(func(ctx *Context) error {
			assert.Equal(t, []string{"light", "turn_on"}, ctx.Config.Args)
			return nil
		})

		ctx := &Context{Args: []string{"call", "light", "turn_on"}}
		err := handler(ctx)
		require.NoError(t, err)
	})

	t.Run("missing second arg", func(t *testing.T) {
		t.Parallel()
		handler := RequireArg2("Usage: call <domain> <service>")(func(_ *Context) error {
			return nil
		})

		ctx := &Context{Args: []string{"call", "light"}}
		err := handler(ctx)
		require.Error(t, err)
	})
}

func TestWithTimeRange(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		defaultHours  int
		hoursArgIndex int
		args          []string
		fromTime      *time.Time
		toTime        *time.Time
		checkFunc     func(t *testing.T, tr *types.TimeRange)
	}{
		{
			name:          "uses default hours",
			defaultHours:  24,
			hoursArgIndex: 2,
			args:          []string{"history", "sensor.temp"},
			checkFunc: func(t *testing.T, tr *types.TimeRange) {
				t.Helper()
				assert.InDelta(t, 24*time.Hour, tr.EndTime.Sub(tr.StartTime), float64(time.Minute))
			},
		},
		{
			name:          "uses arg hours",
			defaultHours:  24,
			hoursArgIndex: 2,
			args:          []string{"history", "sensor.temp", "48"},
			checkFunc: func(t *testing.T, tr *types.TimeRange) {
				t.Helper()
				assert.InDelta(t, 48*time.Hour, tr.EndTime.Sub(tr.StartTime), float64(time.Minute))
			},
		},
		{
			name:          "invalid arg hours uses default",
			defaultHours:  12,
			hoursArgIndex: 2,
			args:          []string{"history", "sensor.temp", "invalid"},
			checkFunc: func(t *testing.T, tr *types.TimeRange) {
				t.Helper()
				assert.InDelta(t, 12*time.Hour, tr.EndTime.Sub(tr.StartTime), float64(time.Minute))
			},
		},
		{
			name:          "uses fromTime if set",
			defaultHours:  24,
			hoursArgIndex: 2,
			args:          []string{"history", "sensor.temp"},
			fromTime:      func() *time.Time { t := now.Add(-48 * time.Hour); return &t }(),
			checkFunc: func(t *testing.T, tr *types.TimeRange) {
				t.Helper()
				assert.InDelta(t, 48*time.Hour, tr.EndTime.Sub(tr.StartTime), float64(time.Minute))
			},
		},
		{
			name:          "uses toTime if set",
			defaultHours:  24,
			hoursArgIndex: 2,
			args:          []string{"history", "sensor.temp"},
			toTime:        func() *time.Time { t := now.Add(-1 * time.Hour); return &t }(),
			checkFunc: func(t *testing.T, tr *types.TimeRange) {
				t.Helper()
				// End time should be 1 hour ago
				assert.InDelta(t, 1*time.Hour, now.Sub(tr.EndTime), float64(time.Minute))
			},
		},
		{
			name:          "hoursArgIndex 0 disables arg parsing",
			defaultHours:  6,
			hoursArgIndex: 0,
			args:          []string{"history", "sensor.temp", "24"},
			checkFunc: func(t *testing.T, tr *types.TimeRange) {
				t.Helper()
				// Should use default, not the "24" arg
				assert.InDelta(t, 6*time.Hour, tr.EndTime.Sub(tr.StartTime), float64(time.Minute))
			},
		},
		{
			name:          "hoursArgIndex beyond args length uses default",
			defaultHours:  4,
			hoursArgIndex: 5,
			args:          []string{"history", "sensor.temp"},
			checkFunc: func(t *testing.T, tr *types.TimeRange) {
				t.Helper()
				assert.InDelta(t, 4*time.Hour, tr.EndTime.Sub(tr.StartTime), float64(time.Minute))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := WithTimeRange(tt.defaultHours, tt.hoursArgIndex)(func(ctx *Context) error {
				tt.checkFunc(t, ctx.Config.TimeRange)
				return nil
			})

			ctx := &Context{
				Args:     tt.args,
				FromTime: tt.fromTime,
				ToTime:   tt.toTime,
			}
			err := handler(ctx)
			require.NoError(t, err)
		})
	}
}

func TestWithOptionalInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		defaultValue int
		argIndex     int
		args         []string
		expected     int
	}{
		{
			name:         "uses default when no arg",
			defaultValue: 60,
			argIndex:     2,
			args:         []string{"watch", "sensor.temp"},
			expected:     60,
		},
		{
			name:         "parses arg value",
			defaultValue: 60,
			argIndex:     2,
			args:         []string{"watch", "sensor.temp", "120"},
			expected:     120,
		},
		{
			name:         "uses default on invalid arg",
			defaultValue: 30,
			argIndex:     2,
			args:         []string{"watch", "sensor.temp", "not-a-number"},
			expected:     30,
		},
		{
			name:         "argIndex 0 uses default",
			defaultValue: 100,
			argIndex:     0,
			args:         []string{"watch", "50"},
			expected:     100,
		},
		{
			name:         "parses negative number",
			defaultValue: 10,
			argIndex:     1,
			args:         []string{"cmd", "-5"},
			expected:     -5,
		},
		{
			name:         "parses zero",
			defaultValue: 10,
			argIndex:     1,
			args:         []string{"cmd", "0"},
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := WithOptionalInt(tt.defaultValue, tt.argIndex)(func(ctx *Context) error {
				assert.Equal(t, tt.expected, ctx.Config.OptionalInt)
				return nil
			})

			ctx := &Context{Args: tt.args}
			err := handler(ctx)
			require.NoError(t, err)
		})
	}
}

func TestWithPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		argIndex    int
		args        []string
		testEntity  string
		shouldMatch bool
		nilPattern  bool
	}{
		{
			name:        "simple wildcard pattern",
			argIndex:    1,
			args:        []string{"states-filter", "light.*"},
			testEntity:  "light.kitchen",
			shouldMatch: true,
		},
		{
			name:        "wildcard at start",
			argIndex:    1,
			args:        []string{"states-filter", "*_temperature"},
			testEntity:  "sensor.outdoor_temperature",
			shouldMatch: true,
		},
		{
			name:        "case insensitive",
			argIndex:    1,
			args:        []string{"states-filter", "LIGHT.*"},
			testEntity:  "light.living_room",
			shouldMatch: true,
		},
		{
			name:        "exact match",
			argIndex:    1,
			args:        []string{"states-filter", "sensor.temp"},
			testEntity:  "sensor.temp",
			shouldMatch: true,
		},
		{
			name:        "no match",
			argIndex:    1,
			args:        []string{"states-filter", "light.*"},
			testEntity:  "switch.kitchen",
			shouldMatch: false,
		},
		{
			name:       "no pattern arg sets nil",
			argIndex:   1,
			args:       []string{"states"},
			nilPattern: true,
		},
		{
			name:       "argIndex 0 sets nil",
			argIndex:   0,
			args:       []string{"pattern"},
			nilPattern: true,
		},
		{
			name:        "multiple wildcards",
			argIndex:    1,
			args:        []string{"filter", "*sensor*temp*"},
			testEntity:  "binary_sensor.bathroom_temperature_high",
			shouldMatch: true,
		},
		{
			name:        "special regex chars escaped",
			argIndex:    1,
			args:        []string{"filter", "sensor.temp[1]"},
			testEntity:  "sensor.temp[1]",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := WithPattern(tt.argIndex)(func(ctx *Context) error {
				if tt.nilPattern {
					assert.Nil(t, ctx.Config.Pattern)
				} else {
					require.NotNil(t, ctx.Config.Pattern)
					assert.Equal(t, tt.shouldMatch, ctx.Config.Pattern.MatchString(tt.testEntity))
				}
				return nil
			})

			ctx := &Context{Args: tt.args}
			err := handler(ctx)
			require.NoError(t, err)
		})
	}
}

func TestWithRequiredPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		argIndex    int
		args        []string
		usage       string
		expectError bool
		testEntity  string
		shouldMatch bool
	}{
		{
			name:        "pattern provided",
			argIndex:    1,
			args:        []string{"states-filter", "light.*"},
			usage:       "Usage: states-filter <pattern>",
			expectError: false,
			testEntity:  "light.kitchen",
			shouldMatch: true,
		},
		{
			name:        "pattern missing",
			argIndex:    1,
			args:        []string{"states-filter"},
			usage:       "Usage: states-filter <pattern>",
			expectError: true,
		},
		{
			name:        "empty args",
			argIndex:    0,
			args:        []string{},
			usage:       "Usage: cmd <pattern>",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handlerCalled := false
			handler := WithRequiredPattern(tt.argIndex, tt.usage)(func(ctx *Context) error {
				handlerCalled = true
				if tt.testEntity != "" {
					require.NotNil(t, ctx.Config.Pattern)
					assert.Equal(t, tt.shouldMatch, ctx.Config.Pattern.MatchString(tt.testEntity))
				}
				return nil
			})

			ctx := &Context{Args: tt.args}
			err := handler(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "missing argument")
				assert.False(t, handlerCalled)
			} else {
				assert.NoError(t, err)
				assert.True(t, handlerCalled)
			}
		})
	}
}

func TestWithAutomationID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		argIndex    int
		args        []string
		usage       string
		expectError bool
		expectedID  string
	}{
		{
			name:        "strips automation prefix",
			argIndex:    1,
			args:        []string{"traces", "automation.kitchen_lights"},
			usage:       "Usage: traces <automation_id>",
			expectError: false,
			expectedID:  "kitchen_lights",
		},
		{
			name:        "no prefix to strip",
			argIndex:    1,
			args:        []string{"traces", "kitchen_lights"},
			usage:       "Usage: traces <automation_id>",
			expectError: false,
			expectedID:  "kitchen_lights",
		},
		{
			name:        "missing argument",
			argIndex:    1,
			args:        []string{"traces"},
			usage:       "Usage: traces <automation_id>",
			expectError: true,
		},
		{
			name:        "full entity ID",
			argIndex:    1,
			args:        []string{"trace", "automation.motion_triggered_lights"},
			usage:       "test",
			expectError: false,
			expectedID:  "motion_triggered_lights",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := WithAutomationID(tt.argIndex, tt.usage)(func(ctx *Context) error {
				assert.Equal(t, tt.expectedID, ctx.Config.AutomationID)
				return nil
			})

			ctx := &Context{Args: tt.args}
			err := handler(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "missing argument")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWithOptionalAutomationID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		argIndex   int
		args       []string
		expectedID string
	}{
		{
			name:       "strips automation prefix",
			argIndex:   1,
			args:       []string{"traces", "automation.test"},
			expectedID: "test",
		},
		{
			name:       "no arg provided",
			argIndex:   1,
			args:       []string{"traces"},
			expectedID: "",
		},
		{
			name:       "argIndex 0 ignores",
			argIndex:   0,
			args:       []string{"traces", "automation.test"},
			expectedID: "",
		},
		{
			name:       "no prefix",
			argIndex:   1,
			args:       []string{"traces", "my_automation"},
			expectedID: "my_automation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := WithOptionalAutomationID(tt.argIndex)(func(ctx *Context) error {
				assert.Equal(t, tt.expectedID, ctx.Config.AutomationID)
				return nil
			})

			ctx := &Context{Args: tt.args}
			err := handler(ctx)
			require.NoError(t, err)
		})
	}
}

func TestEnsureAutomationPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"kitchen_lights", "automation.kitchen_lights"},
		{"automation.kitchen_lights", "automation.kitchen_lights"},
		{"", "automation."},
		{"automation.", "automation."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := EnsureAutomationPrefix(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChain(t *testing.T) {
	t.Parallel()

	t.Run("chains multiple middleware in order", func(t *testing.T) {
		t.Parallel()

		var order []string

		m1 := func(next Handler) Handler {
			return func(ctx *Context) error {
				order = append(order, "m1-before")
				err := next(ctx)
				order = append(order, "m1-after")
				return err
			}
		}

		m2 := func(next Handler) Handler {
			return func(ctx *Context) error {
				order = append(order, "m2-before")
				err := next(ctx)
				order = append(order, "m2-after")
				return err
			}
		}

		handler := Chain(m1, m2)(func(_ *Context) error {
			order = append(order, "handler")
			return nil
		})

		ctx := &Context{}
		err := handler(ctx)
		require.NoError(t, err)

		assert.Equal(t, []string{
			"m1-before",
			"m2-before",
			"handler",
			"m2-after",
			"m1-after",
		}, order)
	})

	t.Run("early return on error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("validation failed")

		m1 := func(_ Handler) Handler {
			return func(_ *Context) error {
				return expectedErr
			}
		}

		m2 := func(next Handler) Handler {
			return func(ctx *Context) error {
				t.Error("m2 should not be called")
				return next(ctx)
			}
		}

		handler := Chain(m1, m2)(func(_ *Context) error {
			t.Error("handler should not be called")
			return nil
		})

		ctx := &Context{}
		err := handler(ctx)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("empty chain", func(t *testing.T) {
		t.Parallel()

		handlerCalled := false
		handler := Chain()(func(_ *Context) error {
			handlerCalled = true
			return nil
		})

		ctx := &Context{}
		err := handler(ctx)
		require.NoError(t, err)
		assert.True(t, handlerCalled)
	})

	t.Run("single middleware", func(t *testing.T) {
		t.Parallel()

		m := func(next Handler) Handler {
			return func(ctx *Context) error {
				ctx.Config = &HandlerConfig{AutomationID: "set-by-middleware"}
				return next(ctx)
			}
		}

		handler := Chain(m)(func(ctx *Context) error {
			assert.Equal(t, "set-by-middleware", ctx.Config.AutomationID)
			return nil
		})

		ctx := &Context{}
		err := handler(ctx)
		require.NoError(t, err)
	})
}

func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("applies middleware to handler", func(t *testing.T) {
		t.Parallel()

		m := RequireArg1("Usage: test <arg>")
		handlerCalled := false

		handler := Apply(m, func(ctx *Context) error {
			handlerCalled = true
			assert.Equal(t, []string{"myarg"}, ctx.Config.Args)
			return nil
		})

		ctx := &Context{Args: []string{"test", "myarg"}}
		err := handler(ctx)
		require.NoError(t, err)
		assert.True(t, handlerCalled)
	})

	t.Run("with chained middleware", func(t *testing.T) {
		t.Parallel()

		m := Chain(
			RequireArg1("Usage: history <entity>"),
			WithTimeRange(24, 2),
		)

		handler := Apply(m, func(ctx *Context) error {
			assert.Equal(t, []string{"sensor.temp"}, ctx.Config.Args)
			assert.NotNil(t, ctx.Config.TimeRange)
			return nil
		})

		ctx := &Context{Args: []string{"history", "sensor.temp", "48"}}
		err := handler(ctx)
		require.NoError(t, err)
	})
}

func TestHandlerConfig(t *testing.T) {
	t.Parallel()

	t.Run("stores all fields", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		cfg := &HandlerConfig{
			Args: []string{"a", "b"},
			TimeRange: &types.TimeRange{
				StartTime: now.Add(-1 * time.Hour),
				EndTime:   now,
			},
			OptionalInt:  42,
			AutomationID: "test_auto",
		}

		assert.Equal(t, []string{"a", "b"}, cfg.Args)
		assert.NotNil(t, cfg.TimeRange)
		assert.Equal(t, 42, cfg.OptionalInt)
		assert.Equal(t, "test_auto", cfg.AutomationID)
	})
}

func TestContext(t *testing.T) {
	t.Parallel()

	now := time.Now()
	ctx := &Context{
		Args:     []string{"cmd", "arg1", "arg2"},
		FromTime: &now,
		ToTime:   &now,
		Config:   &HandlerConfig{AutomationID: "test"},
	}

	assert.Nil(t, ctx.Client) // Would be *client.Client in real usage
	assert.Equal(t, []string{"cmd", "arg1", "arg2"}, ctx.Args)
	assert.NotNil(t, ctx.FromTime)
	assert.NotNil(t, ctx.ToTime)
	assert.Equal(t, "test", ctx.Config.AutomationID)
}

// Benchmark for middleware chain
func BenchmarkChain(b *testing.B) {
	m1 := RequireArg1("test")
	m2 := WithTimeRange(24, 2)
	m3 := WithOptionalInt(60, 3)

	handler := Chain(m1, m2, m3)(func(_ *Context) error {
		return nil
	})

	ctx := &Context{Args: []string{"cmd", "arg1", "24", "120"}}

	b.ResetTimer()
	for range b.N {
		_ = handler(ctx)
	}
}

// Benchmark for pattern matching
func BenchmarkWithPattern(b *testing.B) {
	handler := WithPattern(1)(func(ctx *Context) error {
		_ = ctx.Config.Pattern.MatchString("light.living_room")
		return nil
	})

	ctx := &Context{Args: []string{"filter", "light.*"}}

	b.ResetTimer()
	for range b.N {
		_ = handler(ctx)
	}
}
