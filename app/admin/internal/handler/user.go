package handler

import (
	"github.com/gin-gonic/gin"
	"nunu-layout-monorepo/app/admin/internal/service"
)

type UserHandler struct {
	*Handler
	userService service.UserService
}

func NewUserHandler(
	handler *Handler,
	userService service.UserService,
) *UserHandler {
	return &UserHandler{
		Handler:     handler,
		userService: userService,
	}
}

func (h *UserHandler) GetUsers(ctx *gin.Context) {

}
