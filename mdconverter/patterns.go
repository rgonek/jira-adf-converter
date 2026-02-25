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
	mediaPlaceholderRe = regexp.MustCompile(`\[(Image|File):\s*([^\]]+)\]`)
	dateLayoutTokens   = []dateLayoutToken{
		{token: "January", pattern: `(?i:January|February|March|April|May|June|July|August|September|October|November|December)`},
		{token: "Monday", pattern: `(?i:Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday)`},
		{token: "-07:00", pattern: `[+-][0-9]{2}:[0-9]{2}`},
		{token: "Z07:00", pattern: `(?:Z|[+-][0-9]{2}:[0-9]{2})`},
		{token: "-0700", pattern: `[+-][0-9]{4}`},
		{token: "Z0700", pattern: `(?:Z|[+-][0-9]{4})`},
		{token: "2006", pattern: `[0-9]{4}`},
		{token: "Mon", pattern: `(?i:Mon|Tue|Wed|Thu|Fri|Sat|Sun)`},
		{token: "Jan", pattern: `(?i:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)`},
		{token: "_2", pattern: ` ?(?:[1-9]|[12][0-9]|3[01])`},
		{token: "Z07", pattern: `(?:Z|[+-][0-9]{2})`},
		{token: "15", pattern: `(?:[01][0-9]|2[0-3])`},
		{token: "05", pattern: `[0-5][0-9]`},
		{token: "04", pattern: `[0-5][0-9]`},
		{token: "03", pattern: `(?:0[1-9]|1[0-2])`},
		{token: "02", pattern: `(?:0[1-9]|[12][0-9]|3[01])`},
		{token: "01", pattern: `(?:0[1-9]|1[0-2])`},
		{token: "06", pattern: `[0-9]{2}`},
		{token: "pm", pattern: `(?:am|pm)`},
		{token: "PM", pattern: `(?:AM|PM)`},
		{token: "MST", pattern: `[A-Za-z]{2,5}`},
		{token: "5", pattern: `(?:[0-9]|[1-5][0-9])`},
		{token: "4", pattern: `(?:[0-9]|[1-5][0-9])`},
		{token: "3", pattern: `(?:[1-9]|1[0-2])`},
		{token: "2", pattern: `(?:[1-9]|[12][0-9]|3[01])`},
		{token: "1", pattern: `(?:[1-9]|1[0-2])`},
	}
)

type patternMatch struct {
	kind  string
	start int
	end   int
	value string
	extra string
}

