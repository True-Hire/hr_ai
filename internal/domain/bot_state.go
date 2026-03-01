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

	HRBotStateSharingPhone   = "hr_sharing_phone"
	HRBotStatePostingVacancy = "hr_posting_vacancy"
)
