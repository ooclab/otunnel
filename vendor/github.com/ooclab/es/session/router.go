package session

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrNoHandler = errors.New("no such handler")
)

// RequestHandler define request-response handler func
type RequestHandler interface {
	Handle(*EMSG) *EMSG
}

type RequestHandlerFunc func(*Request) (*Response, error)

type Route struct {
	Action  string
	Handler RequestHandlerFunc
}

type Router struct {
	routes  []Route
	reMatch map[*regexp.Regexp]Route
}

func NewRouter() *Router {
	router := &Router{
		reMatch: make(map[*regexp.Regexp]Route),
	}
	return router
}

func (r *Router) AddRoute(route Route) {
	// Fix me
	pattern := route.Action
	if !strings.HasPrefix(pattern, "^") {
		pattern = "^" + pattern
	}
	if !strings.HasSuffix(pattern, "$") {
		pattern = pattern + "$"
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return
	}
	r.reMatch[re] = route
	r.routes = append(r.routes, route)
}

func (r *Router) AddRoutes(routes []Route) {
	for _, v := range routes {
		r.AddRoute(v)
	}
}

func (r *Router) Dispatch(req *Request) (resp *Response, err error) {

	for re, v := range r.reMatch {
		matchs := re.FindStringSubmatch(req.Action)
		if len(matchs) > 0 {
			return v.Handler(req)
		}
	}

	return nil, ErrNoHandler
}
