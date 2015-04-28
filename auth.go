package middleware

import (
	"github.com/go-martini/martini"
	"net/http"
)

//
//  	urls := []string{"/login", "/users", "/firewall"}
//	    m.Use(middleware.AuthRequest(urls...))
//
func AuthRequest(disurls ...string) martini.Handler {
	return func(ctx martini.Context, c *Context, r *http.Request) {
		var (
			match bool
		)

		if authored := c.S.Get("Authorized"); authored != nil {
			if u := c.S.Get("AuthInfo"); u != nil {
				c.Set("AuthData", u)
				ctx.Next()
				return
			}
		}

		urlLen := len(disurls)

		if urlLen < 1 {
			disurls[0] = "/login"
			urlLen = 1
		}

		for i := 0; i < urlLen; i++ {
			if r.RequestURI == disurls[i] {
				match = true
				break
			}
		}

		if !match && r.RequestURI != "/login" && r.RequestURI != "/users/logined" {
			c.Redirect(disurls[0])
		}

		ctx.Next()
	}
}
