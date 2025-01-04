package util

var emojiNormMap = map[string]string{
	"\u2665": "\u2764\ufe0f",
	"\u2764": "\u2764\ufe0f",
}

// Normalize fixes certain emojis to a particular unicode where
// different platforms use different sequences.
//
// First use-case was the default red-heart emoji, which I found to be:
//   - U+2665 on macOS
//   - U+2764 for a TG react-emoji
//   - U+2764,U+FE0F on iOS
func NormalizeEmoji(e string) string {
	if ee, ok := emojiNormMap[e]; ok {
		return ee
	}

	return e
}
