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
	"en": "👋 Hi and welcome!\n\nI'm your AI-HR manager. I'll help you understand your real market value, find the right job, and earn more.\n\n🌍 Choose your language to get started:",
	"ru": "👋 Привет и добро пожаловать!\n\nЯ твой AI-HR менеджер. Помогу понять твою реальную стоимость на рынке, найти подходящую работу и зарабатывать больше.\n\n🌍 Выбери язык для начала:",
	"uz": "👋 Salom va xush kelibsiz!\n\nMen sizning AI-HR menejeringizman. Bozordagi haqiqiy qiymatni tushunishga, to'g'ri ish topishga va ko'proq ishlashga yordam beraman.\n\n🌍 Boshlash uchun tilni tanlang:",
}

var msgChooseRole = map[string]string{
	"en": "Hi 👋\n\nI'm your AI-HR manager. I'll help you understand your real market value and earn more.\n\nChoose what you need 👇",
	"ru": "Привет 👋\n\nЯ твой AI-HR менеджер. Помогу понять твою реальную стоимость на рынке и зарабатывать больше.\n\nВыбери, что тебе нужно 👇",
	"uz": "Salom 👋\n\nMen sizning AI-HR menejeringizman. Bozordagi haqiqiy qiymatni tushunishga va ko'proq ishlashga yordam beraman.\n\nSizga nima kerakligini tanlang 👇",
}

var msgBtnDetermineSalary = map[string]string{
	"en": "🔘 Determine salary",
	"ru": "🔘 Определить зарплату",
	"uz": "🔘 Maoshni aniqlash",
}

var msgBtnFindJob = map[string]string{
	"en": "🔘 Find a job",
	"ru": "🔘 Найти работу",
	"uz": "🔘 Ish topish",
}

var msgDetermineSalary = map[string]string{
	"en": "To determine your real salary and find offers, just tell me about yourself in free form. You can:\n\n• ✍️ write as text\n• 🎤 send a voice message\n• 📎 attach a resume / portfolio / PDF\n• 🔗 send a link\n\nIt's good to mention:\n— your role\n— years of experience\n— what exactly you do\n— skills\n— current income\n\nI'll analyze everything and show you the result.",
	"ru": "Чтобы определить твою реальную зарплату и подобрать предложения, просто расскажи о себе в свободной форме. Можно:\n\n• ✍️ написать текстом\n• 🎤 отправить голосовое\n• 📎 прикрепить резюме / портфолио / PDF\n• 🔗 отправить ссылку\n\nЖелательно указать:\n— кем работаешь\n— сколько лет опыта\n— чем конкретно занимаешься\n— навыки\n— текущий доход\n\nЯ всё разберу и покажу результат.",
	"uz": "Haqiqiy maoshingizni aniqlash va takliflar topish uchun o'zingiz haqingizda erkin shaklda gapirib bering. Mumkin:\n\n• ✍️ matn yozish\n• 🎤 ovozli xabar yuborish\n• 📎 rezyume / portfolio / PDF biriktirish\n• 🔗 havola yuborish\n\nQuyidagilarni ko'rsatish yaxshi:\n— kim bo'lib ishlaysiz\n— necha yillik tajriba\n— aniq nima bilan shug'ullanasiz\n— ko'nikmalar\n— hozirgi daromad\n\nHammasini tahlil qilib, natijani ko'rsataman.",
}

