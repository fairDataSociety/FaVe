// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// FaveHelloHandlerFunc turns a function with the right signature into a fave hello handler
type FaveHelloHandlerFunc func(FaveHelloParams) middleware.Responder

// Handle executing the request and returning a response
func (fn FaveHelloHandlerFunc) Handle(params FaveHelloParams) middleware.Responder {
	return fn(params)
}

// FaveHelloHandler interface for that can handle valid fave hello params
type FaveHelloHandler interface {
	Handle(FaveHelloParams) middleware.Responder
}

// NewFaveHello creates a new http.Handler for the fave hello operation
func NewFaveHello(ctx *middleware.Context, handler FaveHelloHandler) *FaveHello {
	return &FaveHello{Context: ctx, Handler: handler}
}

/*
	FaveHello swagger:route GET /hello faveHello

hello. Discover the REST API
*/
type FaveHello struct {
	Context *middleware.Context
	Handler FaveHelloHandler
}

func (o *FaveHello) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewFaveHelloParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
