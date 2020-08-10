package model

type LoginType int

const (
	WebLogin = LoginType(iota)
	GoogleLogin
	FacebookLogin
	TwitterLogin
	GithubLogin
	GitlabLogin
)

type User struct {
	//This is a unique field.
	ID         string                 `json:"id"`
	Meta       map[string]interface{} `json:"extra" bson:"extra"`
	SignInType LoginType              `json:"-" bson:"type"`
}
