package mdconverter

import "strings"

func readPandocAttrBlock(line []byte, start int) (string, int, bool) {
	if start < 0 || start >= len(line) || line[start] != '{' {
		return "", 0, false
	}

	var quote byte
	escaped := false
	for idx := start + 1; idx < len(line); idx++ {
		ch := line[idx]
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == quote {
				quote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			quote = ch
			continue
		}
		if ch == '}' {
			return string(line[start+1 : idx]), idx + 1, true
		}
		if ch == '\n' || ch == '\r' {
			return "", 0, false
		}
	}

	return "", 0, false
}

func parsePandocAttributes(raw string) ([]string, map[string]string) {
	classes := make([]string, 0, 2)
	attrs := make(map[string]string, 2)

	for idx := 0; idx < len(raw); {
		for idx < len(raw) && isPandocAttrSpace(raw[idx]) {
			idx++
		}
		if idx >= len(raw) {
			break
		}

		if raw[idx] == '.' {
			idx++
			start := idx
			for idx < len(raw) && !isPandocAttrSpace(raw[idx]) {
				idx++
			}
			className := strings.TrimSpace(raw[start:idx])
			if className != "" {
				classes = append(classes, className)
			}
			continue
		}

		keyStart := idx
		for idx < len(raw) && !isPandocAttrSpace(raw[idx]) && raw[idx] != '=' {
			idx++
		}
		key := strings.TrimSpace(raw[keyStart:idx])
		if key == "" {
			idx++
			continue
		}

		for idx < len(raw) && isPandocAttrSpace(raw[idx]) {
			idx++
		}
		if idx >= len(raw) || raw[idx] != '=' {
			for idx < len(raw) && !isPandocAttrSpace(raw[idx]) {
				idx++
			}
			continue
		}

		idx++
		for idx < len(raw) && isPandocAttrSpace(raw[idx]) {
			idx++
		}
		if idx >= len(raw) {
			attrs[key] = ""
			break
		}

		if raw[idx] == '"' || raw[idx] == '\'' {
			quote := raw[idx]
			idx++
			var value strings.Builder
			for idx < len(raw) {
				ch := raw[idx]
				if ch == '\\' && idx+1 < len(raw) {
					next := raw[idx+1]
					if next == quote || next == '\\' {
						value.WriteByte(next)
						idx += 2
						continue
					}
				}
				if ch == quote {
					idx++
					break
				}
				value.WriteByte(ch)
				idx++
			}
			attrs[key] = value.String()
			continue
		}

		valueStart := idx
		for idx < len(raw) && !isPandocAttrSpace(raw[idx]) {
			idx++
		}
		attrs[key] = raw[valueStart:idx]
	}

	return classes, attrs
}

func isPandocAttrSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func extractTextAlign(style string) string {
	parts := strings.Split(style, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "text-align:") {
			val := strings.TrimSpace(part[len("text-align:"):])
			switch strings.ToLower(val) {
			case "left", "center", "right":
				return strings.ToLower(val)
			}
		}
	}
	return ""
}

func extractStyleColor(style, property string) string {
	parts := strings.Split(style, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), strings.ToLower(property)+":") {
			return strings.TrimSpace(part[len(property)+1:])
		}
	}
	return ""
}
