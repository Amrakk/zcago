package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/cryptox"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/listener"
	"github.com/Amrakk/zcago/session"
)

func New(sc session.MutableContext) (*api, error) {
	a := &api{sc: sc}

	if err := a.initEndpoints(); err != nil {
		return nil, err
	}

	l, err := listener.New(sc, sc.ZPWWebsocket())
	if err != nil {
		return nil, err
	}
	a.l = l

	return a, nil
}

type api struct {
	sc session.MutableContext
	e  endpoints
	l  listener.Listener
}

type endpoints struct {
	//gen:fields

	AcceptFriendRequest         AcceptFriendRequestFn
	AddGroupBlockedMember       AddGroupBlockedMemberFn
	AddGroupDeputy              AddGroupDeputyFn
	AddPollOptions              AddPollOptionsFn
	AddQuickMessage             AddQuickMessageFn
	AddReaction                 AddReactionFn
	AddUnreadMark               AddUnreadMarkFn
	AddUserToGroup              AddUserToGroupFn
	BlockUser                   BlockUserFn
	ChangeGroupOwner            ChangeGroupOwnerFn
	CreateAutoReply             CreateAutoReplyFn
	CreateCatalog               CreateCatalogFn
	CreateGroup                 CreateGroupFn
	CreateNote                  CreateNoteFn
	CreatePoll                  CreatePollFn
	CreateReminder              CreateReminderFn
	DeleteAutoReply             DeleteAutoReplyFn
	DeleteAvatar                DeleteAvatarFn
	DeleteCatalog               DeleteCatalogFn
	DeleteChat                  DeleteChatFn
	DeleteGroup                 DeleteGroupFn
	DeleteMessage               DeleteMessageFn
	DisableGroupLink            DisableGroupLinkFn
	EnableGroupLink             EnableGroupLinkFn
	FindUser                    FindUserFn
	ForwardMessage              ForwardMessageFn
	GetAccountInfo              GetAccountInfoFn
	GetAliasList                GetAliasListFn
	GetAllFriends               GetAllFriendsFn
	GetAllGroups                GetAllGroupsFn
	GetAutoDeleteChat           GetAutoDeleteChatFn
	GetAvatarList               GetAvatarListFn
	GetFriendBoardList          GetFriendBoardListFn
	GetFriendOnlineStatus       GetFriendOnlineStatusFn
	GetFriendRecommendations    GetFriendRecommendationsFn
	GetFriendRequestStatus      GetFriendRequestStatusFn
	GetGroupBlockedMember       GetGroupBlockedMemberFn
	GetGroupBoardList           GetGroupBoardListFn
	GetGroupInfo                GetGroupInfoFn
	GetGroupInviteBoxInfo       GetGroupInviteBoxInfoFn
	GetGroupInviteBoxList       GetGroupInviteBoxListFn
	GetGroupLinkDetail          GetGroupLinkDetailFn
	GetGroupLinkInfo            GetGroupLinkInfoFn
	GetGroupPendingJoinRequests GetGroupPendingJoinRequestsFn
	GetHiddenChat               GetHiddenChatFn
	GetLabels                   GetLabelsFn
	GetMute                     GetMuteFn
	GetPinnedChat               GetPinnedChatFn
	GetPollDetail               GetPollDetailFn
	GetQR                       GetQRFn
	GetQuickMessageList         GetQuickMessageListFn
	GetReminder                 GetReminderFn
	GetReminderList             GetReminderListFn
	GetReminderResponse         GetReminderResponseFn
	GetSentFriendRequest        GetSentFriendRequestFn
	GetSetting                  GetSettingFn
	GetStickerDetail            GetStickerDetailFn
	GetStickers                 GetStickersFn
	GetUnreadMark               GetUnreadMarkFn
	GetUserInfo                 GetUserInfoFn
	GetUserSummaryInfo          GetUserSummaryInfoFn
	InviteUserToGroups          InviteUserToGroupsFn
	JoinGroupInviteBox          JoinGroupInviteBoxFn
	JoinGroupLink               JoinGroupLinkFn
	LastOnline                  LastOnlineFn
	LeaveGroup                  LeaveGroupFn
	LockPoll                    LockPollFn
	ParseLink                   ParseLinkFn
	RejectFriendRequest         RejectFriendRequestFn
	RemoveAlias                 RemoveAliasFn
	RemoveFriend                RemoveFriendFn
	RemoveGroupBlockedMember    RemoveGroupBlockedMemberFn
	RemoveGroupDeputy           RemoveGroupDeputyFn
	RemoveGroupInviteBox        RemoveGroupInviteBoxFn
	RemoveQuickMessage          RemoveQuickMessageFn
	RemoveReminder              RemoveReminderFn
	RemoveUnreadMark            RemoveUnreadMarkFn
	RemoveUserFromGroup         RemoveUserFromGroupFn
	ResetHiddenChatPIN          ResetHiddenChatPINFn
	ReuseAvatar                 ReuseAvatarFn
	ReviewPendingMemberRequest  ReviewPendingMemberRequestFn
	SendBankCard                SendBankCardFn
	SendCard                    SendCardFn
	SendDeliveredEvent          SendDeliveredEventFn
	SendFriendRequest           SendFriendRequestFn
	SendGIF                     SendGIFFn
	SendLink                    SendLinkFn
	SendMessage                 SendMessageFn
	SendReport                  SendReportFn
	SendSeenEvent               SendSeenEventFn
	SendSticker                 SendStickerFn
	SendTypingEvent             SendTypingEventFn
	SendVideo                   SendVideoFn
	SendVoice                   SendVoiceFn
	SetHiddenChat               SetHiddenChatFn
	SetMute                     SetMuteFn
	SetPinChat                  SetPinChatFn
	SetViewFeedBlock            SetViewFeedBlockFn
	SharePoll                   SharePollFn
	UnblockUser                 UnblockUserFn
	UndoFriendRequest           UndoFriendRequestFn
	UndoMessage                 UndoMessageFn
	UpdateAccountAvatar         UpdateAccountAvatarFn
	UpdateActiveStatus          UpdateActiveStatusFn
	UpdateAlias                 UpdateAliasFn
	UpdateAutoDeleteChat        UpdateAutoDeleteChatFn
	UpdateGroupAvatar           UpdateGroupAvatarFn
	UpdateGroupName             UpdateGroupNameFn
	UpdateGroupSetting          UpdateGroupSettingFn
	UpdateHiddenChatPIN         UpdateHiddenChatPINFn
	UpdateLabels                UpdateLabelsFn
	UpdateLanguage              UpdateLanguageFn
	UpdateNote                  UpdateNoteFn
	UpdateProfile               UpdateProfileFn
	UpdateQuickMessage          UpdateQuickMessageFn
	UpdateReminder              UpdateReminderFn
	UpdateSetting               UpdateSettingFn
	UploadAttachment            UploadAttachmentFn
	UploadPhoto                 UploadPhotoFn
	UploadThumbnail             UploadThumbnailFn
	VotePoll                    VotePollFn
}

