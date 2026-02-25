package converter

import "strings"

func escapeMarkdownTextLiteral(value string) string {
	if value == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(value))

	lineStart := true
	for _, r := range value {
		switch r {
		case '\n', '\r':
			builder.WriteRune(r)
			lineStart = true
			continue
		}

		if lineStart && (r == ' ' || r == '\t') {
			builder.WriteRune(r)
			continue
		}

		if lineStart && (r == '#' || r == '>') {
			builder.WriteByte('\\')
			builder.WriteRune(r)
			lineStart = false
			continue
		}

		lineStart = false

		switch r {
		case '\\', '*', '_', '[', ']', '(', ')', '`':
			builder.WriteByte('\\')
			builder.WriteRune(r)
		default:
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func escapeMarkdownLinkDestination(destination string) string {
	destination = strings.TrimSpace(destination)
	if destination == "" {
		return ""
	}

	escaped := strings.ReplaceAll(destination, "\\", "\\\\")
	if strings.ContainsAny(escaped, " \t\n\r") {
		escaped = strings.ReplaceAll(escaped, "<", "\\<")
		escaped = strings.ReplaceAll(escaped, ">", "\\>")
		return "<" + escaped + ">"
	}

	escaped = strings.ReplaceAll(escaped, "(", "\\(")
	escaped = strings.ReplaceAll(escaped, ")", "\\)")
	return escaped
}

func escapeMarkdownLinkTitle(title string) string {
	title = strings.ReplaceAll(title, "\\", "\\\\")
	title = strings.ReplaceAll(title, "\r", " ")
	title = strings.ReplaceAll(title, "\n", " ")
	return strings.ReplaceAll(title, "\"", "\\\"")
}

func hasMarkType(marks []Mark, markType string) bool {
	for _, mark := range marks {
		if mark.Type == markType {
			return true
		}
	}
	return false
}
