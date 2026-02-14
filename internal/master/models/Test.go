package models

type Test struct {
	ID           string   `bson:"_id" json:"id"`
	Name         string   `json:"name"`
	Method       string   `json:"method"`
	Url          string   `json:"url"`
	Body         string   `json:"body"`
	Headers      []Header `json:"headers"`
	ResponseCode int32    `json:"responseCode"`
	ResponseBody string   `json:"responseBody"`
}

type Header struct {
	ID    string `bson:"_id" json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
