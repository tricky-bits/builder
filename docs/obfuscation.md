# Obfuscation

Stage metadata, hints, and completion messages can be obfuscated in the
rendered HTML so puzzle content isn't trivially readable in page source.

!!! warning "Not cryptography"
    This is view-source-grade obfuscation, not security. It defeats casual
    <kbd>Ctrl</kbd>+<kbd>U</kbd> inspection of the generated HTML, it does
    **not** resist an attacker who runs the page's JavaScript or inspects
    values at runtime. Don't rely on it to protect anything sensitive.

    The goal is just to hide the information a little, not to police
    players. Someone who wants to "hack" their way to an answer and enjoys
    that is having exactly as much fun as intended, let them.

## How it works

The build-time obfuscator XORs the UTF-8 plaintext bytes with the UTF-8
`obfuscation_key` bytes (repeating the key as needed), then base64-encodes
the result. A matching JavaScript decoder shipped in the theme performs the
inverse in the browser.

Set `obfuscation_key = ""` in `config.toml` to disable obfuscation entirely
(passthrough), useful during local development when you want to read stage
data straight from the rendered HTML.

## Answers are separate

Stage answers don't use this XOR scheme at all. Only a SHA-256 hash of the
trimmed, lowercased answer ever reaches the client, see
[Stages](stages.md#answers) for details.
