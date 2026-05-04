package obfuscate

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decode mirrors the JavaScript decoder in themes/base/static/progress.js.
// Tests round-trip through this helper so build-time encoding and runtime
// decoding cannot drift out of sync.
func decode(key, encoded string) string {
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

func TestEncoder_RoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		plaintext string
	}{
		{name: "ascii", key: "secret", plaintext: "hello world"},
		{name: "key shorter than input", key: "ab", plaintext: "the quick brown fox jumps over the lazy dog"},
		{name: "key longer than input", key: "this-is-a-very-long-key", plaintext: "short"},
		{name: "key same length as input", key: "abcdef", plaintext: "abcdef"},
		{name: "multi-byte utf8 accents", key: "k", plaintext: "café déjà vu — naïve façade"},
		{name: "multi-byte utf8 emoji", key: "key123", plaintext: "hello 🌍 from 🚀 land"},
		{name: "html payload", key: "K", plaintext: `<div class="hint">Try checking <code>config.toml</code>.</div>`},
		{name: "json payload", key: "abc", plaintext: `[{"slug":"s1","title":"One"},{"slug":"s2","title":"Two"}]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := New(tt.key)

			encoded := enc.Encode(tt.plaintext)
			// Encoded form must be valid base64.
			_, err := base64.StdEncoding.DecodeString(encoded)
			require.NoError(t, err, "encoded value must be valid base64")
			// And must not equal the plaintext (assuming non-trivial input).
			if tt.plaintext != "" {
				assert.NotEqual(t, tt.plaintext, encoded, "non-empty plaintext must be obfuscated")
			}

			got := decode(tt.key, encoded)
			assert.Equal(t, tt.plaintext, got, "round-trip must reproduce plaintext")
		})
	}
}

func TestEncoder_EmptyKeyIsPassthrough(t *testing.T) {
	enc := New("")

	cases := []string{
		"",
		"plain ascii",
		`<div>html</div>`,
		"café 🌍",
	}
	for _, p := range cases {
		assert.Equal(t, p, enc.Encode(p), "empty key must passthrough %q", p)
	}
}

func TestEncoder_EmptyPlaintext(t *testing.T) {
	enc := New("secret")

	encoded := enc.Encode("")
	assert.Equal(t, "", encoded, "empty plaintext encodes to empty string")
	assert.Equal(t, "", decode("secret", encoded))
}

func TestEncoder_StableOutput(t *testing.T) {
	enc1 := New("rotate-me")
	enc2 := New("rotate-me")
	plaintext := "deterministic encoding must be byte-stable across calls and instances"

	first := enc1.Encode(plaintext)
	second := enc1.Encode(plaintext)
	third := enc2.Encode(plaintext)

	assert.Equal(t, first, second, "repeated Encode on same instance must be byte-identical")
	assert.Equal(t, first, third, "Encode across instances with same key must be byte-identical")
}
