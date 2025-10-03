package api

import "github.com/Amrakk/zcago/listener"

func (a *api) Listener() listener.Listener {
	return a.l
}
