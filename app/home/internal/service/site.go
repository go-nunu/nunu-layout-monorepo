package service

import apiV1 "nunu-layout-monorepo/app/home/api/v1"

type SiteService interface {
	Health() apiV1.HealthData
	Meta() apiV1.MetaData
	Manifest() apiV1.ManifestData
}

type siteService struct {
	*Service
}

func NewSiteService(service *Service) SiteService {
	return &siteService{Service: service}
}

func (s *siteService) Health() apiV1.HealthData {
	return apiV1.HealthData{
		App:    "home",
		Status: "ok",
	}
}

func (s *siteService) Meta() apiV1.MetaData {
	return apiV1.MetaData{
		App:   "home",
		Stage: s.config.GetString("env"),
		Entry: "app/home/cmd/server",
		Title: siteTitle(s.config.GetString("site.title")),
	}
}

func (s *siteService) Manifest() apiV1.ManifestData {
	return apiV1.ManifestData{
		App:      "home",
		Stage:    s.config.GetString("env"),
		Entry:    "app/home/cmd/server",
		Title:    siteTitle(s.config.GetString("site.title")),
		Headline: siteHeadline(s.config.GetString("site.headline")),
		Features: []apiV1.ManifestFeature{
			{
				Name:        "Public Shell",
				Description: "Serve public pages and lightweight shell APIs from app/home.",
				Route:       "/",
			},
			{
				Name:        "Runtime Health",
				Description: "Expose a small health endpoint for local development and deploy checks.",
				Route:       "/healthz",
			},
			{
				Name:        "Bootstrap Manifest",
				Description: "Return route and product metadata for future front-end bootstrapping.",
				Route:       "/api/v1/manifest",
			},
		},
	}
}

func siteTitle(title string) string {
	if title == "" {
		return "Home App"
	}

	return title
}

func siteHeadline(headline string) string {
	if headline == "" {
		return "Public-facing shell, ready for real product work."
	}

	return headline
}
