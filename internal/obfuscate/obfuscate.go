// Package obfuscate provides view-source-grade obfuscation of build-time
// strings. It is not cryptography: the goal is to defeat casual Ctrl-U
// inspection of generated HTML, not to resist an attacker who runs the
// page's JavaScript or edits localStorage.
//
// The byte-level contract is: XOR the UTF-8 plaintext bytes with the UTF-8
// key bytes (key index = i mod len(key)), then standard base64-encode the
// result. The matching JavaScript decoder in themes/base/static/progress.js
// performs the inverse and MUST stay byte-compatible with this routine.
package obfuscate

import "encoding/base64"

// Encoder applies the package's XOR+base64 obfuscation using a fixed key.
// When constructed with an empty key, Encode is a passthrough (dev mode).
type Encoder struct {
	key []byte
}

// New returns an Encoder that uses key as the XOR key. An empty key puts
// the Encoder in passthrough mode: Encode returns its input unchanged.
func New(key string) *Encoder {
	return &Encoder{key: []byte(key)}
}

// Encode returns the obfuscated form of plaintext. With an empty key the
// input is returned unchanged.
func (e *Encoder) Encode(plaintext string) string {
	if len(e.key) == 0 {
		return plaintext
	}
	src := []byte(plaintext)
	out := make([]byte, len(src))
	for i, b := range src {
		out[i] = b ^ e.key[i%len(e.key)]
	}
	return base64.StdEncoding.EncodeToString(out)
}
