package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/amrakk/zcago/config"
	"github.com/amrakk/zcago/errs"
	"github.com/amrakk/zcago/internal/httpx"
	"github.com/amrakk/zcago/internal/jsonx"
	"github.com/amrakk/zcago/model"
	"github.com/amrakk/zcago/session"
	"golang.org/x/sync/errgroup"
)

var (
	ErrMessageContentEmpty        = errs.NewZCA("message content cannot be empty", "api.SendMessage")
	ErrInvalidMention             = errs.NewZCA("Invalid mentions: total mention characters exceed message length", "api.SendMessage")
	ErrInvalidWebchatQuote        = errs.NewZCA("invalid quote: content must be string for msgType 'webchat'", "api.SendMessage")
	ErrUnsupportedQuotedGroupPoll = errs.NewZCA("quoted message type 'group.poll' is not supported", "api.SendMessage")
	ErrInvalidAttachmentUpload    = errs.NewZCA("attachment upload returned incomplete data", "api.SendMessage")
)

type TextStyle string

const (
	TextStyleBold          TextStyle = "b"
	TextStyleItalic        TextStyle = "i"
	TextStyleUnderline     TextStyle = "u"
	TextStyleStrikeThrough TextStyle = "s"

	TextStyleRed    TextStyle = "c_db342e"
	TextStyleOrange TextStyle = "c_f27806"
	TextStyleYellow TextStyle = "c_f7b503"
	TextStyleGreen  TextStyle = "c_15a85f"

	TextStyleSmall TextStyle = "f_13"
	TextStyleBig   TextStyle = "f_18"

	TextStyleOrderedList   TextStyle = "lst_2"
	TextStyleUnorderedList TextStyle = "lst_1"

	TextStyleIndent TextStyle = "ind_$"
)

type (
	SendMessageQuote struct {
		MsgID    string `json:"msgId"`
		CliMsgID string `json:"cliMsgId"`
		MsgType  string `json:"msgType"`
		UIDFrom  string `json:"uidFrom"`

		Content     model.Content      `json:"content"`
		PropertyExt *model.PropertyExt `json:"propertyExt,omitempty"`

		TS  string `json:"ts"`
		TTL int    `json:"ttl"`
	}
	MessageStyle struct {
		Start int       `json:"start"`
		Len   int       `json:"len"`
		Style TextStyle `json:"st"`

		IndentSize int `json:"indentSize"` // Used for indent style
	}
	MessageContent struct {
		Msg     string
		Style   []MessageStyle
		Urgency model.Urgency

		Quote       *SendMessageQuote
		Mentions    []model.TMention
		Attachments []model.AttachmentSource
		TTL         int // Time to live in milliseconds
	}

	sendData struct {
		URL     string
		Body    io.Reader
		Headers http.Header
	}
	attachmentSendPayload struct {
		Path   string
		Params map[string]any
	}

	SendMessageResult struct {
		MsgID string `json:"msgId"`
	}
	SendMessageResponse struct {
		Message    *SendMessageResult  `json:"message"`
		Attachment []SendMessageResult `json:"attachment"`
	}
	SendMessageFn = func(ctx context.Context, threadID string, threadType model.ThreadType, message MessageContent) (*SendMessageResponse, error)
)

func (a *api) SendMessage(ctx context.Context, threadID string, threadType model.ThreadType, message MessageContent) (*SendMessageResponse, error) {
	return a.e.SendMessage(ctx, threadID, threadType, message)
}

func prepareMentions(threadType model.ThreadType, msg string, mentions []model.TMention) ([]model.TMention, error) {
	if threadType != model.ThreadTypeGroup {
		return nil, nil
	}

	result := make([]model.TMention, 0, len(mentions))
	totalLen := 0
	for _, mention := range mentions {
		if mention.Pos < 0 || mention.UID == "" || mention.Len <= 0 {
			continue
		}
		if mention.UID == model.MentionAllUID {
			mention.Type = model.MentionAll
		} else {
			mention.Type = model.MentionEach
		}
		totalLen += mention.Len
		result = append(result, mention)
	}
	if totalLen > len(utf16.Encode([]rune(msg))) {
		return nil, ErrInvalidMention
	}
	return result, nil
}

