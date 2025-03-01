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

	// Use a traditional for-loop so we can advance the index when needed.
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
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

		// Process environment declarations.
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

		// Process body elements.
		if inBody && strings.HasPrefix(line, ">") {
			// If the line contains the start of a bracketed block but not its closing,
			// accumulate subsequent lines until we find the closing "]]".
			if strings.Contains(line, "[[") && !strings.Contains(line, "]]") {
				elementLines := []string{line}
				for j := i + 1; j < len(lines); j++ {
					nextLine := lines[j]
					elementLines = append(elementLines, nextLine)
					if strings.Contains(nextLine, "]]") {
						i = j // advance outer loop index past the element
						break
					}
				}
				fullElementText := strings.Join(elementLines, "\n")
				elementMap, err := parseElement(fullElementText, currentEnv)
				if err != nil {
					return nil, err
				}
				result = append(result, elementMap)
			} else {
				elementMap, err := parseElement(line, currentEnv)
				if err != nil {
					return nil, err
				}
				result = append(result, elementMap)
			}
		}
	}

	return result, nil
}

// parseElement parses a single JTL element.
func parseElement(line string, env map[string]string) (interface{}, error) {
	// Remove the leading ">".
	line = strings.TrimPrefix(line, ">")

	// Find the first ">" which separates the attributes from the rest.
	contentStart := strings.Index(line, ">")
	if contentStart == -1 {
		return nil, errors.New("invalid element format: missing separator")
	}

	// Parse attributes using a regex.
	attrRegex := regexp.MustCompile(`(\w+)="([^"]+)"`)
	matches := attrRegex.FindAllStringSubmatch(line, -1)
	if len(matches) == 0 {
		return nil, errors.New("invalid element format: no attributes found")
	}

	elementMap := make(map[string]interface{})
	for _, match := range matches {
		elementMap[match[1]] = match[2]
	}

	// Extract the remainder after the first ">".
	contentPart := line[contentStart+1:]
	// Remove the trailing semicolon if present.
	contentPart = strings.TrimSuffix(contentPart, ";")

	// Split into an element ID and its content.
	parts := strings.SplitN(contentPart, ">", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, errors.New("invalid element format: malformed content")
	}

	id := parts[0]
	content := parts[1]

	// Do not trim any brackets. All brackets are preserved.
	// Replace environment variable references.
	if strings.HasPrefix(content, "$env:") {
		envVar := strings.TrimPrefix(content, "$env:")
		if val, ok := env[envVar]; ok {
			content = val
		}
	}

	// Add the parsed fields to the element map.
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
			break envParsing // Exit the loop once the body begins.
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
