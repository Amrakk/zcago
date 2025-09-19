package session

import (
	"net/http"
)

type Options struct {
	SelfListen  bool
	CheckUpdate bool
	Logging     bool
	APIType     uint
	APIVersion  uint

	Client *http.Client

	ImageMetadataGetter ImageMetadataGetter
}

func ApplyOptions(input *Options) Options {
	defaults := defaultInternalOptions()

	if input == nil {
		return defaults
	}

	result := defaults
	if input.SelfListen {
		result.SelfListen = input.SelfListen
	}
	result.CheckUpdate = input.CheckUpdate
	result.Logging = input.Logging
	if input.APIType != 0 {
		result.APIType = input.APIType
	}
	if input.APIVersion != 0 {
		result.APIVersion = input.APIVersion
	}
	if input.ImageMetadataGetter != nil {
		result.ImageMetadataGetter = input.ImageMetadataGetter
	}

	return result
}

func defaultInternalOptions() Options {
	return Options{
		SelfListen:  false,
		CheckUpdate: true,
		Logging:     true,
		APIType:     30,
		APIVersion:  665,
	}
}

type ImageMetadataGetterResponse struct {
	Width  uint `json:"width"`
	Height uint `json:"height"`
	Size   uint `json:"size"`
}

type ImageMetadataGetter func(filePath string) (*ImageMetadataGetterResponse, error)
