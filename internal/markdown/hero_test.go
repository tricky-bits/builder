package markdown

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderHeroTitle(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want template.HTML
	}{
		{
			name: "bold segment becomes gradient span",
			in:   "DECODE **the unsolvable.**",
			want: `DECODE <span class="tb-gradient">the unsolvable.</span>`,
		},
		{
			name: "no bold renders plain with no gradient span",
			in:   "DECODE the unsolvable",
			want: "DECODE the unsolvable",
		},
		{
			name: "multiple bold segments all become gradient spans",
			in:   "**A** middle **B**",
			want: `<span class="tb-gradient">A</span> middle <span class="tb-gradient">B</span>`,
		},
		{
			name: "empty input yields empty html",
			in:   "",
			want: "",
		},
		{
			name: "whitespace only yields empty html",
			in:   "   \n  ",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderHeroTitle(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
