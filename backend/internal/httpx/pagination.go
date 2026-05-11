package httpx

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

var ErrInvalidPagination = errors.New("invalid pagination")

func PageParams(q url.Values, defaultPageSize int) (page, pageSize int, err error) {
	page, err = parsePositiveInt(q.Get("page"), 1, "page")
	if err != nil {
		return 0, 0, err
	}
	pageSize, err = parsePositiveInt(q.Get("pageSize"), defaultPageSize, "pageSize")
	if err != nil {
		return 0, 0, err
	}

	return page, pageSize, nil
}

func parsePositiveInt(raw string, def int, name string) (int, error) {
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: %s must be a positive integer", ErrInvalidPagination, name)
	}
	if v < 1 {
		return 0, fmt.Errorf("%w: %s must be >= 1", ErrInvalidPagination, name)
	}

	return v, nil
}
