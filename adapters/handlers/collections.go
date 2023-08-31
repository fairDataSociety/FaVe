package handlers

import (
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
	collectionRaw.Name = prefix + collectionRaw.Name
	indexes := make(map[string]collection.IndexType)
	for _, v := range collectionRaw.Indexes {
		switch v.FieldType {
		case "string":
			indexes[v.FieldName] = collection.StringIndex
		case "number":
			indexes[v.FieldName] = collection.NumberIndex
		case "map":
			indexes[v.FieldName] = collection.MapIndex
		case "list":
			indexes[v.FieldName] = collection.ListIndex
		case "bytes":
		default:
			return operations.NewFaveCreateCollectionBadRequest().WithPayload(createErrorResponseObject("Wrong index type"))
		}
	}

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
	request.Collection = prefix + request.Collection
	err := s.doc.DeleteCollection(request.Collection)
	if err != nil {
		return operations.NewFaveDeleteCollectionBadRequest().WithPayload(createErrorResponseObject("Failed to delete collection"))
	}
	return operations.NewFaveDeleteCollectionOK().WithPayload(createOKResponseObject("Collection deleted"))
}
