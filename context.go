package middleware

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
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
		dataLock *sync.RWMutex
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
	self.dataLock.RLock()
	defer self.dataLock.RUnlock()

	return self.Data[key]
}

func (self *Context) Set(key string, val interface{}) {
	self.dataLock.Lock()
	self.init()
	self.Data[key] = val
	self.dataLock.Unlock()
}

func (self *Context) Delete(key string) {
	self.dataLock.Lock()
	delete(self.Data, key)
	self.dataLock.Unlock()
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
func (self *Context) JMessage(code int, url, message string, v ...interface{}) {
	msg := Msg{Code: code, Message: fmt.Sprintf(message, v), Url: url}
	self.JSON(200, msg)
}

// Render HTML message
func (self *Context) HMessage(code int, url, message string, v ...interface{}) {
	self.Set("msg", Msg{Code: code, Message: fmt.Sprintf(message, v), Url: url})
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
			Render:   rnd,
			W:        w,
			R:        r,
			C:        c,
			S:        s,
			dataLock: new(sync.RWMutex),
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
