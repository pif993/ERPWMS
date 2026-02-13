package pii

import "regexp"

var emailRe = regexp.MustCompile(`([a-zA-Z0-9._%+-]{2})[a-zA-Z0-9._%+-]*(@.*)`)

func MaskEmail(v string) string {
	return emailRe.ReplaceAllString(v, "$1***$2")
}
