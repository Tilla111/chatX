package store

import (
	"net/http"
	"strconv"
)

type PaginationQuery struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=20"`
	Offset int    `json:"offset" validate:"gte=0"`
	Search string `json:"search" validate:"max=10"`
}

func (pg PaginationQuery) Parse(r *http.Request) (*PaginationQuery, error) {

	limit := r.URL.Query().Get("limit")

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return nil, err
		}

		pg.Limit = l
	}

	offset := r.URL.Query().Get("offset")

	if offset != "" {
		l, err := strconv.Atoi(offset)
		if err != nil {
			return nil, err
		}

		pg.Offset = l
	}

	search := r.URL.Query().Get("search")

	if search != "" {
		pg.Search = search
	}

	return &pg, nil
}
