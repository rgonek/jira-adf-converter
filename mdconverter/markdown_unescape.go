package mdconverter

import "strings"

func unescapeMarkdownLiteralText(value string) string {
	if value == "" || !strings.Contains(value, "\\") {
		return value
	}

	var builder strings.Builder
	builder.Grow(len(value))

	runes := []rune(value)
	for i := 0; i < len(runes); i++ {
		if runes[i] != '\\' || i+1 >= len(runes) {
			builder.WriteRune(runes[i])
			continue
		}

		next := runes[i+1]
		switch next {
		case '\\', '*', '_', '[', ']', '(', ')', '`', '#', '>':
			builder.WriteRune(next)
			i++
		default:
			builder.WriteRune(runes[i])
		}
	}

	return builder.String()
}

func stackHasMarkType(stack *markStack, markType string) bool {
	if stack == nil {
		return false
	}
	for _, mark := range stack.items {
		if mark.Type == markType {
			return true
		}
	}
	return false
}
