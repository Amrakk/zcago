package events

import (
	"encoding/json"

	"github.com/Amrakk/zcago/model"
)

type ReactionEventData struct {
	Reactions      []model.TReaction
	GroupReactions []model.TReaction
}

func (m *ReactionEventData) UnmarshalJSON(data []byte) error {
	var raw struct {
		Reacts      json.RawMessage `json:"reacts"`
		ReactGroups json.RawMessage `json:"reactGroups"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	reactions, err := unmarshalReactions(raw.Reacts)
	if err != nil {
		return err
	}
	groupReactions, err := unmarshalReactions(raw.ReactGroups)
	if err != nil {
		return err
	}

	m.Reactions = reactions
	m.GroupReactions = groupReactions
	return nil
}

func unmarshalReactions(b json.RawMessage) ([]model.TReaction, error) {
	if len(b) == 0 || string(b) == "null" {
		return nil, nil
	}

	var items []json.RawMessage
	if len(b) > 0 && b[0] == '[' {
		if err := json.Unmarshal(b, &items); err != nil {
			return nil, err
		}
	} else {
		items = []json.RawMessage{b}
	}

	type alias model.TReaction

	out := make([]model.TReaction, 0, len(items))
	for _, it := range items {
		var r model.TReaction
		aux := &struct {
			Content string `json:"content"`
			*alias
		}{
			alias: (*alias)(&r),
		}

		if err := json.Unmarshal(it, &aux); err != nil {
			return nil, err
		}

		if aux.Content == "" {
			r.Content = model.ReactionContent{}
		} else {
			var content model.ReactionContent
			if err := json.Unmarshal([]byte(aux.Content), &content); err != nil {
				return nil, err
			}

			r.Content = content
		}

		out = append(out, r)
	}
	return out, nil
}
