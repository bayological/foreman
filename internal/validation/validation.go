package validation

import (
	"regexp"
	"strings"
)

// Valid branch name pattern: alphanumeric, hyphens, underscores, slashes
// Must not contain ".." to prevent path traversal
var branchNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9/_-]*$`)

// UUID pattern for task IDs
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsValidBranchName validates that a branch name is safe to use in file paths
// and git commands. Prevents path traversal attacks.
func IsValidBranchName(name string) bool {
	if name == "" || len(name) > 200 {
		return false
	}

	// Check for path traversal attempts
	if strings.Contains(name, "..") {
		return false
	}

	// Check against valid pattern
	return branchNameRegex.MatchString(name)
}

// IsValidUUID validates that a string is a valid UUID format.
func IsValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

// SanitizeErrorMessage removes potentially sensitive information from error messages
// before sending them to external systems like Telegram.
func SanitizeErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()

	// Remove absolute paths (Unix and Windows style)
	pathPatterns := []string{
		`/home/[^\s:]+`,
		`/Users/[^\s:]+`,
		`/var/[^\s:]+`,
		`/tmp/[^\s:]+`,
		`C:\\[^\s:]+`,
	}

	for _, pattern := range pathPatterns {
		re := regexp.MustCompile(pattern)
		msg = re.ReplaceAllString(msg, "[path]")
	}

	// Truncate long messages
	if len(msg) > 500 {
		msg = msg[:497] + "..."
	}

	return msg
}
