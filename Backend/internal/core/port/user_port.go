package port

type UserRepository interface {
	CreateUser(email string, password string, firstname string, lastname string) (createStatus bool, err error)
	RemoveUser(uid string) (deleteStatus bool, err error)
	AuthenticateUser(email string, password string) (authStatus bool, uid string, err error)
	OAuthAuthenticateUser(email string, provider string, firstName string, lastName string) (authStatus bool, uid string, err error)
	UpdateUserInfo(uid string, firstname string, lastname string) (updateStatus bool, err error)
}