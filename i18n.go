package middleware

import (
	"github.com/forease/i18n"
	"github.com/go-martini/martini"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	_VERSION              = "0.0.1"
	_LANGUAGE_COOKIE_NAME = "_language"
	_LANGUAGE_PARAM_NAME  = "_language"
	_DEFAULT_DOMAIN       = "martini"
	_DEFAULT_DIRECTORY    = "locale"
	_DEFAULT_LANGUAGE     = "zh_CN"
	_DEFAULT_TMPL_NAME    = "trans"
)

var Locale *i18n.Locale

func Version() string {
	return _VERSION
}

// Options represents a struct for specifying configuration options for the i18n middleware.
type I18nOptions struct {
	Domain      string
	Directory   string
	ZipData     []byte
	DefaultLang string
	// Suburl of path. Default is empty.
	SubURL     string
	CookieName string
	// Name of language parameter name in URL. Default is "lang".
	Parameter string
	// Redirect when user uses get parameter to specify language.
	Redirect bool
	// Name that maps into template variable. Default is "i18n".
	TmplName string
	Inited   bool
}

func prepareOptions(options []I18nOptions) I18nOptions {
	var opt I18nOptions
	if len(options) > 0 {
		opt = options[0]
	}

	opt.SubURL = strings.TrimSuffix(opt.SubURL, "/")
	if len(opt.Domain) == 0 {
		opt.Domain = _DEFAULT_DOMAIN
	}
	if len(opt.Directory) == 0 {
		opt.Directory = _DEFAULT_DIRECTORY
	}
	if len(opt.DefaultLang) == 0 {
		opt.DefaultLang = _DEFAULT_LANGUAGE
	}
	if len(opt.CookieName) == 0 {
		opt.CookieName = _LANGUAGE_COOKIE_NAME
	}
	if len(opt.Parameter) == 0 {
		opt.Parameter = _LANGUAGE_PARAM_NAME
	}
	if len(opt.TmplName) == 0 {
		opt.TmplName = _DEFAULT_TMPL_NAME
	}
	if !opt.Redirect {
		opt.Redirect = true
	}

	return opt
}

func initLocale(opt I18nOptions) {
	var once sync.Once
	onceBody := func() {
		if !i18n.IsExist(opt.DefaultLang) {
			i18n.SetMessage(opt.DefaultLang, opt.Directory+"/locale_"+toLocale(opt.DefaultLang, false)+".ini")
		}
	}
	if !opt.Inited {
		once.Do(onceBody)
	}
}

// I18n is a middleware provides localization layer for your application.
// Paramenter langs must be in the form of "en-US", "zh-CN", etc.
// Otherwise it may not recognize browser input.
func I18n(options ...I18nOptions) martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, c *Context) {
		isNeedRedir := false
		hasCookie := false
		opt := prepareOptions(options)
		initLocale(opt)
		// 1. Check URL arguments.
		lang := req.FormValue(opt.Parameter)
		// 2. Get language information from cookies.
		if len(lang) == 0 {
			var err error
			if lang, err = getCookie(req, opt.CookieName); err == nil {
				hasCookie = true
			}
		} else {
			isNeedRedir = true
		}
		// 3. Get language information from 'Accept-Language'.
		if len(lang) == 0 {
			al := req.Header.Get("Accept-Language")
			if len(al) > 4 {
				al = al[:5] // Only compare first 5 letters.
				lang = al
			}
		}
		// 4. Default language is the first element in the list.
		if len(lang) == 0 {
			lang = opt.DefaultLang
			isNeedRedir = false
		}

		language := toLocale(lang, false)
		if language != opt.DefaultLang && !i18n.IsExist(lang) {
			err := i18n.SetMessage(language, opt.Directory+"/locale_"+language+".ini")
			if err != nil {
				language = opt.DefaultLang
			}
		}

		// Save language information in cookies.
		if !hasCookie {
			setCookie(res, opt.CookieName, language, 1<<31-1, "/"+strings.TrimPrefix(opt.SubURL, "/"))
		}

		Locale = &i18n.Locale{Lang: language}
		c.Set("i18n", Locale)
		if opt.Redirect && isNeedRedir {
			location := opt.SubURL + req.RequestURI[:strings.Index(req.RequestURI, "?")]
			http.Redirect(res, req, location, http.StatusFound)
		}
	}
}

func Tr(format string, args ...interface{}) string {
	return Locale.Tr(format, args...)
}

// GetCookie returns given cookie value from request header.
func getCookie(req *http.Request, name string) (string, error) {
	cookie, err := req.Cookie(name)
	if err != nil {
		return "", err
	}
	return url.QueryUnescape(cookie.Value)
}

// SetCookie sets given cookie value to response header.
func setCookie(resp http.ResponseWriter, name string, value string, others ...interface{}) {
	cookie := http.Cookie{}
	cookie.Name = name
	cookie.Value = url.QueryEscape(value)

	if len(others) > 0 {
		switch v := others[0].(type) {
		case int:
			cookie.MaxAge = v
		case int64:
			cookie.MaxAge = int(v)
		case int32:
			cookie.MaxAge = int(v)
		}
	}

	// default "/"
	if len(others) > 1 {
		if v, ok := others[1].(string); ok && len(v) > 0 {
			cookie.Path = v
		}
	} else {
		cookie.Path = "/"
	}

	// default empty
	if len(others) > 2 {
		if v, ok := others[2].(string); ok && len(v) > 0 {
			cookie.Domain = v
		}
	}

	// default empty
	if len(others) > 3 {
		switch v := others[3].(type) {
		case bool:
			cookie.Secure = v
		default:
			if others[3] != nil {
				cookie.Secure = true
			}
		}
	}

	// default false. for session cookie default true
	if len(others) > 4 {
		if v, ok := others[4].(bool); ok && v {
			cookie.HttpOnly = true
		}
	}

	resp.Header().Add("Set-Cookie", cookie.String())
}

// Turns a language name (en-us) into a locale name (en_US). If 'to_lower' is
// True, the last component is lower-cased (en_us).
func toLocale(language string, to_lower bool) string {
	if p := strings.Index(language, "-"); p >= 0 {
		if to_lower {
			return strings.ToLower(language[:p]) + "_" + strings.ToLower(language[p+1:])
		} else {
			//# Get correct locale for sr-latn
			if len(language[p+1:]) > 2 {
				return strings.ToLower(language[:p]) + "_" + strings.ToUpper(string(language[p+1])) + strings.ToLower(language[p+2:])
			}
			return strings.ToLower(language[:p]) + "_" + strings.ToUpper(language[p+1:])
		}
	} else {
		return strings.ToLower(language)
	}
}

// Turns a locale name (en_US) into a language name (en-us).
func toLanguage(locale string) string {
	if p := strings.Index(locale, "_"); p >= 0 {
		return strings.ToLower(locale[:p]) + "-" + strings.ToLower(locale[p+1:])
	} else {
		return strings.ToLower(locale)
	}
}
