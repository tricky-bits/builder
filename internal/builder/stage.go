package builder

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tricky-bits/builder/internal/helpers"
	"github.com/tricky-bits/builder/internal/markdown"
	"github.com/tricky-bits/builder/internal/obfuscate"
)

// hintData is the per-hint record passed to the stage template. EncContent
// holds the hint's rendered HTML run through the builder's encoder
// (passthrough when the obfuscation key is empty).
type hintData struct {
	WaitSeconds int
	EncContent  string
}

// encodeHints converts the domain Hint slice into template-facing hintData,
// encoding each hint's rendered HTML through enc.
func encodeHints(enc *obfuscate.Encoder, hints []Hint) []hintData {
	out := make([]hintData, len(hints))
	for i, h := range hints {
		out[i] = hintData{
			WaitSeconds: h.WaitSeconds,
			EncContent:  enc.Encode(string(h.Content)),
		}
	}
	return out
}

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

	// Difficulty is the stage difficulty on a 1–5 scale (required).
	Difficulty int `yaml:"difficulty"`

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

	// AnswerSHA256 is a precomputed SHA-256 hex digest of the answer, letting
	// campaigns ship publicly without the plaintext answer. It must be the digest
	// of the trimmed, lowercased answer (see HashAnswer). When Answer is also set,
	// plaintext wins and this value is ignored (a mismatch is logged at build time).
	AnswerSHA256 string `yaml:"answer_sha256,omitempty"`

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

