package session

import (
	"net/http"
)

type Option func(*options)

type options struct {
	selfListen  bool
	checkUpdate bool
	logging     bool
	logLevel    uint8
	apiType     uint
	apiVersion  uint

	client *http.Client

	imageMetadataGetter ImageMetadataGetter
}

func WithSelfListen(v bool) Option {
	return func(o *options) { o.selfListen = v }
}

func WithCheckUpdate(v bool) Option {
	return func(o *options) { o.checkUpdate = v }
}
func WithLogging(v bool) Option {
	return func(o *options) { o.logging = v }
}

func WithLogLevel(level uint8) Option {
	return func(o *options) { o.logLevel = level }
}
func WithAPIType(t uint) Option {
	return func(o *options) {
		if t != 0 {
			o.apiType = t
		}
	}
}
func WithAPIVersion(v uint) Option {
	return func(o *options) {
		if v != 0 {
			o.apiVersion = v
		}
	}
}
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) { o.client = c }
}
func WithImageMetadataGetter(f ImageMetadataGetter) Option {
	return func(o *options) { o.imageMetadataGetter = f }
}

func defaultOptions() options {
	return options{
		selfListen:  false,
		checkUpdate: true,
		logging:     true,
		logLevel:    1, // Debug level by default
		apiType:     30,
		apiVersion:  665,
	}
}

type ImageMetadataGetterResponse struct {
	Width  uint `json:"width"`
	Height uint `json:"height"`
	Size   uint `json:"size"`
}

type ImageMetadataGetter func(filePath string) (*ImageMetadataGetterResponse, error)
