package agent

import (
	"regexp"
	"strings"

	"github.com/orchestra/backend/internal/models"
)

// mentionPattern matches @name sequences (non-space, non-punctuation after @).
var mentionPattern = regexp.MustCompile(`@(\S+)`)

// ParseMentions extracts member IDs from @mention patterns in the content.
// Returns the IDs of mentioned members. "@all" returns all member IDs.
func ParseMentions(content string, members []models.Member) []string {
	matches := mentionPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	// Build name→ID index (case-insensitive)
	nameIndex := make(map[string]string, len(members))
	for _, m := range members {
		nameIndex[strings.ToLower(m.Name)] = m.ID
	}

	seen := make(map[string]struct{})
	var result []string

	for _, match := range matches {
		name := match[1]
		if strings.ToLower(name) == "all" {
			// @all → all members
			for _, m := range members {
				if _, ok := seen[m.ID]; !ok {
					seen[m.ID] = struct{}{}
					result = append(result, m.ID)
				}
			}
			continue
		}

		id, ok := nameIndex[strings.ToLower(name)]
		if !ok {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}

	return result
}
