package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunForwardPrintsWarningsToStderr(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input.json")
	require.NoError(t, os.WriteFile(inputPath, []byte(`{"type":"doc","content":[{"type":"unknownNode"}]}`), 0644))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run([]string{inputPath}, &stdout, &stderr)

	assert.Equal(t, exitCodeOK, exitCode)
	assert.Contains(t, stdout.String(), "[Unknown node: unknownNode]")
	assert.Contains(t, stderr.String(), "warning: type=unknown_node")
	assert.Contains(t, stderr.String(), `message="unknown node rendered as placeholder: unknownNode"`)
}

func TestRunForwardFailOnWarning(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input.json")
	require.NoError(t, os.WriteFile(inputPath, []byte(`{"type":"doc","content":[{"type":"unknownNode"}]}`), 0644))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run([]string{"--fail-on-warning", inputPath}, &stdout, &stderr)

	assert.Equal(t, exitCodeWarnError, exitCode)
	assert.Contains(t, stdout.String(), "[Unknown node: unknownNode]")
	assert.Contains(t, stderr.String(), "warning: type=unknown_node")
}

func TestRunReversePrintsWarningsToStderr(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input.md")
	require.NoError(t, os.WriteFile(inputPath, []byte("- [ ] item ![pic](./a.png)\n"), 0644))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run([]string{"--reverse", inputPath}, &stdout, &stderr)

	assert.Equal(t, exitCodeOK, exitCode)
	assert.Contains(t, stdout.String(), `"type": "doc"`)
	assert.Contains(t, stderr.String(), "warning: type=dropped_feature")
	assert.Contains(t, stderr.String(), "task item only supports inline content")
}

func TestRunWarningFreeWithFailOnWarning(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input.json")
	require.NoError(t, os.WriteFile(inputPath, []byte(`{"version":1,"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hello"}]}]}`), 0644))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run([]string{"--fail-on-warning", inputPath}, &stdout, &stderr)

	assert.Equal(t, exitCodeOK, exitCode)
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "hello\n", stdout.String())
}
