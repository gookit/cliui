package ui

type listWindow struct {
	start int
	end   int
}

func visibleWindow(total, cursorPos, pageSize int) listWindow {
	if total <= 0 {
		return listWindow{}
	}
	if pageSize <= 0 || pageSize > total {
		pageSize = total
	}
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos >= total {
		cursorPos = total - 1
	}

	start := cursorPos - pageSize + 1
	if start < 0 {
		start = 0
	}
	if start+pageSize > total {
		start = total - pageSize
	}
	return listWindow{start: start, end: start + pageSize}
}

func indexPosition(indexes []int, itemIndex int) int {
	for i, idx := range indexes {
		if idx == itemIndex {
			return i
		}
	}
	return -1
}

func firstEnabledFilteredIndex(items []Item, indexes []int) int {
	for _, idx := range indexes {
		if !items[idx].Disabled {
			return idx
		}
	}
	if len(indexes) > 0 {
		return indexes[0]
	}
	return 0
}

func moveFilteredCursor(items []Item, indexes []int, cursor, delta int) int {
	if len(indexes) == 0 {
		return cursor
	}

	pos := indexPosition(indexes, cursor)
	if pos < 0 {
		return firstEnabledFilteredIndex(items, indexes)
	}

	next := pos
	for range indexes {
		next = (next + delta + len(indexes)) % len(indexes)
		idx := indexes[next]
		if !items[idx].Disabled {
			return idx
		}
	}

	return cursor
}
