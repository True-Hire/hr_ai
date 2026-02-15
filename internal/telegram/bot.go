package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"

	tele "gopkg.in/telebot.v3"
)

// -- Localized messages (all keys: en, ru, uz) --

var msgWelcomeNew = map[string]string{
	"en": "👋 Hi and welcome!\n\nThis bot helps people and companies find each other faster using AI.\n\n🔹 Looking for a job?\nUpload your resume and get matched with relevant positions.\n\n🔹 Hiring for your company?\nCreate vacancies and find the right candidates without endless searching.\n\nSmart search, multiple languages, better matches.\n\n🌍 Choose your language to get started:",
	"ru": "👋 Привет и добро пожаловать!\n\nЭтот бот помогает людям и компаниям найти друг друга быстрее с помощью ИИ.\n\n🔹 Ищете работу?\nЗагрузите резюме и получите подходящие вакансии.\n\n🔹 Нанимаете сотрудников?\nСоздавайте вакансии и находите подходящих кандидатов без лишних поисков.\n\nУмный поиск, несколько языков, лучшие совпадения.\n\n🌍 Выберите язык для начала:",
	"uz": "👋 Salom va xush kelibsiz!\n\nBu bot odamlar va kompaniyalarga sun'iy intellekt yordamida bir-birini tezroq topishga yordam beradi.\n\n🔹 Ish qidiryapsizmi?\nRezyumeni yuklang va mos vakansiyalarga ega bo'ling.\n\n🔹 Xodim yollamoqchimisiz?\nVakansiyalar yarating va mos nomzodlarni ortiqcha qidiruvlarsiz toping.\n\nAqlli qidiruv, bir nechta til, yaxshiroq natijalar.\n\n🌍 Boshlash uchun tilni tanlang:",
}

var msgChooseRole = map[string]string{
	"en": "Great! Now tell us — what brings you here?\n\n🔹 Looking for a job? Choose \"Job Seeker\"\n🔹 Hiring for your company? Choose \"Hirer\"",
	"ru": "Отлично! Расскажите нам — что привело вас сюда?\n\n🔹 Ищете работу? Выберите «Соискатель»\n🔹 Нанимаете сотрудников? Выберите «Работодатель»",
	"uz": "Ajoyib! Ayting-chi — sizni bu yerga nima olib keldi?\n\n🔹 Ish qidiryapsizmi? «Ish izlovchi» ni tanlang\n🔹 Xodim yollamoqchimisiz? «Ish beruvchi» ni tanlang",
}

var msgBtnJobSeeker = map[string]string{
	"en": "🔍 Job Seeker",
	"ru": "🔍 Соискатель",
	"uz": "🔍 Ish izlovchi",
}

var msgBtnHirer = map[string]string{
	"en": "🏢 Hirer",
	"ru": "🏢 Работодатель",
	"uz": "🏢 Ish beruvchi",
}

var msgSharePhone = map[string]string{
	"en": "📱 Almost done! Please share your phone number so employers can reach you.",
	"ru": "📱 Почти готово! Пожалуйста, поделитесь номером телефона, чтобы работодатели могли с вами связаться.",
	"uz": "📱 Deyarli tayyor! Iltimos, telefon raqamingizni ulashing, shunda ish beruvchilar siz bilan bog'lanishi mumkin.",
}

var msgSharePhoneHR = map[string]string{
	"en": "📱 Almost done! Please share your phone number so candidates can reach you.",
	"ru": "📱 Почти готово! Пожалуйста, поделитесь номером телефона, чтобы кандидаты могли с вами связаться.",
	"uz": "📱 Deyarli tayyor! Iltimos, telefon raqamingizni ulashing, shunda nomzodlar siz bilan bog'lanishi mumkin.",
}

var msgBtnSharePhone = map[string]string{
	"en": "📞 Share phone number",
	"ru": "📞 Поделиться номером",
	"uz": "📞 Telefon raqamini ulashish",
}

var msgPhoneReminder = map[string]string{
	"en": "Please share your phone number using the button below 👇",
	"ru": "Пожалуйста, поделитесь номером телефона, нажав кнопку ниже 👇",
	"uz": "Iltimos, quyidagi tugma orqali telefon raqamingizni ulashing 👇",
}

