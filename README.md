# client
Simple yet convenient (at least trying to be) HTTP client

# Purposes
Writing it, there were primarly two reasons:
- Testing indigo web-framework by itself
  - Not net/http, because it doesn't provide enough flexibility for enabling/disabling specific protocol mechanics, when testing particular parts of the framework
- Make requests a bit easier, than raw net/http client

# Philosophy
Client is considered to be alike the indigo framework by its structure, programming approach, etc. In some parts it's even similar (e.g. `status` package, `Headers` object, etc.). Also it tends to be more simple yet powerful. Idea is to add a possibility 
for concurrent goroutines to access a single session instance (this'll be using pipelining), proxy, cookies, upgrading, etc.
