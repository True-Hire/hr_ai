package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"

	tele "gopkg.in/telebot.v3"
)


// -- Localized messages for HR bot --

var hrMsgWelcomeNew = map[string]string{
	"en": "Hello 👋\n\nI'm HR AI for recruiting. I'll help you:\n\n• calculate market salary\n• quickly find matching candidates\n• automate hiring\n\nWhat would you like to do?",
	"ru": "Здравствуйте 👋\n\nЯ HR AI для подбора сотрудников. Помогу:\n\n• рассчитать рыночную зарплату\n• быстро найти подходящих кандидатов\n• автоматизировать найм\n\nЧто хотите сделать?",
	"uz": "Salom 👋\n\nMen xodimlarni tanlash uchun HR AI man. Yordam beraman:\n\n• bozor maoshini hisoblash\n• mos nomzodlarni tez topish\n• yollashni avtomatlashtirish\n\nNima qilmoqchisiz?",
}

var hrMsgWelcomeBack = map[string]string{
	"en": "👋 Welcome back, %s! What would you like to do?",
	"ru": "👋 С возвращением, %s! Что хотите сделать?",
	"uz": "👋 Qaytganingiz bilan, %s! Nima qilmoqchisiz?",
}

var hrMsgRegistered = map[string]string{
	"en": "✅ Registration complete, %s! Welcome!\n\nUse the menu below to get started 👇",
	"ru": "✅ Регистрация завершена, %s! Добро пожаловать!\n\nИспользуйте меню ниже 👇",
	"uz": "✅ Ro'yxatdan o'tish yakunlandi, %s! Xush kelibsiz!\n\nBoshlash uchun quyidagi menyudan foydalaning 👇",
}

var hrMsgSharePhone = map[string]string{
	"en": "📱 Please share your phone number to register.",
	"ru": "📱 Пожалуйста, поделитесь номером телефона для регистрации.",
	"uz": "📱 Ro'yxatdan o'tish uchun telefon raqamingizni ulashing.",
}

var hrMsgBtnSharePhone = map[string]string{
	"en": "📞 Share phone number",
	"ru": "📞 Поделиться номером",
	"uz": "📞 Telefon raqamini ulashish",
}

var hrMsgPhoneReminder = map[string]string{
	"en": "Please share your phone number using the button below 👇",
	"ru": "Пожалуйста, поделитесь номером телефона, нажав кнопку ниже 👇",
	"uz": "Iltimos, quyidagi tugma orqali telefon raqamingizni ulashing 👇",
}

var hrMenuBtnPostVacancy = map[string]string{
	"en": "📝 Post Vacancy",
	"ru": "📝 Разместить вакансию",
	"uz": "📝 Vakansiya joylashtirish",
}

var hrMenuBtnMyVacancies = map[string]string{
	"en": "📋 My Vacancies",
	"ru": "📋 Мои вакансии",
	"uz": "📋 Mening vakansiyalarim",
}

var hrMenuBtnFindCandidates = map[string]string{
	"en": "🔍 Find Candidates",
	"ru": "🔍 Найти кандидатов",
	"uz": "🔍 Nomzodlarni topish",
}

var hrMenuBtnChangeLang = map[string]string{
	"en": "🌍 Change Language",
	"ru": "🌍 Сменить язык",
	"uz": "🌍 Tilni o'zgartirish",
}

var hrMsgPostVacancy = map[string]string{
	"en": "📝 Send me the vacancy text.\n\nInclude:\n— Job title\n— Requirements\n— Responsibilities\n— Salary range\n— Work format\n\nI'll parse everything automatically.",
	"ru": "📝 Отправьте мне текст вакансии.\n\nУкажите:\n— Название должности\n— Требования\n— Обязанности\n— Зарплатная вилка\n— Формат работы\n\nЯ всё разберу автоматически.",
	"uz": "📝 Menga vakansiya matnini yuboring.\n\nQuyidagilarni ko'rsating:\n— Lavozim nomi\n— Talablar\n— Mas'uliyatlar\n— Maosh oralig'i\n— Ish formati\n\nHammasini avtomatik tahlil qilaman.",
}

