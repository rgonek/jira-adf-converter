package mdconverter

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
)

func (s *state) applyLinkParseHook(input LinkParseInput) (LinkParseOutput, bool, error) {
	if s.config.LinkHook == nil {
		return LinkParseOutput{}, false, nil
	}

	if err := s.checkContext(); err != nil {
		return LinkParseOutput{}, false, err
	}

	output, err := s.config.LinkHook(s.ctx, input)
	if err != nil {
		if errors.Is(err, ErrUnresolved) {
			if s.config.ResolutionMode == ResolutionStrict {
				return LinkParseOutput{}, false, fmt.Errorf("unresolved link destination %q: %w", input.Destination, err)
			}
			s.addWarning(
				converter.WarningUnresolvedReference,
				"link",
				fmt.Sprintf("unresolved link destination %q; using fallback parsing", input.Destination),
			)
			return LinkParseOutput{}, false, nil
		}
		return LinkParseOutput{}, false, fmt.Errorf("link hook failed: %w", err)
	}

	if !output.Handled {
		return LinkParseOutput{}, false, nil
	}

	output.Destination = strings.TrimSpace(output.Destination)
	output.Title = strings.TrimSpace(output.Title)

	if err := validateLinkParseOutput(output); err != nil {
		return LinkParseOutput{}, false, fmt.Errorf("invalid link hook output: %w", err)
	}

	return output, true, nil
}

func (s *state) applyMediaParseHook(input MediaParseInput) (MediaParseOutput, bool, error) {
	if s.config.MediaHook == nil {
		return MediaParseOutput{}, false, nil
	}

	if err := s.checkContext(); err != nil {
		return MediaParseOutput{}, false, err
	}

	output, err := s.config.MediaHook(s.ctx, input)
	if err != nil {
		if errors.Is(err, ErrUnresolved) {
			if s.config.ResolutionMode == ResolutionStrict {
				return MediaParseOutput{}, false, fmt.Errorf("unresolved media destination %q: %w", input.Destination, err)
			}
			s.addWarning(
				converter.WarningUnresolvedReference,
				"image",
				fmt.Sprintf("unresolved media destination %q; using fallback parsing", input.Destination),
			)
			return MediaParseOutput{}, false, nil
		}
		return MediaParseOutput{}, false, fmt.Errorf("media hook failed: %w", err)
	}

	if !output.Handled {
		return MediaParseOutput{}, false, nil
	}

	output.MediaType = strings.ToLower(strings.TrimSpace(output.MediaType))
	output.ID = strings.TrimSpace(output.ID)
	output.URL = strings.TrimSpace(output.URL)
	output.Alt = strings.TrimSpace(output.Alt)

	if err := validateMediaParseOutput(output); err != nil {
		return MediaParseOutput{}, false, fmt.Errorf("invalid media hook output: %w", err)
	}

	return output, true, nil
}

func validateLinkParseOutput(output LinkParseOutput) error {
	if output.ForceLink && output.ForceCard {
		return errors.New("link parse hook output cannot set both forceLink and forceCard")
	}
	if strings.TrimSpace(output.Destination) == "" {
		return errors.New("handled link parse output requires non-empty destination")
	}
	return nil
}

func validateMediaParseOutput(output MediaParseOutput) error {
	if output.MediaType != "image" && output.MediaType != "file" {
		return fmt.Errorf("handled media parse output requires supported mediaType, got %q", output.MediaType)
	}
	hasID := output.ID != ""
	hasURL := output.URL != ""
	if !hasID && !hasURL {
		return errors.New("handled media parse output requires id or url")
	}
	if hasID && hasURL {
		return errors.New("handled media parse output cannot include both id and url")
	}
	return nil
}

func linkMetadataFromDestination(destination string) LinkMetadata {
	filename, anchor := parseReferenceDetails(destination)
	return LinkMetadata{
		Filename: filename,
		Anchor:   anchor,
	}
}

func mediaMetadataFromDestination(destination string) MediaMetadata {
	filename, anchor := parseReferenceDetails(destination)
	return MediaMetadata{
		Filename: filename,
		Anchor:   anchor,
	}
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
