package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Data         interface{} `json:"data"`
	ErrCode      int         `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
}

func Success(data interface{}, c *gin.Context) {
	Result(0, data, "", c)
}

func Result(code int, data interface{}, msg string, c *gin.Context) {
	result := Response{
		data,
		code,
		msg,
	}
	c.JSON(http.StatusOK, result)
}

func ParamError(message string, c *gin.Context) {
	Result(2, "", message, c)
}
