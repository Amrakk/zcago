package config

import (
	"net/url"
	"time"
)

const (
	DefaultUserAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0"
	DefaultLanguage          = "vi"
	DefaultAPIType           = 30
	DefaultAPIVersion        = 665
	DefaultComputerName      = "Web"
	DefaultUploadCallbackTTL = 5 * time.Minute

	DefaultQRPath  = "qr.png"
	DefaultUIDSelf = "0"

	DefaultEncryptVersion = "v2"
	DefaultZCIDKey        = "3FC4F0D2AB50057BCE0D90D9187A22B1"

	MaxMessagesPerRequest = 50
	MaxRedirects          = 10
)

var DefaultURL = url.URL{Scheme: "https", Host: "chat.zalo.me"}

// ----------------------------------------
// Attachment
// ----------------------------------------

const (
	KiB = 1 << 10
	MiB = 1 << 20
)

const (
	GIFChunkSize         = 512 * KiB // 0.5 MiB
	AttachmentChunksSize = 2 * MiB   // 2 MiB
)

var SupportedImageExtensions = []string{"jpg", "jpeg", "png", "webp"}
