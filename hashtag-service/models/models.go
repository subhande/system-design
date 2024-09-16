package models

type PostWithHashTags struct {
	ID       string   `json:"id"`
	UserID   string   `json:"user_id"`
	Hashtags []string `json:"hashtags"`
	URL      string   `json:"url"`
}

type PostWithHashTag struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Hashtag string `json:"hashtag"`
	URL     string `json:"url"`
}

type Post struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type HashtagCount struct {
	Hashtag  string `json:"hashtag"`
	Count    int    `json:"count"`
	TopPosts []Post `json:"top_posts"`
}
