package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type ReportReason int

const (
	ReportReasonOther ReportReason = iota
	ReportReasonSensitive
	ReportReasonAnnoying
	ReportReasonFraud
)

type (
	SendReportOptions struct {
		Content string // Only used when Reason is ReportReasonOther
		Reason  ReportReason
	}
	SendReportResponse struct {
		ReportID string `json:"reportId"`
	}
	SendReportFn = func(ctx context.Context, threadID string, threadType model.ThreadType, options SendReportOptions) (*SendReportResponse, error)
)

func (a *api) SendReport(ctx context.Context, threadID string, threadType model.ThreadType, options SendReportOptions) (*SendReportResponse, error) {
	return a.e.SendReport(ctx, threadID, threadType, options)
}

var sendReportFactory = apiFactory[*SendReportResponse, SendReportFn]()(
	func(a *api, sc session.Context, u factoryUtils[*SendReportResponse]) (SendReportFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/report/abuse-v2", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/social/profile/reportabuse", nil, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, options SendReportOptions) (*SendReportResponse, error) {
			var payload map[string]any

			if threadType == model.ThreadTypeUser {
				payload = map[string]any{
					"idTo":   threadID,
					"objId":  "person.profile",
					"reason": strconv.Itoa(int(options.Reason)),
				}
			} else {
				payload = map[string]any{
					"uidTo":   threadID,
					"type":    14,
					"reason":  options.Reason,
					"content": "",
					"imei":    sc.IMEI(),
				}
			}

			if options.Reason == ReportReasonOther {
				payload["content"] = options.Content
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.SendReport", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURLs[threadType], &httpx.RequestOptions{
				Method: http.MethodPost,
				Body:   body,
			})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)
