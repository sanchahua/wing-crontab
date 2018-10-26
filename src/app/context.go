package app

type Context struct {
	offline bool
}

func NewContext() *Context {
	return &Context{
		offline:false,
	}
}

func (con *Context) SetOffline(offline bool) {
	con.offline = offline
}

func (con *Context) IsOffline() bool {
	return con.offline
}