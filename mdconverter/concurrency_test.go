package mdconverter

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/stretchr/testify/require"
)

func TestConverterConcurrentConvertSharedInstance(t *testing.T) {
	conv, err := New(ReverseConfig{})
	require.NoError(t, err)

	input := "# Title\n\n- [ ] first\n- [x] second\n\n[Page](https://example.com)\n"

	const workers = 12
	const iterations = 120

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	reportErr := func(err error) {
		select {
		case errCh <- err:
		default:
		}
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				result, err := conv.Convert(input)
				if err != nil {
					reportErr(err)
					return
				}
				doc, err := decodeConcurrentDoc(result.ADF)
				if err != nil {
					reportErr(err)
					return
				}
				if doc.Type != "doc" || len(doc.Content) == 0 {
					reportErr(errors.New("unexpected empty document during concurrent conversion"))
					return
				}
			}
		}()
	}

	wg.Wait()
	select {
	case err := <-errCh:
		require.NoError(t, err)
	default:
	}
}

func TestReverseConverterConfigIsolationConcurrent(t *testing.T) {
	input := "```cpp\nint x = 1;\n```\n\n@username\n"

	cfg := ReverseConfig{
		MentionDetection: MentionDetectAt,
		MentionRegistry: map[string]string{
			"username": "12345",
		},
		LanguageMap: map[string]string{
			"cpp": "c++",
		},
	}

	conv, err := New(cfg)
	require.NoError(t, err)

	const workers = 8
	const iterations = 150

	var mutatorWG sync.WaitGroup
	var workersWG sync.WaitGroup
	errCh := make(chan error, 1)
	stopMutator := make(chan struct{})

	reportErr := func(err error) {
		select {
		case errCh <- err:
		default:
		}
	}

	mutatorWG.Add(1)
	go func() {
		defer mutatorWG.Done()
		for {
			select {
			case <-stopMutator:
				return
			default:
				cfg.MentionRegistry["username"] = "99999"
				cfg.MentionRegistry["username"] = "12345"
				cfg.LanguageMap["cpp"] = "cpp"
				cfg.LanguageMap["cpp"] = "c++"
			}
		}
	}()

	for i := 0; i < workers; i++ {
		workersWG.Add(1)
		go func() {
			defer workersWG.Done()
			for j := 0; j < iterations; j++ {
				result, err := conv.Convert(input)
				if err != nil {
					reportErr(err)
					return
				}
				doc, err := decodeConcurrentDoc(result.ADF)
				if err != nil {
					reportErr(err)
					return
				}
				if !docContainsCodeLanguage(doc, "c++") {
					reportErr(errors.New("converter observed caller language map mutation"))
					return
				}
				if docContainsCodeLanguage(doc, "cpp") {
					reportErr(errors.New("converter emitted unmapped language during concurrent conversion"))
					return
				}
				if !docContainsMentionID(doc, "12345") {
					reportErr(errors.New("converter observed caller mention registry mutation"))
					return
				}
			}
		}()
	}

	workersWG.Wait()
	close(stopMutator)
	mutatorWG.Wait()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	default:
	}
}

func TestReverseConverterConcurrentHooksWithCallerSynchronization(t *testing.T) {
	var mu sync.Mutex
	hookCalls := 0

	conv, err := New(ReverseConfig{
		LinkHook: func(_ context.Context, in LinkParseInput) (LinkParseOutput, error) {
			mu.Lock()
			hookCalls++
			mu.Unlock()

			return LinkParseOutput{
				Destination: in.Destination,
				ForceLink:   true,
				Handled:     true,
			}, nil
		},
	})
	require.NoError(t, err)

	const workers = 10
	const iterations = 80

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	reportErr := func(err error) {
		select {
		case errCh <- err:
		default:
		}
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				result, err := conv.Convert(`[Docs](../docs.md)`)
				if err != nil {
					reportErr(err)
					return
				}

				doc, err := decodeConcurrentDoc(result.ADF)
				if err != nil {
					reportErr(err)
					return
				}
				if !docContainsLinkHref(doc, "../docs.md") {
					reportErr(errors.New("hook-forced link output missing expected href"))
					return
				}
			}
		}()
	}
	wg.Wait()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	default:
	}

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, workers*iterations, hookCalls)
}

func decodeConcurrentDoc(payload []byte) (converter.Doc, error) {
	var doc converter.Doc
	err := json.Unmarshal(payload, &doc)
	return doc, err
}

func docContainsCodeLanguage(doc converter.Doc, language string) bool {
	for _, node := range doc.Content {
		if node.Type == "codeBlock" && node.GetStringAttr("language", "") == language {
			return true
		}
		if nodeContainsCodeLanguage(node, language) {
			return true
		}
	}
	return false
}

func nodeContainsCodeLanguage(node converter.Node, language string) bool {
	if node.Type == "codeBlock" && node.GetStringAttr("language", "") == language {
		return true
	}
	for _, child := range node.Content {
		if nodeContainsCodeLanguage(child, language) {
			return true
		}
	}
	return false
}

func docContainsMentionID(doc converter.Doc, mentionID string) bool {
	for _, node := range doc.Content {
		if nodeContainsMentionID(node, mentionID) {
			return true
		}
	}
	return false
}

func nodeContainsMentionID(node converter.Node, mentionID string) bool {
	if node.Type == "mention" && node.GetStringAttr("id", "") == mentionID {
		return true
	}
	for _, child := range node.Content {
		if nodeContainsMentionID(child, mentionID) {
			return true
		}
	}
	return false
}

func docContainsLinkHref(doc converter.Doc, href string) bool {
	for _, node := range doc.Content {
		if nodeContainsLinkHref(node, href) {
			return true
		}
	}
	return false
}

func nodeContainsLinkHref(node converter.Node, href string) bool {
	for _, mark := range node.Marks {
		if mark.Type == "link" && mark.GetStringAttr("href", "") == href {
			return true
		}
	}
	for _, child := range node.Content {
		if nodeContainsLinkHref(child, href) {
			return true
		}
	}
	return false
}