var msgRegisteredUser = map[string]string{
	"en": "✅ Registration complete, %s! Welcome aboard!\n\nYou can now use the menu below to get started 👇",
	"ru": "✅ Регистрация завершена, %s! Добро пожаловать!\n\nИспользуйте меню ниже, чтобы начать 👇",
	"uz": "✅ Ro'yxatdan o'tish yakunlandi, %s! Xush kelibsiz!\n\nBoshlash uchun quyidagi menyudan foydalaning 👇",
}

var msgRegisteredHR = map[string]string{
	"en": "✅ Registration complete, %s! Welcome aboard!\n\nYou can now use the menu below to get started 👇",
	"ru": "✅ Регистрация завершена, %s! Добро пожаловать!\n\nИспользуйте меню ниже, чтобы начать 👇",
	"uz": "✅ Ro'yxatdan o'tish yakunlandi, %s! Xush kelibsiz!\n\nBoshlash uchun quyidagi menyudan foydalaning 👇",
}

// -- Menu buttons for job seekers --

var menuBtnUpdateResume = map[string]string{
	"en": "📄 Update Resume",
	"ru": "📄 Обновить резюме",
	"uz": "📄 Rezyumeni yangilash",
}

var menuBtnSearchVacancies = map[string]string{
	"en": "🔍 Search Vacancies",
	"ru": "🔍 Поиск вакансий",
	"uz": "🔍 Vakansiyalarni qidirish",
}

// -- Menu buttons for HRs --

var menuBtnCreateVacancy = map[string]string{
	"en": "📝 Create Vacancy",
	"ru": "📝 Создать вакансию",
	"uz": "📝 Vakansiya yaratish",
}

var menuBtnMyVacancies = map[string]string{
	"en": "📋 My Vacancies",
	"ru": "📋 Мои вакансии",
	"uz": "📋 Mening vakansiyalarim",
}

var menuBtnFindCandidates = map[string]string{
	"en": "👥 Find Candidates",
	"ru": "👥 Найти кандидатов",
	"uz": "👥 Nomzodlarni topish",
}

var msgWelcomeBackUser = map[string]string{
	"en": "👋 Welcome back %s! Glad to see you again.\n\nChoose an option from the menu below 👇",
	"ru": "👋 С возвращением, %s! Рады снова вас видеть.\n\nВыберите нужный пункт в меню ниже 👇",
	"uz": "👋 Qaytganingiz bilan, %s! Sizni yana ko'rib turganimizdan xursandmiz.\n\nQuyidagi menyudan tanlang 👇",
}

var msgWelcomeBackHR = map[string]string{
	"en": "👋 Welcome back %s! Glad to see you again.\n\nChoose an option from the menu below 👇",
	"ru": "👋 С возвращением, %s! Рады снова вас видеть.\n\nВыберите нужный пункт в меню ниже 👇",
	"uz": "👋 Qaytganingiz bilan, %s! Sizni yana ko'rib turganimizdan xursandmiz.\n\nQuyidagi menyudan tanlang 👇",
}

var msgChooseLangReminder = map[string]string{
	"en": "Please choose a language from the buttons above ☝️",
	"ru": "Пожалуйста, выберите язык, нажав на кнопку выше ☝️",
	"uz": "Iltimos, yuqoridagi tugmalardan tilni tanlang ☝️",
}

var msgChooseRoleReminder = map[string]string{
	"en": "Please choose your role from the buttons above ☝️",
	"ru": "Пожалуйста, выберите вашу роль, нажав на кнопку выше ☝️",
	"uz": "Iltimos, yuqoridagi tugmalardan rolingizni tanlang ☝️",
}

var msgParsingResume = map[string]string{
	"en": "Parsing your resume... This may take a moment.",
	"ru": "Обрабатываем ваше резюме... Это может занять немного времени.",
	"uz": "Rezyumeni tahlil qilmoqdamiz... Bu biroz vaqt olishi mumkin.",
}

var msgResumeSuccess = map[string]string{
	"en": "Your resume has been parsed successfully!\n\nSource language: %s\nProfile fields: %s\nExperience items: %s\nEducation items: %s",
	"ru": "Ваше резюме успешно обработано!\n\nИсходный язык: %s\nПоля профиля: %s\nОпыт работы: %s\nОбразование: %s",
	"uz": "Rezyumengiz muvaffaqiyatli tahlil qilindi!\n\nManba tili: %s\nProfil maydonlari: %s\nIsh tajribasi: %s\nTa'lim: %s",
}

var msgResumeFailed = map[string]string{
	"en": "Failed to parse your resume. Please try again.",
	"ru": "Не удалось обработать резюме. Пожалуйста, попробуйте ещё раз.",
	"uz": "Rezyumeni tahlil qilib bo'lmadi. Iltimos, qaytadan urinib ko'ring.",
}

