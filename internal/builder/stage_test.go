package builder

import (
	"encoding/base64"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tricky-bits/builder/internal/obfuscate"
)

// decodeHintPayload mirrors the JavaScript decoder in
// themes/base/static/progress.js so the test asserts the runtime contract,
// not just the Go-side internals.
func decodeHintPayload(key, encoded string) string {
	if key == "" {
		return encoded
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err)
	}
	keyBytes := []byte(key)
	out := make([]byte, len(raw))
	for i, b := range raw {
		out[i] = b ^ keyBytes[i%len(keyBytes)]
	}
	return string(out)
}

func TestStage_Validate(t *testing.T) {
	tests := []struct {
		name        string
		stage       *Stage
		wantErr     bool
		errContains string
	}{
		{
			name: "valid stage",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "missing title",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Author: "Test Author",
				},
			},
			wantErr:     true,
			errContains: "title is required",
		},
		{
			name: "missing author",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title: "Test Stage",
				},
			},
			wantErr:     true,
			errContains: "author is required",
		},
		{
			name: "missing both title and author",
			stage: &Stage{
				Filename:    "test.md",
				Frontmatter: StageFrontmatter{},
			},
			wantErr:     true,
			errContains: "title is required",
		},
		{
			name: "valid difficulty - 1",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "valid difficulty - 3",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 3,
				},
			},
			wantErr: false,
		},
		{
			name: "valid difficulty - 5",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 5,
				},
			},
			wantErr: false,
		},
		{
			name: "difficulty below 1",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 0,
				},
			},
			wantErr:     true,
			errContains: "difficulty must be between 1 and 5",
		},
		{
			name: "difficulty above 5",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 6,
				},
			},
			wantErr:     true,
			errContains: "difficulty must be between 1 and 5",
		},
		{
			name: "valid ETA minutes - 0",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1,
					ETAMinutes: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "valid ETA minutes - positive",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1,
					ETAMinutes: 30,
				},
			},
			wantErr: false,
		},
		{
			name: "negative ETA minutes",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1,
					ETAMinutes: -1,
				},
			},
			wantErr:     true,
			errContains: "eta_minutes must be non-negative",
		},
		{
			name: "valid hint wait seconds - 0",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1,
					Hints: []HintFrontmatter{
						{WaitSeconds: 0, Text: "Hint 1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid hint wait seconds - positive",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1,
					Hints: []HintFrontmatter{
						{WaitSeconds: 60, Text: "Hint 1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "negative hint wait seconds",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1,
					Hints: []HintFrontmatter{
						{WaitSeconds: -1, Text: "Hint 1"},
					},
				},
			},
			wantErr:     true,
			errContains: "wait_seconds must be non-negative",
		},
		{
			name: "multiple hints with negative wait seconds",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1,
					Hints: []HintFrontmatter{
						{WaitSeconds: 10, Text: "Hint 1"},
						{WaitSeconds: -5, Text: "Hint 2"},
					},
				},
			},
			wantErr:     true,
			errContains: "hint 1: wait_seconds must be non-negative",
		},
		{
			name: "stage with all optional fields valid",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Slug:       "test-stage",
					Tags:       []string{"tag1", "tag2"},
					Difficulty: 3,
					ETAMinutes: 30,
					Start:      true,
					Next:       "next-stage",
					Answer:     "answer123",
					Hints: []HintFrontmatter{
						{WaitSeconds: 10, Text: "Hint 1"},
						{WaitSeconds: 20, Text: "Hint 2"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stage.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEncodeHints(t *testing.T) {
	t.Run("non-empty key obfuscates and round-trips", func(t *testing.T) {
		key := "secret"
		enc := obfuscate.New(key)
		hints := []Hint{
			{WaitSeconds: 30, Content: template.HTML(`<p>Start by examining the input.</p>`)},
			{WaitSeconds: 120, Content: template.HTML(`<p>Look at <code>config.toml</code>.</p>`)},
		}

		got := encodeHints(enc, hints)
		require.Len(t, got, 2)

		for i, h := range hints {
			assert.Equal(t, h.WaitSeconds, got[i].WaitSeconds)
			assert.NotEqual(t, string(h.Content), got[i].EncContent, "hint %d must not ship as plaintext", i)
			_, err := base64.StdEncoding.DecodeString(got[i].EncContent)
			require.NoError(t, err, "hint %d EncContent must be valid base64", i)
			assert.Equal(t, string(h.Content), decodeHintPayload(key, got[i].EncContent),
				"hint %d must decode back to original HTML", i)
		}
	})

	t.Run("empty key passes through unchanged", func(t *testing.T) {
		enc := obfuscate.New("")
		hints := []Hint{
			{WaitSeconds: 10, Content: template.HTML(`<p>plaintext hint</p>`)},
		}

		got := encodeHints(enc, hints)
		require.Len(t, got, 1)
		assert.Equal(t, 10, got[0].WaitSeconds)
		assert.Equal(t, string(hints[0].Content), got[0].EncContent)
	})

	t.Run("empty hints slice yields empty non-nil result", func(t *testing.T) {
		enc := obfuscate.New("k")
		got := encodeHints(enc, []Hint{})
		require.NotNil(t, got)
		assert.Len(t, got, 0)
	})
}
