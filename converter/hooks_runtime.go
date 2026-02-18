package converter

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

const (
	linkHookCacheResolvedKey = "__jac_link_hook_resolved"
	linkHookCacheHandledKey  = "__jac_link_hook_handled"
	linkHookCacheHrefKey     = "__jac_link_hook_href"
	linkHookCacheTitleKey    = "__jac_link_hook_title"
	linkHookCacheTextOnlyKey = "__jac_link_hook_text_only"
)

func (s *state) applyLinkRenderHook(nodeType string, input LinkRenderInput) (LinkRenderOutput, bool, error) {
	if s.config.LinkHook == nil {
		return LinkRenderOutput{}, false, nil
	}

	if err := s.checkContext(); err != nil {
		return LinkRenderOutput{}, false, err
	}

	output, err := s.config.LinkHook(s.ctx, input)
	if err != nil {
		if errors.Is(err, ErrUnresolved) {
			if s.config.ResolutionMode == ResolutionStrict {
				return LinkRenderOutput{}, false, fmt.Errorf("unresolved link reference %q: %w", input.Href, err)
			}
			s.addWarning(
				WarningUnresolvedReference,
				nodeType,
				fmt.Sprintf("unresolved link reference %q; using fallback rendering", input.Href),
			)
			return LinkRenderOutput{}, false, nil
		}
		return LinkRenderOutput{}, false, fmt.Errorf("link hook failed: %w", err)
	}

	if !output.Handled {
		return LinkRenderOutput{}, false, nil
	}

	if err := validateLinkRenderOutput(output); err != nil {
		return LinkRenderOutput{}, false, fmt.Errorf("invalid link hook output: %w", err)
	}

	output.Href = strings.TrimSpace(output.Href)
	output.Title = strings.TrimSpace(output.Title)

	return output, true, nil
}

func (s *state) applyMediaRenderHook(nodeType string, input MediaRenderInput) (MediaRenderOutput, bool, error) {
	if s.config.MediaHook == nil {
		return MediaRenderOutput{}, false, nil
	}

	if err := s.checkContext(); err != nil {
		return MediaRenderOutput{}, false, err
	}

	output, err := s.config.MediaHook(s.ctx, input)
	if err != nil {
		if errors.Is(err, ErrUnresolved) {
			reference := input.ID
			if reference == "" {
				reference = input.URL
			}
			if s.config.ResolutionMode == ResolutionStrict {
				return MediaRenderOutput{}, false, fmt.Errorf("unresolved media reference %q: %w", reference, err)
			}
			s.addWarning(
				WarningUnresolvedReference,
				nodeType,
				fmt.Sprintf("unresolved media reference %q; using fallback rendering", reference),
			)
			return MediaRenderOutput{}, false, nil
		}
		return MediaRenderOutput{}, false, fmt.Errorf("media hook failed: %w", err)
	}

	if !output.Handled {
		return MediaRenderOutput{}, false, nil
	}

	if err := validateMediaRenderOutput(output); err != nil {
		return MediaRenderOutput{}, false, fmt.Errorf("invalid media hook output: %w", err)
	}

	return output, true, nil
}

func validateLinkRenderOutput(output LinkRenderOutput) error {
	if output.TextOnly {
		return nil
	}
	if strings.TrimSpace(output.Href) == "" {
		return errors.New("handled link render output requires non-empty href unless textOnly is true")
	}
	return nil
}

func validateMediaRenderOutput(output MediaRenderOutput) error {
	if strings.TrimSpace(output.Markdown) == "" {
		return errors.New("handled media render output requires non-empty markdown")
	}
	return nil
}

func linkMetadataFromAttrs(attrs map[string]any, href string) LinkMetadata {
	filename, anchor := parseReferenceDetails(href)

	meta := LinkMetadata{
		PageID:       lookupMetadataValue(attrs, "pageId", "pageID", "contentId"),
		SpaceKey:     lookupMetadataValue(attrs, "spaceKey", "space"),
		AttachmentID: lookupMetadataValue(attrs, "attachmentId", "attachmentID", "mediaId"),
		Filename:     lookupMetadataValue(attrs, "filename", "fileName", "name"),
		Anchor:       lookupMetadataValue(attrs, "anchor", "fragment"),
	}

	if meta.Filename == "" {
		meta.Filename = filename
	}
	if meta.Anchor == "" {
		meta.Anchor = anchor
	}

	return meta
}

