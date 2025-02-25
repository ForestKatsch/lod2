package auth

func Init() {
	initTokens()

	migrateInvites()
	migrateUsers()
	migrateSessions()
}
