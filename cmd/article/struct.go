package article

type Article struct {
	Id    int32  `json:"id"`
	Code  string `json:"code"`
	Title string `json:"title"`
}

type getResponse struct {
	Code    string `json:"code"`
	Data    Data   `json:"data"`
	Success bool   `json:"success"`
}

type Data struct {
	Articles []Article `json:"articles"`
	Total    int64     `json:"total"`
}
