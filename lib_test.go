// SPDX-License-Identifier: MIT
package jtl_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/OrtheSnowJames/jtl"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "basic test",
			input: `>>>DOCTYPE=JTL;
>>>VERSION=0.1;
>>>ENV;
	>>>NAME=developerrrr;
	>>>NAME2=developerrrr2;
>>>BEGIN;
	>class="main", tag="test">test>$env:NAME2;
	>class="main", tag="test">test2>Hello, World!;
>>>END;`,
			expected: map[string]interface{}{
				"test": map[string]interface{}{
					"class":   "main",
					"tag":     "test",
					"content": "developerrrr2",
				},
				"test2": map[string]interface{}{
					"class":   "main",
					"tag":     "test",
					"content": "Hello, World!",
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid input",
			input:   `not a valid JTL document`,
			wantErr: true,
		},
		{
			name: "multiple elements on single line",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
    >>>NAME=dev; >>>ROLE=admin;
>>>BEGIN;
    >class="main", tag="div">test1>Hello; >class="alt", tag="span">test2>World;
>>>END;`,
			expected: map[string]interface{}{
				"test1": map[string]interface{}{
					"class":   "main",
					"tag":     "div",
					"content": "Hello",
				},
				"test2": map[string]interface{}{
					"class":   "alt",
					"tag":     "span",
					"content": "World",
				},
			},
			wantErr: false,
		},
		{
			name: "environment variable replacement",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
    >>>USER=james; >>>APP=myapp; >>>ENV=prod;
>>>BEGIN;
    >class="user", tag="span">userinfo>$env:USER;
    >class="env", tag="div">envinfo>$env:ENV;
>>>END;`,
			expected: map[string]interface{}{
				"userinfo": map[string]interface{}{
					"class":   "user",
					"tag":     "span",
					"content": "james",
				},
				"envinfo": map[string]interface{}{
					"class":   "env",
					"tag":     "div",
					"content": "prod",
				},
			},
			wantErr: false,
		},
		{
			name: "mixed format with multiple attributes",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
    >>>THEME=dark; >>>LANG=en;
>>>BEGIN;
    >class="btn", tag="button", color="blue", size="lg">btn1>Click; >class="text", tag="p", align="center">txt1>$env:LANG;
>>>END;`,
			expected: map[string]interface{}{
				"btn1": map[string]interface{}{
					"class":   "btn",
					"tag":     "button",
					"color":   "blue",
					"size":    "lg",
					"content": "Click",
				},
				"txt1": map[string]interface{}{
					"class":   "text",
					"tag":     "p",
					"align":   "center",
					"content": "en",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid element format",
			input: `>>>DOCTYPE=JTL;
>>>BEGIN;
    >invalid>format>here;
>>>END;`,
			wantErr: true,
		},
		{
			name: "complex document with all features",
			input: `>>>DOCTYPE=JTL;
>>>VERSION=0.1;
>>>ENV;
    >>>DB_HOST=localhost;
    >>>PORT=8080; >>>ENV=prod;
>>>BEGIN;
    >class="container", tag="div", id="main">root>Hello;
    >class="btn", tag="button", color="blue">btn1>$env:ENV;
>>>END;`,
			expected: map[string]interface{}{
				"root": map[string]interface{}{
					"class":   "container",
					"tag":     "div",
					"id":      "main",
					"content": "Hello",
				},
				"btn1": map[string]interface{}{
					"class":   "btn",
					"tag":     "button",
					"color":   "blue",
					"content": "prod",
				},
			},
			wantErr: false,
		},
		{
			name: "missing attributes",
			input: `>>>DOCTYPE=JTL;
>>>BEGIN;
    >invalid>test>content;
>>>END;`,
			wantErr: true,
		},
		{
			name: "missing DOCTYPE",
			input: `>>>BEGIN;
    >class="test", tag="div">id>content;
>>>END;`,
			wantErr: true,
		},
		{
			name: "invalid element format",
			input: `>>>DOCTYPE=JTL;
>>>BEGIN;
    >class="test"incomplete;
>>>END;`,
			wantErr: true,
		},
		{
			name: "empty content",
			input: `>>>DOCTYPE=JTL;
>>>BEGIN;
    >class="test", tag="div">id>;
>>>END;`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := jtl.Parse(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Compare the actual and expected results
			if !reflect.DeepEqual(parsed, tt.expected) {
				actualJSON, _ := json.MarshalIndent(parsed, "", "  ")
				expectedJSON, _ := json.MarshalIndent(tt.expected, "", "  ")
				t.Errorf("Parse() got = %s\nwant = %s", actualJSON, expectedJSON)
			}
		})
	}
}

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "multiple env vars on single line",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
>>>NAME=developer; >>>NAME2=tester; >>>NAME3=admin`,
			expected: map[string]interface{}{
				"NAME":  "developer",
				"NAME2": "tester",
				"NAME3": "admin",
			},
			wantErr: false,
		},
		{
			name: "single env var per line",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
>>>NAME=developer;
>>>NAME2=tester;`,
			expected: map[string]interface{}{
				"NAME":  "developer",
				"NAME2": "tester",
			},
			wantErr: false,
		},
		{
			name: "multiple vars with different formats",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
    >>>USER=john; >>>ROLE=admin; >>>API_KEY=123xyz;
    >>>DB_HOST=localhost; >>>DB_PORT=5432;
    >>>DEBUG=true; >>>MODE=production; >>>VERSION=1.0.0;`,
			expected: map[string]interface{}{
				"USER":    "john",
				"ROLE":    "admin",
				"API_KEY": "123xyz",
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
				"DEBUG":   "true",
				"MODE":    "production",
				"VERSION": "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "environment section without variables",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
>>>BEGIN;`,
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name: "invalid env format",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
    >>>INVALID_FORMAT`,
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name: "multiple declarations with mixed formats",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
    >>>HOST=localhost; >>>PORT=8080;
    >>>USERNAME=admin;
    >>>PASSWORD=secret; >>>DEBUG=true; >>>MODE=development;`,
			expected: map[string]interface{}{
				"HOST":     "localhost",
				"PORT":     "8080",
				"USERNAME": "admin",
				"PASSWORD": "secret",
				"DEBUG":    "true",
				"MODE":     "development",
			},
			wantErr: false,
		},
		{
			name: "special characters in values",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
    >>>URL=https://example.com;
    >>>PATH=/usr/local/bin;`,
			expected: map[string]interface{}{
				"URL":  "https://example.com",
				"PATH": "/usr/local/bin",
			},
			wantErr: false,
		},
		{
			name: "missing DOCTYPE",
			input: `>>>ENV;
    >>>TEST=value;`,
			wantErr: true,
		},
		{
			name: "empty env section",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
>>>BEGIN;`,
			expected: map[string]interface{}{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jtl.ParseEnv(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("ParseEnv() for key %s = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestStringify(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		wantErr  bool
	}{
		{
			name: "simple map",
			input: map[string]interface{}{
				"key": "value",
			},
			wantErr: false,
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes, err := json.MarshalIndent(tt.input, "", "  ")
			got := string(gotBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stringify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// Verify the output is valid JSON
			var parsed interface{}
			if err := json.Unmarshal([]byte(got), &parsed); err != nil {
				t.Errorf("Stringify() produced invalid JSON: %v", err)
			}

			// Verify the structure matches
			expected, _ := json.MarshalIndent(tt.input, "", "  ")
			if got != string(expected) {
				t.Errorf("Stringify() produced unexpected format:\ngot = %v\nwant = %v", got, string(expected))
			}
		})
	}
}
