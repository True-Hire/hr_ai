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

type UploadResult struct {
	ObjectName string
	URL        string
}

func (s *StorageService) Upload(ctx context.Context, folder string, data []byte, contentType string) (*UploadResult, error) {
	objectName, err := s.minio.Upload(ctx, folder, data, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}
	return &UploadResult{
		ObjectName: objectName,
		URL:        s.minio.PublicURL(objectName),
	}, nil
}

func (s *StorageService) UploadProfilePhoto(ctx context.Context, data []byte, contentType string) (*UploadResult, error) {
	return s.Upload(ctx, "profile-photos", data, contentType)
}

func (s *StorageService) UploadResume(ctx context.Context, data []byte, contentType string) (*UploadResult, error) {
	return s.Upload(ctx, "resumes", data, contentType)
}

func (s *StorageService) UploadCompanyLogo(ctx context.Context, data []byte, contentType string) (*UploadResult, error) {
	return s.Upload(ctx, "company-logos", data, contentType)
}

func (s *StorageService) Get(ctx context.Context, objectName string) (*minioclient.FileInfo, error) {
	info, err := s.minio.Get(ctx, objectName)
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}
	return info, nil
}

func (s *StorageService) Delete(ctx context.Context, objectName string) error {
	return s.minio.Delete(ctx, objectName)
}

func (s *StorageService) PublicURL(objectName string) string {
	return s.minio.PublicURL(objectName)
}
