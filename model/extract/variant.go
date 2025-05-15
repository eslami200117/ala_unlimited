package extract

type Variant struct {
	Seller         string `json:"seller"`
	SellerID       int    `json:"seller_id"`
	Price          int    `json:"price"`
	VarWinner      bool   `json:"var_winner"`
	BuyBoxSellerID int    `json:"buy_box_seller_id"`
	Promotion      bool   `json:"promotion"`
}

type ExtProductPrice struct {
	Status      int                   `json:"status"`
	DKP         int                   `json:"dkp"`
	Variants    map[string][]*Variant `json:"variants"`
	BuyBoxPrice int                   `json:"buy_box_price"`
}
