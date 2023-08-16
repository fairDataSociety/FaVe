package handlers

import (
	"fmt"
	"github.com/fairDataSociety/FaVe/pkg/document"
	"github.com/fairDataSociety/FaVe/restapi/operations"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/go-openapi/runtime/middleware"
)

func (s *Handler) FaveCreateCollectionHandler(request operations.FaveCreateCollectionParams) middleware.Responder {
	collectionRaw := request.Body
	if collectionRaw.Name == "" {
		return operations.NewFaveCreateCollectionBadRequest().WithPayload(createErrorResponseObject("Collection name cannot be blank"))
	}
	if collectionRaw.Indexes == nil {
		return operations.NewFaveCreateCollectionBadRequest().WithPayload(createErrorResponseObject("Collection should have at least one index"))
	}
	indexesRaw, ok := collectionRaw.Indexes.(map[string]interface{})
	if !ok {
		return operations.NewFaveCreateCollectionBadRequest().WithPayload(createErrorResponseObject("Wrong indexes format"))
	}
	indexes := make(map[string]collection.IndexType)
	for k, v := range indexesRaw {
		switch v {
		case "string":
			indexes[k] = collection.StringIndex
		case "number":
			indexes[k] = collection.NumberIndex
		case "map":
			indexes[k] = collection.MapIndex
		case "list":
			indexes[k] = collection.ListIndex
		case "bytes":
		default:
			return operations.NewFaveCreateCollectionBadRequest().WithPayload(createErrorResponseObject("Wrong indexes type"))
		}
	}
	fmt.Println(indexes)
	col := &document.Collection{
		Name:    collectionRaw.Name,
		Indexes: indexes,
	}
	err := s.doc.CreateCollection(col)
	if err != nil {
		return operations.NewFaveCreateCollectionBadRequest().WithPayload(createErrorResponseObject("Failed to create collection"))
	}
	// TODO: return success string
	return operations.NewFaveCreateCollectionOK().WithPayload(createOKResponseObject("Collection created"))
}

func (s *Handler) FaveDeleteCollectionHandler(request operations.FaveDeleteCollectionParams) middleware.Responder {
	err := s.doc.DeleteCollection(request.Collection)
	if err != nil {
		return operations.NewFaveDeleteCollectionBadRequest().WithPayload(createErrorResponseObject("Failed to delete collection"))
	}
	return operations.NewFaveDeleteCollectionOK().WithPayload(createOKResponseObject("Collection deleted"))
}
