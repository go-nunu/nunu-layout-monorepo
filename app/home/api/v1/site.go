package v1

type MetaResponse struct {
	Response
	Data MetaData `json:"data"`
}

type HealthResponse struct {
	Response
	Data HealthData `json:"data"`
}

type ManifestResponse struct {
	Response
	Data ManifestData `json:"data"`
}
