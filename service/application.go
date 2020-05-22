package app

import (
	"context"
	"mime/multipart"
)

//FileProcessor ...
type FileProcessor interface {
	CreateFile(context.Context, User, *multipart.FileHeader) error
	// Delete(ctx context.Context, name string) error
}
