package common

import (
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type QueryParams struct {
	Page   int
	Limit  int
	Search string
	Sort   string
	Order  string
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

	search := strings.TrimSpace(c.Query("search"))

	sort := strings.ToLower(strings.TrimSpace(c.Query("sort")))

	allowedSortFields := map[string]string{
		"id":           "id",
		"name":         "website_name",
		"website_name": "website_name",
		"domain":       "domain",
		"created_at":   "created_at",
		"title":        "title",
		"position":     "position",
		"status":       "status",
		"start_time":   "start_time",
		"end_time":     "end_time",
		"platform":     "platform",
	}

	if mappedSort, exists := allowedSortFields[sort]; exists {
		sort = mappedSort
	} else {
		sort = "id" // Default sort field if missing or invalid
	}

	order := strings.ToLower(strings.TrimSpace(c.Query("order")))
	if order != "asc" && order != "desc" {
		order = "desc" // Default sort order
	}

	return QueryParams{
		Page:   page,
		Limit:  limit,
		Search: search,
		Sort:   sort,
		Order:  order,
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
