package request

type Request struct {
	DKP    int      `json:"product_id"`
	Colors []string `json:"colors"`
}