var msgSharePhone = map[string]string{
	"en": "📱 Almost done! Please share your phone number so employers can reach you.",
	"ru": "📱 Почти готово! Пожалуйста, поделитесь номером телефона, чтобы работодатели могли с вами связаться.",
	"uz": "📱 Deyarli tayyor! Iltimos, telefon raqamingizni ulashing, shunda ish beruvchilar siz bilan bog'lanishi mumkin.",
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

var msgWelcomeBackUser = map[string]string{
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
	"en": "Please choose an option from the buttons above ☝️",
	"ru": "Пожалуйста, выбери один из вариантов выше ☝️",
	"uz": "Iltimos, yuqoridagi tugmalardan birini tanlang ☝️",
}

var msgParsingText = map[string]string{
	"en": "Analyzing profile... ⏳",
	"ru": "Анализирую профиль… ⏳",
	"uz": "Profilni tahlil qilmoqdaman… ⏳",
}

var msgParsingFile = map[string]string{
	"en": "I received the file. Analyzing skills and experience…",
	"ru": "Я получил файл. Анализирую навыки и опыт…",
	"uz": "Faylni oldim. Ko'nikmalar va tajribani tahlil qilmoqdaman…",
}

var msgParsingVoice = map[string]string{
	"en": "I transcribed your voice message and created a profile. If anything needs clarifying — just say.",
	"ru": "Я расшифровал твоё голосовое сообщение и сформировал профиль. Если что-то нужно уточнить — скажи.",
	"uz": "Ovozli xabaringizni yozib oldim va profil yaratdim. Biror narsani aniqlashtirish kerak bo'lsa — ayting.",
}

var msgCollectedText = map[string]string{
	"en": "Got it ✅",
	"ru": "Принял ✅",
	"uz": "Qabul qildim ✅",
}

var msgCollectedFile = map[string]string{
	"en": "I received the file ✅",
	"ru": "Я получил файл ✅",
	"uz": "Faylni oldim ✅",
}

var msgCollectedVoice = map[string]string{
	"en": "I transcribed your voice message ✅",
	"ru": "Я расшифровал твоё голосовое сообщение ✅",
	"uz": "Ovozli xabaringizni yozib oldim ✅",
}

var msgAnythingElse = map[string]string{
	"en": "Would you like to add anything else?\nThe more information you provide, the more accurate my answer will be.",
	"ru": "Ещё что-нибудь добавишь?\nЧем больше информации, тем точнее я смогу ответить.",
	"uz": "Yana biror narsa qo'shasizmi?\nMa'lumot qancha ko'p bo'lsa, shuncha aniqroq javob bera olaman.",
}

var msgBtnDone = map[string]string{
	"en": "✅ Everything is ready, continue",
	"ru": "✅ Всё готово, продолжить",
	"uz": "✅ Hammasi tayyor, davom etish",
}

var msgProfileReady = map[string]string{
	"en": "✅ Your profile has been created! Tap below to view it 👇",
	"ru": "✅ Твой профиль готов! Нажми ниже, чтобы посмотреть 👇",
	"uz": "✅ Profilingiz tayyor! Ko'rish uchun quyidagini bosing 👇",
}

var msgBtnViewProfile = map[string]string{
	"en": "👤 View my profile",
	"ru": "👤 Посмотреть мой профиль",
	"uz": "👤 Profilimni ko'rish",
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

var msgOpenProfile = map[string]string{
	"en": "Tap the button below to view your profile 👇",
	"ru": "Нажмите кнопку ниже, чтобы посмотреть свой профиль 👇",
	"uz": "Profilingizni ko'rish uchun quyidagi tugmani bosing 👇",
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

		lang := langOrDefault(result.User.Language)
		return c.Send(fmt.Sprintf(msgWelcomeBackUser[lang], result.User.FirstName), userMenu(lang))
	})

	// Language selection callback — creates user and asks for phone
	bot.Handle(&tele.Btn{Unique: "lang"}, func(c tele.Context) error {
		language := c.Callback().Data
		if language == "" {
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

		lang, err := botSvc.HandleLanguageSelection(ctx, sender.ID, language, sender.FirstName, sender.LastName, sender.Username, photoData)
		if err != nil {
			log.Printf("language selection error for %d: %v", sender.ID, err)
			return c.Respond(&tele.CallbackResponse{Text: msgError["en"]})
		}

		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		// Ask for phone number via reply keyboard
		markup := &tele.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
		btnPhone := markup.Contact(msgBtnSharePhone[lang])
		markup.Reply(markup.Row(btnPhone))

		return c.Send(msgSharePhone[lang], markup)
	})

	// Goal selection callback (salary / job)
	bot.Handle(&tele.Btn{Unique: "goal"}, func(c tele.Context) error {
		goal := c.Callback().Data
		if goal != "salary" && goal != "job" {
			return c.Respond(&tele.CallbackResponse{Text: "Unknown action"})
		}

		lang, err := botSvc.HandleGoalSelection(ctx, c.Sender().ID, goal)
		if err != nil {
			log.Printf("goal selection error for %d: %v", c.Sender().ID, err)
			return c.Respond(&tele.CallbackResponse{Text: msgError[langOrDefault(lang)]})
		}

		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang = langOrDefault(lang)

		if goal == "salary" {
			return c.Send(msgDetermineSalary[lang])
		}

		return c.Send(fmt.Sprintf(msgRegisteredUser[lang], c.Sender().FirstName), userMenu(lang))
	})

	// "Done" button — process collected resume data
	bot.Handle(&tele.Btn{Unique: "done"}, func(c tele.Context) error {
		sender := c.Sender()

		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := getStateLang(ctx, botSvc, sender.ID)
		_ = c.Send(msgParsingText[lang])

		result, err := botSvc.ProcessCollectedResume(ctx, sender.ID)
		if err != nil {
			log.Printf("process collected resume error for %d: %v", sender.ID, err)
			return c.Send(msgResumeFailed[lang])
		}

		_ = result // result stored in DB

		if tb.webAppURL != "" {
			markup := &tele.ReplyMarkup{}
			markup.Inline(
				markup.Row(tele.Btn{Text: msgBtnViewProfile[lang], WebApp: &tele.WebApp{URL: tb.webAppURL + "?view=profile"}}),
			)
			return c.Send(msgProfileReady[lang], markup)
		}
		return c.Send(msgProfileReady[lang])
	})

	// Contact (phone number) handler — after phone, show goal buttons
	bot.Handle(tele.OnContact, func(c tele.Context) error {
		sender := c.Sender()
		contact := c.Message().Contact

		if contact == nil || contact.PhoneNumber == "" {
			return nil
		}

		lang, err := botSvc.HandlePhoneShared(ctx, sender.ID, contact.PhoneNumber)
		if err != nil {
			log.Printf("phone shared error for %d: %v", sender.ID, err)
			return c.Send(msgError[langOrDefault(lang)])
		}

		lang = langOrDefault(lang)

		markup := &tele.ReplyMarkup{}
		btnSalary := markup.Data(msgBtnDetermineSalary[lang], "goal", "salary")
		btnJob := markup.Data(msgBtnFindJob[lang], "goal", "job")
		markup.Inline(
			markup.Row(btnSalary),
			markup.Row(btnJob),
		)
		return c.Send(msgChooseRole[lang], markup)
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
			case domain.BotStateCollectingResume:
				lang = langOrDefault(state.Data["language"])
				if err := botSvc.AddResumeText(ctx, sender.ID, c.Text()); err != nil {
					log.Printf("add resume text error for %d: %v", sender.ID, err)
					return c.Send(msgError[lang])
				}
				_ = c.Send(msgCollectedText[lang])
				return c.Send(msgAnythingElse[lang], anythingElseMarkup(lang))
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
			if tb.webAppURL != "" {
				return c.Send(msgOpenProfile[lang], profileViewInline(lang, tb.webAppURL))
			}
			return c.Send(msgSendResume[lang])
		}
		if isMenuButton(text, menuBtnSearchVacancies) {
			if tb.webAppURL != "" {
				return c.Send(msgOpenSearch[lang], searchVacanciesInline(lang, tb.webAppURL))
			}
			return nil
		}

		// Treat as resume text
		_ = c.Send(msgParsingText[lang])

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
			case domain.BotStateCollectingResume:
				lang = langOrDefault(state.Data["language"])
				mimeType := doc.MIME
				if !isAllowedMIME(mimeType) {
					return c.Send(msgUnsupportedFile[lang])
				}
				reader, err := bot.File(&doc.File)
				if err != nil {
					return c.Send(msgDownloadFailed[lang])
				}
				fileData, _ := io.ReadAll(reader)
				reader.Close()
				if err := botSvc.AddResumeFile(ctx, sender.ID, fileData, mimeType); err != nil {
					log.Printf("add resume file error for %d: %v", sender.ID, err)
					return c.Send(msgError[lang])
				}
				_ = c.Send(msgCollectedFile[lang])
				return c.Send(msgAnythingElse[lang], anythingElseMarkup(lang))
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

		_ = c.Send(msgParsingFile[lang])

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
			case domain.BotStateCollectingResume:
				lang = langOrDefault(state.Data["language"])
				reader, err := bot.File(&photo.File)
				if err != nil {
					return c.Send(msgDownloadFailed[lang])
				}
				fileData, _ := io.ReadAll(reader)
				reader.Close()
				if err := botSvc.AddResumeFile(ctx, sender.ID, fileData, "image/jpeg"); err != nil {
					log.Printf("add resume photo error for %d: %v", sender.ID, err)
					return c.Send(msgError[lang])
				}
				_ = c.Send(msgCollectedFile[lang])
				return c.Send(msgAnythingElse[lang], anythingElseMarkup(lang))
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

		_ = c.Send(msgParsingFile[lang])

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
			case domain.BotStateCollectingResume:
				lang = langOrDefault(state.Data["language"])
				reader, err := bot.File(&voice.File)
				if err != nil {
					return c.Send(msgDownloadFailed[lang])
				}
				fileData, _ := io.ReadAll(reader)
				reader.Close()
				mimeType := voice.MIME
				if mimeType == "" {
					mimeType = "audio/ogg"
				}
				if err := botSvc.AddResumeFile(ctx, sender.ID, fileData, mimeType); err != nil {
					log.Printf("add resume voice error for %d: %v", sender.ID, err)
					return c.Send(msgError[lang])
				}
				_ = c.Send(msgCollectedVoice[lang])
				return c.Send(msgAnythingElse[lang], anythingElseMarkup(lang))
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

		_ = c.Send(msgParsingVoice[lang])

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

func anythingElseMarkup(lang string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	btn := markup.Data(msgBtnDone[lang], "done")
	markup.Inline(markup.Row(btn))
	return markup
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

func profileViewInline(lang, webAppURL string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(tele.Btn{Text: menuBtnUpdateResume[lang], WebApp: &tele.WebApp{URL: webAppURL + "?view=profile"}}),
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