var msgUnsupportedFile = map[string]string{
	"en": "Unsupported file type. Please send a PDF, image (PNG/JPG), or text file.",
	"ru": "Неподдерживаемый тип файла. Пожалуйста, отправьте PDF, изображение (PNG/JPG) или текстовый файл.",
	"uz": "Qo'llab-quvvatlanmaydigan fayl turi. Iltimos, PDF, rasm (PNG/JPG) yoki matn faylini yuboring.",
}

var msgDownloadFailed = map[string]string{
	"en": "Failed to download your file. Please try again.",
	"ru": "Не удалось загрузить файл. Пожалуйста, попробуйте ещё раз.",
	"uz": "Faylni yuklab bo'lmadi. Iltimos, qaytadan urinib ko'ring.",
}

var msgStartFirst = map[string]string{
	"en": "Something went wrong. Please try /start first.",
	"ru": "Что-то пошло не так. Пожалуйста, начните с /start.",
	"uz": "Xatolik yuz berdi. Iltimos, avval /start buyrug'ini yuboring.",
}

var msgError = map[string]string{
	"en": "Something went wrong. Please try again.",
	"ru": "Что-то пошло не так. Пожалуйста, попробуйте ещё раз.",
	"uz": "Xatolik yuz berdi. Iltimos, qaytadan urinib ko'ring.",
}

var msgOpenSearch = map[string]string{
	"en": "Tap the button below to browse vacancies 👇",
	"ru": "Нажмите кнопку ниже для поиска вакансий 👇",
	"uz": "Vakansiyalarni ko'rish uchun quyidagi tugmani bosing 👇",
}

var msgSendResume = map[string]string{
	"en": "📄 Send us your resume as a PDF, photo, or text file.\n\nOr simply tell us about yourself — your experience, skills, and interests. We'll create the resume for you!",
	"ru": "📄 Отправьте нам резюме в формате PDF, фото или текстового файла.\n\nИли просто расскажите о себе — ваш опыт, навыки и интересы. Мы составим резюме за вас!",
	"uz": "📄 Rezyumeni PDF, rasm yoki matn fayli sifatida yuboring.\n\nYoki shunchaki o'zingiz haqingizda gapirib bering — tajribangiz, ko'nikmalaringiz va qiziqishlaringiz. Biz rezyumeni siz uchun tuzib beramiz!",
}

// -- Bot --

type Bot struct {
	bot       *tele.Bot
	botSvc    *application.BotService
	webAppURL string
}

func NewBot(token string, botSvc *application.BotService, webAppURL string) (*Bot, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	tb := &Bot{bot: b, botSvc: botSvc, webAppURL: webAppURL}
	tb.registerHandlers()
	return tb, nil
}

func (tb *Bot) Start() {
	log.Println("telegram bot starting...")
	tb.bot.Start()
}

func (tb *Bot) Stop() {
	log.Println("stopping telegram bot...")
	tb.bot.Stop()
	log.Println("telegram bot stopped")
}