func mediaMetadataFromAttrs(attrs map[string]any, id, mediaURL string) MediaMetadata {
	filename, anchor := parseReferenceDetails(mediaURL)

	meta := MediaMetadata{
		PageID:       lookupMetadataValue(attrs, "pageId", "pageID", "contentId"),
		SpaceKey:     lookupMetadataValue(attrs, "spaceKey", "space"),
		AttachmentID: lookupMetadataValue(attrs, "attachmentId", "attachmentID", "mediaId", "id"),
		Filename:     lookupMetadataValue(attrs, "filename", "fileName", "name"),
		Anchor:       lookupMetadataValue(attrs, "anchor", "fragment"),
	}

	if meta.AttachmentID == "" {
		meta.AttachmentID = strings.TrimSpace(id)
	}
	if meta.Filename == "" {
		meta.Filename = filename
	}
	if meta.Anchor == "" {
		meta.Anchor = anchor
	}

	return meta
}

func lookupMetadataValue(attrs map[string]any, candidates ...string) string {
	if len(attrs) == 0 || len(candidates) == 0 {
		return ""
	}

	maps := collectMetadataMaps(attrs, 0)
	for _, candidate := range candidates {
		normalizedCandidate := normalizeMetadataKey(candidate)
		if normalizedCandidate == "" {
			continue
		}
		for _, candidateMap := range maps {
			if value, ok := lookupMetadataValueInMap(candidateMap, normalizedCandidate); ok {
				return value
			}
		}
	}

	return ""
}

func collectMetadataMaps(attrs map[string]any, depth int) []map[string]any {
	if len(attrs) == 0 || depth > 2 {
		return nil
	}

	result := []map[string]any{attrs}
	for _, value := range attrs {
		nested, ok := value.(map[string]any)
		if !ok || len(nested) == 0 {
			continue
		}
		result = append(result, collectMetadataMaps(nested, depth+1)...)
	}

	return result
}

func lookupMetadataValueInMap(attrs map[string]any, normalizedKey string) (string, bool) {
	for key, raw := range attrs {
		if normalizeMetadataKey(key) != normalizedKey {
			continue
		}

		value, ok := raw.(string)
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		return value, true
	}

	return "", false
}

func normalizeMetadataKey(key string) string {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, " ", "")
	return normalized
}

func parseReferenceDetails(reference string) (string, string) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return "", ""
	}

	parsed, err := url.Parse(reference)
	if err != nil {
		return parseReferenceDetailsFallback(reference)
	}

	anchor := strings.TrimSpace(parsed.Fragment)
	referencePath := parsed.Path
	if referencePath == "" {
		referencePath = reference
	}
	referencePath = strings.ReplaceAll(referencePath, "\\", "/")
	referencePath = strings.TrimRight(referencePath, "/")
	if referencePath == "" {
		return "", anchor
	}

	filename := strings.TrimSpace(path.Base(referencePath))
	if filename == "." || filename == "/" {
		filename = ""
	}

	return filename, anchor
}

func parseReferenceDetailsFallback(reference string) (string, string) {
	anchor := ""
	if hashIndex := strings.LastIndex(reference, "#"); hashIndex >= 0 {
		anchor = strings.TrimSpace(reference[hashIndex+1:])
		reference = reference[:hashIndex]
	}

	reference = strings.ReplaceAll(reference, "\\", "/")
	reference = strings.TrimRight(reference, "/")
	if reference == "" {
		return "", anchor
	}

	filename := strings.TrimSpace(path.Base(reference))
	if filename == "." || filename == "/" {
		filename = ""
	}

	return filename, anchor
}

func cloneAnyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = cloneAnyValue(value)
	}

	return dst
}

func cloneAnyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneAnyMap(typed)
	case []any:
		cloned := make([]any, len(typed))
		for index := range typed {
			cloned[index] = cloneAnyValue(typed[index])
		}
		return cloned
	default:
		return value
	}
}

func loadLinkHookCache(attrs map[string]any) (LinkRenderOutput, bool) {
	if len(attrs) == 0 {
		return LinkRenderOutput{}, false
	}

	resolved, ok := attrs[linkHookCacheResolvedKey].(bool)
	if !ok || !resolved {
		return LinkRenderOutput{}, false
	}

	handled, _ := attrs[linkHookCacheHandledKey].(bool)
	out := LinkRenderOutput{Handled: handled}
	if !handled {
		return out, true
	}

	out.Href, _ = attrs[linkHookCacheHrefKey].(string)
	out.Title, _ = attrs[linkHookCacheTitleKey].(string)
	out.TextOnly, _ = attrs[linkHookCacheTextOnlyKey].(bool)
	return out, true
}

func storeLinkHookCache(attrs map[string]any, output LinkRenderOutput, handled bool) {
	if attrs == nil {
		return
	}

	attrs[linkHookCacheResolvedKey] = true
	attrs[linkHookCacheHandledKey] = handled
	if !handled {
		return
	}

	attrs[linkHookCacheHrefKey] = output.Href
	attrs[linkHookCacheTitleKey] = output.Title
	attrs[linkHookCacheTextOnlyKey] = output.TextOnly
}
