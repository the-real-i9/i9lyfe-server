package utilServices

import (
	"regexp"
)

func ExtractHashtags(description string) []string {
	re := regexp.MustCompile("#[[:alnum:]][[:alnum:]_]+[[:alnum:]]+")

	matches := re.FindAllString(description, -1)

	res := make([]string, len(matches))

	for i, m := range matches {
		res[i] = m[1:]
	}

	return res
}

func ExtractMentions(description string) []string {
	re := regexp.MustCompile("@[[:alnum:]][[:alnum:]_-]+[[:alnum:]]+")

	matches := re.FindAllString(description, -1)

	res := make([]string, len(matches))

	for i, m := range matches {
		res[i] = m[1:]
	}

	return res
}
