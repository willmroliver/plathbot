package util

import "unicode/utf8"

func IsEmoji(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)

	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Miscellaneous Symbols and Pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map Symbols
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
		(r >= 0x1F1E6 && r <= 0x1F1FF) || // Flags (regional indicators)
		(r >= 0x2600 && r <= 0x26FF) || // Miscellaneous Symbols (e.g., ♥, ☀, ☔)
		(r >= 0x2700 && r <= 0x27BF) // Dingbats (e.g., ✂, ✈, ✉)
}
