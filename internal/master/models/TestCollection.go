package models

type TestCollection struct {
	ID          string `bson:"_id" json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tests       []Test `json:"tests"`
	GroupId     string `json:"groupId"`
}
