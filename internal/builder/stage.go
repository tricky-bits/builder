package builder

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/tricky-bits/builder/internal/helpers"
	"github.com/tricky-bits/builder/internal/markdown"
)

// HintFrontmatter defines a timed, click-to-reveal hint associated with a stage.
type HintFrontmatter struct {
	// WaitSeconds is the delay, in seconds, after the stage becomes visible
	// before this hint is allowed to be revealed.
	WaitSeconds int `yaml:"wait_seconds"`

	// Text is the hint content presented to the user once revealed.
	Text string `yaml:"text"`
}

// Hint represents the parsed click-to-reveal hint associated with a stage.
type Hint struct {
	// WaitSeconds is the delay, in seconds, after the stage becomes visible
	// before this hint is allowed to be revealed.
	WaitSeconds int

	// Text is the hint content presented to the user once revealed.
	Content template.HTML
}

// StageFrontmatter is the YAML frontmatter declared at the top of a stage
// markdown file.
type StageFrontmatter struct {
	// Title is the human-readable stage title (required).
	Title string `yaml:"title"`

	// Author represents the author of the stage (required).
	Author string `yaml:"author"`

	// Slug is the logical route identifier for the stage (optional). If omitted,
	// it is typically derived from the filename.
	Slug string `yaml:"slug,omitempty"`

	// Tags is a set of free-form labels used for discovery, filtering, or grouping
	// (optional).
	Tags []string `yaml:"tags,omitempty"`

	// Difficulty is an optional numeric indicator used for sorting or display.
	// The scale is project-defined (e.g., 1.0–5.0).
	// TODO Discussion: Should we stick with a float value or a int range 1-5 is enough?
	Difficulty float64 `yaml:"difficulty,omitempty"`

	// ETAMinutes is an optional estimate of the time required to complete the stage.
	ETAMinutes int `yaml:"eta_minutes,omitempty"`

	// Theme specifies the theme to use for this stage, overriding global configuration if present.
	Theme string `yaml:"theme"`

	// Start marks this stage as an entry point for the campaign.
	Start bool `yaml:"start,omitempty"`

	// Next identifies the slug (or route identifier) of the next stage (optional, if last).
	// This is used to build linear navigation.
	Next string `yaml:"next,omitempty"`

	// Answer is the expected solution token/value for the stage (optional), used
	// for validation or gating progression depending on the application logic.
	Answer string `yaml:"answer,omitempty"`

	// Hints is the ordered list of timed hints for the stage (optional).
	Hints []HintFrontmatter `yaml:"hints,omitempty"`

	// PublishedAt is the stage publication timestamp (optional).
	PublishedAt *time.Time `yaml:"published_at,omitempty"`

	// LastUpdatedAt is the stage last update timestamp (optional).
	LastUpdatedAt *time.Time `yaml:"last_update_at,omitempty"`

	// CompletionMessage represents the message displayed on stage completion.
	CompletionMessage string `yaml:"completion_message"`

	// Inputs is a list of file names that should be copied alongside the stage.
	// These files should be present in the same directory as the stage markdown file.
	// TODO Discussion: Define where the input files should be stored.
	Inputs []string `yaml:"inputs,omitempty"`

	// Assets is a list of file names (e.g., images) that should be copied alongside the stage.
	// These files should be present in the same directory as the stage markdown file.
	// Unlike inputs, assets are not displayed in the stage info section.
	// TODO Discussion: Define where the assets files should be stored.
	Assets []string `yaml:"assets,omitempty"`
}

// Stage represents a parsed and rendered stage produced from a markdown file.
type Stage struct {
	// Filename is the source markdown file path (or name), used primarily for
	// traceability and error reporting.
	Filename string

	// Frontmatter contains the parsed YAML header for the stage.
	Frontmatter StageFrontmatter

	// Content is the rendered HTML body of the stage, intended to be injected into
	// a larger HTML template.
	Content template.HTML

	// CompletionMessage is the rendered HTML message, to be displayed upon completion.
	CompletionMessage template.HTML

	// Hints is the ordered list of parsed hints for the stage (optional).
	Hints []Hint
}