var hrMsgParsingVacancy = map[string]string{
	"en": "Parsing vacancy... ⏳",
	"ru": "Разбираю вакансию… ⏳",
	"uz": "Vakansiyani tahlil qilmoqdaman… ⏳",
}

var hrMsgVacancyCreated = map[string]string{
	"en": "✅ Vacancy created!\n\n**%s**\n\nSalary: %s – %s %s\nFormat: %s | Schedule: %s",
	"ru": "✅ Вакансия создана!\n\n**%s**\n\nЗарплата: %s – %s %s\nФормат: %s | График: %s",
	"uz": "✅ Vakansiya yaratildi!\n\n**%s**\n\nMaosh: %s – %s %s\nFormat: %s | Jadval: %s",
}

var hrMsgVacancyFailed = map[string]string{
	"en": "❌ Failed to parse vacancy. Please try again.",
	"ru": "❌ Не удалось разобрать вакансию. Попробуйте ещё раз.",
	"uz": "❌ Vakansiyani tahlil qilib bo'lmadi. Qaytadan urinib ko'ring.",
}

var hrMsgNoVacancies = map[string]string{
	"en": "📋 You haven't posted any vacancies yet.",
	"ru": "📋 У вас пока нет размещённых вакансий.",
	"uz": "📋 Siz hali vakansiya joylashtirgansiz.",
}

var hrMsgFindCandidates = map[string]string{
	"en": "🔍 Enter your search query.\n\nExample: \"Senior Go developer with 3+ years experience\"",
	"ru": "🔍 Введите поисковый запрос.\n\nПример: «Senior Go разработчик с опытом от 3 лет»",
	"uz": "🔍 Qidiruv so'rovini kiriting.\n\nMisol: «3+ yillik tajribali Senior Go dasturchi»",
}

var hrMsgSearching = map[string]string{
	"en": "Searching candidates... 🔍",
	"ru": "Ищу кандидатов… 🔍",
	"uz": "Nomzodlarni qidirmoqdaman… 🔍",
}

var hrMsgNoCandidates = map[string]string{
	"en": "No matching candidates found. Try a different query.",
	"ru": "Подходящих кандидатов не найдено. Попробуйте другой запрос.",
	"uz": "Mos nomzodlar topilmadi. Boshqa so'rov bilan urinib ko'ring.",
}

var hrMsgChooseLang = map[string]string{
	"en": "🌍 Choose your new language:",
	"ru": "🌍 Выберите новый язык:",
	"uz": "🌍 Yangi tilni tanlang:",
}

var hrMsgLangChanged = map[string]string{
	"en": "✅ Language changed to English",
	"ru": "✅ Язык изменён на русский",
	"uz": "✅ Til o'zbekchaga o'zgartirildi",
}

var hrMsgUseMenu = map[string]string{
	"en": "🤔 I didn't quite understand that. Use the menu below 👇",
	"ru": "🤔 Не совсем понял. Используйте меню ниже 👇",
	"uz": "🤔 Tushunmadim. Quyidagi menyudan foydalaning 👇",
}

var hrMsgMenuUpdated = map[string]string{
	"en": "🔄 Bot updated! Check out the menu below 👇",
	"ru": "🔄 Бот обновлён! Используйте меню ниже 👇",
	"uz": "🔄 Bot yangilandi! Quyidagi menyudan foydalaning 👇",
}

var hrMsgError = map[string]string{
	"en": "Something went wrong. Please try again.",
	"ru": "Что-то пошло не так. Пожалуйста, попробуйте ещё раз.",
	"uz": "Xatolik yuz berdi. Iltimos, qaytadan urinib ko'ring.",
}

// -- HR Bot --

type HRBot struct {
	bot      *tele.Bot
	hrBotSvc *application.HRBotService
}

func NewHRBot(token string, hrBotSvc *application.HRBotService) (*HRBot, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("create hr telegram bot: %w", err)
	}

	hb := &HRBot{bot: b, hrBotSvc: hrBotSvc}
	hb.registerHandlers()
	return hb, nil
}

