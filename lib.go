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

	// Process the document line by line.
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
			fullElement := line
			// If the current line does not end with ";" then it's a multi-line element.
			if !strings.HasSuffix(line, ";") {
				for j := i + 1; j < len(lines); j++ {
					nextLine := lines[j]
					fullElement += "\n" + nextLine
					if strings.HasSuffix(strings.TrimSpace(nextLine), ";") {
						i = j // move outer loop index past the complete element
						break
					}
				}
			}
			elementMap, err := parseElement(fullElement, currentEnv)
			if err != nil {
				return nil, err
			}
			result = append(result, elementMap)
		}
	}

	return result, nil
}

// parseElement parses a single JTL element by locating the first two ">" separators.
func parseElement(elementText string, env map[string]string) (interface{}, error) {
	// Remove the leading ">" and the trailing ";".
	elementText = strings.TrimPrefix(elementText, ">")
	elementText = strings.TrimSuffix(elementText, ";")

	// Find the first ">" separator that ends the attribute part.
	firstSep := strings.Index(elementText, ">")
	if firstSep == -1 {
		return nil, errors.New("invalid element format: missing first separator")
	}
	attributesPart := elementText[:firstSep]
	remainder := elementText[firstSep+1:]

	// Find the second ">" separator that separates the element ID from its content.
	secondSep := strings.Index(remainder, ">")
	if secondSep == -1 {
		return nil, errors.New("invalid element format: missing second separator")
	}
	id := strings.TrimSpace(remainder[:secondSep])
	content := remainder[secondSep+1:]

	// For non-bracketed content, trim surrounding whitespace.
	if !strings.HasPrefix(content, "[[") {
		content = strings.TrimSpace(content)
	}

	// Ensure neither the element id nor the content is empty.
	if id == "" || content == "" {
		return nil, errors.New("invalid element format: malformed content")
	}

	// Replace environment variable references.
	if strings.HasPrefix(content, "$env:") {
		// Trim any surrounding whitespace from the env key.
		envVar := strings.TrimSpace(strings.TrimPrefix(content, "$env:"))
		if val, ok := env[envVar]; ok {
			content = val
		}
	}

	// Process attributes using a regex.
	attrRegex := regexp.MustCompile(`(\w+)="([^"]+)"`)
	matches := attrRegex.FindAllStringSubmatch(attributesPart, -1)
	if len(matches) == 0 {
		return nil, errors.New("invalid element format: no attributes found")
	}
	elementMap := make(map[string]interface{})
	for _, match := range matches {
		elementMap[match[1]] = match[2]
	}

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
			break envParsing // exit once the body begins
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
