package model

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/Amrakk/zcago/errs"
)

type AttachmentSource struct {
	str *string
	obj *AttachmentObject
}

type AttachmentObject struct {
	Data     []byte             `json:"data"`
	Filename string             `json:"filename"`
	Metadata AttachmentMetadata `json:"metadata"`
}

type AttachmentMetadata struct {
	TotalSize int64 `json:"totalSize"`
	Width     *int  `json:"width,omitempty"`
	Height    *int  `json:"height,omitempty"`
}

func NewStringAttachment(s string) AttachmentSource { return AttachmentSource{str: &s} }
func NewObjectAttachment(data []byte, filename string, meta AttachmentMetadata) (*AttachmentSource, error) {
	if !strings.Contains(filename, ".") {
		return nil, errs.NewZCA("filename must include an extension", "NewObjectAttachment")
	}
	o := &AttachmentObject{Data: data, Filename: filename, Metadata: meta}
	return &AttachmentSource{obj: o}, nil
}

func (a AttachmentSource) IsString() bool { return a.str != nil }
func (a AttachmentSource) IsObject() bool { return a.obj != nil }
func (a AttachmentSource) StringValue() (string, bool) {
	if a.str == nil {
		return "", false
	}
	return *a.str, true
}

func (a AttachmentSource) ObjectValue() (*AttachmentObject, bool) {
	if a.obj == nil {
		return nil, false
	}
	return a.obj, true
}

func (a AttachmentSource) MarshalJSON() ([]byte, error) {
	switch {
	case a.str != nil && a.obj != nil:
		return nil, errs.NewZCA("both str and obj are set", "AttachmentSource.MarshalJSON")
	case a.str != nil:
		return json.Marshal(*a.str)
	case a.obj != nil:
		return json.Marshal(a.obj)
	default:
		return []byte("null"), nil
	}
}

func (a *AttachmentSource) UnmarshalJSON(b []byte) error {
	trim := bytes.TrimSpace(b)
	if len(trim) == 0 || bytes.Equal(trim, []byte("null")) {
		*a = AttachmentSource{}
		return nil
	}
	switch trim[0] {
	case '"':
		var s string
		if err := json.Unmarshal(trim, &s); err != nil {
			return err
		}
		*a = AttachmentSource{str: &s}
		return nil
	case '{':
		var o AttachmentObject
		if err := json.Unmarshal(trim, &o); err != nil {
			return err
		}

		if o.Filename == "" || !strings.Contains(o.Filename, ".") {
			return errs.NewZCA("invalid filename", "AttachmentSource.UnmarshalJSON")
		}
		*a = AttachmentSource{obj: &o}
		return nil
	default:
		return errs.NewZCA("value must be a string or an object", "AttachmentSource.UnmarshalJSON")
	}
}