// ReadAllStages discovers all .md files under basepath, parses each with
// ReadStageFile, and returns them keyed by their slug. Returns an error if any
// file fails to parse or if two stages share the same slug.
func ReadAllStages(basepath string) (map[string]*Stage, error) {
	filenames, err := helpers.ListFiles(basepath, ".md")
	if err != nil {
		return nil, fmt.Errorf("unable to list stage files: %w", err)
	}

	stages := make(map[string]*Stage)
	for _, filename := range filenames {
		stage, err := ReadStageFile(filename)
		if err != nil {
			return nil, fmt.Errorf("unable to read stage file %q: %w", filename, err)
		}

		if _, exists := stages[stage.Frontmatter.Slug]; exists {
			return nil, fmt.Errorf("[%s] duplicate stage slug: %s", filename, stage.Frontmatter.Slug)
		}

		stages[stage.Frontmatter.Slug] = stage
	}

	return stages, nil
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

	// Normalize a precomputed answer hash so it matches the lowercase hex emitted
	// by HashAnswer and the in-browser SHA-256 comparison.
	stage.Frontmatter.AnswerSHA256 = strings.ToLower(strings.TrimSpace(stage.Frontmatter.AnswerSHA256))

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

	if s.Frontmatter.Difficulty < 1 || s.Frontmatter.Difficulty > 5 {
		return fmt.Errorf("difficulty must be between 1 and 5, got %d", s.Frontmatter.Difficulty)
	}

	if s.Frontmatter.ETAMinutes < 0 {
		return fmt.Errorf("eta_minutes must be non-negative, got %d", s.Frontmatter.ETAMinutes)
	}

	// A precomputed answer hash must be a SHA-256 hex digest, otherwise the stage
	// ships silently unsolvable. Absence stays legal (informational stages).
	if s.Frontmatter.AnswerSHA256 != "" && !sha256HexRE.MatchString(s.Frontmatter.AnswerSHA256) {
		return fmt.Errorf("answer_sha256 must be a 64-character SHA-256 hex digest, got %q", s.Frontmatter.AnswerSHA256)
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

// sha256HexRE matches a lowercase SHA-256 hex digest.
var sha256HexRE = regexp.MustCompile(`^[0-9a-f]{64}$`)

// ResolveAnswerHash returns the answer hash shipped to the client. Plaintext
// wins: when Answer is set, its HashAnswer digest is used and any precomputed
// AnswerSHA256 is ignored. mismatch reports that both were set but disagreed,
// signalling a stale precomputed hash worth warning about. When Answer is empty
// the precomputed AnswerSHA256 is returned as-is (empty for informational stages).
func (s *Stage) ResolveAnswerHash() (hash string, mismatch bool) {
	if s.Frontmatter.Answer != "" {
		h := HashAnswer(s.Frontmatter.Answer)
		if s.Frontmatter.AnswerSHA256 != "" && s.Frontmatter.AnswerSHA256 != h {
			return h, true
		}
		return h, false
	}
	return s.Frontmatter.AnswerSHA256, false
}

// HashAnswer returns the SHA-256 hex digest of s after trimming spaces and
// lowercasing. Returns an empty string if s is empty.
func HashAnswer(s string) string {
	if s == "" {
		return ""
	}
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// Build renders the stage to an HTML file and copies input/asset files.
func (s *Stage) Build(b *Builder, c *Campaign) error {
	basename := filepath.Base(s.Filename)

	outputDir := filepath.Join(
		b.config.Build.OutputDir,
		"campaigns",
		c.Frontmatter.Slug,
		s.Frontmatter.Slug,
	)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("[%s] create stage output directory: %w", basename, err)
	}

	if err := s.buildStagePage(b, c, outputDir); err != nil {
		return err
	}

	if err := s.buildStageInputs(b, outputDir); err != nil {
		return err
	}

	if err := s.buildStageAssets(b, outputDir); err != nil {
		return err
	}

	b.logger.Info("built stage", "campaign", c.Frontmatter.Slug, "stage", s.Frontmatter.Slug)
	return nil
}

// buildStagePage renders the stage HTML template and writes index.html to outputDir.
func (s *Stage) buildStagePage(b *Builder, c *Campaign, outputDir string) error {
	basename := filepath.Base(s.Filename)

	t, err := b.themeMgr.Load(s.Frontmatter.Theme, c.Frontmatter.Theme, b.config.Build.Theme)
	if err != nil {
		return fmt.Errorf("[%s] load theme: %w", basename, err)
	}

	// Locate this stage's index in the campaign's linear chain so the
	// sidebar can mark the current entry without leaking its slug.
	orderIndex := -1
	current := c.StartSlug
	for i := 0; current != ""; i++ {
		if current == s.Frontmatter.Slug {
			orderIndex = i
			break
		}
		next, ok := c.Stages[current]
		if !ok {
			break
		}
		current = next.Frontmatter.Next
	}

	payloads, err := encodeStagePayloads(b, c)
	if err != nil {
		return fmt.Errorf("[%s] encode stage payloads: %w", basename, err)
	}

	type stageData struct {
		Title                string
		Slug                 string
		Author               string
		Tags                 []string
		Difficulty           int
		ETAMinutes           int
		Content              template.HTML
		EncCompletionMessage string
		Next                 string
		AnswerHash           string
		Hints                []hintData
		Inputs               []string
		OrderIndex           int
	}

	type campaignData struct {
		Slug                 string
		Title                string
		Category             string
		StageStartSlug       string
		StageCount           int
		Payloads             string
		HasFeaturedImage     bool
		EncCompletionMessage string
	}

	answerHash, mismatch := s.ResolveAnswerHash()
	if mismatch {
		b.logger.Warn("answer_sha256 does not match plaintext answer; using plaintext",
			"campaign", c.Frontmatter.Slug, "stage", s.Frontmatter.Slug)
	}

	encStageCompletion := ""
	if s.CompletionMessage != "" {
		encStageCompletion = b.encoder.Encode(string(s.CompletionMessage))
	}
	encCampaignCompletion := ""
	if c.CompletionMessage != "" {
		encCampaignCompletion = b.encoder.Encode(string(c.CompletionMessage))
	}

	data := struct {
		Site     SiteConfig
		Stage    stageData
		Campaign campaignData
		K        string
	}{
		Site: b.config.Site,
		Stage: stageData{
			Title:                s.Frontmatter.Title,
			Slug:                 s.Frontmatter.Slug,
			Author:               s.Frontmatter.Author,
			Tags:                 s.Frontmatter.Tags,
			Difficulty:           s.Frontmatter.Difficulty,
			ETAMinutes:           s.Frontmatter.ETAMinutes,
			Content:              s.Content,
			EncCompletionMessage: encStageCompletion,
			Next:                 s.Frontmatter.Next,
			AnswerHash:           answerHash,
			Hints:                encodeHints(b.encoder, s.Hints),
			Inputs:               s.Frontmatter.Inputs,
			OrderIndex:           orderIndex,
		},
		Campaign: campaignData{
			Slug:                 c.Frontmatter.Slug,
			Title:                c.Frontmatter.Title,
			Category:             c.Frontmatter.Category,
			StageStartSlug:       c.StartSlug,
			StageCount:           len(c.Stages),
			Payloads:             payloads,
			HasFeaturedImage:     c.HasFeaturedImage,
			EncCompletionMessage: encCampaignCompletion,
		},
		K: b.config.Build.ObfuscationKey,
	}

	var buffer bytes.Buffer
	if err := t.Render(&buffer, "stage.html", data); err != nil {
		return fmt.Errorf("[%s] render stage: %w", basename, err)
	}

	outputPath := filepath.Join(outputDir, "index.html")
	if err := os.WriteFile(outputPath, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("[%s] write rendered stage: %w", basename, err)
	}

	return nil
}

// buildStageInputs copies the stage's input files to outputDir.
func (s *Stage) buildStageInputs(b *Builder, outputDir string) error {
	basename := filepath.Base(s.Filename)
	stageDir := filepath.Dir(s.Filename)

	for _, inputFile := range s.Frontmatter.Inputs {
		srcPath := filepath.Join(stageDir, inputFile)
		dstPath := filepath.Join(outputDir, inputFile)
		if err := helpers.CopyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("[%s] copy input file %q: %w", basename, inputFile, err)
		}
		b.logger.Info("copied input file", "stage", s.Frontmatter.Slug, "file", inputFile)
	}

	return nil
}

// buildStageAssets copies the stage's asset files to outputDir.
func (s *Stage) buildStageAssets(b *Builder, outputDir string) error {
	basename := filepath.Base(s.Filename)
	stageDir := filepath.Dir(s.Filename)

	for _, assetFile := range s.Frontmatter.Assets {
		srcPath := filepath.Join(stageDir, assetFile)
		dstPath := filepath.Join(outputDir, assetFile)
		if err := helpers.CopyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("[%s] copy asset file %q: %w", basename, assetFile, err)
		}
		b.logger.Info("copied asset file", "stage", s.Frontmatter.Slug, "file", assetFile)
	}

	return nil
}