func (hb *HRBot) Start() {
	log.Println("hr telegram bot starting...")
	hb.bot.Start()
}

func (hb *HRBot) Stop() {
	log.Println("stopping hr telegram bot...")
	hb.bot.Stop()
	log.Println("hr telegram bot stopped")
}

func (hb *HRBot) BroadcastMenu() {
	ctx := context.Background()
	hrs, err := hb.hrBotSvc.ListAllHRs(ctx)
	if err != nil {
		log.Printf("hr broadcast menu: failed to list hrs: %v", err)
		return
	}

	log.Printf("hr broadcast menu: sending to %d HRs", len(hrs))
	sent := 0
	for _, hr := range hrs {
		if hr.TelegramID == "" {
			continue
		}
		tgID, err := strconv.ParseInt(hr.TelegramID, 10, 64)
		if err != nil {
			continue
		}
		lang := langOrDefault(hr.Language)
		_, err = hb.bot.Send(&tele.User{ID: tgID}, hrMsgMenuUpdated[lang], hrMenu(lang))
		if err != nil {
			log.Printf("hr broadcast menu: failed to send to %s: %v", hr.TelegramID, err)
			continue
		}
		sent++
		time.Sleep(50 * time.Millisecond)
	}
	log.Printf("hr broadcast menu: sent to %d/%d HRs", sent, len(hrs))
}

func (hb *HRBot) registerHandlers() {
	ctx := context.Background()
	bot := hb.bot
	hrBotSvc := hb.hrBotSvc

	// /update_menu — admin-only command to broadcast updated menu to all HRs
	bot.Handle("/update_menu", func(c tele.Context) error {
		if c.Sender().ID != adminTelegramID {
			return nil
		}
		_ = c.Send("Broadcasting menu to all HRs...")
		go hb.BroadcastMenu()
		return nil
	})

	// /start handler
	bot.Handle("/start", func(c tele.Context) error {
		sender := c.Sender()

		result, err := hrBotSvc.HandleStart(ctx, sender.ID)
		if err != nil {
			log.Printf("hr handle /start error for %d: %v", sender.ID, err)
			return c.Send(hrMsgError["en"])
		}

		if result.IsNew {
			lang := detectLang(sender.LanguageCode)
			markup := &tele.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
			btnPhone := markup.Contact(hrMsgBtnSharePhone[lang])
			markup.Reply(markup.Row(btnPhone))
			_ = c.Send(hrMsgWelcomeNew[lang])
			return c.Send(hrMsgSharePhone[lang], markup)
		}

		lang := result.HR.Language
		if lang == "" {
			lang = detectLang(sender.LanguageCode)
		}
		lang = langOrDefault(lang)
		name := strings.TrimSpace(result.HR.FirstName + " " + result.HR.LastName)
		return c.Send(fmt.Sprintf(hrMsgWelcomeBack[lang], name), hrMenu(lang))
	})

	// Contact (phone number) handler — create HR record
	bot.Handle(tele.OnContact, func(c tele.Context) error {
		sender := c.Sender()
		contact := c.Message().Contact

		if contact == nil || contact.PhoneNumber == "" {
			return nil
		}

		lang := detectLang(sender.LanguageCode)

		hr, err := hrBotSvc.HandlePhoneShared(ctx, sender.ID, contact.PhoneNumber, sender.FirstName, sender.LastName, sender.Username, lang)
		if err != nil {
			log.Printf("hr phone shared error for %d: %v", sender.ID, err)
			return c.Send(hrMsgError[lang])
		}

		lang = langOrDefault(hr.Language)
		return c.Send(fmt.Sprintf(hrMsgRegistered[lang], hr.FirstName), hrMenu(lang))
	})

	// Language change callback for HR
	bot.Handle(&tele.Btn{Unique: "hr_chg_lang"}, func(c tele.Context) error {
		sender := c.Sender()
		newLang := c.Callback().Data
		if newLang != "en" && newLang != "ru" && newLang != "uz" {
			return c.Respond(&tele.CallbackResponse{Text: "Unknown language"})
		}

		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		tgID := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgID)
		if err != nil {
			return c.Send(hrMsgError[newLang])
		}

		if _, err := hrBotSvc.UpdateLanguage(ctx, hr.ID, newLang); err != nil {
			log.Printf("hr change language error for %d: %v", sender.ID, err)
			return c.Send(hrMsgError[newLang])
		}

		return c.Send(hrMsgLangChanged[newLang], hrMenu(newLang))
	})

	// Text message handler
	bot.Handle(tele.OnText, func(c tele.Context) error {
		sender := c.Sender()

		state, _ := hrBotSvc.GetBotState(ctx, sender.ID)
		if state != nil {
			switch state.State {
			case domain.HRBotStateSharingPhone:
				lang := langOrDefault(state.Data["language"])
				return c.Send(hrMsgPhoneReminder[lang])
			case domain.HRBotStatePostingVacancy:
				return hb.handleVacancyText(ctx, c, state)
			case "hr_searching":
				return hb.handleSearchQuery(ctx, c, state)
			}
		}

		tgID := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgID)
		if err != nil {
			return c.Send(hrMsgError["en"])
		}
		lang := langOrDefault(hr.Language)

		text := c.Text()

		if isMenuButton(text, hrMenuBtnPostVacancy) {
			if err := hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStatePostingVacancy, map[string]string{"language": lang, "hr_id": hr.ID.String()}); err != nil {
				log.Printf("hr set posting state error for %d: %v", sender.ID, err)
			}
			return c.Send(hrMsgPostVacancy[lang])
		}
		if isMenuButton(text, hrMenuBtnMyVacancies) {
			return hb.handleMyVacancies(ctx, c, hr)
		}
		if isMenuButton(text, hrMenuBtnFindCandidates) {
			return hb.handleFindCandidatesStart(ctx, c, hr)
		}
		if isMenuButton(text, hrMenuBtnChangeLang) {
			markup := &tele.ReplyMarkup{}
			markup.Inline(
				markup.Row(
					markup.Data("🇬🇧 English", "hr_chg_lang", "en"),
					markup.Data("🇷🇺 Русский", "hr_chg_lang", "ru"),
					markup.Data("🇺🇿 O'zbek", "hr_chg_lang", "uz"),
				),
			)
			return c.Send(hrMsgChooseLang[lang], markup)
		}

		return c.Send(hrMsgUseMenu[lang], hrMenu(lang))
	})
}

