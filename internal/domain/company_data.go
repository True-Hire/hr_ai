package domain

type CompanyData struct {
	EmployeeCount   int32             `json:"employee_count,omitempty"`
	Country         string            `json:"country,omitempty"`
	Address         string            `json:"address,omitempty"`
	Phone           string            `json:"phone,omitempty"`
	Telegram        string            `json:"telegram,omitempty"`
	TelegramChannel string            `json:"telegram_channel,omitempty"`
	Email           string            `json:"email,omitempty"`
	LogoURL         string            `json:"logo_url,omitempty"`
	WebSite         string            `json:"web_site,omitempty"`
	Instagram       string            `json:"instagram,omitempty"`
	SourceLang      string            `json:"source_lang,omitempty"`
	Texts           []CompanyDataText `json:"texts,omitempty"`
}

type CompanyDataText struct {
	Lang         string `json:"lang"`
	Name         string `json:"name"`
	ActivityType string `json:"activity_type,omitempty"`
	CompanyType  string `json:"company_type,omitempty"`
	About        string `json:"about,omitempty"`
	Market       string `json:"market,omitempty"`
	IsSource     bool   `json:"is_source"`
}
