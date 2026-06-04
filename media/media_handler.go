package media

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type MediaRepository interface {
	UploadMedia(ctx context.Context, mediaData []byte, fileName, mediaType string) (string, error)
}

type MediaHandler struct {
}

func NewMediaHandler() *MediaHandler {
	return &MediaHandler{}
}

func (mh *MediaHandler) UploadMedia(ctx context.Context, mediaData []byte, fileName, mediaType string) (string, error) {
	filePath := fmt.Sprintf("./uploads/message/%s_%s", uuid.NewString(), fileName)
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = outFile.Write(mediaData)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
