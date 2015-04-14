package middleware

import (
	"github.com/go-martini/martini"
	"reflect"
)

const (
	SignInRequired = 100
)

func AuthRequest(req interface{}, redirect string) martini.Handler {
	return func(c *Context) {
		if redirect == "" {
			redirect = "/login"
		}

		switch req {
		case SignInRequired:
			if user := c.S.Get("Authorized"); user != nil {
				// pass
				return
			}
			c.Redirect(redirect)
			return
		default:
			if user := c.S.Get("Authorized"); user != nil {
				// pass
				return
				if reflect.TypeOf(req).Kind() == reflect.Int {

					c.HTML(403, "error/403", c)
					return
				}
			} else {
				c.Redirect(redirect)
				return
			}
			c.HTML(403, "error/403", c)
			return
		}
	}
}
