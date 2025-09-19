package jsonx

import (
	"encoding/json"

	"github.com/Amrakk/zcago/internal/errs"
)

type OneOrMany[T any] struct {
	Values []T
}

func NewOne[T any](v T) OneOrMany[T]      { return OneOrMany[T]{Values: []T{v}} }
func NewMany[T any](vs ...T) OneOrMany[T] { return OneOrMany[T]{Values: append([]T(nil), vs...)} }

func (o OneOrMany[T]) IsSingle() bool { return len(o.Values) == 1 }
func (o OneOrMany[T]) First() (T, bool) {
	var zero T
	if len(o.Values) == 0 {
		return zero, false
	}
	return o.Values[0], true
}
func (o OneOrMany[T]) AsSlice() []T { return append([]T(nil), o.Values...) }

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
