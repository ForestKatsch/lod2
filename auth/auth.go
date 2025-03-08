package auth

func Init() {
	initTokens()

	migrate()
}
