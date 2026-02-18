package mdconverter

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rgonek/jira-adf-converter/converter"
)

var (
	emojiShortcodeRe   = regexp.MustCompile(`:[A-Za-z0-9_+\-]+:`)
	statusBracketRe    = regexp.MustCompile(`\[Status:\s*([^\]]+)\]`)
	dateISORe          = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`)
	mediaPlaceholderRe = regexp.MustCompile(`\[(Image|File):\s*([^\]]+)\]`)
)

type patternMatch struct {
	kind  string
	start int
	end   int
	value string
	extra string
}

func (s *state) shouldDetectMentionLink() bool {
	return s.config.MentionDetection == MentionDetectLink || s.config.MentionDetection == MentionDetectAll
}

func (s *state) shouldDetectMentionAt() bool {
	return s.config.MentionDetection == MentionDetectAt || s.config.MentionDetection == MentionDetectAll
}

func (s *state) shouldDetectMentionHTML() bool {
	return s.config.MentionDetection == MentionDetectHTML || s.config.MentionDetection == MentionDetectAll
}

func (s *state) shouldDetectEmoji() bool {
	return s.config.EmojiDetection == EmojiDetectShortcode || s.config.EmojiDetection == EmojiDetectAll
}

func (s *state) shouldDetectStatus() bool {
	return s.config.StatusDetection == StatusDetectBracket || s.config.StatusDetection == StatusDetectAll
}

func (s *state) shouldDetectDate() bool {
	return s.config.DateDetection == DateDetectISO || s.config.DateDetection == DateDetectAll
}

func (s *state) expandTextPatterns(textValue string, marks []converter.Mark) []converter.Node {
	if textValue == "" {
		return nil
	}
	if len(marks) > 0 {
		return []converter.Node{newTextNode(textValue, marks)}
	}

	var content []converter.Node
	remaining := textValue

	for remaining != "" {
		match, ok := s.findNextPattern(remaining)
		if !ok {
			content = appendInlineNode(content, newTextNode(remaining, nil))
			break
		}

		if match.start > 0 {
			content = appendInlineNode(content, newTextNode(remaining[:match.start], nil))
		}

		switch match.kind {
		case "emoji":
			content = append(content, converter.Node{
				Type: "emoji",
				Attrs: map[string]interface{}{
					"shortName": match.value,
				},
			})

		case "status":
			content = append(content, converter.Node{
				Type: "status",
				Attrs: map[string]interface{}{
					"text": strings.TrimSpace(match.value),
				},
			})

		case "date":
			layout := s.config.DateFormat
			if strings.TrimSpace(layout) == "" {
				layout = "2006-01-02"
			}
			parsedDate, err := time.Parse(layout, match.value)
			if err != nil {
				parsedDate, err = time.Parse("2006-01-02", match.value)
			}
			if err != nil {
				content = appendInlineNode(content, newTextNode(match.value, nil))
				break
			}
			content = append(content, converter.Node{
				Type: "date",
				Attrs: map[string]interface{}{
					"timestamp": strconv.FormatInt(parsedDate.Unix(), 10),
				},
			})

		case "media":
			mediaType := strings.ToLower(strings.TrimSpace(match.value))
			id := strings.TrimSpace(match.extra)
			if id == "" {
				content = appendInlineNode(content, newTextNode(remaining[match.start:match.end], nil))
				break
			}
			content = append(content, converter.Node{
				Type: "mediaSingle",
				Content: []converter.Node{
					{
						Type: "media",
						Attrs: map[string]interface{}{
							"type": mediaType,
							"id":   id,
						},
					},
				},
			})

		case "mentionAt":
			content = append(content, converter.Node{
				Type: "mention",
				Attrs: map[string]interface{}{
					"id":   match.extra,
					"text": match.value,
				},
			})
		}

		remaining = remaining[match.end:]
	}

	return content
}

func (s *state) findNextPattern(textValue string) (patternMatch, bool) {
	candidates := make([]patternMatch, 0, 5)

	if s.shouldDetectEmoji() {
		if loc := emojiShortcodeRe.FindStringIndex(textValue); loc != nil {
			candidates = append(candidates, patternMatch{
				kind:  "emoji",
				start: loc[0],
				end:   loc[1],
				value: textValue[loc[0]:loc[1]],
			})
		}
	}

	if s.shouldDetectStatus() {
		if match := statusBracketRe.FindStringSubmatchIndex(textValue); match != nil {
			candidates = append(candidates, patternMatch{
				kind:  "status",
				start: match[0],
				end:   match[1],
				value: textValue[match[2]:match[3]],
			})
		}
	}

	if s.shouldDetectDate() {
		if loc := dateISORe.FindStringIndex(textValue); loc != nil {
			candidates = append(candidates, patternMatch{
				kind:  "date",
				start: loc[0],
				end:   loc[1],
				value: textValue[loc[0]:loc[1]],
			})
		}
	}

	if match := mediaPlaceholderRe.FindStringSubmatchIndex(textValue); match != nil {
		candidates = append(candidates, patternMatch{
			kind:  "media",
			start: match[0],
			end:   match[1],
			value: textValue[match[2]:match[3]],
			extra: textValue[match[4]:match[5]],
		})
	}

	if s.shouldDetectMentionAt() {
		if mention, ok := s.findMentionRegistryMatch(textValue); ok {
			candidates = append(candidates, mention)
		}
	}

	if len(candidates) == 0 {
		return patternMatch{}, false
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].start != candidates[j].start {
			return candidates[i].start < candidates[j].start
		}
		return candidates[i].end > candidates[j].end
	})

	return candidates[0], true
}

func (s *state) findMentionRegistryMatch(textValue string) (patternMatch, bool) {
	if len(s.config.MentionRegistry) == 0 {
		return patternMatch{}, false
	}

	type mentionCandidate struct {
		name string
		id   string
	}

	candidates := make([]mentionCandidate, 0, len(s.config.MentionRegistry))
	for name, id := range s.config.MentionRegistry {
		cleanName := strings.TrimSpace(name)
		cleanID := strings.TrimSpace(id)
		if cleanName == "" || cleanID == "" {
			continue
		}
		candidates = append(candidates, mentionCandidate{name: cleanName, id: cleanID})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return len(candidates[i].name) > len(candidates[j].name)
	})

	best := patternMatch{start: len(textValue) + 1}
	found := false

	for _, candidate := range candidates {
		token := "@" + candidate.name
		start := strings.Index(textValue, token)
		if start < 0 {
			continue
		}

		end := start + len(token)
		if !isMentionBoundary(textValue, start, end) {
			continue
		}

		if !found || start < best.start || (start == best.start && end > best.end) {
			best = patternMatch{
				kind:  "mentionAt",
				start: start,
				end:   end,
				value: candidate.name,
				extra: candidate.id,
			}
			found = true
		}
	}

	return best, found
}

func isMentionBoundary(textValue string, start, end int) bool {
	if start > 0 {
		prev := textValue[start-1]
		if !isBoundaryChar(prev) {
			return false
		}
	}
	if end < len(textValue) {
		next := textValue[end]
		if !isBoundaryChar(next) {
			return false
		}
	}
	return true
}

func isBoundaryChar(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r', '.', ',', '!', '?', ':', ';', ')', ']', '}', '(', '[', '{', '"', '\'':
		return true
	default:
		return false
	}
}
