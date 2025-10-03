package api

func (a *api) GetOwnID() string {
	return a.sc.UID()
}
