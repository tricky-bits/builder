package markdown

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_SimpleTypes(t *testing.T) {
	content := `---
title: Test
count: 42
published: true
price: 19.99
---
# Content`
	tmpfile, err := os.CreateTemp("", "test-simple-*.md")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	tmpfile.Close()

	fm := &struct {
		Title     string  `yaml:"title"`
		Count     int     `yaml:"count"`
		Published bool    `yaml:"published"`
		Price     float64 `yaml:"price"`
	}{}

	var buf bytes.Buffer
	err = Parse(tmpfile.Name(), &buf, fm)
	require.NoError(t, err)

	assert.Equal(t, "Test", fm.Title)
	assert.Equal(t, 42, fm.Count)
	assert.True(t, fm.Published)
	assert.InDelta(t, 19.99, fm.Price, 0.001)
	assert.Contains(t, strings.TrimSpace(buf.String()), "Content")
}

func TestParse_Map(t *testing.T) {
	content := `---
metadata:
  key1: value1
  key2: value2
---
# Document`
	tmpfile, err := os.CreateTemp("", "test-map-*.md")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	tmpfile.Close()

	fm := &struct {
		Metadata map[string]string `yaml:"metadata"`
	}{}

	var buf bytes.Buffer
	err = Parse(tmpfile.Name(), &buf, fm)
	require.NoError(t, err)

	assert.Equal(t, "value1", fm.Metadata["key1"])
	assert.Equal(t, "value2", fm.Metadata["key2"])
	assert.Contains(t, strings.TrimSpace(buf.String()), "Document")
}

func TestParse_Slice(t *testing.T) {
	content := `---
tags:
  - go
  - test
  - markdown
---
# Post`
	tmpfile, err := os.CreateTemp("", "test-slice-*.md")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	tmpfile.Close()

	fm := &struct {
		Tags []string `yaml:"tags"`
	}{}

	var buf bytes.Buffer
	err = Parse(tmpfile.Name(), &buf, fm)
	require.NoError(t, err)

	assert.Equal(t, []string{"go", "test", "markdown"}, fm.Tags)
	assert.Contains(t, strings.TrimSpace(buf.String()), "Post")
}

func TestParse_Struct(t *testing.T) {
	content := `---
author:
  name: John
  email: john@example.com
---
# Article`
	tmpfile, err := os.CreateTemp("", "test-struct-*.md")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	tmpfile.Close()

	fm := &struct {
		Author struct {
			Name  string `yaml:"name"`
			Email string `yaml:"email"`
		} `yaml:"author"`
	}{}

	var buf bytes.Buffer
	err = Parse(tmpfile.Name(), &buf, fm)
	require.NoError(t, err)

	assert.Equal(t, "John", fm.Author.Name)
	assert.Equal(t, "john@example.com", fm.Author.Email)
	assert.Contains(t, strings.TrimSpace(buf.String()), "Article")
}

func TestRender(t *testing.T) {
	html, err := Render("# Hello World")
	require.NoError(t, err)
	assert.Contains(t, string(html), "<h1 id=\"hello-world\">Hello World</h1>")
}

func TestRender_SimpleMarkdown(t *testing.T) {
	html, err := Render("**bold** and *italic*")
	require.NoError(t, err)
	assert.Contains(t, string(html), "<strong>bold</strong>")
	assert.Contains(t, string(html), "<em>italic</em>")
}
