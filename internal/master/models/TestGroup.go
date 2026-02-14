package models

type TestGroup struct {
	ID              string           `bson:"_id" json:"id"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	TestCollections []TestCollection `json:"testCollections"`
}
