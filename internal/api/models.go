package api

type LocateRequest struct {
	Text      string `json:"text"`
	Placement string `json:"placement"`
}
type ValidateRequest struct {
	Text string `json:"text"`
}
