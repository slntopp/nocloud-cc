package graph

import (
	"context"

	"github.com/slntopp/nocloud/pkg/nocloud/access"
	"google.golang.org/protobuf/types/known/structpb"
)

func (c *ChatsMessagesController) FetchEntities(ctx context.Context, entities []string) (map[string]*structpb.Value, error) {
	result := make(map[string]*structpb.Value)
	for _, entity := range entities {
		document, err := GetByDocumentId(ctx, c.db, entity)
		if err != nil {
			continue
		}

		m, err := StructToMap(document)
		if err != nil {
			continue
		}
		if val, ok := m["access_level"]; !ok || val.(float64) < float64(access.READ) {
			continue
		}

		str, err := structpb.NewStruct(m)
		if err != nil {
			continue
		}

		result[entity] = structpb.NewStructValue(str)
	}
	return result, nil
}