func prepareAttachmentPayloads(
	uploads UploadAttachmentResponse,
	threadID string,
	threadType model.ThreadType,
	message MessageContent,
	mentions []model.TMention,
	canBeDescription bool,
	groupLayoutID string,
	clientID int64,
) ([]attachmentSendPayload, error) {
	isGroup := threadType == model.ThreadTypeGroup
	isMultiFile := len(uploads) > 1
	result := make([]attachmentSendPayload, 0, len(uploads))

	for index, upload := range uploads {
		var path string
		var payload map[string]any

		switch upload.FileType {
		case model.FileTypeImage:
			if upload.Image == nil {
				return nil, ErrInvalidAttachmentUpload
			}
			image := upload.Image
			payload = map[string]any{
				"photoId":  image.PhotoID,
				"clientId": strconv.FormatInt(clientID, 10),
				"desc":     message.Msg,
				"width":    image.Width,
				"height":   image.Height,
				"rawUrl":   image.NormalURL,
				"hdUrl":    image.HDURL,
				"thumbUrl": image.ThumbURL,
				"hdSize":   strconv.FormatInt(upload.TotalSize, 10),
				"zsource":  -1,
				"ttl":      message.TTL,
				"jcp":      `{"convertible":"jxl"}`,
			}
			clientID++
			if isGroup {
				payload["grid"] = threadID
				payload["oriUrl"] = image.NormalURL
			} else {
				payload["toid"] = threadID
				payload["normalUrl"] = image.NormalURL
			}
			if isMultiFile {
				payload["groupLayoutId"] = groupLayoutID
				payload["isGroupLayout"] = 1
				payload["idInGroup"] = index
				payload["totalItemInGroup"] = len(uploads)
				payload["extMsgProp"] = fmt.Sprintf(`{"groupMediaMsg":{"groupLayoutId":"%s"}}`, groupLayoutID)
			}
			if len(mentions) > 0 && canBeDescription && message.Quote == nil {
				payload["mentionInfo"] = jsonx.Stringify(mentions)
			}
			path = "photo_original/send"

		case model.FileTypeVideo, model.FileTypeOther:
			if upload.File == nil {
				return nil, ErrInvalidAttachmentUpload
			}
			file := upload.File
			payload = map[string]any{
				"fileId":      file.FileID,
				"checksum":    file.Checksum,
				"checksumSha": "",
				"extention":   strings.TrimPrefix(strings.ToLower(filepath.Ext(file.FileName)), "."),
				"totalSize":   upload.TotalSize,
				"fileName":    file.FileName,
				"clientId":    upload.ClientFileID,
				"fType":       1,
				"fileCount":   0,
				"fdata":       "{}",
				"fileUrl":     file.FileURL,
				"zsource":     -1,
				"ttl":         message.TTL,
			}
			if isGroup {
				payload["grid"] = threadID
			} else {
				payload["toid"] = threadID
			}
			path = "asyncfile/msg"

		default:
			return nil, ErrInvalidAttachmentUpload
		}

		if message.Urgency == model.UrgImportant || message.Urgency == model.UrgUrgent {
			payload["metaData"] = map[string]any{"urgency": message.Urgency}
		}
		result = append(result, attachmentSendPayload{Path: path, Params: payload})
	}

	return result, nil
}

