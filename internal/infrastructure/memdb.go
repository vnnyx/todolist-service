package infrastructure

import (
	"github.com/hashicorp/go-memdb"
	"github.com/sirupsen/logrus"
)

func NewMemDB() *memdb.MemDB {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"todos": {
				Name: "todos",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "ID"},
					},
					"activity_group_id": {
						Name:    "activity_group_id",
						Unique:  false,
						Indexer: &memdb.IntFieldIndex{Field: "ActivityGroupID"},
					},
				},
			},
			"activities": {
				Name: "activities",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "ID"},
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		logrus.Fatal(err)
	}
	return db
}
