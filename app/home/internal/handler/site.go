package handler

import (
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
	apiV1 "nunu-layout-monorepo/app/home/api/v1"
	"nunu-layout-monorepo/app/home/internal/service"
)

type SiteHandler struct {
	*Handler
	siteService service.SiteService
	indexFile   string
}

func NewSiteHandler(handler *Handler, siteService service.SiteService) *SiteHandler {
	return &SiteHandler{
		Handler:     handler,
		siteService: siteService,
		indexFile:   resolveHomeIndexFile(),
	}
}

func (h *SiteHandler) Index(ctx *gin.Context) {
	ctx.File(h.indexFile)
}

func (h *SiteHandler) Health(ctx *gin.Context) {
	apiV1.HandleSuccess(ctx, h.siteService.Health())
}

func (h *SiteHandler) Meta(ctx *gin.Context) {
	apiV1.HandleSuccess(ctx, h.siteService.Meta())
}

func (h *SiteHandler) Manifest(ctx *gin.Context) {
	apiV1.HandleSuccess(ctx, h.siteService.Manifest())
}

func resolveHomeIndexFile() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Clean("app/home/web/index.html")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../web/index.html"))
}
