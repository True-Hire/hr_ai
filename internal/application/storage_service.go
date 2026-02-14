package application

import (
	"context"
	"fmt"

	minioclient "github.com/ruziba3vich/hr-ai/internal/infrastructure/minio"
)

type StorageService struct {
	minio *minioclient.Client
}

func NewStorageService(minio *minioclient.Client) *StorageService {
	return &StorageService{minio: minio}
}

func (s *StorageService) UploadProfilePhoto(ctx context.Context, data []byte, contentType string) (string, error) {
	url, err := s.minio.Upload(ctx, "profile-photos", data, contentType)
	if err != nil {
		return "", fmt.Errorf("upload profile photo: %w", err)
	}
	return url, nil
}

func (s *StorageService) UploadResume(ctx context.Context, data []byte, contentType string) (string, error) {
	url, err := s.minio.Upload(ctx, "resumes", data, contentType)
	if err != nil {
		return "", fmt.Errorf("upload resume: %w", err)
	}
	return url, nil
}

func (s *StorageService) UploadCompanyLogo(ctx context.Context, data []byte, contentType string) (string, error) {
	url, err := s.minio.Upload(ctx, "company-logos", data, contentType)
	if err != nil {
		return "", fmt.Errorf("upload company logo: %w", err)
	}
	return url, nil
}

func (s *StorageService) Delete(ctx context.Context, objectName string) error {
	return s.minio.Delete(ctx, objectName)
}
