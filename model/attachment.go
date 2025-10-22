package model

import (
	"io"
	"strings"

	"github.com/Amrakk/zcago/errs"
)

type AttachmentSource struct {
	str *string
	obj *AttachmentObject
}

type AttachmentObject struct {
	Data     io.Reader
	Filename string
	Metadata AttachmentMetadata
}

type AttachmentMetadata struct {
	Size   int64
	Width  int
	Height int
}

func NewStringAttachment(s string) AttachmentSource { return AttachmentSource{str: &s} }
func NewObjectAttachment(filename string, meta AttachmentMetadata, data io.Reader) (*AttachmentSource, error) {
	if !strings.Contains(filename, ".") {
		return nil, errs.NewZCA("filename must include an extension", "NewObjectAttachment")
	}
	o := &AttachmentObject{Data: data, Filename: filename, Metadata: meta}
	return &AttachmentSource{obj: o}, nil
}

func (a AttachmentSource) IsString() bool { return a.str != nil }
func (a AttachmentSource) IsObject() bool { return a.obj != nil }
func (a AttachmentSource) String() string {
	if a.str == nil {
		return ""
	}
	return *a.str
}

func (a AttachmentSource) Object() *AttachmentObject {
	if a.obj == nil {
		return nil
	}
	return a.obj
}
