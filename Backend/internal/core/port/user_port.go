package port

type UserRepository interface {
	CreateUser(email string, password string, firstname string, lastname string) (createUserStatus bool, err error)
	AuthenticateUser(email string, password string) (authStatus bool, uid string, err error)
}