package entities

type Token struct {
	Token string
}

type StoredUser struct {
	Id         string
	Username   string
	StoredPass string
	AuthToken  Token
}

type User struct {
	Id       string
	Username string
}