func (tb *Bot) registerHandlers() {
	ctx := context.Background()
	bot := tb.bot
	botSvc := tb.botSvc

	// /start handler
	bot.Handle("/start", func(c tele.Context) error {
		sender := c.Sender()

		result, err := botSvc.HandleStart(ctx, sender.ID)
		if err != nil {
			log.Printf("handle /start error for %d: %v", sender.ID, err)
			return c.Send(msgError["en"])
		}

		if result.IsNew {
			markup := &tele.ReplyMarkup{}
			btnEn := markup.Data("🇬🇧 English", "lang", "en")
			btnRu := markup.Data("🇷🇺 Русский", "lang", "ru")
			btnUz := markup.Data("🇺🇿 O'zbek", "lang", "uz")
			markup.Inline(
				markup.Row(btnEn),
				markup.Row(btnRu),
				markup.Row(btnUz),
			)
			return c.Send(msgWelcomeNew["en"], markup)
		}

		if result.IsHR {
			lang := langOrDefault(result.HR.Language)
			return c.Send(fmt.Sprintf(msgWelcomeBackHR[lang], result.HR.FirstName), hrMenu(lang))
		}

		lang := langOrDefault(result.User.Language)
		return c.Send(fmt.Sprintf(msgWelcomeBackUser[lang], result.User.FirstName), userMenu(lang))
	})

	// Language selection callback
	bot.Handle(&tele.Btn{Unique: "lang"}, func(c tele.Context) error {
		language := c.Callback().Data
		if language == "" {
			return c.Respond(&tele.CallbackResponse{Text: "Unknown action"})
		}

		sender := c.Sender()

		lang, err := botSvc.HandleLanguageSelection(ctx, sender.ID, language)
		if err != nil {
			log.Printf("language selection error for %d: %v", sender.ID, err)
			return c.Respond(&tele.CallbackResponse{Text: msgError["en"]})
		}

		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		markup := &tele.ReplyMarkup{}
		btnSeeker := markup.Data(msgBtnJobSeeker[lang], "role", "seeker")
		btnHirer := markup.Data(msgBtnHirer[lang], "role", "hr")
		markup.Inline(
			markup.Row(btnSeeker),
			markup.Row(btnHirer),
		)
		return c.Send(msgChooseRole[lang], markup)
	})

	// Role selection callback
	bot.Handle(&tele.Btn{Unique: "role"}, func(c tele.Context) error {
		role := c.Callback().Data
		if role != "seeker" && role != "hr" {
			return c.Respond(&tele.CallbackResponse{Text: "Unknown action"})
		}

		sender := c.Sender()

		var photoData []byte
		photos, err := bot.ProfilePhotosOf(sender)
		if err == nil && len(photos) > 0 {
			reader, err := bot.File(&photos[0].File)
			if err == nil {
				photoData, _ = io.ReadAll(reader)
				reader.Close()
			}
		}

		lang, isHR, err := botSvc.HandleRoleSelection(ctx, sender.ID, role, sender.FirstName, sender.LastName, sender.Username, photoData)
		if err != nil {
			log.Printf("role selection error for %d: %v", sender.ID, err)
			return c.Respond(&tele.CallbackResponse{Text: msgError[langOrDefault(lang)]})
		}

		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang = langOrDefault(lang)

		// Ask for phone number via reply keyboard
		phoneMsg := msgSharePhone[lang]
		if isHR {
			phoneMsg = msgSharePhoneHR[lang]
		}

		markup := &tele.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
		btnPhone := markup.Contact(msgBtnSharePhone[lang])
		markup.Reply(markup.Row(btnPhone))

		return c.Send(phoneMsg, markup)
	})

	// Contact (phone number) handler
	bot.Handle(tele.OnContact, func(c tele.Context) error {
		sender := c.Sender()
		contact := c.Message().Contact

		if contact == nil || contact.PhoneNumber == "" {
			return nil
		}

		lang, err := botSvc.HandlePhoneShared(ctx, sender.ID, contact.PhoneNumber)
		if err != nil {
			log.Printf("phone shared error for %d: %v", sender.ID, err)
			lang = langOrDefault(lang)
			return c.Send(msgError[lang])
		}

		lang = langOrDefault(lang)

		// Check if HR or user to send correct success message with menu
		result, err := botSvc.HandleStart(ctx, sender.ID)
		if err != nil {
			return c.Send(fmt.Sprintf(msgRegisteredUser[lang], sender.FirstName), userMenu(lang))
		}

		if result.IsHR {
			return c.Send(fmt.Sprintf(msgRegisteredHR[lang], result.HR.FirstName), hrMenu(lang))
		}
		return c.Send(fmt.Sprintf(msgRegisteredUser[lang], result.User.FirstName), userMenu(lang))
	})

	// Text message handler
	bot.Handle(tele.OnText, func(c tele.Context) error {
		sender := c.Sender()
		lang := getStateLang(ctx, botSvc, sender.ID)

		state, _ := botSvc.GetBotState(ctx, sender.ID)
		if state != nil {
			switch state.State {
			case domain.BotStateChoosingLanguage:
				return c.Send(msgChooseLangReminder["en"] + "\n" + msgChooseLangReminder["ru"] + "\n" + msgChooseLangReminder["uz"])
			case domain.BotStateChoosingRole:
				return c.Send(msgChooseRoleReminder[langOrDefault(state.Data["language"])])
			case domain.BotStateSharingPhone:
				return c.Send(msgPhoneReminder[langOrDefault(state.Data["language"])])
			}
		}

		user, err := ensureUser(ctx, botSvc, sender)
		if err != nil {
			return c.Send(msgStartFirst[lang])
		}
		lang = langOrDefault(user.Language)

		// Handle menu button taps
		text := c.Text()
		if isMenuButton(text, menuBtnUpdateResume) {
			return c.Send(msgSendResume[lang])
		}
		if isMenuButton(text, menuBtnCreateVacancy) ||
			isMenuButton(text, menuBtnMyVacancies) || isMenuButton(text, menuBtnFindCandidates) {
			// Placeholder for future functionality
			return nil
		}
		if isMenuButton(text, menuBtnSearchVacancies) {
			if tb.webAppURL != "" {
				return c.Send(msgOpenSearch[lang], searchVacanciesInline(lang, tb.webAppURL))
			}
			return nil
		}

		// Treat as resume text
		_ = c.Send(msgParsingResume[lang])

		result, err := botSvc.HandleResumeText(ctx, user.ID, c.Text())
		if err != nil {
			log.Printf("parse resume text error for %s: %v", user.ID, err)
			return c.Send(msgResumeFailed[lang])
		}

		return c.Send(fmt.Sprintf(msgResumeSuccess[lang],
			result.SourceLang,
			itoa(len(result.Fields)),
			itoa(len(result.Experience)),
			itoa(len(result.Education))))
	})

	// Document handler
	bot.Handle(tele.OnDocument, func(c tele.Context) error {
		sender := c.Sender()
		doc := c.Message().Document
		lang := getStateLang(ctx, botSvc, sender.ID)

		state, _ := botSvc.GetBotState(ctx, sender.ID)
		if state != nil {
			switch state.State {
			case domain.BotStateChoosingLanguage:
				return c.Send(msgChooseLangReminder["en"] + "\n" + msgChooseLangReminder["ru"] + "\n" + msgChooseLangReminder["uz"])
			case domain.BotStateChoosingRole:
				return c.Send(msgChooseRoleReminder[langOrDefault(state.Data["language"])])
			case domain.BotStateSharingPhone:
				return c.Send(msgPhoneReminder[langOrDefault(state.Data["language"])])
			}
		}

		user, err := ensureUser(ctx, botSvc, sender)
		if err != nil {
			return c.Send(msgStartFirst[lang])
		}
		lang = langOrDefault(user.Language)

		mimeType := doc.MIME
		if !isAllowedMIME(mimeType) {
			return c.Send(msgUnsupportedFile[lang])
		}

		reader, err := bot.File(&doc.File)
		if err != nil {
			log.Printf("download file error: %v", err)
			return c.Send(msgDownloadFailed[lang])
		}
		defer reader.Close()

		fileData, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("read file error: %v", err)
			return c.Send(msgDownloadFailed[lang])
		}

		_ = c.Send(msgParsingResume[lang])

		result, err := botSvc.HandleResumeFile(ctx, user.ID, fileData, mimeType)
		if err != nil {
			log.Printf("parse resume file error for %s: %v", user.ID, err)
			return c.Send(msgResumeFailed[lang])
		}

		return c.Send(fmt.Sprintf(msgResumeSuccess[lang],
			result.SourceLang,
			itoa(len(result.Fields)),
			itoa(len(result.Experience)),
			itoa(len(result.Education))))
	})

	// Photo handler
	bot.Handle(tele.OnPhoto, func(c tele.Context) error {
		sender := c.Sender()
		photo := c.Message().Photo
		lang := getStateLang(ctx, botSvc, sender.ID)

		state, _ := botSvc.GetBotState(ctx, sender.ID)
		if state != nil {
			switch state.State {
			case domain.BotStateChoosingLanguage:
				return c.Send(msgChooseLangReminder["en"] + "\n" + msgChooseLangReminder["ru"] + "\n" + msgChooseLangReminder["uz"])
			case domain.BotStateChoosingRole:
				return c.Send(msgChooseRoleReminder[langOrDefault(state.Data["language"])])
			case domain.BotStateSharingPhone:
				return c.Send(msgPhoneReminder[langOrDefault(state.Data["language"])])
			}
		}

		user, err := ensureUser(ctx, botSvc, sender)
		if err != nil {
			return c.Send(msgStartFirst[lang])
		}
		lang = langOrDefault(user.Language)

		reader, err := bot.File(&photo.File)
		if err != nil {
			log.Printf("download photo error: %v", err)
			return c.Send(msgDownloadFailed[lang])
		}
		defer reader.Close()

		fileData, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("read photo error: %v", err)
			return c.Send(msgDownloadFailed[lang])
		}

		_ = c.Send(msgParsingResume[lang])

		result, err := botSvc.HandleResumeFile(ctx, user.ID, fileData, "image/jpeg")
		if err != nil {
			log.Printf("parse resume photo error for %s: %v", user.ID, err)
			return c.Send(msgResumeFailed[lang])
		}

		return c.Send(fmt.Sprintf(msgResumeSuccess[lang],
			result.SourceLang,
			itoa(len(result.Fields)),
			itoa(len(result.Experience)),
			itoa(len(result.Education))))
	})

	// Voice message handler
	bot.Handle(tele.OnVoice, func(c tele.Context) error {
		sender := c.Sender()
		voice := c.Message().Voice
		lang := getStateLang(ctx, botSvc, sender.ID)

		state, _ := botSvc.GetBotState(ctx, sender.ID)
		if state != nil {
			switch state.State {
			case domain.BotStateChoosingLanguage:
				return c.Send(msgChooseLangReminder["en"] + "\n" + msgChooseLangReminder["ru"] + "\n" + msgChooseLangReminder["uz"])
			case domain.BotStateChoosingRole:
				return c.Send(msgChooseRoleReminder[langOrDefault(state.Data["language"])])
			case domain.BotStateSharingPhone:
				return c.Send(msgPhoneReminder[langOrDefault(state.Data["language"])])
			}
		}

		user, err := ensureUser(ctx, botSvc, sender)
		if err != nil {
			return c.Send(msgStartFirst[lang])
		}
		lang = langOrDefault(user.Language)

		reader, err := bot.File(&voice.File)
		if err != nil {
			log.Printf("download voice error: %v", err)
			return c.Send(msgDownloadFailed[lang])
		}
		defer reader.Close()

		fileData, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("read voice error: %v", err)
			return c.Send(msgDownloadFailed[lang])
		}

		mimeType := voice.MIME
		if mimeType == "" {
			mimeType = "audio/ogg"
		}

		_ = c.Send(msgParsingResume[lang])

		result, err := botSvc.HandleResumeFile(ctx, user.ID, fileData, mimeType)
		if err != nil {
			log.Printf("parse resume voice error for %s: %v", user.ID, err)
			return c.Send(msgResumeFailed[lang])
		}

		return c.Send(fmt.Sprintf(msgResumeSuccess[lang],
			result.SourceLang,
			itoa(len(result.Fields)),
			itoa(len(result.Experience)),
			itoa(len(result.Education))))
	})
}

