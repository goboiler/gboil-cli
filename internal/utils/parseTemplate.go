package utils

import "regexp"

func ParseTemplate(template string) string {
	official := regexp.MustCompile(`^[a-zA-Z\d\_\-]+$`)
	unofficial := regexp.MustCompile(`^([a-zA-Z\d\_\-]+\/)+[a-zA-Z\d\_\-]+$`)

	if official.MatchString(template) {
		return "https://github.com/goboiler/templates/tree/main/" + template
	} else if unofficial.MatchString(template) {
		return "https://github.com/" + template
	} else {
		panic("Invalid Template")
	}
}
