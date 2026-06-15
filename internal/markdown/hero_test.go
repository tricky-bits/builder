package markdown

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeroTitle(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantHTML template.HTML
		wantText string
	}{
		{
			name:     "bold segment becomes gradient span, plain text strips markers",
			in:       "DECODE **the unsolvable.**",
			wantHTML: `DECODE <span class="tb-gradient">the unsolvable.</span>`,
			wantText: "DECODE the unsolvable.",
		},
		{
			name:     "no bold renders plain in both forms",
			in:       "DECODE the unsolvable",
			wantHTML: "DECODE the unsolvable",
			wantText: "DECODE the unsolvable",
		},
		{
			name:     "multiple bold segments all become gradient spans",
			in:       "**A** middle **B**",
			wantHTML: `<span class="tb-gradient">A</span> middle <span class="tb-gradient">B</span>`,
			wantText: "A middle B",
		},
		{
			name:     "ampersand stays escaped in html, unescaped in text",
			in:       "Salt **&** Pepper",
			wantHTML: `Salt <span class="tb-gradient">&amp;</span> Pepper`,
			wantText: "Salt & Pepper",
		},
		{
			name:     "empty input yields zero values",
			in:       "",
			wantHTML: "",
			wantText: "",
		},
		{
			name:     "whitespace only yields zero values",
			in:       "   \n  ",
			wantHTML: "",
			wantText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHTML, gotText, err := HeroTitle(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.wantHTML, gotHTML)
			assert.Equal(t, tt.wantText, gotText)
		})
	}
}
