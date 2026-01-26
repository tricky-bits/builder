package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
					Title:  "Test Stage",
					Author: "Test Author",
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
			name: "valid difficulty - 0",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "valid difficulty - 0.5",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 0.5,
				},
			},
			wantErr: false,
		},
		{
			name: "valid difficulty - 1.0",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1.0,
				},
			},
			wantErr: false,
		},
		{
			name: "valid difficulty - 2.5",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 2.5,
				},
			},
			wantErr: false,
		},
		{
			name: "valid difficulty - 5.0",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 5.0,
				},
			},
			wantErr: false,
		},
		{
			name: "difficulty below 0",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: -0.5,
				},
			},
			wantErr:     true,
			errContains: "difficulty must be between 0 and 5",
		},
		{
			name: "difficulty above 5",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 5.5,
				},
			},
			wantErr:     true,
			errContains: "difficulty must be between 0 and 5",
		},
		{
			name: "difficulty not multiple of 0.5 - 0.3",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 0.3,
				},
			},
			wantErr:     true,
			errContains: "difficulty must be in steps of 0.5",
		},
		{
			name: "difficulty not multiple of 0.5 - 1.7",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
					Difficulty: 1.7,
				},
			},
			wantErr:     true,
			errContains: "difficulty must be in steps of 0.5",
		},
		{
			name: "valid ETA minutes - 0",
			stage: &Stage{
				Filename: "test.md",
				Frontmatter: StageFrontmatter{
					Title:      "Test Stage",
					Author:     "Test Author",
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
					Title:  "Test Stage",
					Author: "Test Author",
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
					Title:  "Test Stage",
					Author: "Test Author",
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
					Title:  "Test Stage",
					Author: "Test Author",
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
					Title:  "Test Stage",
					Author: "Test Author",
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
					Difficulty: 2.5,
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
