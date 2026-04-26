package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type MetaData struct {
	App   string `json:"app"`
	Stage string `json:"stage"`
	Entry string `json:"entry"`
	Title string `json:"title"`
}

type HealthData struct {
	App    string `json:"app"`
	Status string `json:"status"`
}

type ManifestFeature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Route       string `json:"route"`
}

type ManifestData struct {
	App      string            `json:"app"`
	Stage    string            `json:"stage"`
	Entry    string            `json:"entry"`
	Title    string            `json:"title"`
	Headline string            `json:"headline"`
	Features []ManifestFeature `json:"features"`
}

func HandleSuccess(ctx *gin.Context, data interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}

	ctx.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

func HandleError(ctx *gin.Context, httpCode int, message string, data interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}

	ctx.JSON(httpCode, Response{Code: httpCode, Message: message, Data: data})
}
