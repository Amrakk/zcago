package model

type CatalogItem struct {
	ID        string `json:"id"`
	OwnerID   string `json:"ownerId"`
	Name      string `json:"name"`
	Version   int    `json:"version"`
	IsDefault bool   `json:"isDefault"`

	// Relative path used to build the catalog URL
	//
	// Example: https://catalog.zalo.me/{path}
	Path         string  `json:"path"`
	CatalogPhoto *string `json:"catalogPhoto"`
	TotalProduct int     `json:"totalProduct"`
	CreatedTime  int64   `json:"created_time"`
}

type ProductCatalogItem struct {
	ProductID   string `json:"product_id"`
	CatalogID   string `json:"catalog_id"`
	OwnerID     string `json:"owner_id"`
	Price       string `json:"price"`
	Description string `json:"description"`

	// Relative path used to build the product URL
	//
	// Example: https://catalog.zalo.me/{path}
	Path          string   `json:"path"`
	ProductName   string   `json:"product_name"`
	CurrencyUnit  string   `json:"currency_unit"`
	ProductPhotos []string `json:"product_photos"`
	CreateTime    int64    `json:"create_time"`
}
