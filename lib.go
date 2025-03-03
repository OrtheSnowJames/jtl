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

	type stackItem struct {
		element interface{}
		indent  int
	}
	stack := []stackItem{}

	for i := 0; i < len(lines); i++ {
		indent := countIndentation(lines[i])
		line := strings.TrimSpace(lines[i])

		if line == "" || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*/") || strings.HasPrefix(line, ">//>") {
			continue
		}

		// Handle section markers
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
			continue
		}

		// Handle body elements
		if inBody && strings.HasPrefix(line, ">") {
			// Collect multi-line content
			fullContent := line
			if !strings.HasSuffix(line, ";") {
				for j := i + 1; j < len(lines); j++ {
					nextLine := lines[j]
					fullContent += "\n" + nextLine
					if strings.HasSuffix(strings.TrimSpace(nextLine), ";") {
						i = j // Skip processed lines
						break
					}
				}
			}

			element, err := parseElement(fullContent, currentEnv)
			if err != nil {
				return nil, err
			}

			// Pop stack items with greater or equal indentation
			for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
				stack = stack[:len(stack)-1]
			}

			if len(stack) > 0 {
				// Add as child to parent
				parent := stack[len(stack)-1].element.(map[string]interface{})
				if _, exists := parent["children"]; !exists {
					parent["children"] = make([]interface{}, 0)
				}
				parent["children"] = append(parent["children"].([]interface{}), element)
			} else {
				// Root level element
				result = append(result, element)
			}

			// Push current element to stack
			stack = append(stack, stackItem{element, indent})
		}
	}

	return result, nil
}

func countIndentation(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

// parseElement parses a single JTL element by locating the first two ">" separators.
func parseElement(elementText string, env map[string]string) (interface{}, error) {
	// Remove the leading ">" and the trailing ";"
	elementText = strings.TrimPrefix(elementText, ">")
	elementText = strings.TrimSuffix(strings.TrimSpace(elementText), ";")

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

	// Handle multi-line and bracketed content
	if strings.Contains(content, "\n") {
		lines := strings.Split(content, "\n")
		// Find minimum indentation level
		minIndent := -1
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			indent := len(line) - len(strings.TrimLeft(line, " \t"))
			if minIndent == -1 || indent < minIndent {
				minIndent = indent
			}
		}

		// Remove common indentation and trim each line
		var processedLines []string
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			if minIndent > 0 && len(line) >= minIndent {
				line = line[minIndent:]
			}
			processedLines = append(processedLines, strings.TrimRight(line, " \t"))
		}
		content = strings.Join(processedLines, "\n")
		content = strings.TrimSpace(content)
	} else {
		content = strings.TrimSpace(content)
	}

	// Ensure the element id is not empty
	if id == "" {
		return nil, errors.New("invalid element format: empty element id")
	}

	// Replace environment variable references only if content is not empty
	if content != "" && strings.HasPrefix(content, "$env:") {
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
