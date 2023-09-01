package handlers

import (
	"github.com/fairDataSociety/FaVe/models"
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
	err := s.doc.DeleteCollection(request.Collection)
	if err != nil {
		return operations.NewFaveDeleteCollectionBadRequest().WithPayload(createErrorResponseObject("Failed to delete collection"))
	}
	return operations.NewFaveDeleteCollectionOK().WithPayload(createOKResponseObject("Collection deleted"))
}

func (s *Handler) FaveGetCollectionsHandler(operations.FaveGetCollectionsParams) middleware.Responder {
	collections, err := s.doc.GetCollections()
	if err != nil {
		return operations.NewFaveGetCollectionsBadRequest().WithPayload(createErrorResponseObject("Failed to get collections"))
	}
	collectionResp := make([]*models.Collection, len(collections))
	for i, v := range collections {
		collectionResp[i] = &models.Collection{
			Name:    v.Name,
			Indexes: []*models.Index{},
		}
		for j, k := range v.Indexes {
			switch k {
			case collection.StringIndex:
				collectionResp[i].Indexes = append(collectionResp[i].Indexes, &models.Index{
					FieldName: j,
					FieldType: "string",
				})
			case collection.NumberIndex:
				collectionResp[i].Indexes = append(collectionResp[i].Indexes, &models.Index{
					FieldName: j,
					FieldType: "number",
				})
			case collection.MapIndex:
				collectionResp[i].Indexes = append(collectionResp[i].Indexes, &models.Index{
					FieldName: j,
					FieldType: "map",
				})
			case collection.ListIndex:
				collectionResp[i].Indexes = append(collectionResp[i].Indexes, &models.Index{
					FieldName: j,
					FieldType: "list",
				})
			}
		}
	}
	return operations.NewFaveGetCollectionsOK().WithPayload(collectionResp)
}
