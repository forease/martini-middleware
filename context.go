package middleware

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"net/http"
)

type (
	Context struct {
		render.Render
		C        martini.Context
		P        martini.Params
		S        sessions.Session
		R        *http.Request
		W        http.ResponseWriter
		FormErr  binding.Errors
		Messages []string
		Errors   []string
		Data     map[string]interface{}
		//DbUtil   *model.DbUtil
	}

	HTMLOptions struct {
		Layout   string
		Template string
	}
)

func (self *Context) init() {
	if self.Data == nil {
		self.Data = make(map[string]interface{})
	}

}

func (self *Context) Get(key string) interface{} {
	return self.Data[key]
}

func (self *Context) Set(key string, val interface{}) {
	self.init()
	self.Data[key] = val
}

func (self *Context) Delete(key string) {
	delete(self.Data, key)
}

func (self *Context) Clear() {
	self.Data = make(map[string]interface{})
}

func (self *Context) AddMessage(message string) {
	self.Messages = append(self.Messages, message)
}

func (self *Context) ClearMessages() {
	self.Messages = self.Messages[:0]
}

func (self *Context) HasMessage() bool {
	return (len(self.Messages) > 0)
}

// Render JSON message
func (self *Context) JMessage(code int, message, url string) {
	msg := Msg{Code: code, Message: message, Url: url}
	self.JSON(200, msg)
}

// Render HTML message
func (self *Context) HMessage(code int, message, url string) {
	self.Set("msg", Msg{Code: code, Message: message, Url: url})
	self.HTML(200, "message", self.Data)
}

// Parse HTML code
func (self *Context) HTML(code int, name string, binding interface{}, htmlOpt ...HTMLOptions) {

	if len(htmlOpt) > 0 {
		if len(htmlOpt[0].Template) > 0 {
			self.Template().Parse(htmlOpt[0].Template)
		}

		opt := render.HTMLOptions{Layout: htmlOpt[0].Layout}
		self.Render.HTML(200, name, binding, opt)
	} else {
		self.Render.HTML(200, name, binding)
	}
}

func InitContext() martini.Handler {
	return func(c martini.Context, rnd render.Render, r *http.Request,
		w http.ResponseWriter, s sessions.Session) {
		ctx := &Context{
			Render: rnd,
			W:      w,
			R:      r,
			C:      c,
			S:      s,
			//DbUtil: &model.DbUtil{},
		}

		//lang := s.Get("Lang")
		//if lang == nil {
		//	s.Set("Lang", Cfg.MustValue("", "locale", "en"))
		//}

		//s.Set("Settings", model.GetSettings())
		c.Map(ctx)
	}
}

func InitContextNoSess() martini.Handler {
	return func(c martini.Context, rnd render.Render, r *http.Request,
		w http.ResponseWriter) {
		ctx := &Context{
			Render: rnd,
			W:      w,
			R:      r,
			C:      c,
		}

		c.Map(ctx)
	}
}

func InitSession(name, key string) martini.Handler {
	var store sessions.Store
	store = sessions.NewCookieStore([]byte(key))

	return sessions.Sessions(name, store)
}