type dateLayoutToken struct {
	token   string
	pattern string
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

func (s *state) shouldDetectMentionPandoc() bool {
	return s.config.MentionDetection == MentionDetectPandoc || s.config.MentionDetection == MentionDetectAll
}

func (s *state) shouldDetectUnderlineHTML() bool {
	return s.config.UnderlineDetection == UnderlineDetectHTML || s.config.UnderlineDetection == UnderlineDetectAll
}

func (s *state) shouldDetectUnderlinePandoc() bool {
	return s.config.UnderlineDetection == UnderlineDetectPandoc || s.config.UnderlineDetection == UnderlineDetectAll
}

func (s *state) shouldDetectSubSupHTML() bool {
	return s.config.SubSupDetection == SubSupDetectHTML || s.config.SubSupDetection == SubSupDetectAll
}

func (s *state) shouldDetectSubSupPandoc() bool {
	return s.config.SubSupDetection == SubSupDetectPandoc || s.config.SubSupDetection == SubSupDetectAll
}

func (s *state) shouldDetectColorHTML() bool {
	return s.config.ColorDetection == ColorDetectHTML || s.config.ColorDetection == ColorDetectAll
}

func (s *state) shouldDetectColorPandoc() bool {
	return s.config.ColorDetection == ColorDetectPandoc || s.config.ColorDetection == ColorDetectAll
}

func (s *state) shouldDetectAlignHTML() bool {
	return s.config.AlignmentDetection == AlignDetectHTML || s.config.AlignmentDetection == AlignDetectAll
}

func (s *state) shouldDetectAlignPandoc() bool {
	return s.config.AlignmentDetection == AlignDetectPandoc || s.config.AlignmentDetection == AlignDetectAll
}

func (s *state) shouldDetectExpandPandoc() bool {
	return s.config.ExpandDetection == ExpandDetectPandoc || s.config.ExpandDetection == ExpandDetectAll
}

func (s *state) shouldDetectLayoutSectionHTML() bool {
	return s.config.LayoutSectionDetection == LayoutSectionDetectHTML || s.config.LayoutSectionDetection == LayoutSectionDetectAll
}

func (s *state) shouldDetectLayoutSectionPandoc() bool {
	return s.config.LayoutSectionDetection == LayoutSectionDetectPandoc || s.config.LayoutSectionDetection == LayoutSectionDetectAll
}

func (s *state) shouldDetectInlineCardPandoc() bool {
	return s.config.InlineCardDetection == InlineCardDetectPandoc || s.config.InlineCardDetection == InlineCardDetectAll
}

func (s *state) shouldDetectAnnotationPandoc() bool {
	return s.config.AnnotationDetection == AnnotationDetectPandoc
}

func (s *state) shouldDetectMediaInlinePandoc() bool {
	return s.config.MediaInlineDetection == MediaInlineDetectPandoc
}

func (s *state) shouldDetectBlockCardPandoc() bool {
	return s.config.BlockCardDetection == BlockCardDetectPandoc
}

func (s *state) shouldDetectEmbedCardPandoc() bool {
	return s.config.EmbedCardDetection == EmbedCardDetectPandoc
}

func (s *state) shouldDetectCaptionPandoc() bool {
	return s.config.CaptionDetection == CaptionDetectPandoc
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
			timestamp := strings.TrimSpace(match.extra)
			if timestamp == "" {
				parsedDate, ok := s.parseDatePatternValue(match.value)
				if !ok {
					content = appendInlineNode(content, newTextNode(match.value, nil))
					break
				}
				timestamp = strconv.FormatInt(parsedDate.Unix(), 10)
			}
			content = append(content, converter.Node{
				Type: "date",
				Attrs: map[string]interface{}{
					"timestamp": timestamp,
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
		if dateMatch, ok := s.findDatePattern(textValue); ok {
			candidates = append(candidates, dateMatch)
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

func (s *state) findDatePattern(textValue string) (patternMatch, bool) {
	best := patternMatch{}
	found := false

	for _, layout := range s.dateDetectionLayouts() {
		match, ok := findDatePatternForLayout(textValue, layout)
		if !ok {
			continue
		}

		if !found || match.start < best.start || (match.start == best.start && match.end > best.end) {
			best = match
			found = true
		}
	}

	return best, found
}

func (s *state) parseDatePatternValue(value string) (time.Time, bool) {
	for _, layout := range s.dateDetectionLayouts() {
		parsedDate, err := time.Parse(layout, value)
		if err == nil {
			return parsedDate, true
		}
	}
	return time.Time{}, false
}

func (s *state) dateDetectionLayouts() []string {
	layouts := make([]string, 0, 2)
	seen := map[string]struct{}{}

	appendLayout := func(layout string) {
		layout = strings.TrimSpace(layout)
		if layout == "" {
			return
		}
		if _, ok := seen[layout]; ok {
			return
		}
		seen[layout] = struct{}{}
		layouts = append(layouts, layout)
	}

	appendLayout(s.config.DateFormat)
	appendLayout("2006-01-02")

	return layouts
}

func findDatePatternForLayout(textValue, layout string) (patternMatch, bool) {
	dateRe, ok := dateRegexForLayout(layout)
	if !ok {
		return patternMatch{}, false
	}

	matches := dateRe.FindAllStringSubmatchIndex(textValue, -1)
	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		start := match[4]
		end := match[5]
		if start < 0 || end <= start {
			continue
		}

		candidate := textValue[start:end]
		parsedDate, err := time.Parse(layout, candidate)
		if err != nil {
			continue
		}

		return patternMatch{
			kind:  "date",
			start: start,
			end:   end,
			value: candidate,
			extra: strconv.FormatInt(parsedDate.Unix(), 10),
		}, true
	}

	return patternMatch{}, false
}

func dateRegexForLayout(layout string) (*regexp.Regexp, bool) {
	tokenPattern, ok := dateTokenPatternForLayout(layout)
	if !ok {
		return nil, false
	}

	compiled, err := regexp.Compile(`(^|[^0-9A-Za-z])(` + tokenPattern + `)($|[^0-9A-Za-z])`)
	if err != nil {
		return nil, false
	}

	return compiled, true
}

func dateTokenPatternForLayout(layout string) (string, bool) {
	layout = strings.TrimSpace(layout)
	if layout == "" {
		return "", false
	}

	var builder strings.Builder
	for i := 0; i < len(layout); {
		matched := false
		for _, token := range dateLayoutTokens {
			if strings.HasPrefix(layout[i:], token.token) {
				builder.WriteString(token.pattern)
				i += len(token.token)
				matched = true
				break
			}
		}

		if matched {
			continue
		}

		if layout[i] == ' ' {
			builder.WriteByte(' ')
		} else {
			builder.WriteString(regexp.QuoteMeta(string(layout[i])))
		}
		i++
	}

	if builder.Len() == 0 {
		return "", false
	}

	return builder.String(), true
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
		searchOffset := 0
		for searchOffset < len(textValue) {
			relative := strings.Index(textValue[searchOffset:], token)
			if relative < 0 {
				break
			}

			start := searchOffset + relative
			end := start + len(token)
			searchOffset = start + 1

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

			break
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
