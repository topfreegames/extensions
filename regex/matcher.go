package regex

import "regexp"

// PrivateIPRegex matches private ips in RFC1918 private IPV4 address ranges
const PrivateIPRegex = "^(?:10|127|172\\.(?:1[6-9]|2[0-9]|3[01])|192\\.168)\\..*"

// IsPrivateIP returns whether a given ip is private or not
func IsPrivateIP(ip string) bool {
	match, _ := regexp.MatchString(PrivateIPRegex, ip)
	return match
}
