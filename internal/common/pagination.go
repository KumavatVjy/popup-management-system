package common

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type QueryParams struct {
	Page  int
	Limit int
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func GetQueryParams(c *gin.Context) QueryParams {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	return QueryParams{
		Page:  page,
		Limit: limit,
	}
}

func BuildPagination(total int64, params QueryParams) Pagination {
	totalPages := 0
	if params.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(params.Limit)))
	}
	return Pagination{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
