package model

import "github.com/Amrakk/zcago/config"

type ReactionIcon string

const (
	ReactionHeart       ReactionIcon = `/-heart`
	ReactionLike        ReactionIcon = `/-strong`
	ReactionHaha        ReactionIcon = `:>`
	ReactionWow         ReactionIcon = `:o`
	ReactionCry         ReactionIcon = `:-((`
	ReactionAngry       ReactionIcon = `:-h`
	ReactionKiss        ReactionIcon = `:-*`
	ReactionTearsOfJoy  ReactionIcon = `:\')`
	ReactionShit        ReactionIcon = `/-shit`
	ReactionRose        ReactionIcon = `/-rose`
	ReactionBrokenHeart ReactionIcon = `/-break`
	ReactionDislike     ReactionIcon = `/-weak`
	ReactionLove        ReactionIcon = `;xx`
	ReactionConfused    ReactionIcon = `;-/`
	ReactionWink        ReactionIcon = `;-)`
	ReactionFade        ReactionIcon = `/-fade`
	ReactionSun         ReactionIcon = `/-li`
	ReactionBirthday    ReactionIcon = `/-bd`
	ReactionBomb        ReactionIcon = `/-bome`
	ReactionOK          ReactionIcon = `/-ok`
	ReactionPeace       ReactionIcon = `/-v`
	ReactionThanks      ReactionIcon = `/-thanks`
	ReactionPunch       ReactionIcon = `/-punch`
	ReactionShare       ReactionIcon = `/-share`
	ReactionPray        ReactionIcon = `_()_`
	ReactionNo          ReactionIcon = `/-no`
	ReactionBad         ReactionIcon = `/-bad`
	ReactionLoveYou     ReactionIcon = `/-loveu`
	ReactionSad         ReactionIcon = `--b`
	ReactionVerySad     ReactionIcon = `:((`
	ReactionCool        ReactionIcon = `x-)`
	ReactionNerd        ReactionIcon = `8-)`
	ReactionBigSmile    ReactionIcon = `;-d`
	ReactionSunglasses  ReactionIcon = `b-)`
	ReactionNeutral     ReactionIcon = `:--|`
	ReactionSadFace     ReactionIcon = `p-(`
	ReactionBye         ReactionIcon = `:-bye`
	ReactionSleepy      ReactionIcon = `|-)`
	ReactionWipe        ReactionIcon = `:wipe`
	ReactionDig         ReactionIcon = `:-dig`
	ReactionAnguish     ReactionIcon = `&-(`
	ReactionHandclap    ReactionIcon = `:handclap`
	ReactionAngryFace   ReactionIcon = `>-|`
	ReactionFChair      ReactionIcon = `:-f`
	ReactionLChair      ReactionIcon = `:-l`
	ReactionRChair      ReactionIcon = `:-r`
	ReactionSilent      ReactionIcon = `;-x`
	ReactionSurprise    ReactionIcon = `:-o`
	ReactionEmbarrassed ReactionIcon = `;-s`
	ReactionAfraid      ReactionIcon = `;-a`
	ReactionSad2        ReactionIcon = `:-<`
	ReactionBigLaugh    ReactionIcon = `:))`
	ReactionRich        ReactionIcon = `$-)`
	ReactionBeer        ReactionIcon = `/-beer`
	ReactionNone        ReactionIcon = ``

	// more...

	DefaultReactionSource = 6
)

var reactionType = map[ReactionIcon]int{
	ReactionHaha:        0,
	ReactionLike:        3,
	ReactionHeart:       5,
	ReactionWow:         32,
	ReactionCry:         2,
	ReactionAngry:       20,
	ReactionKiss:        8,
	ReactionTearsOfJoy:  7,
	ReactionShit:        66,
	ReactionRose:        120,
	ReactionBrokenHeart: 65,
	ReactionDislike:     4,
	ReactionLove:        29,
	ReactionConfused:    51,
	ReactionWink:        45,
	ReactionFade:        121,
	ReactionSun:         67,
	ReactionBirthday:    126,
	ReactionBomb:        127,
	ReactionOK:          68,
	ReactionPeace:       69,
	ReactionThanks:      70,
	ReactionPunch:       71,
	ReactionShare:       72,
	ReactionPray:        73,
	ReactionNo:          131,
	ReactionBad:         132,
	ReactionLoveYou:     133,
	ReactionSad:         1,
	ReactionVerySad:     16,
	ReactionCool:        21,
	ReactionNerd:        22,
	ReactionBigSmile:    23,
	ReactionSunglasses:  26,
	ReactionNeutral:     30,
	ReactionSadFace:     35,
	ReactionBye:         36,
	ReactionSleepy:      38,
	ReactionWipe:        39,
	ReactionDig:         42,
	ReactionAnguish:     44,
	ReactionHandclap:    46,
	ReactionAngryFace:   47,
	ReactionFChair:      48,
	ReactionLChair:      49,
	ReactionRChair:      50,
	ReactionSilent:      52,
	ReactionSurprise:    53,
	ReactionEmbarrassed: 54,
	ReactionAfraid:      60,
	ReactionSad2:        61,
	ReactionBigLaugh:    62,
	ReactionRich:        63,
	ReactionBeer:        99,
}

func (icon ReactionIcon) TypeCode() int {
	if t, ok := reactionType[icon]; ok {
		return t
	}
	return -1
}

type Reaction struct {
	Type     ThreadType
	Data     TReaction
	ThreadID string
	IsSelf   bool
}

func NewReaction(uid string, data TReaction, threadType ThreadType) Reaction {
	if data.IDTo == config.DefaultUIDSelf {
		data.IDTo = uid
	}
	if data.UIDFrom == config.DefaultUIDSelf {
		data.UIDFrom = uid
	}

	isSelf := data.UIDFrom == config.DefaultUIDSelf

	threadID := data.UIDFrom
	if threadType == ThreadTypeGroup || isSelf {
		threadID = data.IDTo
	}

	return Reaction{
		Type:     threadType,
		Data:     data,
		ThreadID: threadID,
		IsSelf:   isSelf,
	}
}

type TReaction struct {
	ActionID string          `json:"actionId"`
	MsgID    string          `json:"msgId"`
	CliMsgID string          `json:"cliMsgId"`
	MsgType  string          `json:"msgType"`
	UIDFrom  string          `json:"uidFrom"`
	IDTo     string          `json:"idTo"`
	DName    *string         `json:"dName,omitempty"`
	Content  ReactionContent `json:"content"`
	TS       string          `json:"ts"`
	TTL      int             `json:"ttl"`
}

type ReactionData struct {
	RIcon  ReactionIcon `json:"rIcon"`
	RType  int          `json:"r.RType"`
	Source int          `json:"r.Source"`
}

func NewReactionData(icon ReactionIcon) ReactionData {
	return ReactionData{
		RIcon:  icon,
		Source: DefaultReactionSource,
		RType:  icon.TypeCode(),
	}
}

func (rid ReactionData) IsValid() bool {
	return len(rid.RIcon) != 0
}

type ReactionContent struct {
	RMsg []ReactionMessageRef `json:"rMsg"`
	ReactionData
}

type ReactionMessageRef struct {
	GMsgID  string `json:"gMsgID"`
	CMsgID  string `json:"cMsgID"`
	MsgType int    `json:"msgType"`
}