// ReadStageFile reads and parses a stage markdown file from the given filename.
// It extracts the YAML frontmatter, renders the markdown content to HTML, processes
// hints and completion messages, derives the slug if not provided, and validates
// the stage structure.
func ReadStageFile(filename string) (*Stage, error) {
	stage := &Stage{
		Filename: filename,
	}

	// Parse markdown file and extract frontmatter, writing HTML to buffer
	var contentBuf bytes.Buffer
	if err := markdown.Parse(filename, &contentBuf, &stage.Frontmatter); err != nil {
		return nil, fmt.Errorf("unable to parse markdown file: %w", err)
	}

	stage.Content = template.HTML(contentBuf.String())

	// Render completion message to HTML
	var err error
	if stage.Frontmatter.CompletionMessage != "" {
		stage.CompletionMessage, err = markdown.Render(stage.Frontmatter.CompletionMessage)
		if err != nil {
			return nil, fmt.Errorf("render stage completion message: %w", err)
		}
	}

	// Render hints
	stage.Hints = make([]Hint, len(stage.Frontmatter.Hints))
	for i, hintFrontmatter := range stage.Frontmatter.Hints {
		if hintFrontmatter.Text == "" {
			return nil, fmt.Errorf("hint %d has empty text", i)
		}

		content, err := markdown.Render(hintFrontmatter.Text)
		if err != nil {
			return nil, fmt.Errorf("unable to render hint %d: %w", i, err)
		}

		stage.Hints[i] = Hint{
			WaitSeconds: hintFrontmatter.WaitSeconds,
			Content:     content,
		}
	}

	// Derive slug from frontmatter or filename
	if stage.Frontmatter.Slug == "" {
		stage.Frontmatter.Slug = helpers.DeriveSlug(filename)
	}

	if err := stage.Validate(); err != nil {
		return nil, err
	}

	return stage, nil
}

// Validate performs stage-local validation.
func (s *Stage) Validate() error {
	if s.Frontmatter.Title == "" {
		return fmt.Errorf("title is required")
	}
	if s.Frontmatter.Author == "" {
		return fmt.Errorf("author is required")
	}

	// Validate difficulty range (0-5 with 0.5 steps)
	if s.Frontmatter.Difficulty < 0 || s.Frontmatter.Difficulty > 5 {
		return fmt.Errorf("difficulty must be between 0 and 5, got %.1f", s.Frontmatter.Difficulty)
	}

	// Check if difficulty is a multiple of 0.5 (i.e., 0, 0.5, 1, 1.5, ..., 5)
	doubledDifficulty := s.Frontmatter.Difficulty * 2
	if doubledDifficulty != float64(int(doubledDifficulty)) {
		return fmt.Errorf("difficulty must be in steps of 0.5, got %.2f", s.Frontmatter.Difficulty)
	}

	if s.Frontmatter.ETAMinutes < 0 {
		return fmt.Errorf("eta_minutes must be non-negative, got %d", s.Frontmatter.ETAMinutes)
	}

	// Validate hints
	for i, hint := range s.Frontmatter.Hints {
		if hint.WaitSeconds < 0 {
			return fmt.Errorf("hint %d: wait_seconds must be non-negative, got %d", i, hint.WaitSeconds)
		}
	}

	// Validate input files exist
	// TODO Discussion: Where are input / assert files located?
	stageDir := filepath.Dir(s.Filename)
	for _, inputFile := range s.Frontmatter.Inputs {
		inputPath := filepath.Join(stageDir, inputFile)
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			return fmt.Errorf("[%s] input file not found: %s", s.Filename, inputPath)
		}
	}

	// Validate asset files exist
	for _, assetFile := range s.Frontmatter.Assets {
		assetPath := filepath.Join(stageDir, assetFile)
		if _, err := os.Stat(assetPath); os.IsNotExist(err) {
			return fmt.Errorf("[%s] asset file not found: %s", s.Filename, assetPath)
		}
	}

	return nil
}
