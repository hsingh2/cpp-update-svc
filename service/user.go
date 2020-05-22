package app

import "context"

//User ...
type User struct {
	UserNTID string `json:"userNTID"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

//FileRequest ...
type FileRequest struct {
	FileName string
	User
}

//Notifier ...
type Notifier interface {
	NotifyUser(ctx context.Context, user User, message string) error
}
