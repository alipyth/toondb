package parser

import (
        "encoding/json"
        "fmt"
        "regexp"
        "strconv"
        "strings"
)

type Parser struct{}

func NewParser() *Parser {
        return &Parser{}
}

type ToonData struct {
        Fields map[string]interface{}
}

func (p *Parser) ParseToon(toon string) (*ToonData, error) {
        data := &ToonData{
                Fields: make(map[string]interface{}),
        }

        lines := strings.Split(toon, "\n")
        var currentObject map[string]interface{}
        var currentArray []interface{}
        var inArray bool
        var inObject bool

        for _, line := range lines {
                line = strings.TrimSpace(line)
                if line == "" || strings.HasPrefix(line, "#") {
                        continue
                }

                // Handle array syntax: key[n]: value1,value2,value3
                if arrayMatch := regexp.MustCompile(`^([^\[]+)\[(\d+)\]:\s*(.+)$`).FindStringSubmatch(line); arrayMatch != nil {
                        key := arrayMatch[1]
                        size, _ := strconv.Atoi(arrayMatch[2])
                        values := strings.Split(arrayMatch[3], ",")
                        
                        // Trim whitespace from each value
                        for i, val := range values {
                                values[i] = strings.TrimSpace(val)
                        }
                        
                        // Ensure we don't exceed the specified size
                        if len(values) > size {
                                values = values[:size]
                        }
                        
                        data.Fields[key] = values
                        continue
                }

                // Handle object array syntax: key[n]{fields}: data
                if objectArrayMatch := regexp.MustCompile(`^([^\[]+)\[(\d+)\]\{([^\}]+)\}:\s*(.+)$`).FindStringSubmatch(line); objectArrayMatch != nil {
                        key := objectArrayMatch[1]
                        size, _ := strconv.Atoi(objectArrayMatch[2])
                        fields := strings.Split(objectArrayMatch[3], ",")
                        dataLines := strings.Split(objectArrayMatch[4], "\n")
                        
                        objects := make([]map[string]interface{}, 0)
                        
                        for _, dataLine := range dataLines {
                                dataLine = strings.TrimSpace(dataLine)
                                if dataLine == "" {
                                        continue
                                }
                                
                                values := strings.Split(dataLine, ",")
                                if len(values) == len(fields) {
                                        obj := make(map[string]interface{})
                                        for i, field := range fields {
                                                obj[strings.TrimSpace(field)] = strings.TrimSpace(values[i])
                                        }
                                        objects = append(objects, obj)
                                }
                        }
                        
                        // Ensure we don't exceed the specified size
                        if len(objects) > size {
                                objects = objects[:size]
                        }
                        
                        data.Fields[key] = objects
                        continue
                }

                // Handle nested objects (indented lines)
                if strings.HasPrefix(line, "  ") && (inObject || inArray) {
                        nestedLine := strings.TrimSpace(line)
                        if strings.Contains(nestedLine, ":") {
                                parts := strings.SplitN(nestedLine, ":", 2)
                                if len(parts) == 2 {
                                        nestedKey := strings.TrimSpace(parts[0])
                                        nestedValue := strings.TrimSpace(parts[1])
                                        
                                        if inObject {
                                                currentObject[nestedKey] = nestedValue
                                        } else if inArray {
                                                // For arrays, we'd need more complex handling
                                                // For now, we'll treat it as a simple value
                                                if currentArray != nil {
                                                        currentArray = append(currentArray, nestedValue)
                                                }
                                        }
                                }
                        }
                        continue
                }

                // Reset nested context
                inObject = false
                inArray = false
                currentObject = nil
                currentArray = nil

                // Handle simple key-value pairs
                if strings.Contains(line, ":") {
                        parts := strings.SplitN(line, ":", 2)
                        if len(parts) == 2 {
                                key := strings.TrimSpace(parts[0])
                                value := strings.TrimSpace(parts[1])
                                
                                // Check if the value indicates a nested object
                                if value == "" {
                                        currentObject = make(map[string]interface{})
                                        data.Fields[key] = currentObject
                                        inObject = true
                                } else {
                                        data.Fields[key] = value
                                }
                        }
                }
        }

        return data, nil
}

func (p *Parser) ToonToJSON(toon string) (string, error) {
        data, err := p.ParseToon(toon)
        if err != nil {
                return "", err
        }

        jsonData, err := json.MarshalIndent(data.Fields, "", "  ")
        if err != nil {
                return "", err
        }

        return string(jsonData), nil
}

func (p *Parser) JSONToTOON(jsonStr string) (string, error) {
        var data map[string]interface{}
        err := json.Unmarshal([]byte(jsonStr), &data)
        if err != nil {
                return "", err
        }

        return p.mapToTOON(data, 0), nil
}

func (p *Parser) mapToTOON(data interface{}, indent int) string {
        var result strings.Builder
        indentStr := strings.Repeat("  ", indent)

        switch v := data.(type) {
        case map[string]interface{}:
                for key, value := range v {
                        result.WriteString(indentStr)
                        result.WriteString(key)
                        result.WriteString(": ")
                        
                        switch val := value.(type) {
                        case string:
                                result.WriteString(val)
                        case map[string]interface{}:
                                result.WriteString("\n")
                                result.WriteString(p.mapToTOON(val, indent+1))
                                continue
                        case []interface{}:
                                result.WriteString(p.arrayToTOON(val, key, indent))
                                continue
                        default:
                                result.WriteString(fmt.Sprintf("%v", val))
                        }
                        
                        result.WriteString("\n")
                }
        case []interface{}:
                result.WriteString(p.arrayToTOON(v, "", indent))
        }

        return result.String()
}

func (p *Parser) arrayToTOON(array []interface{}, key string, indent int) string {
        if len(array) == 0 {
                return ""
        }

        // Check if it's an array of objects (table format)
        if _, ok := array[0].(map[string]interface{}); ok && key != "" {
                var result strings.Builder
                result.WriteString(fmt.Sprintf("%s[%d]{", key, len(array)))
                
                // Get all unique field names from all objects
                fieldSet := make(map[string]bool)
                for _, item := range array {
                        if obj, ok := item.(map[string]interface{}); ok {
                                for k := range obj {
                                        fieldSet[k] = true
                                }
                        }
                }
                
                // Convert field set to sorted slice
                fields := make([]string, 0, len(fieldSet))
                for field := range fieldSet {
                        fields = append(fields, field)
                }
                
                result.WriteString(strings.Join(fields, ","))
                result.WriteString("}:\n")
                
                // Write each object as a comma-separated line
                for _, item := range array {
                        if obj, ok := item.(map[string]interface{}); ok {
                                var values []string
                                for _, field := range fields {
                                        if val, exists := obj[field]; exists {
                                                values = append(values, fmt.Sprintf("%v", val))
                                        } else {
                                                values = append(values, "")
                                        }
                                }
                                result.WriteString(strings.Repeat("  ", indent+1))
                                result.WriteString(strings.Join(values, ","))
                                result.WriteString("\n")
                        }
                }
                
                return result.String()
        }
        
        // Simple array format
        if key != "" {
                return fmt.Sprintf("%s[%d]: %s", key, len(array), p.interfaceArrayToString(array))
        }
        return p.interfaceArrayToString(array)
}

func (p *Parser) interfaceArrayToString(array []interface{}) string {
        var values []string
        for _, item := range array {
                values = append(values, fmt.Sprintf("%v", item))
        }
        return strings.Join(values, ",")
}