func (hb *HRBot) handleVacancyText(ctx context.Context, c tele.Context, state *domain.BotState) error {
	sender := c.Sender()
	lang := langOrDefault(state.Data["language"])
	hrIDStr := state.Data["hr_id"]

	_ = c.Send(hrMsgParsingVacancy[lang])

	hrID, err := uuid.Parse(hrIDStr)
	if err != nil {
		log.Printf("hr parse vacancy: invalid hr_id %s: %v", hrIDStr, err)
		return c.Send(hrMsgError[lang])
	}

	result, err := hb.hrBotSvc.ParseVacancy(ctx, hrID, c.Text())
	if err != nil {
		log.Printf("hr parse vacancy error for %d: %v", sender.ID, err)
		return c.Send(hrMsgVacancyFailed[lang], hrMenu(lang))
	}

	if err := hb.hrBotSvc.ClearState(ctx, sender.ID); err != nil {
		log.Printf("hr clear state error for %d: %v", sender.ID, err)
	}

	title := vacancyTitle(result, lang)
	minStr := formatNumber(int64(result.Vacancy.SalaryMin))
	maxStr := formatNumber(int64(result.Vacancy.SalaryMax))

	msg := fmt.Sprintf(hrMsgVacancyCreated[lang], title, minStr, maxStr, result.Vacancy.SalaryCurrency, result.Vacancy.Format, result.Vacancy.Schedule)
	return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: hrMenu(lang)})
}

