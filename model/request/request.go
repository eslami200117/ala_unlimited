package request

type Request struct {
	DKP    string   `json:"product_id"`
	Colors []string `json:"colors"`
}