func (a *api) initEndpoints() error {
	if !a.sc.SecretKey().IsValid() {
		return errs.NewZCA("secret key missing or invalid", "api.initEndpoints")
	}

	return firstErr(
		//gen:binds

		bind(a.sc, a, &a.e.AcceptFriendRequest, acceptFriendRequestFactory),
		bind(a.sc, a, &a.e.AddGroupBlockedMember, addGroupBlockedMemberFactory),
		bind(a.sc, a, &a.e.AddGroupDeputy, addGroupDeputyFactory),
		bind(a.sc, a, &a.e.AddPollOptions, addPollOptionsFactory),
		bind(a.sc, a, &a.e.AddQuickMessage, addQuickMessageFactory),
		bind(a.sc, a, &a.e.AddReaction, addReactionFactory),
		bind(a.sc, a, &a.e.AddUnreadMark, addUnreadMarkFactory),
		bind(a.sc, a, &a.e.AddUserToGroup, addUserToGroupFactory),
		bind(a.sc, a, &a.e.BlockUser, blockUserFactory),
		bind(a.sc, a, &a.e.ChangeGroupOwner, changeGroupOwnerFactory),
		bind(a.sc, a, &a.e.CreateAutoReply, createAutoReplyFactory),
		bind(a.sc, a, &a.e.CreateCatalog, createCatalogFactory),
		bind(a.sc, a, &a.e.CreateGroup, createGroupFactory),
		bind(a.sc, a, &a.e.CreateNote, createNoteFactory),
		bind(a.sc, a, &a.e.CreatePoll, createPollFactory),
		bind(a.sc, a, &a.e.CreateReminder, createReminderFactory),
		bind(a.sc, a, &a.e.DeleteAutoReply, deleteAutoReplyFactory),
		bind(a.sc, a, &a.e.DeleteAvatar, deleteAvatarFactory),
		bind(a.sc, a, &a.e.DeleteCatalog, deleteCatalogFactory),
		bind(a.sc, a, &a.e.DeleteChat, deleteChatFactory),
		bind(a.sc, a, &a.e.DeleteGroup, deleteGroupFactory),
		bind(a.sc, a, &a.e.DeleteMessage, deleteMessageFactory),
		bind(a.sc, a, &a.e.DisableGroupLink, disableGroupLinkFactory),
		bind(a.sc, a, &a.e.EnableGroupLink, enableGroupLinkFactory),
		bind(a.sc, a, &a.e.FindUser, findUserFactory),
		bind(a.sc, a, &a.e.ForwardMessage, forwardMessageFactory),
		bind(a.sc, a, &a.e.GetAccountInfo, getAccountInfoFactory),
		bind(a.sc, a, &a.e.GetAliasList, getAliasListFactory),
		bind(a.sc, a, &a.e.GetAllFriends, getAllFriendsFactory),
		bind(a.sc, a, &a.e.GetAllGroups, getAllGroupsFactory),
		bind(a.sc, a, &a.e.GetAutoDeleteChat, getAutoDeleteChatFactory),
		bind(a.sc, a, &a.e.GetAvatarList, getAvatarListFactory),
		bind(a.sc, a, &a.e.GetFriendBoardList, getFriendBoardListFactory),
		bind(a.sc, a, &a.e.GetFriendOnlineStatus, getFriendOnlineStatusFactory),
		bind(a.sc, a, &a.e.GetFriendRecommendations, getFriendRecommendationsFactory),
		bind(a.sc, a, &a.e.GetFriendRequestStatus, getFriendRequestStatusFactory),
		bind(a.sc, a, &a.e.GetGroupBlockedMember, getGroupBlockedMemberFactory),
		bind(a.sc, a, &a.e.GetGroupBoardList, getGroupBoardListFactory),
		bind(a.sc, a, &a.e.GetGroupInfo, getGroupInfoFactory),
		bind(a.sc, a, &a.e.GetGroupInviteBoxInfo, getGroupInviteBoxInfoFactory),
		bind(a.sc, a, &a.e.GetGroupInviteBoxList, getGroupInviteBoxListFactory),
		bind(a.sc, a, &a.e.GetGroupLinkDetail, getGroupLinkDetailFactory),
		bind(a.sc, a, &a.e.GetGroupLinkInfo, getGroupLinkInfoFactory),
		bind(a.sc, a, &a.e.GetGroupPendingJoinRequests, getGroupPendingJoinRequestsFactory),
		bind(a.sc, a, &a.e.GetHiddenChat, getHiddenChatFactory),
		bind(a.sc, a, &a.e.GetLabels, getLabelsFactory),
		bind(a.sc, a, &a.e.GetMute, getMuteFactory),
		bind(a.sc, a, &a.e.GetPinnedChat, getPinnedChatFactory),
		bind(a.sc, a, &a.e.GetPollDetail, getPollDetailFactory),
		bind(a.sc, a, &a.e.GetQR, getQRFactory),
		bind(a.sc, a, &a.e.GetQuickMessageList, getQuickMessageListFactory),
		bind(a.sc, a, &a.e.GetReminder, getReminderFactory),
		bind(a.sc, a, &a.e.GetReminderList, getReminderListFactory),
		bind(a.sc, a, &a.e.GetReminderResponse, getReminderResponseFactory),
		bind(a.sc, a, &a.e.GetSentFriendRequest, getSentFriendRequestFactory),
		bind(a.sc, a, &a.e.GetSetting, getSettingFactory),
		bind(a.sc, a, &a.e.GetStickerDetail, getStickerDetailFactory),
		bind(a.sc, a, &a.e.GetStickers, getStickersFactory),
		bind(a.sc, a, &a.e.GetUnreadMark, getUnreadMarkFactory),
		bind(a.sc, a, &a.e.GetUserInfo, getUserInfoFactory),
		bind(a.sc, a, &a.e.GetUserSummaryInfo, getUserSummaryInfoFactory),
		bind(a.sc, a, &a.e.InviteUserToGroups, inviteUserToGroupsFactory),
		bind(a.sc, a, &a.e.JoinGroupInviteBox, joinGroupInviteBoxFactory),
		bind(a.sc, a, &a.e.JoinGroupLink, joinGroupLinkFactory),
		bind(a.sc, a, &a.e.LastOnline, lastOnlineFactory),
		bind(a.sc, a, &a.e.LeaveGroup, leaveGroupFactory),
		bind(a.sc, a, &a.e.LockPoll, lockPollFactory),
		bind(a.sc, a, &a.e.ParseLink, parseLinkFactory),
		bind(a.sc, a, &a.e.RejectFriendRequest, rejectFriendRequestFactory),
		bind(a.sc, a, &a.e.RemoveAlias, removeAliasFactory),
		bind(a.sc, a, &a.e.RemoveFriend, removeFriendFactory),
		bind(a.sc, a, &a.e.RemoveGroupBlockedMember, removeGroupBlockedMemberFactory),
		bind(a.sc, a, &a.e.RemoveGroupDeputy, removeGroupDeputyFactory),
		bind(a.sc, a, &a.e.RemoveGroupInviteBox, removeGroupInviteBoxFactory),
		bind(a.sc, a, &a.e.RemoveQuickMessage, removeQuickMessageFactory),
		bind(a.sc, a, &a.e.RemoveReminder, removeReminderFactory),
		bind(a.sc, a, &a.e.RemoveUnreadMark, removeUnreadMarkFactory),
		bind(a.sc, a, &a.e.RemoveUserFromGroup, removeUserFromGroupFactory),
		bind(a.sc, a, &a.e.ResetHiddenChatPIN, resetHiddenChatPINFactory),
		bind(a.sc, a, &a.e.ReuseAvatar, reuseAvatarFactory),
		bind(a.sc, a, &a.e.ReviewPendingMemberRequest, reviewPendingMemberRequestFactory),
		bind(a.sc, a, &a.e.SendBankCard, sendBankCardFactory),
		bind(a.sc, a, &a.e.SendCard, sendCardFactory),
		bind(a.sc, a, &a.e.SendDeliveredEvent, sendDeliveredEventFactory),
		bind(a.sc, a, &a.e.SendFriendRequest, sendFriendRequestFactory),
		bind(a.sc, a, &a.e.SendGIF, sendGIFFactory),
		bind(a.sc, a, &a.e.SendLink, sendLinkFactory),
		bind(a.sc, a, &a.e.SendMessage, sendMessageFactory),
		bind(a.sc, a, &a.e.SendReport, sendReportFactory),
		bind(a.sc, a, &a.e.SendSeenEvent, sendSeenEventFactory),
		bind(a.sc, a, &a.e.SendSticker, sendStickerFactory),
		bind(a.sc, a, &a.e.SendTypingEvent, sendTypingEventFactory),
		bind(a.sc, a, &a.e.SendVideo, sendVideoFactory),
		bind(a.sc, a, &a.e.SendVoice, sendVoiceFactory),
		bind(a.sc, a, &a.e.SetHiddenChat, setHiddenChatFactory),
		bind(a.sc, a, &a.e.SetMute, setMuteFactory),
		bind(a.sc, a, &a.e.SetPinChat, setPinChatFactory),
		bind(a.sc, a, &a.e.SetViewFeedBlock, setViewFeedBlockFactory),
		bind(a.sc, a, &a.e.SharePoll, sharePollFactory),
		bind(a.sc, a, &a.e.UnblockUser, unblockUserFactory),
		bind(a.sc, a, &a.e.UndoFriendRequest, undoFriendRequestFactory),
		bind(a.sc, a, &a.e.UndoMessage, undoMessageFactory),
		bind(a.sc, a, &a.e.UpdateAccountAvatar, updateAccountAvatarFactory),
		bind(a.sc, a, &a.e.UpdateActiveStatus, updateActiveStatusFactory),
		bind(a.sc, a, &a.e.UpdateAlias, updateAliasFactory),
		bind(a.sc, a, &a.e.UpdateAutoDeleteChat, updateAutoDeleteChatFactory),
		bind(a.sc, a, &a.e.UpdateGroupAvatar, updateGroupAvatarFactory),
		bind(a.sc, a, &a.e.UpdateGroupName, updateGroupNameFactory),
		bind(a.sc, a, &a.e.UpdateGroupSetting, updateGroupSettingFactory),
		bind(a.sc, a, &a.e.UpdateHiddenChatPIN, updateHiddenChatPINFactory),
		bind(a.sc, a, &a.e.UpdateLabels, updateLabelsFactory),
		bind(a.sc, a, &a.e.UpdateLanguage, updateLanguageFactory),
		bind(a.sc, a, &a.e.UpdateNote, updateNoteFactory),
		bind(a.sc, a, &a.e.UpdateProfile, updateProfileFactory),
		bind(a.sc, a, &a.e.UpdateQuickMessage, updateQuickMessageFactory),
		bind(a.sc, a, &a.e.UpdateReminder, updateReminderFactory),
		bind(a.sc, a, &a.e.UpdateSetting, updateSettingFactory),
		bind(a.sc, a, &a.e.UploadAttachment, uploadAttachmentFactory),
		bind(a.sc, a, &a.e.UploadPhoto, uploadPhotoFactory),
		bind(a.sc, a, &a.e.UploadThumbnail, uploadThumbnailFactory),
		bind(a.sc, a, &a.e.VotePoll, votePollFactory),
	)
}