var sendMessageFactory = apiFactory[*SendMessageResponse, SendMessageFn]()(
	func(a *api, sc session.Context, u factoryUtils[*SendMessageResponse]) (SendMessageFn, error) {
		fileBase := jsonx.FirstOr(sc.GetZpwService("file"), "")
		userBase := jsonx.FirstOr(sc.GetZpwService("chat"), "")
		groupBase := jsonx.FirstOr(sc.GetZpwService("group"), "")
		defaultParams := map[string]any{"nretry": 0}
		shareFile := sc.Settings().Features.ShareFile

		serviceURLs := struct {
			Message    map[model.ThreadType]string
			Attachment map[model.ThreadType]string
		}{
			Message: map[model.ThreadType]string{
				model.ThreadTypeUser:  u.MakeURL(userBase+"/api/message", defaultParams, true),
				model.ThreadTypeGroup: u.MakeURL(groupBase+"/api/group", defaultParams, true),
			},
			Attachment: map[model.ThreadType]string{
				model.ThreadTypeUser:  fileBase + "/api/message",
				model.ThreadTypeGroup: fileBase + "/api/group",
			},
		}

		isExceedMaxFile := func(totalFile int) bool {
			return totalFile > shareFile.MaxFile
		}

		handleStyles := func(payload map[string]any, mStyles []MessageStyle) {
			if len(mStyles) == 0 {
				return
			}

			styles := make([]map[string]any, 0, len(mStyles))
			for _, s := range mStyles {
				st := s.Style
				if st == TextStyleIndent {
					size := s.IndentSize
					if size <= 0 {
						size = 1
					}
					repl := fmt.Sprintf("%d0", size)
					st = TextStyle(strings.ReplaceAll(string(TextStyleIndent), "$", repl))
				}

				styles = append(styles, map[string]any{
					"start": s.Start,
					"len":   s.Len,
					"st":    st,
				})
			}

			payload["textProperties"] = jsonx.Stringify(map[string]any{
				"styles": styles,
				"ver":    0,
			})
		}

		handleQuoteMessage := func(payload map[string]any, quote *SendMessageQuote, isGroup bool) {
			payload["qmsgOwner"] = quote.UIDFrom
			payload["qmsgId"] = quote.MsgID
			payload["qmsgCliId"] = quote.CliMsgID
			payload["qmsgType"] = quote.GetMessageType()
			payload["qmsgTs"] = quote.TS
			payload["qmsgTTL"] = quote.TTL

			if quote.Content.String != nil {
				payload["qmsg"] = quote.Content.String
			} else {
				payload["qmsg"] = quote.BuildMessagePayload()
			}

			if isGroup {
				payload["qmsgAttach"] = jsonx.Stringify(quote.BuildAttachmentMessagePayload())
			}
		}

		handleMessage := func(threadID string, threadType model.ThreadType, message MessageContent) (*sendData, error) {
			quote := message.Quote
			isGroup := threadType == model.ThreadTypeGroup
			mentions, err := prepareMentions(threadType, message.Msg, message.Mentions)
			if err != nil {
				return nil, err
			}
			if message.Quote != nil {
				if quote.Content.String == nil && quote.MsgType == "webchat" {
					return nil, ErrInvalidWebchatQuote
				}

				if quote.MsgType == "group.poll" {
					return nil, ErrUnsupportedQuotedGroupPoll
				}
			}

			payload := map[string]any{
				"message":  message.Msg,
				"clientId": time.Now().UnixMilli(),
				"ttl":      message.TTL,
			}

			if isGroup {
				payload["grid"] = threadID
				payload["visibility"] = 0
				if len(mentions) > 0 {
					payload["mentionInfo"] = jsonx.Stringify(mentions)
				}
			} else {
				payload["toid"] = threadID
				payload["imei"] = sc.IMEI()
			}

			path := "/sendmsg"
			if quote != nil {
				handleQuoteMessage(payload, quote, isGroup)
				path = "/quote"
			} else if !isGroup {
				path = "/sms"
			} else if len(mentions) > 0 {
				path = "/mention"
			}

			if message.Urgency == model.UrgImportant || message.Urgency == model.UrgUrgent {
				payload["metaData"] = map[string]any{"urgency": message.Urgency}
			}

			handleStyles(payload, message.Style)

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.SendMessage", err)
			}

			url, err := url.Parse(serviceURLs.Message[threadType])
			if err != nil {
				return nil, errs.WrapZCA("failed to parse message URL", "api.SendMessage", err)
			}

			url.Path += path
			body := httpx.BuildFormBody(map[string]string{"params": enc})

			return &sendData{
				URL:  url.String(),
				Body: body,
			}, nil
		}

		handleAttachment := func(ctx context.Context, threadID string, threadType model.ThreadType, message MessageContent) ([]sendData, []model.AttachmentSource, error) {
			if len(message.Attachments) == 0 {
				return nil, nil, errs.ErrSourceEmpty
			}

			canBeDescription := message.IsPhotoDescription()
			attachments := make([]model.AttachmentSource, 0, len(message.Attachments))
			gifs := make([]model.AttachmentSource, 0, len(message.Attachments))
			for _, source := range message.Attachments {
				if source.GetExtension() == "gif" {
					gifs = append(gifs, source)
				} else {
					attachments = append(attachments, source)
				}
			}

			var uploads UploadAttachmentResponse
			if len(attachments) > 0 {
				var err error
				uploads, err = a.UploadAttachment(ctx, threadID, threadType, attachments...)
				if err != nil {
					return nil, nil, err
				}
			}

			mentions, err := prepareMentions(threadType, message.Msg, message.Mentions)
			if err != nil {
				return nil, nil, err
			}
			payloads, err := prepareAttachmentPayloads(
				uploads,
				threadID,
				threadType,
				message,
				mentions,
				canBeDescription,
				strconv.FormatInt(time.Now().UnixMilli(), 10),
				time.Now().UnixMilli(),
			)
			if err != nil {
				return nil, nil, err
			}

			result := make([]sendData, 0, len(payloads))
			for _, payload := range payloads {
				enc, err := u.EncodeAES(jsonx.Stringify(payload.Params))
				if err != nil {
					return nil, nil, errs.WrapZCA("failed to encrypt params", "api.SendMessage", err)
				}
				result = append(result, sendData{
					URL: u.MakeURL(
						serviceURLs.Attachment[threadType]+"/"+payload.Path,
						map[string]any{"nretry": "0"},
						true,
					),
					Body: httpx.BuildFormBody(map[string]string{"params": enc}),
				})
			}
			return result, gifs, nil
		}

		sendMessage := func(ctx context.Context, sendData []sendData) ([]SendMessageResult, error) {
			var (
				g, gctx = errgroup.WithContext(ctx)
				results = make([]SendMessageResult, len(sendData))
			)

			for index, data := range sendData {
				g.Go(func() error {
					resp, err := u.Request(gctx, data.URL, &httpx.RequestOptions{
						Method:  http.MethodPost,
						Body:    data.Body,
						Headers: data.Headers,
					})
					if err != nil {
						return err
					}
					defer resp.Body.Close()

					res, err := resolveResponse[SendMessageResult](sc, resp, true)
					if err != nil {
						return err
					}

					results[index] = res

					return nil
				})
			}

			if err := g.Wait(); err != nil {
				return nil, err
			}
			return results, nil
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, message MessageContent) (*SendMessageResponse, error) {
			if len(message.Msg) == 0 && (len(message.Attachments) == 0) {
				return nil, ErrMessageContentEmpty
			}
			if isExceedMaxFile(len(message.Attachments)) {
				return nil, errs.ErrExceedMaxFile
			}

			hasText := func() bool { return len(message.Msg) > 0 }
			hasAttachments := func() bool { return len(message.Attachments) > 0 }

			results := &SendMessageResponse{}

			sendText := func() error {
				data, err := handleMessage(threadID, threadType, message)
				if err != nil {
					return err
				}
				resps, err := sendMessage(ctx, []sendData{*data})
				if err != nil {
					return err
				}
				if len(resps) > 0 {
					results.Message = &resps[0]
				}
				return nil
			}

			sendAttachments := func() error {
				data, gifs, err := handleAttachment(ctx, threadID, threadType, message)
				if err != nil {
					return err
				}
				resps, err := sendMessage(ctx, data)
				if err != nil {
					return err
				}
				results.Attachment = append(results.Attachment, resps...)
				for _, source := range gifs {
					resp, err := a.SendGIF(ctx, threadID, threadType, GIFContent{
						Attachment: source,
						Msg:        message.Msg,
						TTL:        message.TTL,
						Urgency:    message.Urgency,
					})
					if err != nil {
						return err
					}
					if resp == nil {
						return ErrInvalidAttachmentUpload
					}
					results.Attachment = append(results.Attachment, SendMessageResult{MsgID: resp.MsgID})
				}
				return nil
			}

			if hasAttachments() && hasText() && (!message.IsPhotoDescription() || message.Quote != nil) {
				if err := sendText(); err != nil {
					return nil, err
				}
				message.Msg = ""
				message.Mentions = nil
			}
			if hasAttachments() {
				if err := sendAttachments(); err != nil {
					return nil, err
				}
				message.Msg = ""
			}
			if hasText() {
				if err := sendText(); err != nil {
					return nil, err
				}
			}

			return results, nil
		}, nil
	},
)

