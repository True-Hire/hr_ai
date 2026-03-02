package application

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type AccountDeletionService struct {
	userRepo       domain.UserRepository
	hrRepo         domain.CompanyHRRepository
	sessionRepo    domain.SessionRepository
	hrSessionRepo  domain.HRSessionRepository
	vacancyRepo    domain.VacancyRepository
	vacancyAppRepo domain.VacancyApplicationRepository
	pfRepo         domain.ProfileFieldRepository
	pftRepo        domain.ProfileFieldTextRepository
	expRepo        domain.ExperienceItemRepository
	eduRepo        domain.EducationItemRepository
	itemTextRepo   domain.ItemTextRepository
	skillRepo      domain.SkillRepository
	vectorIndexSvc *VectorIndexService
}

func NewAccountDeletionService(
	userRepo domain.UserRepository,
	hrRepo domain.CompanyHRRepository,
	sessionRepo domain.SessionRepository,
	hrSessionRepo domain.HRSessionRepository,
	vacancyRepo domain.VacancyRepository,
	vacancyAppRepo domain.VacancyApplicationRepository,
	pfRepo domain.ProfileFieldRepository,
	pftRepo domain.ProfileFieldTextRepository,
	expRepo domain.ExperienceItemRepository,
	eduRepo domain.EducationItemRepository,
	itemTextRepo domain.ItemTextRepository,
	skillRepo domain.SkillRepository,
	vectorIndexSvc *VectorIndexService,
) *AccountDeletionService {
	return &AccountDeletionService{
		userRepo:       userRepo,
		hrRepo:         hrRepo,
		sessionRepo:    sessionRepo,
		hrSessionRepo:  hrSessionRepo,
		vacancyRepo:    vacancyRepo,
		vacancyAppRepo: vacancyAppRepo,
		pfRepo:         pfRepo,
		pftRepo:        pftRepo,
		expRepo:        expRepo,
		eduRepo:        eduRepo,
		itemTextRepo:   itemTextRepo,
		skillRepo:      skillRepo,
		vectorIndexSvc: vectorIndexSvc,
	}
}

// DeleteUserByPhone finds a user by phone and deletes everything associated with them.
func (s *AccountDeletionService) DeleteUserByPhone(ctx context.Context, phone string) error {
	phone = normalizePhone(phone)
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		// try with opposite format
		user, err = s.userRepo.GetByPhone(ctx, flipPhonePrefix(phone))
		if err != nil {
			return fmt.Errorf("find user by phone: %w", err)
		}
	}
	return s.deleteUser(ctx, user.ID)
}

// DeleteHRByPhone finds an HR by phone, nullifies their vacancy references, and deletes the HR.
func (s *AccountDeletionService) DeleteHRByPhone(ctx context.Context, phone string) error {
	phone = normalizePhone(phone)
	hr, err := s.hrRepo.GetByPhone(ctx, phone)
	if err != nil {
		// try with opposite format
		hr, err = s.hrRepo.GetByPhone(ctx, flipPhonePrefix(phone))
		if err != nil {
			return fmt.Errorf("find hr by phone: %w", err)
		}
	}
	return s.deleteHR(ctx, hr.ID)
}

func normalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}

func flipPhonePrefix(phone string) string {
	if strings.HasPrefix(phone, "+") {
		return strings.TrimPrefix(phone, "+")
	}
	return "+" + phone
}

func (s *AccountDeletionService) deleteUser(ctx context.Context, userID uuid.UUID) error {
	// 1. Soft-delete user sessions
	if err := s.sessionRepo.SoftDeleteByUser(ctx, userID); err != nil {
		log.Printf("delete user %s: soft delete sessions: %v", userID, err)
	}

	// 2. Delete vacancy applications
	if err := s.vacancyAppRepo.DeleteByUser(ctx, userID); err != nil {
		log.Printf("delete user %s: delete vacancy applications: %v", userID, err)
	}

	// 3. Delete item_texts for experience items, then experience items
	expItems, _ := s.expRepo.ListByUser(ctx, userID)
	for _, item := range expItems {
		if err := s.itemTextRepo.DeleteByItemID(ctx, item.ID); err != nil {
			log.Printf("delete user %s: delete experience item texts %s: %v", userID, item.ID, err)
		}
	}
	if err := s.expRepo.DeleteByUser(ctx, userID); err != nil {
		log.Printf("delete user %s: delete experience items: %v", userID, err)
	}

	// 4. Delete item_texts for education items, then education items
	eduItems, _ := s.eduRepo.ListByUser(ctx, userID)
	for _, item := range eduItems {
		if err := s.itemTextRepo.DeleteByItemID(ctx, item.ID); err != nil {
			log.Printf("delete user %s: delete education item texts %s: %v", userID, item.ID, err)
		}
	}
	if err := s.eduRepo.DeleteByUser(ctx, userID); err != nil {
		log.Printf("delete user %s: delete education items: %v", userID, err)
	}

	// 5. Delete profile_field_texts for each profile field, then profile fields
	fields, _ := s.pfRepo.ListByUser(ctx, userID)
	for _, f := range fields {
		if err := s.pftRepo.DeleteByField(ctx, f.ID); err != nil {
			log.Printf("delete user %s: delete profile field texts %s: %v", userID, f.ID, err)
		}
	}
	if err := s.pfRepo.DeleteByUser(ctx, userID); err != nil {
		log.Printf("delete user %s: delete profile fields: %v", userID, err)
	}

	// 6. Remove user skills
	if err := s.skillRepo.RemoveUserSkills(ctx, userID); err != nil {
		log.Printf("delete user %s: remove user skills: %v", userID, err)
	}

	// 7. Delete Qdrant vectors
	if s.vectorIndexSvc != nil {
		if err := s.vectorIndexSvc.DeleteUser(ctx, userID); err != nil {
			log.Printf("delete user %s: delete qdrant vectors: %v", userID, err)
		}
	}

	// 8. Delete the user row
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete user row: %w", err)
	}

	log.Printf("user %s deleted successfully", userID)
	return nil
}

func (s *AccountDeletionService) deleteHR(ctx context.Context, hrID uuid.UUID) error {
	// 1. Soft-delete HR sessions
	if err := s.hrSessionRepo.SoftDeleteByHR(ctx, hrID); err != nil {
		log.Printf("delete hr %s: soft delete sessions: %v", hrID, err)
	}

	// 2. For each vacancy owned by this HR, clean up applications in Qdrant
	vacancyIDs, _ := s.vacancyRepo.ListIDsByHR(ctx, hrID)
	for _, vid := range vacancyIDs {
		if err := s.vacancyAppRepo.DeleteByVacancy(ctx, vid); err != nil {
			log.Printf("delete hr %s: delete vacancy applications for %s: %v", hrID, vid, err)
		}
		if s.vectorIndexSvc != nil {
			if err := s.vectorIndexSvc.DeleteVacancy(ctx, vid); err != nil {
				log.Printf("delete hr %s: delete vacancy vector %s: %v", hrID, vid, err)
			}
		}
	}

	// 3. Nullify hr_id on all vacancies owned by this HR (vacancies remain)
	if err := s.vacancyRepo.NullifyHRID(ctx, hrID); err != nil {
		log.Printf("delete hr %s: nullify vacancy hr_id: %v", hrID, err)
	}

	// 4. Delete the HR row
	if err := s.hrRepo.Delete(ctx, hrID); err != nil {
		return fmt.Errorf("delete hr row: %w", err)
	}

	log.Printf("hr %s deleted successfully", hrID)
	return nil
}
