package config

const MaxMessagesPerRequest = 50

const (
	KiB = 1 << 10
	MiB = 1 << 20
)

const (
	GIFChunkSize         = 512 * KiB // 0.5 MiB
	AttachmentChunksSize = 2 * MiB   // 2 MiB
)

var SupportedImageExtensions = []string{"jpg", "jpeg", "png", "webp"}
