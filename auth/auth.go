package auth

func Init() {
	initTokens()
	migrateUsers()
}
