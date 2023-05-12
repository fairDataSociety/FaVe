package handlers

import (
	"github.com/fairDataSociety/FaVe/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
)

type Handlers struct{}

func (s *Handlers) FaveRootHandler(_ operations.FaveRootParams) middleware.Responder {
	return operations.NewFaveRootOK()
}
