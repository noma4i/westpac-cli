package utils

var MaskMode bool

func MaskText(s string) string {
	return maskFrom(s, 0)
}

func MaskPartial(s string, keepFirst int) string {
	return maskFrom(s, keepFirst)
}

func MaskStars(s string) string {
	if !MaskMode || s == "" {
		return s
	}
	runes := []rune(s)
	for i, r := range runes {
		if r != ' ' && r != '-' {
			runes[i] = '*'
		}
	}
	return string(runes)
}

func maskFrom(s string, keepFirst int) string {
	if !MaskMode || s == "" {
		return s
	}
	runes := []rune(s)
	for i, r := range runes {
		if i < keepFirst {
			continue
		}
		switch {
		case r >= 'A' && r <= 'Z':
			runes[i] = '\u2588'
		case r >= 'a' && r <= 'z':
			runes[i] = '\u2593'
		case r >= '0' && r <= '9':
			runes[i] = '\u2592'
		}
	}
	return string(runes)
}