func (q *SendMessageQuote) BuildAttachmentMessagePayload() any {
	if q.Content.String != nil {
		return q.PropertyExt
	}

	if q.MsgType == "chat.todo" {
		return map[string]any{
			"properties": model.PropertyExt{
				Color:   0,
				Size:    0,
				Type:    0,
				SubType: 0,
				Ext:     `{"shouldParseLinkOrContact":0}`,
			},
		}
	}

	a := q.Content.Attachment
	return map[string]any{
		"title":       a.Title,
		"description": a.Description,
		"href":        a.Href,
		"thumbUrl":    a.Thumb,
		"oriUrl":      a.Href,
		"normalUrl":   a.Href,
		"childnumber": a.ChildNumber,
		"action":      a.Action,
		"params":      a.Params,
		"type":        a.Type,
	}
}

func (q *SendMessageQuote) BuildMessagePayload() any {
	if q.MsgType != "chat.todo" {
		return ""
	}

	hasRef := (q.Content.Attachment != nil) || (q.Content.Other != nil)
	if !hasRef {
		return ""
	}

	s := ""
	if q.Content.Attachment != nil {
		s = q.Content.Attachment.Params
	} else if q.Content.Other != nil {
		if val, ok := q.Content.Other["params"].(string); ok {
			s = val
		}
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	var payload struct {
		Item struct {
			Content any `json:"content"`
		} `json:"item"`
	}
	if err := json.Unmarshal([]byte(s), &payload); err != nil {
		return ""
	}
	return payload.Item.Content
}

func (q *SendMessageQuote) GetMessageType() int {
	switch q.MsgType {
	case "webchat":
		return 1
	case "chat.voice":
		return 31
	case "chat.photo":
		return 32
	case "chat.sticker":
		return 36
	case "chat.doodle":
		return 37
	case "chat.recommended":
		return 38
	case "chat.lin:":
		return 38 // don't know || if (msgType === "chat.link") return 1;
	case "chat.video.ms":
		return 44 // not sure
	case "share.fil:":
		return 46
	case "chat.gif":
		return 49
	case "chat.location.ne":
		return 43
	default:
		return 1
	}
}

func (c *MessageContent) IsPhotoDescription() bool {
	return len(c.Attachments) == 1 && slices.Contains(config.SupportedImageExtensions, c.Attachments[0].GetExtension())
}
