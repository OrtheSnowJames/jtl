// SPDX-License-Identifier: MIT
package jtl

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

// Parse parses JTL content into a structured slice of interfaces.
func Parse(text string) ([]interface{}, error) {
	var result []interface{}
	lines := strings.Split(text, "\n")

	if len(lines) == 0 || !strings.Contains(lines[0], "DOCTYPE=JTL") {
		return nil, errors.New("invalid JTL document: missing DOCTYPE")
	}

	inBody := false
	inEnv := false
	currentEnv := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*/") || strings.HasPrefix(line, ">//>") {
			continue
		}

		switch line {
		case ">>>ENV;":
			inEnv = true
			continue
		case ">>>BEGIN;":
			inEnv = false
			inBody = true
			continue
		case ">>>END;":
			inBody = false
			continue
		}

		// Handle environment variables
		if inEnv && strings.HasPrefix(line, ">>>") {
			declarations := strings.Split(line, ";")
			for _, declaration := range declarations {
				declaration = strings.TrimSpace(declaration)
				if strings.HasPrefix(declaration, ">>>") {
					parts := strings.SplitN(declaration[3:], "=", 2)
					if len(parts) == 2 {
						varName := strings.TrimSpace(parts[0])
						varValue := strings.TrimSpace(parts[1])
						currentEnv[varName] = varValue
					}
				}
			}
		}

		// Handle body elements
		if inBody && strings.HasPrefix(line, ">") {
			elementMap, err := parseElement(line, currentEnv)
			if err != nil {
				return nil, err
			}
			result = append(result, elementMap)
		}
	}

	return result, nil
}

// parseElement parses a single JTL element.
func parseElement(line string, env map[string]string) (interface{}, error) {
	line = strings.TrimPrefix(line, ">")

	if !strings.Contains(line, ">") {
		return nil, errors.New("invalid element format: missing separator")
	}

	attrRegex := regexp.MustCompile(`(\w+)="([^"]+)"`)
	matches := attrRegex.FindAllStringSubmatch(line, -1)
	if len(matches) == 0 {
		return nil, errors.New("invalid element format: no attributes found")
	}

	elementMap := make(map[string]interface{})
	for _, match := range matches {
		elementMap[match[1]] = match[2]
	}

	// Extract content and ID
	contentStart := strings.Index(line, ">")
	if contentStart == -1 {
		return nil, errors.New("invalid element format: missing content separator")
	}

	contentPart := line[contentStart+1:]
	contentPart = strings.TrimSuffix(contentPart, ";")

	parts := strings.SplitN(contentPart, ">", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, errors.New("invalid element format: malformed content")
	}

	id := parts[0]
	content := parts[1]

	// Handle environment variable replacement
	if strings.HasPrefix(content, "$env:") {
		envVar := strings.TrimPrefix(content, "$env:")
		if val, ok := env[envVar]; ok {
			content = val
		}
	}

	// Add new fields as per the updated Rust code
	elementMap["KEY"] = id
	elementMap["Content"] = content
	elementMap["Contents"] = content

	return elementMap, nil
}

// Stringify converts a slice of maps to a JSON string.
func Stringify(data []interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}


// ParseEnv extracts environment variables from JTL text.
func ParseEnv(text string) (map[string]interface{}, error) {
	envMap := make(map[string]interface{})
	lines := strings.Split(text, "\n")

	if len(lines) == 0 || !strings.Contains(lines[0], "DOCTYPE=JTL") {
		return nil, errors.New("invalid JTL document: missing DOCTYPE")
	}

	inEnv := false

	envParsing:
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*/") || strings.HasPrefix(line, ">//>") {
			continue
		}

		switch line {
		case ">>>ENV;":
			inEnv = true
			continue
		case ">>>BEGIN;":
			break envParsing // Properly exits the outer loop
		}

		if inEnv && strings.HasPrefix(line, ">>>") {
			declarations := strings.Split(line, ";")
			for _, declaration := range declarations {
				declaration = strings.TrimSpace(declaration)
				if strings.HasPrefix(declaration, ">>>") {
					parts := strings.SplitN(declaration[3:], "=", 2)
					if len(parts) == 2 {
						varName := strings.TrimSpace(parts[0])
						varValue := strings.TrimSpace(parts[1])
						envMap[varName] = varValue
					}
				}
			}
		}
	}

	return envMap, nil
}
