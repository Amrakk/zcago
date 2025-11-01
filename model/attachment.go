package model

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Amrakk/zcago/config"
	"github.com/Amrakk/zcago/errs"
)

type FileType string

const (
	FileTypeImage FileType = "image"
	FileTypeVideo FileType = "video"
	FileTypeGif   FileType = "gif"
	FileTypeOther FileType = "others"
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

func (s AttachmentSource) IsString() bool { return s.str != nil }
func (s AttachmentSource) IsObject() bool { return s.obj != nil }
func (s AttachmentSource) String() string {
	if s.str == nil {
		return ""
	}
	return *s.str
}

func (s AttachmentSource) Object() *AttachmentObject {
	if s.obj == nil {
		return nil
	}
	return s.obj
}

func (s AttachmentSource) GetExtension() string {
	var name string
	switch {
	case s.IsString():
		name = s.String()
	case s.IsObject():
		name = s.Object().Filename
	default:
		return ""
	}
	if name == "" {
		return ""
	}
	ext := filepath.Ext(name)
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}
	return strings.ToLower(ext)
}

type MD5ChecksumResult struct {
	CurrentChunk int
	Checksum     string
}

func (s *AttachmentSource) GetLargeFileMD5() *MD5ChecksumResult {
	var (
		reader io.Reader
		closer io.Closer
	)

	defer func() {
		if closer != nil {
			_ = closer.Close()
		}
	}()

	if f := s.String(); f != "" {
		r, err := os.Open(f)
		if err != nil {
			return nil
		}

		reader = r
		closer = r
	} else if f := s.Object(); f != nil {
		reader = f.Data
	}

	h := md5.New()
	buf := make([]byte, config.AttachmentChunksSize)
	chunks := 0

	for {
		n, err := io.ReadFull(reader, buf)
		switch err {
		case nil:
			_, _ = h.Write(buf[:n])
			chunks++
		case io.ErrUnexpectedEOF:
			if n > 0 {
				_, _ = h.Write(buf[:n])
				chunks++
			}
			sum := hex.EncodeToString(h.Sum(nil))
			return &MD5ChecksumResult{CurrentChunk: chunks, Checksum: sum}
		case io.EOF:
			sum := hex.EncodeToString(h.Sum(nil))
			return &MD5ChecksumResult{CurrentChunk: chunks, Checksum: sum}
		default:
			return nil
		}
	}
}

type UploadAttachment struct {
	FileID  string `json:"fileId"`
	FileURL string `json:"fileUrl"`
}

func NewUploadAttachment(fileID, fileURL string) UploadAttachment {
	return UploadAttachment{FileID: fileID, FileURL: fileURL}
}
