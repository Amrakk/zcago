package jsonx

import (
	"encoding/json"

	"github.com/Amrakk/zcago/internal/errs"
)

type OneOrMany[T any] struct {
	Values []T
}

func (o OneOrMany[T]) Single() (T, bool) {
	var zero T
	if len(o.Values) == 1 {
		return o.Values[0], true
	}
	return zero, false
}

func (o OneOrMany[T]) Slice() []T {
	if len(o.Values) == 0 {
		return nil
	}
	cp := make([]T, len(o.Values))
	copy(cp, o.Values)
	return cp
}

func (o *OneOrMany[T]) UnmarshalJSON(b []byte) error {
	var single T
	if err := json.Unmarshal(b, &single); err == nil {
		o.Values = []T{single}
		return nil
	}
	var many []T
	if err := json.Unmarshal(b, &many); err == nil {
		o.Values = many
		return nil
	}
	return errs.NewZCAError("value must be a single item or an array", "OneOrMany.UnmarshalJSON", nil)
}

func (o OneOrMany[T]) MarshalJSON() ([]byte, error) {
	if len(o.Values) == 1 {
		return json.Marshal(o.Values[0])
	}
	return json.Marshal(o.Values)
}
