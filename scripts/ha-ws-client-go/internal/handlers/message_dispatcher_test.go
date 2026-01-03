package handlers

import (
	"testing"
)

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		data    map[string]any
	}{
		{
			name:    "simple request without data",
			msgType: "get_states",
			data:    nil,
		},
		{
			name:    "request with data",
			msgType: "trace/get",
			data: map[string]any{
				"domain":  "automation",
				"item_id": "test_automation",
				"run_id":  "12345",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest[map[string]any](tt.msgType, tt.data)
			if req.Type != tt.msgType {
				t.Errorf("NewRequest().Type = %v, want %v", req.Type, tt.msgType)
			}
			if tt.data != nil {
				for k, v := range tt.data {
					if req.Data[k] != v {
						t.Errorf("NewRequest().Data[%s] = %v, want %v", k, req.Data[k], v)
					}
				}
			}
		})
	}
}

func TestOutputConfig(t *testing.T) {
	tests := []struct {
		name        string
		opts        []OutputOption
		wantCommand string
		wantSummary string
		wantCount   int
	}{
		{
			name:        "with command",
			opts:        []OutputOption{WithOutputCommand("test-cmd")},
			wantCommand: "test-cmd",
		},
		{
			name:        "with summary",
			opts:        []OutputOption{WithOutputSummary("test summary")},
			wantSummary: "test summary",
		},
		{
			name:      "with count",
			opts:      []OutputOption{WithOutputCount(42)},
			wantCount: 42,
		},
		{
			name: "combined options",
			opts: []OutputOption{
				WithOutputCommand("combined-cmd"),
				WithOutputSummary("combined summary"),
				WithOutputCount(100),
			},
			wantCommand: "combined-cmd",
			wantSummary: "combined summary",
			wantCount:   100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := OutputConfig{}
			for _, opt := range tt.opts {
				opt(&cfg)
			}
			if cfg.Command != tt.wantCommand {
				t.Errorf("OutputConfig.Command = %v, want %v", cfg.Command, tt.wantCommand)
			}
			if cfg.Summary != tt.wantSummary {
				t.Errorf("OutputConfig.Summary = %v, want %v", cfg.Summary, tt.wantSummary)
			}
			if cfg.Count != tt.wantCount {
				t.Errorf("OutputConfig.Count = %v, want %v", cfg.Count, tt.wantCount)
			}
		})
	}
}

func TestDispatchCreation(t *testing.T) {
	d := Dispatch[[]string]("get_states", nil)
	if d.request.Type != "get_states" {
		t.Errorf("Dispatch request.Type = %v, want get_states", d.request.Type)
	}
}

func TestDispatchTransform(t *testing.T) {
	called := false
	d := Dispatch[[]string]("test", nil).Transform(func(s []string) ([]string, error) {
		called = true
		return append(s, "transformed"), nil
	})

	// Transform function should be set
	if d.transform == nil {
		t.Fatal("Transform function not set")
	}

	// Verify transform works by calling it directly
	result, err := d.transform([]string{"original"})
	if err != nil {
		t.Errorf("Transform error: %v", err)
	}
	if !called {
		t.Error("Transform function was not called")
	}
	if len(result) != 2 || result[1] != "transformed" {
		t.Errorf("Transform result = %v, want [original transformed]", result)
	}
}

func TestDispatchOutput(t *testing.T) {
	called := false
	d := Dispatch[string]("test", nil).Output(func(_ string) error {
		called = true
		return nil
	})

	// Output function should be set
	if d.outputFn == nil {
		t.Fatal("Output function not set")
	}

	// Verify output works by calling it directly
	err := d.outputFn("test")
	if err != nil {
		t.Errorf("Output error: %v", err)
	}
	if !called {
		t.Error("Output function was not called")
	}
}

func TestListRequest(t *testing.T) {
	lr := &ListRequest[string]{
		MessageType: "test/list",
		Title:       "Test Title",
		Command:     "test-cmd",
		Formatter: func(s string, _ int) string {
			return s
		},
		Filter: func(s string) bool {
			return s != ""
		},
	}

	if lr.MessageType != "test/list" {
		t.Errorf("ListRequest.MessageType = %v, want test/list", lr.MessageType)
	}
	if lr.Title != "Test Title" {
		t.Errorf("ListRequest.Title = %v, want Test Title", lr.Title)
	}
	if lr.Command != "test-cmd" {
		t.Errorf("ListRequest.Command = %v, want test-cmd", lr.Command)
	}
	if lr.Formatter == nil {
		t.Error("ListRequest.Formatter should not be nil")
	}
	if lr.Filter == nil {
		t.Error("ListRequest.Filter should not be nil")
	}
	// Test filter
	if !lr.Filter("hello") {
		t.Error("Filter should return true for non-empty string")
	}
	if lr.Filter("") {
		t.Error("Filter should return false for empty string")
	}
}

func TestTimelineRequest(t *testing.T) {
	tr := &TimelineRequest[string]{
		MessageType: "test/timeline",
		Title:       "Timeline Title",
		Command:     "timeline-cmd",
		Formatter: func(s string) string {
			return "formatted: " + s
		},
	}

	if tr.MessageType != "test/timeline" {
		t.Errorf("TimelineRequest.MessageType = %v, want test/timeline", tr.MessageType)
	}
	if tr.Title != "Timeline Title" {
		t.Errorf("TimelineRequest.Title = %v, want Timeline Title", tr.Title)
	}
	if tr.Command != "timeline-cmd" {
		t.Errorf("TimelineRequest.Command = %v, want timeline-cmd", tr.Command)
	}
	if tr.Formatter == nil {
		t.Error("TimelineRequest.Formatter should not be nil")
	}
	// Test formatter
	result := tr.Formatter("test")
	if result != "formatted: test" {
		t.Errorf("Formatter result = %v, want 'formatted: test'", result)
	}
}

func TestMapRequest(t *testing.T) {
	mr := &MapRequest[string]{
		MessageType:  "test/map",
		Key:          "testkey",
		EmptyMessage: "No data found",
	}

	if mr.MessageType != "test/map" {
		t.Errorf("MapRequest.MessageType = %v, want test/map", mr.MessageType)
	}
	if mr.Key != "testkey" {
		t.Errorf("MapRequest.Key = %v, want testkey", mr.Key)
	}
	if mr.EmptyMessage != "No data found" {
		t.Errorf("MapRequest.EmptyMessage = %v, want 'No data found'", mr.EmptyMessage)
	}
}
