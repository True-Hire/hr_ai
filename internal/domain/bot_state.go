package domain

type BotState struct {
	State string            `json:"state"`
	Data  map[string]string `json:"data,omitempty"`
}

const (
	BotStateChoosingLanguage = "choosing_language"
	BotStateChoosingRole     = "choosing_role"
	BotStateSharingPhone     = "sharing_phone"
	BotStateCollectingResume = "collecting_resume"

	HRBotStateSharingPhone          = "hr_sharing_phone"
	HRBotStateCollectingCompanyData = "hr_collecting_company_data"
	HRBotStateCompanyReview         = "hr_company_review"
	HRBotStateCollectingCompanyLogo = "hr_collecting_company_logo"
	HRBotStatePostingVacancy        = "hr_posting_vacancy"
	HRBotStateVacancyReview         = "hr_vacancy_review"
	HRBotStateAddingVacancyInfo     = "hr_adding_vacancy_info"
	HRBotStateEditingPublishedVacancy = "hr_editing_published_vacancy"
)
