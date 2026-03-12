package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveSlug(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "simple filename",
			filename: "hello.md",
			want:     "hello",
		},
		{
			name:     "filename with path",
			filename: "/path/to/stage.md",
			want:     "stage",
		},
		{
			name:     "filename with spaces",
			filename: "hello world.md",
			want:     "hello-world",
		},
		{
			name:     "filename with underscores",
			filename: "hello_world.md",
			want:     "hello-world",
		},
		{
			name:     "filename with spaces and underscores",
			filename: "hello_world test.md",
			want:     "hello-world-test",
		},
		{
			name:     "filename with special characters",
			filename: "hello@world#test.md",
			want:     "helloworldtest",
		},
		{
			name:     "filename with mixed case",
			filename: "HelloWorld.md",
			want:     "helloworld",
		},
		{
			name:     "filename with numbers",
			filename: "stage123.md",
			want:     "stage123",
		},
		{
			name:     "filename with consecutive hyphens",
			filename: "hello---world.md",
			want:     "hello-world",
		},
		{
			name:     "filename starting with hyphen",
			filename: "-hello.md",
			want:     "hello",
		},
		{
			name:     "filename ending with hyphen",
			filename: "hello-.md",
			want:     "hello",
		},
		{
			name:     "filename with hyphens at both ends",
			filename: "-hello-world-.md",
			want:     "hello-world",
		},
		{
			name:     "filename with multiple special characters",
			filename: "hello!@#$%world.md",
			want:     "helloworld",
		},
		{
			name:     "filename with unicode characters",
			filename: "héllo-wörld.md",
			want:     "hllo-wrld",
		},
		{
			name:     "filename with dots in name",
			filename: "hello.world.md",
			want:     "helloworld",
		},
		{
			name:     "filename with no extension",
			filename: "hello",
			want:     "hello",
		},
		{
			name:     "filename with path and no extension",
			filename: "/path/to/hello",
			want:     "hello",
		},
		{
			name:     "filename with only special characters",
			filename: "!@#$%.md",
			want:     "",
		},
		{
			name:     "empty filename",
			filename: "",
			want:     "",
		},
		{
			name:     "filename with only hyphens",
			filename: "---.md",
			want:     "",
		},
		{
			name:     "filename with mixed separators",
			filename: "hello_world-test file.md",
			want:     "hello-world-test-file",
		},
		{
			name:     "filename with numbers and special chars",
			filename: "stage-1_v2.0.md",
			want:     "stage-1-v20",
		},
		{
			name:     "filename with multiple extensions",
			filename: "file.tar.gz",
			want:     "filetar",
		},
		{
			name:     "filename with Windows path",
			filename: "C:\\Users\\test\\stage.md",
			want:     "cusersteststage",
		},
		{
			name:     "filename with relative path",
			filename: "./stages/intro.md",
			want:     "intro",
		},
		{
			name:     "filename with parent directory",
			filename: "../stages/intro.md",
			want:     "intro",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveSlug(tt.filename)
			assert.Equal(t, tt.want, got, "DeriveSlug(%q) = %q, want %q", tt.filename, got, tt.want)
		})
	}
}
