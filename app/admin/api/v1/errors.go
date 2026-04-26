package v1

var (
	// common errors
	ErrSuccess             = newError(0, "ok")
	ErrBadRequest          = newError(400, "参数错误")
	ErrUnauthorized        = newError(401, "登录失效，请重新登录~")
	ErrNotFound            = newError(404, "数据不存在")
	ErrForbidden           = newError(403, "权限不足，请联系管理员开通权限~")
	ErrInternalServerError = newError(500, "服务器错误~")

	// more biz errors
	ErrUsernameAlreadyUse = newError(1001, "The username is already in use.")
	ErrRoleAlreadyUse     = newError(1002, "The role sid is already in use.")
)
