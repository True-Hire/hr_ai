package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/repository"
	"gopkg.in/telebot.v3"
)

type VacancyAIHandlers struct {
	bot        *telebot.Bot
	aiService  *application.VacancyAIService
	hrBotSvc   *application.HRBotService
	vacancyRepo *repository.VacancyAIRepository
}

func NewVacancyAIHandlers(bot *telebot.Bot, aiSvc *application.VacancyAIService, hrBotSvc *application.HRBotService, repo *repository.VacancyAIRepository) *VacancyAIHandlers {
	return &VacancyAIHandlers{
		bot:         bot,
		aiService:   aiSvc,
		hrBotSvc:    hrBotSvc,
		vacancyRepo: repo,
	}
}

func (h *VacancyAIHandlers) Register() {
	// Menyu tugmasi yoki bironta komanda bosilganda (Buni main menyuga ulashingiz mumkin)
	// h.bot.Handle(&btnCreateVacancyAI, h.handleStartAI)
	
	h.bot.Handle(&telebot.Btn{Unique: "hr_ai_vac_confirm"}, h.handleConfirm)
	h.bot.Handle(&telebot.Btn{Unique: "hr_ai_vac_cancel"}, h.handleCancel)
}

func (h *VacancyAIHandlers) HandleAIInput(ctx context.Context, telegramID int64, vacancyID uuid.UUID, text string) error {
	log.Printf("[VacancyAIHandlers] Starting background AI taxonomy extraction for vacancy %s", vacancyID)
	// 1. AI ga so'rov yuborish
	parsed, err := h.aiService.ParseVacancyText(ctx, text)
	if err != nil {
		log.Printf("[VacancyAIHandlers] Silent AI Parse error for %s: %v", vacancyID, err)
		return nil
	}
	
	// 2. Mavjud vakansiyani yuklash
	vacancy, err := h.vacancyRepo.GetByID(ctx, vacancyID)
	if err != nil {
		log.Printf("[VacancyAIHandlers] Silent AI: vacancy not found %s: %v", vacancyID, err)
		return nil
	}

	// 3. Uni tahlil natijalari bilan yangilash (Taxonomy update)
	if err := h.vacancyRepo.SaveParsedVacancy(ctx, vacancy, *parsed); err != nil {
		log.Printf("[VacancyAIHandlers] Silent SaveParsedVacancy error for %s: %v", vacancyID, err)
		return nil
	}
	
	log.Printf("[VacancyAIHandlers] Successfully completed background taxonomy extraction for vacancy %s", vacancyID)
	return nil
}

func (h *VacancyAIHandlers) handleConfirm(c telebot.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	// 1. Redis dan draftni o'qib olamiz
	parsed, err := h.hrBotSvc.GetAIVacancyDraft(ctx, telegramID)
	if err != nil || parsed == nil {
		return c.Send("❌ Xatolik: Vakansiya ma'lumotlari topilmadi yoki muddati o'tgan.")
	}

	// 2. HR ma'lumotlarini olamiz
	hr, err := h.hrBotSvc.GetHRByTelegramID(ctx, fmt.Sprintf("%d", telegramID))
	if err != nil {
		return c.Send("❌ HR ma'lumotlari topilmadi.")
	}

	vacancy := &domain.Vacancy{
		ID:             uuid.New(),
		HRID:           hr.ID,
		Format:         "hybrid", // Default
		SalaryCurrency: "USD",
		Status:         domain.VacancyStatusActive,
		CompanyData:    hr.CompanyData,
	}

	// 4. Bazaga saqlaymiz (Zanjirband bog'liqliklar bilan)
	if err := h.vacancyRepo.SaveParsedVacancy(ctx, vacancy, *parsed); err != nil {
		log.Printf("SaveParsedVacancy error: %v", err)
		return c.Send("❌ Bazaga saqlashda xatolik yuz berdi.")
	}

	// 5. Tozalash
	h.hrBotSvc.ClearAIVacancyDraft(ctx, telegramID)

	c.Respond(&telebot.CallbackResponse{Text: "Muvaffaqiyatli saqlandi!"})
	c.Delete()
	return c.Send("✅ Vakansiya muvaffaqiyatli saqlandi! \nMain Category, Sub Category, Texnologiyalar va Ko'nikmalar avtomatik bog'landi.")
}

func (h *VacancyAIHandlers) handleCancel(c telebot.Context) error {
	c.Respond(&telebot.CallbackResponse{Text: "Bekor qilindi"})
	c.Delete()
	return c.Send("❌ Vakansiya bekor qilindi.")
}

func (h *VacancyAIHandlers) formatPreview(parsed *domain.AIParsedVacancy) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>Sarlavha:</b> %s\n", parsed.Title))
	b.WriteString(fmt.Sprintf("<b>Tavsif:</b> %s\n\n", parsed.Description))
	
	b.WriteString("<b>🔗 Bazaga bog'langan ID lar:</b>\n")
	if parsed.MatchedMainCatID != "" {
		b.WriteString(fmt.Sprintf("- Main Cat ID: %s\n", parsed.MatchedMainCatID))
	} else if parsed.NewMainCategory != "" {
		b.WriteString(fmt.Sprintf("- Main Cat (YANGI): %s\n", parsed.NewMainCategory))
	} else {
		b.WriteString("- Main Cat: ❌ Topilmadi\n")
	}

	if parsed.MatchedSubCatID != "" {
		b.WriteString(fmt.Sprintf("- Sub Cat ID: %s\n", parsed.MatchedSubCatID))
	} else if parsed.NewSubCategory != "" {
		b.WriteString(fmt.Sprintf("- Sub Cat (YANGI): %s\n", parsed.NewSubCategory))
	} else {
		b.WriteString("- Sub Cat: ❌ Topilmadi\n")
	}

	b.WriteString(fmt.Sprintf("- Tech IDs: %d ta\n", len(parsed.MatchedTechIDs)))
	b.WriteString(fmt.Sprintf("- Skill IDs: %d ta\n\n", len(parsed.MatchedSkillIDs)))

	if len(parsed.NewTechnologies) > 0 {
		b.WriteString("<b>🆕 BAZADA YO'Q (Yangi Yaratiladigan) Texnologiyalar:</b>\n")
		for _, t := range parsed.NewTechnologies {
			b.WriteString(fmt.Sprintf("- %s\n", t))
		}
		b.WriteString("\n")
	}

	if len(parsed.NewSkills) > 0 {
		b.WriteString("<b>🆕 BAZADA YO'Q (Yangi Yaratiladigan) Ko'nikmalar:</b>\n")
		for _, s := range parsed.NewSkills {
			b.WriteString(fmt.Sprintf("- %s\n", s))
		}
	}

	b.WriteString("\n<i>Tasdiqlaysizmi?</i>")
	return b.String()
}
