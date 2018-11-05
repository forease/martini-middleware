package middleware

import (
	"github.com/go-martini/martini"
	"net/http"
)

// setup not need check authorize URL
//
//  	urls := []string{"/login", "/users", "/firewall"}
//	    m.Use(middleware.AuthRequest(urls...))
//
func AuthRequest(disUrls ...string) martini.Handler {
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

		urlLen := len(disUrls)

		if urlLen < 1 {
			disUrls[0] = "/login"
			urlLen = 1
		}

		for _, url := range disUrls {
			if c.R.RequestURI == url {
				match = true
				break
			}
		}

		if !match && c.R.RequestURI != "/login" && c.R.RequestURI != "/users/logined" {
			c.Redirect(disUrls[0])
		}

		ctx.Next()
	}
}
