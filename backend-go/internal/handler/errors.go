package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
)

// userFacingError returns an error message suitable for display to end users.
// In release mode, internal error details are hidden behind a generic message
// to avoid leaking implementation details (column names, file paths, etc.).
// In debug/test mode, the original error message is preserved for development.
func userFacingError(err error) string {
	if gin.Mode() == gin.ReleaseMode {
		return "Ocorreu um erro interno. Tente novamente mais tarde."
	}
	return err.Error()
}

// respondInternalError sends a 500 JSON response with a sanitized error message.
// Use this for all internal server errors in handlers.
func respondInternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"message": userFacingError(err)})
}

// respondListOrPaginated writes a JSON list response with optional pagination.
// When the request includes paginated=true query params, listPaginated is called
// and the response includes pagination metadata; otherwise listAll is used.
func respondListOrPaginated[T any](
	c *gin.Context,
	listAll func() ([]T, error),
	listPaginated func(offset, limit int) ([]T, int64, error),
) {
	var params dto.PaginationParams
	if err := c.ShouldBindQuery(&params); err == nil && params.Paginated {
		params.Normalize()
		items, total, err := listPaginated(params.Offset(), params.PageSize)
		if err != nil {
			respondInternalError(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.NewPaginatedResponse(items, total, params))
		return
	}

	items, err := listAll()
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}
