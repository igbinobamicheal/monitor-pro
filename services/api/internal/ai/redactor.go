package ai

import (
	"regexp"
)

var (
	ipv4Regex    = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	ipv6Regex    = regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`)
	macRegex     = regexp.MustCompile(`\b(?:[0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}\b`)
	emailRegex   = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	winUserRegex = regexp.MustCompile(`(?i)(?:C:\\Users\\)([a-zA-Z0-9_.-]+)`)
	tokenRegex   = regexp.MustCompile(`(?i)(bearer\s+[A-Za-z0-9\-._~+/]+=*|token\s*[=:]\s*["']?[A-Za-z0-9\-._~+/]+["']?)`)
)

// RedactPII redacts sensitive identifiers from arbitrary text before advisory ingestion
func RedactPII(input string) string {
	if input == "" {
		return ""
	}
	res := ipv4Regex.ReplaceAllString(input, "[REDACTED_IP]")
	res = ipv6Regex.ReplaceAllString(res, "[REDACTED_IPV6]")
	res = macRegex.ReplaceAllString(res, "[REDACTED_MAC]")
	res = emailRegex.ReplaceAllString(res, "[REDACTED_EMAIL]")
	res = winUserRegex.ReplaceAllString(res, `C:\Users\[REDACTED_USER]`)
	res = tokenRegex.ReplaceAllString(res, "[REDACTED_TOKEN]")
	return res
}
