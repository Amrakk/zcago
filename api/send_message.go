package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
	"golang.org/x/sync/errgroup"
)

var (
	ErrMessageContentEmpty        = errs.NewZCA("message content cannot be empty", "api.SendMessage")
	ErrInvalidMention             = errs.NewZCA("Invalid mentions: total mention characters exceed message length", "api.SendMessage")
	ErrInvalidWebchatQuote        = errs.NewZCA("invalid quote: content must be string for msgType 'webchat'", "api.SendMessage")
	ErrUnsupportedQuotedGroupPoll = errs.NewZCA("quoted message type 'group.poll' is not supported", "api.SendMessage")
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
				model.ThreadTypeUser:  u.MakeURL(fileBase+"/api/message", nil, true),
				model.ThreadTypeGroup: u.MakeURL(fileBase+"/api/group", nil, true),
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
			if len(message.Msg) == 0 {
				return nil, ErrMessageContentEmpty
			}

			quote := message.Quote
			isGroup := threadType == model.ThreadTypeGroup
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
				payload["mentionInfo"] = jsonx.Stringify(message.Mentions)
				payload["visibility"] = 0
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
			} else if payload["mentionInfo"] != nil {
				path = "/mention"
			}

			if message.Urgency != model.UrgDefault {
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

		// handleAttachment := func(threadID string, threadType model.ThreadType, message MessageContent) ([]sendData, error) {
		// }

		sendMessage := func(ctx context.Context, sendData []sendData) ([]SendMessageResult, error) {
			var (
				mu      sync.Mutex
				g, gctx = errgroup.WithContext(ctx)
				results = make([]SendMessageResult, 0, len(sendData))
			)

			for _, data := range sendData {
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

					mu.Lock()
					results = append(results, res)
					mu.Unlock()

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

			results := &SendMessageResponse{}

			if len(message.Msg) > 0 {
				data, err := handleMessage(threadID, threadType, message)
				if err != nil {
					return nil, err
				}

				resps, err := sendMessage(ctx, []sendData{*data})
				if err != nil {
					return nil, err
				}

				if len(resps) > 0 {
					results.Message = &resps[0]
				} else {
					panic("expected at least one message response")
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
