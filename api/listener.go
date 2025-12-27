package api

import "github.com/amrakk/zcago/listener"

func (a *api) Listener() listener.Listener {
	return a.l
}
