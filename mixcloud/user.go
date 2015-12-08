package mixcloud

type User struct {
	FollowingCount        int                    `json:"following_count,omitempty"`
	IsPremium             bool                   `json:"is_premium,omitempty"`
	Follower              bool                   `json:"follow,omitempty"`
	created_time          string                 `json:"created_time,omitempty"`
	favorite_count        int                    `json:"favorite_count,omitempty"`
	city                  string                 `json:"city,omitempty"`
	biog                  string                 `json:"biog,omitempty"`
	pictures              map[string]interface{} `json:"pictures,omitempty"`
	is_current_user       bool                   `json:"is_current_user,omitempty"`
	follower_count        int                    `json:"follower_count,omitempty"`
	Username              string                 `json:"username,omitempty"`
	listen_count          int                    `json:"listen_count,omitempty"`
	cover_pictures        map[string]interface{} `json:"cover_pictures,omitempty"`
	key                   string                 `json:"key,omitempty"`
	cloudcast_count       int                    `json:"cloudcast_count,omitempty"`
	name                  string                 `json:"name,omitempty"`
	url                   string                 `json:"url,omitempty"`
	country               string                 `json:"country,omitempty"`
	updated_time          string                 `json:"updated_time,omitempty"`
	picture_primary_color string                 `json:"picture_primary_color,omitempty"`
	following             bool                   `json:"follow,omitempty"`
	IsPro                 bool                   `json:"is_pro,omitempty"`
}
