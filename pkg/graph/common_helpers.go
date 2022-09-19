package graph

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/arangodb/go-driver"
	nograph "github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	noschema "github.com/slntopp/nocloud/pkg/nocloud/schema"
)

func StructToMap(obj interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(obj)

	if err != nil {
		return nil, err
	}

	newMap := make(map[string]interface{})

	err = json.Unmarshal(data, &newMap)
	return newMap, err
}

func GetByDocumentId(ctx context.Context, db driver.Database, id string) (*interface{}, error) {
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	var document interface{}

	// Want uuid in format collection/uuid
	tokens := strings.Split(id, "/")
	if len(tokens) != 2 {
		return nil, errors.New("id contains extra tokens")
	}
	collectionTitle := tokens[0]
	uuid := tokens[1]

	exists, err := db.CollectionExists(ctx, collectionTitle)
	if err != nil || !exists {
		return nil, errors.New("collection doesn't exist")
	}

	err = nograph.GetWithAccess(ctx, db,
		driver.NewDocumentID(noschema.ACCOUNTS_COL, requestor),
		driver.NewDocumentID(collectionTitle, uuid), &document)
	if err != nil {
		return nil, err
	}

	return &document, nil
}

const edgeQuery = `
FOR edge IN @@collection
    FILTER edge._from == @fromDocID && edge._to == @toDocID
    LIMIT 1
    RETURN edge
`

func HasAccess(ctx context.Context, db driver.Database, collection, fromKey, toKey string, level int32) bool {
	collections := strings.Split(collection, "2")
	if len(collections) != 2 {
		return false
	}

	fromDocID := driver.NewDocumentID(collections[0], fromKey)
	toDocID := driver.NewDocumentID(collections[1], toKey)

	c, err := db.Query(ctx, edgeQuery, map[string]interface{}{
		"@collection": collection,
		"fromDocID":   fromDocID,
		"toDocID":     toDocID,
	})
	if err != nil {
		return false
	}
	defer c.Close()

	access := &nograph.Access{}
	c.ReadDocument(ctx, access)

	return access.Level >= level
}