type factoryUtils[T any] struct {
	MakeURL   func(baseURL string, params map[string]any, includeDefaults bool) string
	EncodeAES func(data string) (string, error)
	Request   func(ctx context.Context, url string, options *httpx.RequestOptions) (*http.Response, error)
	Logger    *logger.Logger
	Resolve   func(res *http.Response, isEncrypted bool) (T, error)
}

type (
	handler[T any, R any]         func(api *api, sc session.Context, utils factoryUtils[T]) (R, error)
	endpointFactory[T any, R any] func(sc session.MutableContext, api *api) (R, error)
)

func apiFactory[T any, R any]() func(
	callback handler[T, R],
) endpointFactory[T, R] {
	return func(callback handler[T, R]) endpointFactory[T, R] {
		return func(sc session.MutableContext, a *api) (R, error) {
			utils := factoryUtils[T]{
				MakeURL: func(url string, params map[string]any, includeDefaults bool) string {
					return httpx.MakeURL(sc, url, params, includeDefaults)
				},
				EncodeAES: func(data string) (string, error) {
					key := sc.SecretKey().Bytes()
					return cryptox.EncodeAESCBC(key, data, cryptox.EncryptTypeBase64)
				},
				Request: func(ctx context.Context, url string, opts *httpx.RequestOptions) (*http.Response, error) {
					return httpx.Request(ctx, sc, url, opts)
				},
				Logger: logger.Log(sc),
				Resolve: func(res *http.Response, isEncrypted bool) (T, error) {
					return resolveResponse[T](sc, res, isEncrypted)
				},
			}

			return callback(a, sc, utils)
		}
	}
}

func resolveResponse[T any](
	sc session.Context,
	res *http.Response,
	isEncrypted bool,
) (T, error) {
	var zero T

	r := httpx.HandleZaloResponse[T](sc, res, isEncrypted)
	if r == nil {
		return zero, errs.NewZCA("empty response", "api.resolveResponse")
	}
	if r.Meta.Code != 0 {
		var zero T
		code := errs.ZaloErrorCode(r.Meta.Code)
		return zero, errs.NewZaloAPIError(r.Meta.Message, &code)
	}

	return r.Data, nil
}

func bind[T any, F any](sc session.MutableContext, a *api, target *F, factory endpointFactory[T, F]) error {
	fn, err := factory(sc, a)
	if err != nil {
		return err
	}
	*target = fn
	return nil
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
