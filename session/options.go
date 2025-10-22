package session

import (
	"net/http"

	"github.com/Amrakk/zcago/model"
)

type Option func(*options)

type OptionsSnapshot struct {
	SelfListen          bool
	CheckUpdate         bool
	Logging             bool
	LogLevel            uint8
	APIType             uint
	APIVersion          uint
	Client              *http.Client
	ImageMetadataGetter ImageMetadataGetter
}

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

func WithSelfListen(v bool) Option         { return func(o *options) { o.selfListen = v } }
func WithCheckUpdate(v bool) Option        { return func(o *options) { o.checkUpdate = v } }
func WithLogging(v bool) Option            { return func(o *options) { o.logging = v } }
func WithLogLevel(level uint8) Option      { return func(o *options) { o.logLevel = level } }
func WithHTTPClient(c *http.Client) Option { return func(o *options) { o.client = c } }

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
		client:      http.DefaultClient,
	}
}

type ImageMetadataGetter func(filePath string) (model.AttachmentMetadata, error)
