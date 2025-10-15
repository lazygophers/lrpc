package lrpc

import (
	"fmt"
	"testing"
)

func TestRouteParser(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"root", "/", false},
		{"static", "/api/users", false},
		{"param", "/api/users/{id}", false},
		{"typed param", "/api/users/{id:int}", false},
		{"regex param", "/api/users/{id:^[0-9]+$}", false},
		{"wildcard", "/api/*", false},
		{"catchall", "/api/**", false},
		{"mixed params", "/api/{uid}-{cid}", false},
		{"colon param", "/api/:id", false},
		{"constraint", "/api/{id:int,min=1,max=100}", false},
		{"invalid", "api/users", true},
		{"invalid wildcard position", "/api/*/users", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := parseRoute(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				t.Logf("Pattern %q parsed into %d nodes", tt.pattern, len(nodes))
				for i, node := range nodes {
					t.Logf("  [%d] type=%d segment=%q priority=%d", i, node.typ, node.segment, node.priority)
				}
			}
		})
	}
}

func TestRouteMatching(t *testing.T) {
	router := NewRouter()

	// Add routes
	routes := map[string][]HandlerFunc{
		"/api/users/list":       {func(c *Ctx) error { return fmt.Errorf("handler1") }},
		"/api/users/{id}":       {func(c *Ctx) error { return fmt.Errorf("handler2") }},
		"/api/users/{id:int}":   {func(c *Ctx) error { return fmt.Errorf("handler3") }},
		"/api/users/{id:digit}": {func(c *Ctx) error { return fmt.Errorf("handler4") }},
		"/api/posts/*":          {func(c *Ctx) error { return fmt.Errorf("handler5") }},
		"/api/**":               {func(c *Ctx) error { return fmt.Errorf("handler6") }},
	}

	for pattern, handlers := range routes {
		if err := router.addRoute(pattern, handlers); err != nil {
			t.Fatalf("Failed to add route %q: %v", pattern, err)
		}
	}

	tests := []struct {
		path       string
		wantMatch  bool
		wantParams map[string]string
	}{
		{"/api/users/list", true, map[string]string{}},
		{"/api/users/123", true, map[string]string{"id": "123"}},
		{"/api/users/abc", true, map[string]string{"id": "abc"}},
		{"/api/posts/trending", true, map[string]string{"*": "trending"}},
		{"/api/anything/else", true, map[string]string{"*": "anything/else"}},
		{"/notfound", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result, err := router.findRoute(tt.path)
			if err != nil {
				t.Fatalf("findRoute() error = %v", err)
			}

			if tt.wantMatch && result == nil {
				t.Errorf("Expected match for %q, got nil", tt.path)
				return
			}

			if !tt.wantMatch && result != nil {
				t.Errorf("Expected no match for %q, got result", tt.path)
				return
			}

			if result != nil && tt.wantParams != nil {
				for k, v := range tt.wantParams {
					if got, ok := result.Params[k]; !ok || got != v {
						t.Errorf("Param %q: want %q, got %q", k, v, got)
					}
				}
			}
		})
	}
}

func TestParameterValidation(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		constraint *ParamConstraint
		wantErr    bool
	}{
		{
			"int valid",
			"123",
			&ParamConstraint{Type: "int"},
			false,
		},
		{
			"int invalid",
			"abc",
			&ParamConstraint{Type: "int"},
			true,
		},
		{
			"int with min",
			"5",
			&ParamConstraint{Type: "int", Min: intPtr(10)},
			true,
		},
		{
			"uuid valid",
			"550e8400-e29b-41d4-a716-446655440000",
			&ParamConstraint{Type: "uuid"},
			false,
		},
		{
			"uuid invalid",
			"not-a-uuid",
			&ParamConstraint{Type: "uuid"},
			true,
		},
		{
			"string len",
			"hello",
			&ParamConstraint{Type: "string", Len: intPtr(5)},
			false,
		},
		{
			"string len invalid",
			"hi",
			&ParamConstraint{Type: "string", Len: intPtr(5)},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParam(tt.value, tt.constraint)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateParam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func intPtr(v int) *int {
	return &v
}
