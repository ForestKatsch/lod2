package auth

type Group int

const (
	GroupUserEdit Group = iota
	GroupUserView
	GroupMediaEdit
	GroupMediaView
)
