package service

const (
	maxPageSize               = 100
	defaultAssistantsPageSize = 10
	defaultRunsPageSize       = 20
)

func normalizePage(page, pageSize, defaultSize, maxSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}

	return page, pageSize
}
