// SPDX-License-Identifier: MIT
package jtl

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

// Parse parses JTL content into a structured map.
func Parse(text string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	lines := strings.Split(text, "\n")

	if !strings.Contains(lines[0], "DOCTYPE=JTL") {
		return nil, errors.New("invalid JTL document: missing DOCTYPE")
	}

	inBody := false
	inEnv := false
	currentEnv := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*/") || strings.HasPrefix(line, ">//>") {
			continue // Skip empty lines, block comments, and // comments.
		}

		// Handle sections
		if line == ">>>ENV;" {
			inEnv = true
			continue
		}
		if line == ">>>BEGIN;" {
			inEnv = false
			inBody = true
			continue
		}
		if line == ">>>END;" {
			inBody = false
			continue
		}

		// Handle multiple declarations per line
		declarations := strings.Split(line, ";")
		for _, declaration := range declarations {
			declaration = strings.TrimSpace(declaration)
			if declaration == "" || strings.HasPrefix(declaration, ">//>") {
				continue
			}

			if inEnv && strings.HasPrefix(declaration, ">>>") {
				parts := strings.SplitN(declaration[3:], "=", 2)
				if len(parts) == 2 {
					varName := strings.TrimSpace(parts[0])
					varValue := strings.TrimSpace(parts[1])
					currentEnv[varName] = varValue
				}
			} else if inBody && strings.HasPrefix(declaration, ">") {
				// Validate minimum element length
				if len(declaration) < 5 { // ">a>b" is minimum valid length
					return nil, errors.New("invalid element format: too short")
				}

				elementMap, id, err := parseElement(declaration, currentEnv)
				if err != nil {
					return nil, err
				}
				result[id] = elementMap
			}
		}
	}

	return result, nil
}

// Stringify converts a map to a JSON string.
func Stringify(data map[string]interface{}) (string, error) {
	b, err := json.Marshal(data) // Remove Indent to get compact JSON
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// parseElement parses a single JTL element.
func parseElement(line string, env map[string]string) (map[string]interface{}, string, error) {
	line = strings.TrimPrefix(line, ">")

	// Validate basic format
	if !strings.Contains(line, ">") {
		return nil, "", errors.New("invalid element format: missing separator")
	}

	// Find attribute pairs
	attrRegex := regexp.MustCompile(`(\w+)="([^"]+)"`)
	matches := attrRegex.FindAllStringSubmatch(line, -1)

	if len(matches) == 0 {
		return nil, "", errors.New("invalid element format: no attributes found")
	}

	elementMap := make(map[string]interface{})
	for _, match := range matches {
		elementMap[match[1]] = match[2]
	}

	// Extract content and ID
	contentStart := strings.Index(line, ">")
	if contentStart == -1 {
		return nil, "", errors.New("invalid element format: missing content separator")
	}

	contentPart := line[contentStart+1:]
	contentPart = strings.TrimSuffix(contentPart, ";")

	parts := strings.SplitN(contentPart, ">", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, "", errors.New("invalid element format: malformed content")
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

	elementMap["content"] = content

	return elementMap, id, nil
}

// ParseEnv extracts environment variables from JTL text.
func ParseEnv(text string) (map[string]interface{}, error) {
	envMap := make(map[string]interface{})
	lines := strings.Split(text, "\n")

	if !strings.Contains(lines[0], "DOCTYPE=JTL") {
		return nil, errors.New("invalid JTL document: missing DOCTYPE")
	}

	inEnv := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*/") || strings.HasPrefix(line, ">//>") {
			continue
		}

		if line == ">>>ENV;" {
			inEnv = true
			continue
		}
		if line == ">>>BEGIN;" {
			break
		}

		if inEnv && strings.HasPrefix(line, ">>>") {
			// Handle multiple declarations in one line
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
