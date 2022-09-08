package graph

import (
	"context"
	"strings"

	"github.com/arangodb/go-driver"
	nograph "github.com/slntopp/nocloud/pkg/graph"
)

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
