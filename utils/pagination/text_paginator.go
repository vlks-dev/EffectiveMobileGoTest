package pagination

import (
	"errors"
	"strconv"
	"strings"
)

// SplitTextIntoVerses Разделить текст на куплеты
func SplitTextIntoVerses(text string) []string {
	return strings.Split(text, "\n\n")
}

// PaginateVerses Пагинация куплетов
func PaginateVerses(verses []string, page string) ([]string, error) {
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		return nil, errors.New("invalid page number")
	}

	versesPerPage := 1 // Количество куплетов на странице
	start := (pageNum - 1) * versesPerPage
	if start >= len(verses) {
		return nil, errors.New("page out of range")
	}

	end := start + versesPerPage
	if end > len(verses) {
		end = len(verses)
	}

	return verses[start:end], nil
}
