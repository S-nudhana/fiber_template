package port

type UserRepository interface {
	CreateUser(email string, password string, firstName string, lastName string) (createStatus bool, err error)
	RemoveUser(uid string) (deleteStatus bool, err error)
	AuthenticateUser(email string, password string) (authStatus bool, uid string, err error)
	OAuthAuthenticateUser(email string, provider string, firstName string, lastName string) (authStatus bool, uid string, err error)
	UpdateUserInfo(uid string, firstName string, lastName string) (updateStatus bool, err error)
}