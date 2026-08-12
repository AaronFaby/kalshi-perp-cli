package cli

import "fmt"

const maxPaginationPages = 100

// continueCursor returns the next cursor, or "" when paging should stop.
func continueCursor(all bool, prev, next string, page int) (string, error) {
	if !all || next == "" {
		return "", nil
	}
	if next == prev {
		return "", fmt.Errorf("pagination cursor did not advance")
	}
	if page >= maxPaginationPages {
		return "", fmt.Errorf("pagination exceeded %d pages; pass a larger --limit or a --cursor", maxPaginationPages)
	}
	return next, nil
}
