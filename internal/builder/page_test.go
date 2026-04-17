package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPage_Validate(t *testing.T) {
	tests := []struct {
		name        string
		page        *Page
		wantErr     bool
		errContains string
	}{
		{
			name: "valid page",
			page: &Page{
				Filename: "test.md",
				Frontmatter: PageFrontmatter{
					Title: "Test Page",
					Slug:  "test-page",
				},
			},
			wantErr: false,
		},
		{
			name: "missing title",
			page: &Page{
				Filename: "test.md",
				Frontmatter: PageFrontmatter{
					Slug: "test-page",
				},
			},
			wantErr:     true,
			errContains: "title is required",
		},
		{
			name: "missing slug",
			page: &Page{
				Filename: "test.md",
				Frontmatter: PageFrontmatter{
					Title: "Test Page",
				},
			},
			wantErr:     true,
			errContains: "slug is required",
		},
		{
			name: "negative menu order",
			page: &Page{
				Filename: "test.md",
				Frontmatter: PageFrontmatter{
					Title:     "Test Page",
					Slug:      "test-page",
					MenuOrder: -1,
				},
			},
			wantErr:     true,
			errContains: "menu_order must be non-negative",
		},
		{
			name: "valid page with all optional fields",
			page: &Page{
				Filename: "test.md",
				Frontmatter: PageFrontmatter{
					Title:     "Test Page",
					Slug:      "test-page",
					Theme:     "dark",
					MenuLabel: "Test",
					MenuOrder: 1,
					MenuIcon:  "icon-test",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.page.Validate()

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
