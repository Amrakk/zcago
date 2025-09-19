package auth

type LoginQROption struct {
	UserAgent *string `json:"userAgent,omitempty"`
	Language  *string `json:"language,omitempty"`
	QRPath    *string `json:"qrPath,omitempty"`
}

type LoginQRCallback func(event any) (any, error)
