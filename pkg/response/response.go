package response

import (
	"log"
	"net/http"
	"warehouse-management/pkg/apperror"

	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{
		"data": data,
	})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Paginated(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"meta": meta,
	})
}

func Error(c *gin.Context, err error) {
	if appErr, ok := apperror.As(err); ok {
		c.AbortWithStatusJSON(appErr.Status, appErr)
		return
	}

	log.Printf("Internal server error: %v", err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"message": "something went wrong",
	})
}
