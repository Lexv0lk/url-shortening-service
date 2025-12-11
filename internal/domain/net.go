package domain

const (
	BaseUrl = `http://localhost:8080`
)

const (
	AddUrlAddress   = "POST /add-url"
	UrlTokenStr     = "urlToken"
	RedirectAddress = "/{" + UrlTokenStr + "}"
)
