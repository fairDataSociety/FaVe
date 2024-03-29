// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// FaveGetDocumentsHandlerFunc turns a function with the right signature into a fave get documents handler
type FaveGetDocumentsHandlerFunc func(FaveGetDocumentsParams) middleware.Responder

// Handle executing the request and returning a response
func (fn FaveGetDocumentsHandlerFunc) Handle(params FaveGetDocumentsParams) middleware.Responder {
	return fn(params)
}

// FaveGetDocumentsHandler interface for that can handle valid fave get documents params
type FaveGetDocumentsHandler interface {
	Handle(FaveGetDocumentsParams) middleware.Responder
}

// NewFaveGetDocuments creates a new http.Handler for the fave get documents operation
func NewFaveGetDocuments(ctx *middleware.Context, handler FaveGetDocumentsHandler) *FaveGetDocuments {
	return &FaveGetDocuments{Context: ctx, Handler: handler}
}

/*
	FaveGetDocuments swagger:route GET /documents faveGetDocuments

Retrieve a document based on query parameters
*/
type FaveGetDocuments struct {
	Context *middleware.Context
	Handler FaveGetDocumentsHandler
}

func (o *FaveGetDocuments) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewFaveGetDocumentsParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
