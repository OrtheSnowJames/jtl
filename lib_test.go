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
		expected []interface{}
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
	>class="main" tag="test">test>$env:NAME2;
	>class="main" tag="test">test2>Hello, World!;
>>>END;`,
			expected: []interface{}{
				map[string]interface{}{
					"KEY":      "test",
					"class":    "main",
					"tag":      "test",
					"Content":  "developerrrr2",
					"Contents": "developerrrr2",
				},
				map[string]interface{}{
					"KEY":      "test2",
					"class":    "main",
					"tag":      "test",
					"Content":  "Hello, World!",
					"Contents": "Hello, World!",
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
			name: "environment variable replacement",
			input: `>>>DOCTYPE=JTL;
>>>ENV;
    >>>USER=james; >>>APP=myapp; >>>ENV=prod;
>>>BEGIN;
    >class="user" tag="span">userinfo>$env:USER;
    >class="env" tag="div">envinfo>$env:ENV;
>>>END;`,
			expected: []interface{}{
				map[string]interface{}{
					"KEY":      "userinfo",
					"class":    "user",
					"tag":      "span",
					"Content":  "james",
					"Contents": "james",
				},
				map[string]interface{}{
					"KEY":      "envinfo",
					"class":    "env",
					"tag":      "div",
					"Content":  "prod",
					"Contents": "prod",
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
    >class="test" tag="div">id>content;
>>>END;`,
			wantErr: true,
		},
		{
			name: "empty content",
			input: `>>>DOCTYPE=JTL;
>>>BEGIN;
    >class="test" tag="div">id>;
>>>END;`,
			expected: []interface{}{
				map[string]interface{}{
					"KEY":      "id",
					"class":    "test",
					"tag":      "div",
					"Content":  "",
					"Contents": "",
				},
			},
			wantErr: false,
		},
		{
			name: "bracketed content",
			input: `>>>DOCTYPE=JTL;
>>>BEGIN;
    >type="lua">script>[[
        document.onEvent(".buttontest", "click", [[
            print("Button clicked!")
            -- Do more stuff here
        ]];
>>>END;`,
			// Note: In the updated lib, the outer bracket markers ([[ and ]]) are preserved.
			expected: []interface{}{
				map[string]interface{}{
					"KEY":  "script",
					"type": "lua",
					"Content": `[[
        document.onEvent(".buttontest", "click", [[
            print("Button clicked!")
            -- Do more stuff here
        ]]`,
					"Contents": `[[
        document.onEvent(".buttontest", "click", [[
            print("Button clicked!")
            -- Do more stuff here
        ]]`,
				},
			},
			wantErr: false,
		},
		{
			name: "lua script with button",
			input: `>>>DOCTYPE=JTL

>>>ENV;
    >>>NAME=me
>>>version=1.0;

>>>BEGIN;
    >type="lua">script>
        document.onEvent(".buttontest", "click", [[
            print("Button clicked!")
            -- Do more stuff here
        ]]);
    >class="buttontest">button>Test Button;
>>>END;`,
			expected: []interface{}{
				map[string]interface{}{
					"KEY":  "script",
					"type": "lua",
					"Content": `document.onEvent(".buttontest", "click", [[
            print("Button clicked!")
            -- Do more stuff here
        ]])`,
					"Contents": `document.onEvent(".buttontest", "click", [[
            print("Button clicked!")
            -- Do more stuff here
        ]])`,
				},
				map[string]interface{}{
					"KEY":      "button",
					"class":    "buttontest",
					"Content":  "Test Button",
					"Contents": "Test Button",
				},
			},
			wantErr: false,
		},
		{
			name: "empty tag content",
			input: `>>>DOCTYPE=JTL;
>>>BEGIN;
    >src="test.lua">script>;
    >class="empty">div>;
>>>END;`,
			expected: []interface{}{
				map[string]interface{}{
					"KEY":      "script",
					"src":      "test.lua",
					"Content":  "",
					"Contents": "",
				},
				map[string]interface{}{
					"KEY":      "div",
					"class":    "empty",
					"Content":  "",
					"Contents": "",
				},
			},
			wantErr: false,
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
			if !reflect.DeepEqual(got, tt.expected) {
				gotJSON, _ := json.MarshalIndent(got, "", "  ")
				expectedJSON, _ := json.MarshalIndent(tt.expected, "", "  ")
				t.Errorf("ParseEnv() got = %s\nwant = %s", gotJSON, expectedJSON)
			}
		})
	}
}

func TestStringify(t *testing.T) {
	tests := []struct {
		name    string
		input   []interface{}
		wantErr bool
	}{
		{
			name: "simple vector",
			input: []interface{}{
				map[string]interface{}{
					"KEY":      "example",
					"Content":  "test content",
					"Contents": "test content",
				},
			},
			wantErr: false,
		},
		{
			name:    "empty vector",
			input:   []interface{}{},
			wantErr: false,
		},
		{
			name: "nested map",
			input: []interface{}{
				map[string]interface{}{
					"KEY": "parent",
					"nested": map[string]interface{}{
						"KEY":      "child",
						"Content":  "child content",
						"Contents": "child content",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jtl.Stringify(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stringify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var parsed []interface{}
			if err := json.Unmarshal([]byte(got), &parsed); err != nil {
				t.Errorf("Stringify() produced invalid JSON: %v", err)
			}

			if !reflect.DeepEqual(parsed, tt.input) {
				gotJSON, _ := json.MarshalIndent(parsed, "", "  ")
				expectedJSON, _ := json.MarshalIndent(tt.input, "", "  ")
				t.Errorf("Stringify() got = %s\nwant = %s", gotJSON, expectedJSON)
			}
		})
	}
}
