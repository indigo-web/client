package method

type Method = string

const (
	Unknown Method = ""
	GET     Method = "GET"
	HEAD    Method = "HEAD"
	POST    Method = "POST"
	PUT     Method = "PUT"
	DELETE  Method = "DELETE"
	CONNECT Method = "CONNECT"
	OPTIONS Method = "OPTIONS"
	TRACE   Method = "TRACE"
	PATCH   Method = "PATCH"
)
