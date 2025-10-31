package zcago

import (
	"context"

	"github.com/Amrakk/zcago/api"
	"github.com/Amrakk/zcago/listener"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type API interface {
	GetContext() (session.Context, error)
	GetOwnID() string
	Listener() listener.Listener

	//gen:methods

	// AcceptFriendRequest accepts a friend request from a user.
	//
	// Params:
	//  - ctx — cancel/deadline control
	//  - friendID — ID of the user whose friend request is being accepted
	//
	// Errors: errs.ZaloAPIError
	AcceptFriendRequest(ctx context.Context, friendID string) (api.AcceptFriendRequestResponse, error)
	// AddGroupBlockedMember adds users to the group's blocked list.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - groupID — group ID
	//   - memberID — member ID(s)
	//
	// Errors: errs.ZaloAPIError
	AddGroupBlockedMember(ctx context.Context, groupID string, memberID ...string) (api.AddGroupBlockedMemberResponse, error)
	// AddGroupDeputy adds one or more users as group deputies.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - groupID — group ID
	//   - memberID — user ID(s)
	//
	// Errors: ZaloApiError
	AddGroupDeputy(ctx context.Context, groupID string, memberID ...string) (api.AddGroupDeputyResponse, error)
	// AddPollOptions adds new options to a poll.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - options — poll options
	//
	// Errors: errs.ZaloAPIError
	AddPollOptions(ctx context.Context, options api.AddPollOptionsRequest) (*api.AddPollOptionsResponse, error)
	// AddQuickMessage adds a new quick message.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - message — payload containing data to add the quick message
	//
	// Note: Zalo may return error code 821 if the quick message limit is reached.
	//
	// Errors: errs.ZaloAPIError
	AddQuickMessage(ctx context.Context, message api.AddQuickMessageRequest) (*api.AddQuickMessageResponse, error)
	// AddReaction adds a reaction to a message.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - dest - destination data including message IDs and thread information
	//   - icon - reaction icon
	//
	// Note: Use model.NewReactionData(icon) to create a valid
	// reaction payload from a predefined model.ReactionIcon enum.
	//
	// Errors: errs.ZaloAPIError, api.ErrInvalidReaction
	AddReaction(ctx context.Context, dest api.AddReactionDestination, reaction model.ReactionData) (*api.AddReactionResponse, error)
	// AddUnreadMark adds an unread mark to a conversation.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - thread ID
	//   - threadType - thread type
	//
	// Errors: errs.ZaloAPIError
	AddUnreadMark(ctx context.Context, threadID string, threadType model.ThreadType) (*api.AddUnreadMarkResponse, error)
	// AddUserToGroup adds one or more users to a group.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - groupID — group ID
	//   - userID — user ID(s)
	//
	//  Errors: errs.ZaloAPIError
	AddUserToGroup(ctx context.Context, groupID string, userID ...string) (*api.AddUserToGroupResponse, error)
	// BlockUser blocks a user by their ID.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - userID — ID of the user to block
	//
	// Errors: errs.ZaloAPIError
	BlockUser(ctx context.Context, userID string) (api.BlockUserResponse, error)
	// ChangeGroupOwner changes the owner of a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//   - memberID - user ID of the new group owner
	//
	// Note: Changing the group owner will result in losing current admin rights.
	//
	// Errors: errs.ZaloAPIError
	ChangeGroupOwner(ctx context.Context, groupID string, memberID string) (*api.ChangeGroupOwnerResponse, error)
	// CreateAutoReply creates an auto reply.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - message - payload containing data to create the auto reply message
	//
	// Note: This API is used for zBusiness.
	//
	// Errors: errs.ZaloAPIError
	//
	// @TODO: test this api
	CreateAutoReply(ctx context.Context, message api.CreateAutoReplyRequest) (*api.CreateAutoReplyResponse, error)
	// CreateCatalog creates a new catalog.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - name - catalog name
	//
	// Note: This API is used for zBusiness.
	//
	// Errors: errs.ZaloAPIError
	//
	// @TODO: test this api
	CreateCatalog(ctx context.Context, name string) (*api.CreateCatalogResponse, error)
	// CreateGroup creates a new group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - options - group options
	//
	// Errors: errs.ZaloAPIError
	CreateGroup(ctx context.Context, options api.CreateGroupOptions) (*api.CreateGroupResponse, error)
	// CreateGroupNote creates a note in a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//   - options - note options
	//
	// Errors: errs.ZaloAPIError
	CreateNote(ctx context.Context, groupID string, options api.CreateNoteOptions) (*api.CreateNoteResponse, error)
	// CreateGroupPoll creates a poll in a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID to create the poll in
	//   - options - poll options
	//
	// Errors: errs.ZaloAPIError
	CreatePoll(ctx context.Context, groupID string, options api.CreatePollOptions) (*api.CreatePollResponse, error)
	// CreateReminder creates a reminder in a thread.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - group/user ID to create the reminder in
	//   - threadType - thread type
	//   - options - reminder options
	//
	// Errors: errs.ZaloAPIError
	CreateReminder(ctx context.Context, threadID string, threadType model.ThreadType, options api.CreateReminderOptions) (*api.CreateReminderResponse, error)
	// DeleteAutoReply deletes an auto reply.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - id - ID of the auto reply
	//
	// Note: This API is used for zBusiness.
	//
	// Errors: errs.ZaloAPIError
	//
	// @TODO: test this api
	DeleteAutoReply(ctx context.Context, id int) (*api.DeleteAutoReplyResponse, error)
	// DeleteAvatar removes one or more avatars from the avatar list.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - photoID - avatar photo ID(s) to delete
	//
	// Errors: errs.ZaloAPIError
	DeleteAvatar(ctx context.Context, photoID ...string) (*api.DeleteAvatarResponse, error)
	// DeleteCatalog deletes a catalog.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - catalogID - catalog ID
	//
	// Note: This API is used for zBusiness.
	//
	// Errors: errs.ZaloAPIError
	//
	// @TODO: test this api
	DeleteCatalog(ctx context.Context, catalogID string) (api.DeleteCatalogResponse, error)
	// DeleteChat deletes a chat thread.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - thread ID
	//   - threadType - thread type
	//   - lastMsg - last message information
	//
	// Note: If lastMsg is incorrect or stale, the conversation (thread entry) is removed
	// but the underlying chat messages are NOT deleted. No error is returned in this case.
	//
	// Errors: errs.ZaloAPIError
	DeleteChat(ctx context.Context, threadID string, threadType model.ThreadType, lastMsg api.DeleteChatLastMessage) (*api.DeleteChatResponse, error)
	// DeleteGroup disbands a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID to disband
	//
	// Errors: errs.ZaloAPIError
	DeleteGroup(ctx context.Context, groupID string) (api.DeleteGroupResponse, error)
	// DeleteMessage deletes a message.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - dest - delete target
	//   - onlyMe - when true, hides the message only for you (others still see it)
	//
	// Note:
	//   - Self-delete (onlyMe=true) is always allowed.
	//   - Senders deleting for everyone (onlyMe=false) should use [api.UndoMessage].
	//   - DMs: Deletion for everyone is not supported.
	//   - Groups: Deleting for everyone requires higher privilege
	// 			   and applies only to messages within 24 hours.
	//
	// Errors: errs.ZaloAPIError, api.ErrSenderRecall, api.ErrDMRecipientsDeleteUnsupported
	DeleteMessage(ctx context.Context, dest api.DeleteMessageDestination, onlyMe bool) (*api.DeleteMessageResponse, error)
	// DisableGroupLink disables the group link.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//
	// Errors: errs.ZaloAPIError
	DisableGroupLink(ctx context.Context, groupID string) (api.DisableGroupLinkResponse, error)
	// EnableGroupLink enables and creates a new group link.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//
	// Errors: errs.ZaloAPIError
	EnableGroupLink(ctx context.Context, groupID string) (*api.EnableGroupLinkResponse, error)
	// FindUser finds users by their phone number.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - phoneNumber - phone number(s) (prefix with country code, e.g., "84123456789")
	//
	// Note: This test is not boundary-validated yet. Use with caution.
	//
	// Errors: errs.ZaloAPIError, api.ErrPhoneNumberEmpty
	FindUser(ctx context.Context, phoneNumber ...string) (*api.FindUserResponse, error)
	// ForwardMessage forwards a message to multiple threads.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadIDs - thread IDs
	//   - threadType - thread type
	//   - message - forward message payload
	//
	// Errors: errs.ZaloAPIError, api.ErrMessageEmpty, api.ErrThreadIDEmpty
	//
	// @TODO: test this api, still incomplete
	ForwardMessage(ctx context.Context, threadIDs []string, threadType model.ThreadType, message api.ForwardMessagePayload) (*api.ForwardMessageResponse, error)
	// GetAccountInfo retrieves the account information.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetAccountInfo(ctx context.Context) (*api.GetAccountInfoResponse, error)
	// GetAliasList retrieves the alias list.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - options - offset pagination options (default Count:100, Page:1)
	//
	// Errors: errs.ZaloAPIError
	GetAliasList(ctx context.Context, options model.OffsetPaginationOptions) (*api.GetAliasListResponse, error)
	// GetAllFriends retrieves all friends.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - options - offset pagination options (default Count:20000, Page:1)
	//
	// Errors: errs.ZaloAPIError
	GetAllFriends(ctx context.Context, options model.OffsetPaginationOptions) (*api.GetAllFriendsResponse, error)
	// GetAllGroups retrieves all groups.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetAllGroups(ctx context.Context) (*api.GetAllGroupsResponse, error)
	// GetAutoDeleteChat retrieves the auto-delete conversations.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetAutoDeleteChat(ctx context.Context) (*api.GetAutoDeleteChatResponse, error)
	// GetAvatarList retrieves the list of avatars.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - options - offset pagination options (default: Count:50, Page:1)
	//
	// Errors: errs.ZaloAPIError
	GetAvatarList(ctx context.Context, options model.OffsetPaginationOptions) (*api.GetAvatarListResponse, error)
	// GetFriendBoardList retrieves the friend board list for a conversation.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - friendID - friend ID
	//
	// Errors: errs.ZaloAPIError
	GetFriendBoardList(ctx context.Context, friendID string) (*api.GetFriendBoardListResponse, error)
	// GetOnlineFriends retrieves the list of online friends' status.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetFriendOnlineStatus(ctx context.Context) (*api.GetFriendOnlineStatusResponse, error)
	// GetFriendRecommendations retrieves friend recommendations.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetFriendRecommendations(ctx context.Context) (*api.GetFriendRecommendationsResponse, error)
	// GetFriendRequestStatus retrieves the status of a friend request.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - friendID - friend ID
	//
	// Errors: errs.ZaloAPIError
	GetFriendRequestStatus(ctx context.Context, friendID string) (*api.GetFriendRequestStatusResponse, error)
	// GetGroupBlockedMember retrieves the list of blocked members in a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//   - options - offset pagination options (default: Count:50, Page:1)
	//
	// Errors: errs.ZaloAPIError
	GetGroupBlockedMember(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*api.GetGroupBlockedMemberResponse, error)
	// GetGroupBoardList retrieves the list of board items in a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - ID of the group
	//   - options - offset pagination options (default: Count:20, Page:1)
	//
	// Errors: errs.ZaloAPIError
	GetGroupBoardList(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*api.GetGroupBoardListResponse, error)
	// GetGroupInfo retrieves information about one or more groups.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID(s)
	//
	// Errors: errs.ZaloAPIError
	GetGroupInfo(ctx context.Context, groupID ...string) (*api.GetGroupInfoResponse, error)
	// GetGroupInviteBoxInfo retrieves information about a group invite box.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//   - options - offset pagination options (default: Count:10, Page:1)
	//
	// Errors: errs.ZaloAPIError
	GetGroupInviteBoxInfo(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*api.GetGroupInviteBoxInfoResponse, error)
	// GetGroupInviteBoxList retrieves the list of group invite boxes.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - options - get group invite box list options
	//
	// Errors: errs.ZaloAPIError
	GetGroupInviteBoxList(ctx context.Context, options api.GetGroupInviteBoxListOptions) (*api.GetGroupInviteBoxListResponse, error)
	// GetGroupLinkDetail retrieves information about a group link.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//
	// Errors: errs.ZaloAPIError
	GetGroupLinkDetail(ctx context.Context, groupID string) (*api.GetGroupLinkDetailResponse, error)
	// GetGroupLinkInfo retrieves group information from a group link.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - link - group link
	//   - memberPage - member page number
	//
	// Errors: errs.ZaloAPIError
	GetGroupLinkInfo(ctx context.Context, link string, memberPage int) (*api.GetGroupLinkInfoResponse, error)
	// GetGroupPendingJoinRequests retrieves the list of pending group membership requests.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - ID of the group to get pending members
	//
	// Note: Only the group leader and deputy group leader can view pending members.
	//
	// Errors: errs.ZaloAPIError
	GetGroupPendingJoinRequests(ctx context.Context, groupID string) (*api.GetGroupPendingJoinRequestsResponse, error)
	// GetHiddenChat retrieves the list of hidden conversations.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetHiddenChat(ctx context.Context) (*api.GetHiddenChatResponse, error)
	// GetLabels retrieves all labels.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetLabels(ctx context.Context) (*api.GetLabelsResponse, error)
	// GetMute retrieves mute status of threads.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetMute(ctx context.Context) (*api.GetMuteResponse, error)
	// GetPinnedChat retrieves the list of pinned conversations.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetPinnedChat(ctx context.Context) (*api.GetPinnedChatResponse, error)
	// GetPollDetail retrieves detailed information about a poll.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - pollID - poll ID
	//
	// Errors: errs.ZaloAPIError
	GetPollDetail(ctx context.Context, pollID int) (*api.GetPollDetailResponse, error)
	// GetQR retrieves the QR code for one or more users.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - userID - user ID or list of user IDs
	//
	// Errors: errs.ZaloAPIError
	GetQR(ctx context.Context, userID ...string) (*api.GetQRResponse, error)
	// GetQuickMessageList retrieves the list of quick messages.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetQuickMessageList(ctx context.Context) (*api.GetQuickMessageListResponse, error)
	// GetReminder retrieves the details of a reminder from a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - reminderID - reminder ID
	//
	// Errors: errs.ZaloAPIError
	GetReminder(ctx context.Context, reminderID string) (*api.GetReminderResponse, error)
	// GetReminderList retrieves the list of reminders.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - ID of the thread
	//   - threadType - thread type
	//   - options - offset pagination options (default: Count:20, Page:1)
	//
	// Note: This will return model.ErrAmbiguousReminder if Zalo return empty list
	//
	// Errors: errs.ZaloAPIError, model.ErrAmbiguousReminder
	GetReminderList(ctx context.Context, threadID string, threadType model.ThreadType, options model.OffsetPaginationOptions) (*api.GetReminderListResponse, error)
	// GetReminderResponse retrieves the responses for a specific reminder.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - reminderID - reminder ID
	//
	// Errors: errs.ZaloAPIError
	GetReminderResponse(ctx context.Context, reminderID string) (*api.GetReminderResponseResponse, error)
	// GetSentFriendRequest retrieves the list of friend requests you have sent.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Note: Zalo may return error code 112 if there are no friend requests.
	//
	// Errors: errs.ZaloAPIError
	GetSentFriendRequest(ctx context.Context) (*api.GetSentFriendRequestResponse, error)
	// GetSetting retrieves the current account settings.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetSetting(ctx context.Context) (*api.GetSettingResponse, error)
	// GetStickerDetail retrieves detailed information for the specified sticker ID.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - stickerID - sticker ID
	//
	// Errors: errs.ZaloAPIError
	GetStickerDetail(ctx context.Context, stickerID int) (*api.GetStickerDetailResponse, error)
	// GetStickers retrieves sticker IDs by keyword.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - keyword - keyword to search for
	//
	// Returns: sticker IDs
	//
	// Errors: errs.ZaloAPIError
	GetStickers(ctx context.Context, keyword string) (*api.GetStickersResponse, error)
	// GetUnreadMark retrieves the current unread mark status for conversations.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	GetUnreadMark(ctx context.Context) (*api.GetUnreadMarkResponse, error)
	// GetUserInfo returns the profile for userID.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - userID — user id(s)
	//
	// Errors: errs.ZaloAPIError
	GetUserInfo(ctx context.Context, userID ...string) (*api.GetUserInfoResponse, error)
	// GetUserSummaryInfo retrieves user summary information by user ID.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - userID - user ID(s)
	//
	// Errors: errs.ZaloAPIError
	GetUserSummaryInfo(ctx context.Context, userID ...string) (*api.GetUserSummaryInfoResponse, error)
	// InviteUserToGroups invites a user to one or more groups.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - userID - user ID
	//   - groupID - group ID(s)
	//
	// Errors: errs.ZaloAPIError
	InviteUserToGroups(ctx context.Context, userID string, groupID ...string) (*api.InviteUserToGroupsResponse, error)
	// JoinGroupInviteBox joins a group via an invite box.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//
	// Note:
	//   - If membership approval is enabled and the invite is from a regular member, admin approval is required.
	//   - Invites from the owner or a deputy allow immediate joining.
	//
	// Errors: errs.ZaloAPIError
	JoinGroupInviteBox(ctx context.Context, groupID string) (api.JoinGroupInviteBoxResponse, error)
	// JoinGroupLink joins a group using an invite link.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - link - invite link to join the group
	//
	// Note: Zalo may return error code 240 if the group requires membership approval,
	//       or 178 if you are already a member.
	//
	// Errors: errs.ZaloAPIError
	JoinGroupLink(ctx context.Context, link string) (api.JoinGroupLinkResponse, error)
	// LastOnline retrieves the last online time of a user.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - userID - user ID
	//
	// Errors: errs.ZaloAPIError
	LastOnline(ctx context.Context, userID string) (*api.LastOnlineResponse, error)
	// LeaveGroup leaves a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID to leave
	//   - isSilent - whether to leave the group silently
	//
	// Note: Zalo may return error code 166 if you are not a member of the group.
	//
	// Errors: errs.ZaloAPIError
	LeaveGroup(ctx context.Context, groupID string, isSilent bool) (*api.LeaveGroupResponse, error)
	// LockPoll locks a poll, preventing further votes.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - pollID - poll ID to lock
	//
	// Errors: errs.ZaloAPIError
	LockPoll(ctx context.Context, pollID int) (api.LockPollResponse, error)
	// ParseLink parses a link and returns its metadata.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - link - link to parse
	//
	// Errors: errs.ZaloAPIError
	ParseLink(ctx context.Context, link string) (*api.ParseLinkResponse, error)
	// RejectFriendRequest rejects a friend request from a user.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - friendID - friend ID
	//
	// Errors: errs.ZaloAPIError
	RejectFriendRequest(ctx context.Context, friendID string) (api.RejectFriendRequestResponse, error)
	// RemoveAlias removes a user's alias.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - userID - user ID
	//
	// Errors: errs.ZaloAPIError
	RemoveAlias(ctx context.Context, userID string) (api.RemoveAliasResponse, error)
	// RemoveFriend removes a friend by their ID.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - friendID - ID of the friend to remove
	//
	// Errors: errs.ZaloAPIError
	RemoveFriend(ctx context.Context, friendID string) (api.RemoveFriendResponse, error)
	// RemoveGroupBlockedMember removes one or more members from a group's blocked list.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//   - memberID - member ID(s)
	//
	// Errors: errs.ZaloAPIError
	RemoveGroupBlockedMember(ctx context.Context, groupID string, memberID ...string) (api.RemoveGroupBlockedMemberResponse, error)
	// RemoveGroupDeputy removes one or more users from group deputy roles.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//   - memberID - user ID(s)
	//
	// Errors: errs.ZaloAPIError
	RemoveGroupDeputy(ctx context.Context, groupID string, memberID ...string) (api.RemoveGroupDeputyResponse, error)
	// RemoveGroupInviteBox removes pending group invitations for one or more groups.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - blockFutureInvite - if true, prevents future invitations from the specified groups
	//   - groupID - one or more group IDs
	//
	// Notes:
	//   - When blockFutureInvite is true, future invitations from these groups will be automatically blocked.
	//   - To re-enable invitations, join the group again via link or QR code.
	//
	// Errors: errs.ZaloAPIError
	RemoveGroupInviteBox(ctx context.Context, blockFutureInvite bool, groupID ...string) (*api.RemoveGroupInviteBoxResponse, error)
	// RemoveQuickMessage removes one or more quick messages.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - messageID - quick message id(s)
	//
	// Note: Zalo may return error code 212 if the specified item does not exist.
	//
	// Errors: errs.ZaloAPIError
	RemoveQuickMessage(ctx context.Context, messageID ...int) (*api.RemoveQuickMessageResponse, error)
	// RemoveReminder removes a reminder from a user or group thread.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - user or group ID to remove the reminder from
	//   - threadType - thread type
	//   - reminderID - reminder ID to remove
	//
	// Errors: errs.ZaloAPIError
	RemoveReminder(ctx context.Context, threadID string, threadType model.ThreadType, reminderID string) (api.RemoveReminderResponse, error)
	// RemoveUnreadMark removes the unread mark from a conversation.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - thread ID
	//   - threadType - thread type
	//
	// Errors: errs.ZaloAPIError
	RemoveUnreadMark(ctx context.Context, threadID string, threadType model.ThreadType) (*api.RemoveUnreadMarkResponse, error)
	// RemoveUserFromGroup removes one or more users from an existing group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//   - memberID - user ID(s)
	//
	// Note: Zalo may return error code 165 if the user is not in the group,
	//       or 166 if you lack permissions or are not in the group.
	//
	// Errors: errs.ZaloAPIError
	RemoveUserFromGroup(ctx context.Context, groupID string, memberID ...string) (*api.RemoveUserFromGroupResponse, error)
	// ResetHiddenChatPIN resets the PIN for hidden conversations.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//
	// Errors: errs.ZaloAPIError
	ResetHiddenChatPIN(ctx context.Context) (api.ResetHiddenChatPINResponse, error)
	// ReuseAvatar reuses an existing avatar from the avatar list.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - photoID - photo ID obtained from the GetAvatarList API
	//
	// Errors: errs.ZaloAPIError
	ReuseAvatar(ctx context.Context, photoID string) (*api.ReuseAvatarResponse, error)
	// ReviewPendingMemberRequest reviews pending membership requests for a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group id
	//   - data - data to review the pending member(s) request
	//
	// Note: Only the group leader and group deputy can perform this action.
	//
	// Errors: errs.ZaloAPIError
	ReviewPendingMemberRequest(ctx context.Context, groupID string, data api.ReviewPendingMemberData) (*api.ReviewPendingMemberRequestResponse, error)
	// SendBankCard sends a bank card to a specified thread.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - thread id
	//   - threadType - thread type
	//   - data - data containing the bank card information
	//
	// Errors: errs.ZaloAPIError
	SendBankCard(ctx context.Context, threadID string, threadType model.ThreadType, data api.SendBankCardData) (api.SendBankCardResponse, error)
	// SendCard sends a card to a user or group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - ID of the conversation
	//   - threadType - thread type
	//   - options - card options
	//
	// Errors: errs.ZaloAPIError
	SendCard(ctx context.Context, threadID string, threadType model.ThreadType, options api.SendCardOptions) (*api.SendCardResponse, error)
	// SendDeliveredEvent sends a message delivered event.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - messages - list of messages to send the delivered event for
	//   - threadType - message type (User or Group); defaults to User
	//   - isSeen - whether the message is seen or not
	//
	// Errors: errs.ZaloAPIError, errs.ErrInvalidMessageCount, errs.ErrInconsistentGroupRecipient
	//
	// @TODO: test this api
	SendDeliveredEvent(ctx context.Context, messages []model.OutboundMessage, threadType model.ThreadType, isSeen bool) (api.SendDeliveredEventResponse, error)
	// SendFriendRequest sends a friend request to a user.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - message - message sent with the friend request
	//   - userID - user ID to send the friend request to
	//
	// Note: Zalo may return error code 225 if the user is already your friend,
	//       215 if the user has blocked you, or 222 if the user has already sent you a friend request
	//       (your request will then be treated as acceptance).
	//
	// Errors: errs.ZaloAPIError
	SendFriendRequest(ctx context.Context, message string, userID string) (api.SendFriendRequestResponse, error)
	// SendGIF sends a GIF message.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - thread ID
	//   - threadType - thread type
	//   - gif - GIF content
	//
	// Errors: errs.ZaloAPIError, errs.ErrMissingImageMetadataGetter
	SendGIF(ctx context.Context, threadID string, threadType model.ThreadType, gif api.GIFContent) (*api.SendGIFResponse, error)
	// SendLink sends a link message.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - thread ID
	//   - threadType - thread type
	//   - options - link and TTL information
	//
	// Errors: errs.ZaloAPIError
	SendLink(ctx context.Context, threadID string, threadType model.ThreadType, options api.SendLinkOptions) (*api.SendLinkResponse, error)
	// SendMessage sends a message to a thread.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - group or user ID
	//   - threadType - thread type
	//   - message - message content
	//
	// Errors: errs.ZaloAPIError, errs.ErrMissingImageMetadataGetter
	//
	// TODO: test this api
	SendMessage(ctx context.Context, threadID string, threadType model.ThreadType, message api.MessageContent) (*api.SendMessageResponse, error)
	// SendReport sends a report to Zalo.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - thread ID to report
	//   - threadType - thread type
	//   - options - report options
	//
	// Errors: errs.ZaloAPIError
	SendReport(ctx context.Context, threadID string, threadType model.ThreadType, options api.SendReportOptions) (*api.SendReportResponse, error)
	// SendSeenEvent sends a message seen event.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - messages - list of messages to send the seen event for
	//   - threadType - thread type
	//
	// Errors: errs.ZaloAPIError, errs.ErrInvalidMessageCount, errs.ErrInconsistentGroupRecipient
	//
	// @TODO: test this api
	SendSeenEvent(ctx context.Context, messages []model.OutboundMessage, threadType model.ThreadType) (*api.SendSeenEventResponse, error)
	// SendSticker sends a sticker to a thread.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - group or user ID
	//   - threadType - thread type
	//   - sticker - sticker object
	//
	// Errors: errs.ZaloAPIError
	SendSticker(ctx context.Context, threadID string, threadType model.ThreadType, sticker api.SendStickerPayload) (*api.SendStickerResponse, error)
	// SendTypingEvent sends a typing event to a user or group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - ID of the user or group to send the typing event to
	//   - threadType - thread type
	//   - destType - destination type (User or Page); for user threads only, defaults to User
	//
	// Errors: errs.ZaloAPIError
	SendTypingEvent(ctx context.Context, threadID string, threadType model.ThreadType, destType model.DestType) (*api.SendTypingEventResponse, error)
	// SendVideo sends a video to a user or group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - ID of the user or group to send the video to
	//   - threadType - thread type
	//   - options - send video options
	//
	// Errors: errs.ZaloAPIError
	//
	// Examples:
	//   - Standard Videos:
	//      - 3840x2160 (4K UHD): width 3840px, height 2160px
	//      - 1920x1080 (Full HD): width 1920px, height 1080px
	//      - 1280x720 (HD): width 1280px, height 720px
	//   - Document-Oriented Videos (Portrait):
	//      - 3840x2160 (4K UHD): width 3840px, height 2160px
	//      - 720x1280 (HD): width 720px, height 1280px
	//      - 1440x2560 (2K): width 1440px, height 2560px
	SendVideo(ctx context.Context, threadID string, threadType model.ThreadType, options api.SendVideoOptions) (*api.SendVideoResponse, error)
	// SendVoice sends a voice message to a user or group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - ID of the user or group to send the voice message to
	//   - threadType - thread type
	//   - options - voice message options
	//
	// Errors: errs.ZaloAPIError
	SendVoice(ctx context.Context, threadID string, threadType model.ThreadType, options api.SendVoiceOptions) (*api.SendVoiceResponse, error)
	// SetHiddenChat hides or unhides conversations.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - thread IDs
	//   - threadType - thread type
	//   - isHidden - whether to hide or unhide conversations
	//
	// Errors: errs.ZaloAPIError
	SetHiddenChat(ctx context.Context, threadID []string, threadType model.ThreadType, isHidden bool) (api.SetHiddenChatResponse, error)
	// SetMute sets mute preferences for a thread.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - ID of the thread to mute
	//   - threadType - thread type
	//   - options - mute options
	//
	// Errors: errs.ZaloAPIError
	SetMute(ctx context.Context, threadID string, threadType model.ThreadType, options api.SetMuteOptions) (api.SetMuteResponse, error)
	// SetPinChat pins or unpins conversations.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - ID(s) of the thread
	//   - threadType - thread type (default: User)
	//   - isPinned - whether to pin conversations
	//
	// Errors: errs.ZaloAPIError
	SetPinChat(ctx context.Context, threadID []string, threadType model.ThreadType, isPinned bool) (api.SetPinChatResponse, error)
	// SetViewFeedBlock blocks or unblocks a user's ability to view the feed.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - userID — user ID to block or unblock feed view
	//   - isBlock — boolean to block or unblock feed view
	//
	// Errors: errs.ZaloAPIError
	SetViewFeedBlock(ctx context.Context, userID string, isBlock bool) (api.SetViewFeedBlockResponse, error)
	// SharePoll shares a poll.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - pollID - poll ID to share
	//
	// Errors: errs.ZaloAPIError
	SharePoll(ctx context.Context, pollID int) (api.SharePollResponse, error)
	// UnblockUser unblocks a user by their ID.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - userID — ID of the user to unblock
	//
	// Errors: errs.ZaloAPIError
	UnblockUser(ctx context.Context, userID string) (api.UnblockUserResponse, error)
	// UndoFriendRequest cancels a previously sent friend request.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - userID - user id
	//
	// Errors: errs.ZaloAPIError
	UndoFriendRequest(ctx context.Context, userID string) (api.UndoFriendRequestResponse, error)
	// UndoMessage recalls a message you previously sent.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - target thread ID
	//   - threadType - type of the thread
	//   - data - message data to recall
	//
	// Note: Applies only to messages you sent, and only if they were sent within the last 24 hours.
	//
	// Errors: errs.ZaloAPIError
	UndoMessage(ctx context.Context, threadID string, threadType model.ThreadType, data api.UndoMessageData) (*api.UndoMessageResponse, error)
	// UpdateAccountAvatar changes the account avatar.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - source - attachment source
	//
	// Errors: errs.ZaloAPIError, errs.ZaloAPIMissingImageMetadataGetter
	UpdateAccountAvatar(ctx context.Context, source model.AttachmentSource) (api.UpdateAccountAvatarResponse, error)
	// UpdateActiveStatus updates the active status of the account.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - isActive - whether the account is active
	//
	// Errors: errs.ZaloAPIError
	UpdateActiveStatus(ctx context.Context, isActive bool) (*api.UpdateActiveStatusResponse, error)
	// UpdateAlias changes a user's alias.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - userID - user ID
	//   - alias - new alias
	//
	// Errors: errs.ZaloAPIError
	UpdateAlias(ctx context.Context, userID string, alias string) (api.UpdateAliasResponse, error)
	// UpdateAutoDeleteChat enables automatic deletion for a conversation.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - thread ID to auto-delete
	//   - threadType - thread type
	//   - ttl - time to live of the conversation
	//
	// Errors: errs.ZaloAPIError
	UpdateAutoDeleteChat(ctx context.Context, threadID string, threadType model.ThreadType, ttl api.ChatTTL) (api.UpdateAutoDeleteChatResponse, error)
	// UpdateGroupAvatar changes the group avatar.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//   - source - attachment source
	//
	// Errors: errs.ZaloAPIError, errs.ErrMissingImageMetadataGetter
	UpdateGroupAvatar(ctx context.Context, groupID string, source model.AttachmentSource) (api.UpdateGroupAvatarResponse, error)
	// UpdateGroupName changes the name of a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//   - name - new group name
	//
	// Errors: errs.ZaloAPIError
	UpdateGroupName(ctx context.Context, groupID string, name string) (*api.UpdateGroupNameResponse, error)
	// UpdateGroupSetting updates group settings.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID
	//   - options - settings options
	//
	// Note: Zalo may return error code 166 if you lack permissions to change the settings.
	//
	// Errors: errs.ZaloAPIError
	UpdateGroupSetting(ctx context.Context, groupID string, options api.UpdateGroupSettingOptions) (api.UpdateGroupSettingResponse, error)
	// UpdateHiddenChatPIN updates the PIN for a hidden conversation.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - pin - PIN to update (must be a 4-digit number between 0000–9999)
	//
	// Errors: errs.ZaloAPIError
	UpdateHiddenChatPIN(ctx context.Context, pin string) (api.UpdateHiddenChatPINResponse, error)
	// UpdateLabels updates label data.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - label - label data
	//
	// Errors: errs.ZaloAPIError
	UpdateLabels(ctx context.Context, labels api.UpdateLabelsData) (*api.UpdateLabelsResponse, error)
	// UpdateLanguage sets the user's language.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - lang — target language ("VI", "EN")
	//
	// Note: Calling this endpoint alone will not update the user's language.
	//
	// Errors: errs.ZaloAPIError
	UpdateLanguage(ctx context.Context, lang api.Language) (api.UpdateLanguageResponse, error)
	// UpdateNote edits an existing note in a group.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - groupID - group ID to edit the note in
	//   - options - options for updating the note
	//
	// Errors: errs.ZaloAPIError
	UpdateNote(ctx context.Context, groupID string, options api.UpdateNoteOptions) (*api.UpdateNoteResponse, error)
	// UpdateProfile changes the account setting information.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - data - payload data
	//
	// Note:
	//   - If the account is a Business Account, include the biz.cate field; otherwise, the category will be removed.
	//   - You may leave other biz fields empty if no changes are required.
	//
	// Errors: errs.ZaloAPIError
	UpdateProfile(ctx context.Context, data api.UpdateProfileData) (api.UpdateProfileResponse, error)
	// UpdateQuickMessage updates an existing quick message.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - itemID - ID of the quick message to update
	//   - data - payload containing data to update the quick message
	//
	// Note: Zalo may return error code 212 if the specified itemID does not exist.
	//
	// Errors: errs.ZaloAPIError
	UpdateQuickMessage(ctx context.Context, itemID string, data api.UpdateQuickMessageData) (*api.UpdateQuickMessageResponse, error)
	// UpdateReminder edits an existing reminder.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - thread ID
	//   - threadType - thread type
	//   - options - reminder options
	//
	// Errors: errs.ZaloAPIError
	UpdateReminder(ctx context.Context, threadID string, threadType model.ThreadType, options api.UpdateReminderOptions) (*api.UpdateReminderResponse, error)
	// UpdateSetting sets account configuration options.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - sType - setting type
	//   - value - new value for the setting
	//
	// Note:
	//   - Ensure the provided value is valid for the specified setting type.
	//   - Refer to [api.UpdateSettingsType] documentation for allowed value ranges.
	//
	// Errors: errs.ZaloAPIError
	UpdateSetting(ctx context.Context, sType api.UpdateSettingType, value int) (api.UpdateSettingResponse, error)
	// UploadAttachment uploads one or more attachments to a thread.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - threadID - group or user ID
	//   - type - message type (User or Group)
	//   - sources - path to files or attachment sources
	//
	// Errors:
	//   - errs.ZaloAPIError, errs.ErrMissingImageMetadataGetter
	//   - errs.ErrSourceEmpty, errs.ErrExceedMaxFile, errs.ErrInvalidExtension, errs.ErrExceedMaxFileSize
	UploadAttachment(ctx context.Context, threadID string, threadType model.ThreadType, sources ...model.AttachmentSource) (api.UploadAttachmentResponse, error)
	// UploadPhoto uploads a product photo for quick message,
	// product catalog, or custom local storage.
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - source — file path or attachment source
	//
	// Errors: errs.ZaloAPIError, errs.ErrMissingImageMetadataGetter
	UploadPhoto(ctx context.Context, source model.AttachmentSource) (*api.UploadPhotoResponse, error)
	// UploadThumbnail uploads an attachment thumbnail
	//
	// Params:
	//   - ctx — cancel/deadline control
	//   - source — file path or attachment source
	//
	// Errors: errs.ZaloAPIError, errs.ErrMissingImageMetadataGetter
	UploadThumbnail(ctx context.Context, source model.AttachmentSource) (*api.UploadThumbnailResponse, error)
	// VotePoll submits a vote on a poll.
	//
	// Params:
	//   - ctx - cancel/deadline control
	//   - pollID - ID of the poll to vote on
	//   - optionID - ID(s) of the option to vote on
	//
	// Errors: errs.ZaloAPIError
	VotePoll(ctx context.Context, pollID string, optionID ...int) (*api.VotePollResponse, error)
}