func userMenu(lang string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}
	markup.Reply(
		markup.Row(tele.Btn{Text: menuBtnUpdateResume[lang]}),
		markup.Row(tele.Btn{Text: menuBtnSearchVacancies[lang]}),
	)
	return markup
}

func searchVacanciesInline(lang, webAppURL string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(tele.Btn{Text: menuBtnSearchVacancies[lang], WebApp: &tele.WebApp{URL: webAppURL}}),
	)
	return markup
}

func hrMenu(lang string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{ResizeKeyboard: true}
	markup.Reply(
		markup.Row(tele.Btn{Text: menuBtnCreateVacancy[lang]}),
		markup.Row(tele.Btn{Text: menuBtnMyVacancies[lang]}),
		markup.Row(tele.Btn{Text: menuBtnFindCandidates[lang]}),
	)
	return markup
}

// getStateLang tries to get the language from bot state data, falls back to "en".
func getStateLang(ctx context.Context, botSvc *application.BotService, senderID int64) string {
	state, err := botSvc.GetBotState(ctx, senderID)
	if err == nil && state != nil && state.Data["language"] != "" {
		return langOrDefault(state.Data["language"])
	}
	return "en"
}

func ensureUser(ctx context.Context, botSvc *application.BotService, sender *tele.User) (*domain.User, error) {
	result, err := botSvc.HandleStart(ctx, sender.ID)
	if err != nil {
		return nil, err
	}
	if result.User != nil {
		return result.User, nil
	}
	return nil, fmt.Errorf("user not found")
}

func isAllowedMIME(mime string) bool {
	switch mime {
	case "application/pdf", "image/png", "image/jpeg", "text/plain",
		"audio/ogg", "audio/mpeg", "audio/wav", "audio/mp4", "audio/webm":
		return true
	}
	return false
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func langOrDefault(lang string) string {
	if lang == "en" || lang == "ru" || lang == "uz" {
		return lang
	}
	return "en"
}

func isMenuButton(text string, btnTexts map[string]string) bool {
	for _, v := range btnTexts {
		if text == v {
			return true
		}
	}
	return false
}