func (hb *HRBot) handleMyVacancies(ctx context.Context, c tele.Context, hr *domain.CompanyHR) error {
	lang := langOrDefault(hr.Language)

	vacancies, err := hb.hrBotSvc.ListMyVacancies(ctx, hr.ID)
	if err != nil {
		log.Printf("hr list vacancies error: %v", err)
		return c.Send(hrMsgError[lang])
	}

	if len(vacancies) == 0 {
		return c.Send(hrMsgNoVacancies[lang])
	}

	var sb strings.Builder
	for i, v := range vacancies {
		title := vacancyTitle(&v, lang)
		minStr := formatNumber(int64(v.Vacancy.SalaryMin))
		maxStr := formatNumber(int64(v.Vacancy.SalaryMax))
		sb.WriteString(fmt.Sprintf("%d. **%s**\n   %s – %s %s | %s\n\n", i+1, title, minStr, maxStr, v.Vacancy.SalaryCurrency, v.Vacancy.Status))
	}

	return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func (hb *HRBot) handleFindCandidatesStart(ctx context.Context, c tele.Context, hr *domain.CompanyHR) error {
	lang := langOrDefault(hr.Language)

	// Set state so next text message is treated as search query
	if err := hb.hrBotSvc.SetState(ctx, c.Sender().ID, "hr_searching", map[string]string{"language": lang, "hr_id": hr.ID.String()}); err != nil {
		log.Printf("hr set search state error: %v", err)
	}

	return c.Send(hrMsgFindCandidates[lang])
}

func (hb *HRBot) handleSearchQuery(ctx context.Context, c tele.Context, state *domain.BotState) error {
	sender := c.Sender()
	lang := langOrDefault(state.Data["language"])

	_ = c.Send(hrMsgSearching[lang])

	results, err := hb.hrBotSvc.SearchCandidates(ctx, c.Text())
	if err != nil {
		log.Printf("hr search candidates error for %d: %v", sender.ID, err)
		return c.Send(hrMsgError[lang], hrMenu(lang))
	}

	if err := hb.hrBotSvc.ClearState(ctx, sender.ID); err != nil {
		log.Printf("hr clear search state error for %d: %v", sender.ID, err)
	}

	if len(results) == 0 {
		return c.Send(hrMsgNoCandidates[lang], hrMenu(lang))
	}

	var sb strings.Builder
	for i, r := range results {
		user, err := hb.hrBotSvc.GetUser(ctx, r.UserID)
		if err != nil {
			continue
		}
		name := strings.TrimSpace(user.FirstName + " " + user.LastName)
		if name == "" {
			name = "—"
		}
		contact := ""
		if user.Telegram != "" {
			contact = user.Telegram
		} else if user.Phone != "" {
			contact = user.Phone
		}
		sb.WriteString(fmt.Sprintf("%d. **%s** (score: %.1f%%)", i+1, name, r.Score*100))
		if contact != "" {
			sb.WriteString(fmt.Sprintf(" — %s", contact))
		}
		sb.WriteString("\n")
	}

	return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: hrMenu(lang)})
}

func hrMenu(lang string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}
	markup.Reply(
		markup.Row(tele.Btn{Text: hrMenuBtnPostVacancy[lang]}, tele.Btn{Text: hrMenuBtnMyVacancies[lang]}),
		markup.Row(tele.Btn{Text: hrMenuBtnFindCandidates[lang]}, tele.Btn{Text: hrMenuBtnChangeLang[lang]}),
	)
	return markup
}

func vacancyTitle(v *application.VacancyWithDetails, lang string) string {
	for _, t := range v.Texts {
		if t.Lang == lang && t.Title != "" {
			return t.Title
		}
	}
	for _, t := range v.Texts {
		if t.Lang == "en" && t.Title != "" {
			return t.Title
		}
	}
	for _, t := range v.Texts {
		if t.Title != "" {
			return t.Title
		}
	}
	return "Untitled"
}

func detectLang(langCode string) string {
	switch {
	case strings.HasPrefix(langCode, "ru"):
		return "ru"
	case strings.HasPrefix(langCode, "uz"):
		return "uz"
	default:
		return "en"
	}
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
