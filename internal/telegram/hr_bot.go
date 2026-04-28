package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"

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
	"en": "Thank you for registering, %s! 🎉\n\nNow you can use the menu below to manage vacancies and find candidates 👇",
	"ru": "Спасибо за регистрацию, %s! 🎉\n\nТеперь вы можете использовать меню ниже для управления вакансиями и поиска кандидатов 👇",
	"uz": "Ro'yxatdan o'tganingiz uchun rahmat, %s! 🎉\n\nEndi quyidagi menyu orqali vakansiyalarni boshqarish va nomzodlarni topish mumkin 👇",
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

var hrMenuBtnCreateVacancy = map[string]string{
	"en": "📝 Create Vacancy",
	"ru": "📝 Создать вакансию",
	"uz": "📝 Vakansiya yaratish",
}

var hrMenuBtnActiveVacancies = map[string]string{
	"en": "📋 Active Vacancies",
	"ru": "📋 Активные вакансии",
	"uz": "📋 Faol vakansiyalar",
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

var hrMsgCompanyDataPrompt = map[string]string{
	"en": "Before creating a vacancy, tell about your company.\n\nThis will help me:\n— automatically create proper vacancies\n— show candidates information about the company\n— find candidates more accurately for your field of activity\n\nYou can add information now or skip this step.",
	"ru": "Перед созданием вакансии расскажите о вашей компании.\n\nЭто поможет мне:\n— автоматически формировать правильные вакансии\n— показывать кандидатам информацию о компании\n— точнее подбирать кандидатов под вашу сферу деятельности\n\nВы можете добавить информацию сейчас или пропустить этот шаг.",
	"uz": "Vakansiya yaratishdan oldin kompaniyangiz haqida gapirib bering.\n\nBu menga yordam beradi:\n— vakansiyalarni avtomatik to'g'ri shakllantirish\n— nomzodlarga kompaniya haqida ma'lumot ko'rsatish\n— faoliyat sohasiga mos nomzodlarni aniqroq tanlash\n\nSiz hozir ma'lumot qo'shishingiz yoki bu bosqichni o'tkazib yuborishingiz mumkin.",
}

var hrBtnAddCompanyInfo = map[string]string{
	"en": "📋 Add company information",
	"ru": "📋 Добавить информацию о компании",
	"uz": "📋 Kompaniya haqida ma'lumot qo'shish",
}

var hrBtnSkipCompanyInfo = map[string]string{
	"en": "⏩ Skip and create vacancy",
	"ru": "⏩ Пропустить и создать вакансию",
	"uz": "⏩ O'tkazib yuborish va vakansiya yaratish",
}

var hrMsgCompanyDataStep = map[string]string{
	"en": "Send information about your company.\n\nYou can:\n\n✍️ write as text\n🎤 send a voice message\n📎 attach files\n\nPlease include:\n— Company name\n— Logo\n— City / address\n— What the company does\n— etc.",
	"ru": "Отправьте информацию о вашей компании.\n\nВы можете:\n\n✍️ написать текстом\n🎤 отправить голосовое\n📎 прикрепить файлы\n\nУкажите, пожалуйста:\n— Название компании\n— Логотип\n— Город / адрес\n— Чем занимается компания\n— д.р.",
	"uz": "Kompaniyangiz haqida ma'lumot yuboring.\n\nSiz quyidagilarni yuborishingiz mumkin:\n\n✍️ matn yozish\n🎤 ovozli xabar yuborish\n📎 fayllar biriktirish\n\nIltimos, quyidagilarni ko'rsating:\n— Kompaniya nomi\n— Logotip\n— Shahar / manzil\n— Kompaniya nima bilan shug'ullanadi\n— va boshqalar",
}

var hrMsgParsingCompany = map[string]string{
	"en": "🔎 Got the information. Analyzing company data...",
	"ru": "🔎 Получил информацию. Анализирую данные компании…",
	"uz": "🔎 Ma'lumotni oldim. Kompaniya ma'lumotlarini tahlil qilmoqdaman…",
}

var hrMsgCompanySaved = map[string]string{
	"en": "✅ Company information saved!\n\nNow let's create your vacancy.",
	"ru": "✅ Информация о компании сохранена!\n\nТеперь давайте создадим вакансию.",
	"uz": "✅ Kompaniya ma'lumotlari saqlandi!\n\nEndi vakansiya yaratamiz.",
}

var hrMsgCompanyAutoAdd = map[string]string{
	"en": "\n\nThis information will be automatically added to your vacancies.",
	"ru": "\n\nЭта информация будет автоматически добавляться в ваши вакансии.",
	"uz": "\n\nBu ma'lumot vakansiyalaringizga avtomatik qo'shiladi.",
}

var hrBtnCreateVacancyFromCompany = map[string]string{
	"en": "📝 Create vacancy",
	"ru": "📝 Создать вакансию",
	"uz": "📝 Vakansiya yaratish",
}

var hrBtnEditCompanyInfo = map[string]string{
	"en": "✏️ Add/Edit",
	"ru": "✏️ Добавить/Изменить",
	"uz": "✏️ Qo'shish/O'zgartirish",
}

var hrBtnAddLogo = map[string]string{
	"en": "🖼 Add logo",
	"ru": "🖼 Добавить логотип",
	"uz": "🖼 Logotip qo'shish",
}

var hrMsgSendLogo = map[string]string{
	"en": "Send your company logo as an image.",
	"ru": "Отправьте логотип вашей компании в виде изображения.",
	"uz": "Kompaniyangiz logotipini rasm sifatida yuboring.",
}

var hrMsgLogoSaved = map[string]string{
	"en": "📎 Logo saved",
	"ru": "📎 Логотип сохранён",
	"uz": "📎 Logotip saqlandi",
}

var hrMsgCompanyFailed = map[string]string{
	"en": "❌ Failed to process company information. Please try again.",
	"ru": "❌ Не удалось обработать информацию о компании. Попробуйте ещё раз.",
	"uz": "❌ Kompaniya ma'lumotlarini qayta ishlashda xatolik. Qaytadan urinib ko'ring.",
}

var hrMsgPostVacancy = map[string]string{
	"en": "Send the information in one message:\n\n— Who are you looking for (Position)\n— City\n— Work format\n— Salary range\n— Key responsibilities\n— Requirements\n— Level (Junior / Middle / Senior)\n\nYou can:\n\n✍️ write\n🎤 send a voice message\n📎 attach a file\n\n⚡ The more detail — the more accurate the match.",
	"ru": "Отправьте информацию одним сообщением:\n\n— Кого ищете (Должность)\n— Город\n— Формат работы\n— Зарплатная вилка\n— Основные обязанности\n— Требования\n— Уровень (Junior / Middle / Senior)\n\nВы можете:\n\n✍️ написать\n🎤 отправить голосовое\n📎 прикрепить файл\n\n⚡ Чем подробнее — тем точнее подбор.",
	"uz": "Ma'lumotlarni bitta xabarda yuboring:\n\n— Kimni qidiryapsiz (Lavozim)\n— Shahar\n— Ish formati\n— Maosh oralig'i\n— Asosiy vazifalar\n— Talablar\n— Daraja (Junior / Middle / Senior)\n\nSiz quyidagilarni yuborishingiz mumkin:\n\n✍️ yozish\n🎤 ovozli xabar yuborish\n📎 fayl biriktirish\n\n⚡ Qancha batafsil bo'lsa — tanlov shuncha aniq.",
}

var hrMsgParsingText = map[string]string{
	"en": "🔎 Got the information.\nCreating vacancy…\n\n⏳ Please wait...",
	"ru": "🔎 Получил информацию.\nСоздаю вакансию…\n\n⏳ Ожидайте...",
	"uz": "🔎 Ma'lumotni oldim.\nVakansiya yaratmoqdaman…\n\n⏳ Kuting...",
}

var hrMsgParsingVoice = map[string]string{
	"en": "I transcribed your voice message and I'm creating the vacancy.\n\n⏳ Please wait...",
	"ru": "Я расшифровал твоё голосовое сообщение и создаю вакансию.\n\n⏳ Ожидайте...",
	"uz": "Ovozli xabaringizni yozib oldim va vakansiya yaratmoqdaman.\n\n⏳ Kuting...",
}

var hrMsgParsingFile = map[string]string{
	"en": "I received your file and I'm creating the vacancy.\n\n⏳ Please wait...",
	"ru": "Я получил ваш файл и создаю вакансию.\n\n⏳ Ожидайте...",
	"uz": "Faylingizni oldim va vakansiya yaratmoqdaman.\n\n⏳ Kuting...",
}

var hrMsgVacancyCreated = map[string]string{
	"en": "✅ Vacancy created successfully!",
	"ru": "✅ Вакансия успешно создана!",
	"uz": "✅ Vakansiya muvaffaqiyatli yaratildi!",
}

var hrMsgVacancyFailed = map[string]string{
	"en": "❌ Failed to parse vacancy. Please try again.",
	"ru": "❌ Не удалось разобрать вакансию. Попробуйте ещё раз.",
	"uz": "❌ Vakansiyani tahlil qilib bo'lmadi. Qaytadan urinib ko'ring.",
}

var hrMsgMissing = map[string]string{
	"en": "Currently missing:",
	"ru": "Сейчас не хватает:",
	"uz": "Hozircha yetishmayapti:",
}

var hrMsgHowToContinue = map[string]string{
	"en": "How shall we continue? ⬇️",
	"ru": "Как продолжим? ⬇️",
	"uz": "Qanday davom etamiz? ⬇️",
}

var hrMsgBtnAddMissingInfo = map[string]string{
	"en": "✏️ Fill in missing information",
	"ru": "✏️ Заполнить недостающую информацию",
	"uz": "✏️ Yetishmayotgan ma'lumotlarni to'ldirish",
}

var hrMsgBtnContinueCurrent = map[string]string{
	"en": "▶️ Continue with current data",
	"ru": "▶️ Продолжить с текущими данными",
	"uz": "▶️ Joriy ma'lumotlar bilan davom etish",
}

var hrMsgBtnCreateDescription = map[string]string{
	"en": "✨ Create proper vacancy description",
	"ru": "✨ Создать правильное описание вакансии",
	"uz": "✨ To'g'ri vakansiya tavsifini yaratish",
}

var hrMsgEnhancing = map[string]string{
	"en": "✨ Creating a professional vacancy description...\n\nThis may take a moment.",
	"ru": "✨ Создаю профессиональное описание вакансии...\n\nЭто может занять немного времени.",
	"uz": "✨ Professional vakansiya tavsifi yaratilmoqda...\n\nBu biroz vaqt olishi mumkin.",
}

var hrMsgEnhanceFailed = map[string]string{
	"en": "❌ Failed to create description. Creating vacancy with current data instead.",
	"ru": "❌ Не удалось создать описание. Создаю вакансию с текущими данными.",
	"uz": "❌ Tavsif yaratib bo'lmadi. Vakansiya joriy ma'lumotlar bilan yaratilmoqda.",
}

var hrMsgBtnAddInfo = map[string]string{
	"en": "✏️ Add additional information",
	"ru": "✏️ Добавить дополнительную информацию",
	"uz": "✏️ Qo'shimcha ma'lumot qo'shish",
}

var hrMsgBtnConfirmCreate = map[string]string{
	"en": "✅ Create vacancy",
	"ru": "✅ Создать вакансию",
	"uz": "✅ Vakansiya yaratish",
}

var hrMsgBtnAddMoreInfo = map[string]string{
	"en": "✏️ Add or change information",
	"ru": "✏️ Добавить или изменить информацию",
	"uz": "✏️ Ma'lumot qo'shish yoki o'zgartirish",
}

var hrMsgSendAdditionalInfo = map[string]string{
	"en": "Send additional information about the vacancy.\n\nYou can:\n✍️ write\n🎤 send a voice message\n📎 attach a file",
	"ru": "Отправьте дополнительную информацию о вакансии.\n\nВы можете:\n✍️ написать\n🎤 отправить голосовое\n📎 прикрепить файл",
	"uz": "Vakansiya haqida qo'shimcha ma'lumot yuboring.\n\nSiz quyidagilarni yuborishingiz mumkin:\n✍️ yozish\n🎤 ovozli xabar yuborish\n📎 fayl biriktirish",
}

var hrMsgMaxAttemptsReached = map[string]string{
	"en": "⚠️ Maximum number of additions reached. The operation has been cancelled.\n\nYou can start again from the menu.",
	"ru": "⚠️ Достигнуто максимальное количество дополнений. Операция отменена.\n\nВы можете начать заново из меню.",
	"uz": "⚠️ Maksimal qo'shimchalar soniga yetildi. Amal bekor qilindi.\n\nMenyudan qaytadan boshlashingiz mumkin.",
}

var hrMsgUnsupportedFile = map[string]string{
	"en": "Unsupported file type. Please send a PDF, image (PNG/JPG), or text file.",
	"ru": "Неподдерживаемый тип файла. Пожалуйста, отправьте PDF, изображение (PNG/JPG) или текстовый файл.",
	"uz": "Qo'llab-quvvatlanmaydigan fayl turi. Iltimos, PDF, rasm (PNG/JPG) yoki matn faylini yuboring.",
}

var hrMsgDownloadFailed = map[string]string{
	"en": "Failed to download your file. Please try again.",
	"ru": "Не удалось загрузить файл. Пожалуйста, попробуйте ещё раз.",
	"uz": "Faylni yuklab bo'lmadi. Iltimos, qaytadan urinib ko'ring.",
}

var hrMsgFieldPosition = map[string]string{"en": "Position", "ru": "Должность", "uz": "Lavozim"}
var hrMsgFieldCity = map[string]string{"en": "City", "ru": "Город", "uz": "Shahar"}
var hrMsgFieldFormat = map[string]string{"en": "Format", "ru": "Формат", "uz": "Format"}
var hrMsgFieldStack = map[string]string{"en": "Stack", "ru": "Стек", "uz": "Stek"}
var hrMsgFieldExperience = map[string]string{"en": "Experience", "ru": "Опыт", "uz": "Tajriba"}
var hrMsgFieldSalary = map[string]string{"en": "Salary", "ru": "Зарплата", "uz": "Maosh"}
var hrMsgFieldUrgency = map[string]string{"en": "Urgency", "ru": "Срочность", "uz": "Shoshilinchlik"}

var hrMsgMissingSalary = map[string]string{"en": "Salary range", "ru": "Зарплатной вилки", "uz": "Maosh oralig'i"}
var hrMsgMissingFormat = map[string]string{"en": "Work format", "ru": "Формата работы", "uz": "Ish formati"}
var hrMsgMissingResponsibilities = map[string]string{"en": "Responsibilities", "ru": "Обязанностей", "uz": "Vazifalar"}
var hrMsgMissingRequirements = map[string]string{"en": "Requirements", "ru": "Требований", "uz": "Talablar"}
var hrMsgMissingExperience = map[string]string{"en": "Experience", "ru": "Опыта", "uz": "Tajriba"}

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
	bot        *tele.Bot
	hrBotSvc   *application.HRBotService
	storageSvc *application.StorageService
	webAppURL  string
}

func NewHRBot(token string, hrBotSvc *application.HRBotService, storageSvc *application.StorageService, webAppURL string) (*HRBot, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("create hr telegram bot: %w", err)
	}

	hb := &HRBot{bot: b, hrBotSvc: hrBotSvc, storageSvc: storageSvc, webAppURL: webAppURL}
	_ = b.RemoveWebhook()
	hb.registerHandlers()

	if webAppURL != "" {
		_, err = b.Raw("setChatMenuButton", map[string]interface{}{
			"menu_button": tele.MenuButton{
				Type:   tele.MenuButtonWebApp,
				Text:   "Open",
				WebApp: &tele.WebApp{URL: webAppURL},
			},
		})
		if err != nil {
			log.Printf("hr bot: failed to set menu button: %v", err)
		}
	}

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
		log.Printf("HR Bot: Received /start from %d (%s)", sender.ID, sender.Username)

		// Clear any leftover state from previous flow
		_ = hrBotSvc.ClearState(ctx, sender.ID)
		_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)

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
		return c.Send(hrMsgWelcomeNew[lang], hrMenu(lang))
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

	// Inline menu callback
	bot.Handle(&tele.Btn{Unique: "hr_menu"}, func(c tele.Context) error {
		sender := c.Sender()
		action := c.Callback().Data
		_ = c.Respond(&tele.CallbackResponse{})

		// Clear any previous state/draft when navigating via menu
		_ = hrBotSvc.ClearState(ctx, sender.ID)
		_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)

		tgID := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgID)
		if err != nil {
			return c.Send(hrMsgError["en"])
		}
		lang := langOrDefault(hr.Language)

		switch action {
		case "create_vacancy":
			// Check if company data is filled before allowing vacancy creation
			if !hrBotSvc.HasCompanyData(hr) {
				markup := &tele.ReplyMarkup{}
				markup.Inline(
					markup.Row(
						markup.Data(hrBtnAddCompanyInfo[lang], "hr_company_add"),
					),
					markup.Row(
						markup.Data(hrBtnSkipCompanyInfo[lang], "hr_company_skip"),
					),
				)
				return c.Send(hrMsgCompanyDataPrompt[lang], markup)
			}
			if err := hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStatePostingVacancy, map[string]string{"language": lang, "hr_id": hr.ID.String()}); err != nil {
				log.Printf("hr set posting state error for %d: %v", sender.ID, err)
			}
			return c.Send(hrMsgPostVacancy[lang])
		case "active_vacancies":
			return hb.handleMyVacancies(ctx, c, hr)
		case "find_candidates":
			return hb.handleFindCandidatesStart(ctx, c, hr)
		case "change_lang":
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
		return nil
	})

	// Company data prompt callbacks
	bot.Handle(&tele.Btn{Unique: "hr_company_add"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)
		tgID := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgID)
		if err != nil {
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		if err := hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStateCollectingCompanyData, map[string]string{"language": lang, "hr_id": hr.ID.String()}); err != nil {
			log.Printf("hr set company data state error for %d: %v", sender.ID, err)
		}
		return c.Send(hrMsgCompanyDataStep[lang])
	})

	bot.Handle(&tele.Btn{Unique: "hr_company_skip"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)
		tgID := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgID)
		if err != nil {
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		if err := hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStatePostingVacancy, map[string]string{"language": lang, "hr_id": hr.ID.String()}); err != nil {
			log.Printf("hr set posting state error for %d: %v", sender.ID, err)
		}
		_ = hr // ensure hr is fetched for state data
		return c.Send(hrMsgPostVacancy[lang])
	})

	// Company profile review callbacks
	bot.Handle(&tele.Btn{Unique: "hr_company_create_vacancy"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)
		tgID := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgID)
		if err != nil {
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		// Save company draft to Postgres
		draft, err := hrBotSvc.GetCompanyDraft(ctx, sender.ID)
		if err != nil || draft == nil {
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}
		if _, err := hrBotSvc.SaveCompanyDataFromDraft(ctx, hr.ID, draft); err != nil {
			log.Printf("hr save company data error for %d: %v", sender.ID, err)
			return c.Send(hrMsgCompanyFailed[lang], hrMenu(lang))
		}
		_ = hrBotSvc.ClearCompanyDraft(ctx, sender.ID)

		// Transition to vacancy posting
		if err := hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStatePostingVacancy, map[string]string{"language": lang, "hr_id": hr.ID.String()}); err != nil {
			log.Printf("hr set posting state error for %d: %v", sender.ID, err)
		}
		return c.Send(hrMsgPostVacancy[lang])
	})

	bot.Handle(&tele.Btn{Unique: "hr_company_edit"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)
		tgID := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgID)
		if err != nil {
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		if err := hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStateCollectingCompanyData, map[string]string{"language": lang, "hr_id": hr.ID.String()}); err != nil {
			log.Printf("hr set company data state error for %d: %v", sender.ID, err)
		}
		return c.Send(hrMsgCompanyDataStep[lang])
	})

	bot.Handle(&tele.Btn{Unique: "hr_company_logo"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)
		tgID := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgID)
		if err != nil {
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		if err := hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStateCollectingCompanyLogo, map[string]string{"language": lang, "hr_id": hr.ID.String()}); err != nil {
			log.Printf("hr set company logo state error for %d: %v", sender.ID, err)
		}
		return c.Send(hrMsgSendLogo[lang])
	})

	// Vacancy review callbacks
	bot.Handle(&tele.Btn{Unique: "hr_vac_continue"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)

		draft, err := hrBotSvc.GetVacancyDraft(ctx, sender.ID)
		if err != nil || draft == nil {
			_ = hrBotSvc.ClearState(ctx, sender.ID)
			_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		return hb.sendVacancyConfirmation(c, draft, lang)
	})

	bot.Handle(&tele.Btn{Unique: "hr_vac_create_desc"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)

		draft, err := hrBotSvc.GetVacancyDraft(ctx, sender.ID)
		if err != nil || draft == nil {
			_ = hrBotSvc.ClearState(ctx, sender.ID)
			_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		// Send "enhancing..." feedback
		waitMsg, _ := c.Bot().Send(c.Recipient(), hrMsgEnhancing[lang])

		// Enhance via Gemini
		enhanced, err := hrBotSvc.EnhanceVacancyDraft(ctx, draft)
		if waitMsg != nil {
			_ = c.Bot().Delete(waitMsg)
		}
		if err != nil {
			log.Printf("hr_vac_create_desc: enhance failed for %d: %v", sender.ID, err)
			// Fall back to original draft
			enhanced = draft
			_ = c.Send(hrMsgEnhanceFailed[lang])
		}

		// Save enhanced draft to Redis and show confirmation
		_ = hrBotSvc.SaveVacancyDraft(ctx, sender.ID, enhanced)

		return hb.sendVacancyConfirmation(c, enhanced, lang)
	})

	bot.Handle(&tele.Btn{Unique: "hr_vac_confirm"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)

		draft, err := hrBotSvc.GetVacancyDraft(ctx, sender.ID)
		if err != nil || draft == nil {
			_ = hrBotSvc.ClearState(ctx, sender.ID)
			_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		tgIDStr := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgIDStr)
		if err != nil || hr == nil {
			log.Printf("hr_vac_confirm: failed to get HR for tg_id %d: %v", sender.ID, err)
			_ = hrBotSvc.ClearState(ctx, sender.ID)
			_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		result, err := hrBotSvc.CreateVacancyFromDraft(ctx, hr.ID, hr.CompanyData, draft)
		if err != nil {
			log.Printf("hr_vac_confirm: create vacancy error for %d: %v", sender.ID, err)
			_ = hrBotSvc.ClearState(ctx, sender.ID)
			_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)
			return c.Send(hrMsgVacancyFailed[lang], hrMenu(lang))
		}

		_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)
		_ = hrBotSvc.ClearState(ctx, sender.ID)

		// Count matching candidates
		var skills []string
		for _, sk := range result.Skills {
			skills = append(skills, sk.Name)
		}
		matchCount := hrBotSvc.CountMatchingCandidates(ctx, vacancyTitle(result, "en"), skills)

		msg := buildVacancyCreatedMessage(result, lang, matchCount)
		return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: vacancyPublishedMenu(lang, result.Vacancy.ID.String(), hb.webAppURL)})
	})

	bot.Handle(&tele.Btn{Unique: "hr_vac_add_info"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)

		state, _ := hrBotSvc.GetBotState(ctx, sender.ID)
		attempts := 0
		hrIDStr := ""
		if state != nil {
			hrIDStr = state.Data["hr_id"]
			if a, err := strconv.Atoi(state.Data["add_attempts"]); err == nil {
				attempts = a
			}
		}

		if attempts >= 3 {
			_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)
			_ = hrBotSvc.ClearState(ctx, sender.ID)
			return c.Send(hrMsgMaxAttemptsReached[lang], hrMenu(lang))
		}

		if err := hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStateAddingVacancyInfo, map[string]string{
			"language":     lang,
			"hr_id":        hrIDStr,
			"add_attempts": strconv.Itoa(attempts + 1),
		}); err != nil {
			log.Printf("hr set adding info state error for %d: %v", sender.ID, err)
		}

		// Show current draft content + missing fields, then ask for additional info
		draft, _ := hrBotSvc.GetVacancyDraft(ctx, sender.ID)
		if draft != nil {
			_ = c.Send(hb.buildAddInfoMessage(draft, lang), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		} else {
			_ = c.Send(hrMsgSendAdditionalInfo[lang])
		}
		return nil
	})

	// -- Vacancy published action buttons (placeholders) --
	bot.Handle(&tele.Btn{Unique: "hr_pub_candidates"}, func(c tele.Context) error {
		// Only fires when webAppURL is not configured (WebApp buttons don't trigger callbacks)
		_ = c.Respond(&tele.CallbackResponse{})
		return nil
	})
	bot.Handle(&tele.Btn{Unique: "hr_pub_view"}, func(c tele.Context) error {
		_ = c.Respond(&tele.CallbackResponse{Text: "Coming soon..."})
		return nil
	})
	bot.Handle(&tele.Btn{Unique: "hr_pub_edit"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)
		vacancyID := c.Callback().Data

		tgID := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgID)
		if err != nil {
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		if err := hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStateEditingPublishedVacancy, map[string]string{
			"language":   lang,
			"hr_id":      hr.ID.String(),
			"vacancy_id": vacancyID,
		}); err != nil {
			log.Printf("hr set editing published vacancy state error for %d: %v", sender.ID, err)
		}

		return c.Send(hrMsgSendAdditionalInfo[lang])
	})
	bot.Handle(&tele.Btn{Unique: "hr_pub_stop"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)
		vacancyID, err := uuid.Parse(c.Callback().Data)
		if err != nil {
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		if err := hrBotSvc.StopVacancy(ctx, vacancyID); err != nil {
			log.Printf("hr stop vacancy error for %d: %v", sender.ID, err)
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		shortID := vacancyID.String()[:8]
		msg := buildVacancyStoppedMessage(shortID, lang)

		markup := &tele.ReplyMarkup{}
		markup.Inline(
			markup.Row(markup.Data(hrBtnActivateVacancy[lang], "hr_pub_activate", vacancyID.String())),
		)
		return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: markup})
	})

	bot.Handle(&tele.Btn{Unique: "hr_pub_activate"}, func(c tele.Context) error {
		sender := c.Sender()
		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := hb.resolveHRLang(ctx, sender)
		vacancyID, err := uuid.Parse(c.Callback().Data)
		if err != nil {
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		if err := hrBotSvc.ActivateVacancy(ctx, vacancyID); err != nil {
			log.Printf("hr activate vacancy error for %d: %v", sender.ID, err)
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		result, err := hrBotSvc.GetVacancy(ctx, vacancyID)
		if err != nil {
			return c.Send(hrMsgError[lang], hrMenu(lang))
		}

		var skills []string
		for _, sk := range result.Skills {
			skills = append(skills, sk.Name)
		}
		matchCount := hrBotSvc.CountMatchingCandidates(ctx, vacancyTitle(result, "en"), skills)

		msg := buildVacancyCreatedMessage(result, lang, matchCount)
		return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: vacancyPublishedMenu(lang, vacancyID.String(), hb.webAppURL)})
	})

	// Text message handler
	bot.Handle(tele.OnText, func(c tele.Context) error {
		sender := c.Sender()
		text := strings.TrimSpace(c.Text())

		// 1. Check if the text is a menu button first
		isMenu := isMenuButton(text, hrMenuBtnCreateVacancy) ||
			isMenuButton(text, hrMenuBtnActiveVacancies) ||
			isMenuButton(text, hrMenuBtnFindCandidates) ||
			isMenuButton(text, hrMenuBtnChangeLang)

		if isMenu {
			// If it's a menu button, clear any active state
			_ = hrBotSvc.ClearState(ctx, sender.ID)
			_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)
		} else {
			// 2. If it's not a menu button, handle active states
			state, _ := hrBotSvc.GetBotState(ctx, sender.ID)
			if state != nil {
				switch state.State {
				case domain.HRBotStateSharingPhone:
					lang := langOrDefault(state.Data["language"])
					return c.Send(hrMsgPhoneReminder[lang])
				case domain.HRBotStateCollectingCompanyData:
					return hb.handleCompanyDataInput(ctx, c, state, "text", text, nil, "")
				case domain.HRBotStatePostingVacancy:
					return hb.handleVacancyInput(ctx, c, state, "text", text, nil, "")
				case domain.HRBotStateAddingVacancyInfo:
					return hb.handleVacancyInput(ctx, c, state, "text", text, nil, "")
				case domain.HRBotStateEditingPublishedVacancy:
					return hb.handleEditPublishedVacancy(ctx, c, state, "text", text, nil, "")
				case "hr_searching":
					return hb.handleSearchQuery(ctx, c, state)
				}
			}
		}

		// 3. Process menu command (or handle unknown text)
		tgID := strconv.FormatInt(sender.ID, 10)
		hr, err := hrBotSvc.GetHRByTelegramID(ctx, tgID)
		if err != nil {
			return c.Send(hrMsgError["en"])
		}
		lang := langOrDefault(hr.Language)

		if isMenuButton(text, hrMenuBtnCreateVacancy) {
			_ = hrBotSvc.ClearVacancyDraft(ctx, sender.ID)
			// Check if company data is filled before allowing vacancy creation
			if !hrBotSvc.HasCompanyData(hr) {
				markup := &tele.ReplyMarkup{}
				markup.Inline(
					markup.Row(
						markup.Data(hrBtnAddCompanyInfo[lang], "hr_company_add"),
					),
					markup.Row(
						markup.Data(hrBtnSkipCompanyInfo[lang], "hr_company_skip"),
					),
				)
				return c.Send(hrMsgCompanyDataPrompt[lang], markup)
			}
			if err := hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStatePostingVacancy, map[string]string{"language": lang, "hr_id": hr.ID.String()}); err != nil {
				log.Printf("hr set posting state error for %d: %v", sender.ID, err)
			}
			return c.Send(hrMsgPostVacancy[lang])
		}
		if isMenuButton(text, hrMenuBtnActiveVacancies) {
			return hb.handleMyVacancies(ctx, c, hr)
		}
		if isMenuButton(text, hrMenuBtnFindCandidates) {
			if hb.webAppURL != "" {
				markup := &tele.ReplyMarkup{}
				markup.Inline(
					markup.Row(markup.WebApp(hrMenuBtnFindCandidates[lang], &tele.WebApp{URL: hb.webAppURL})),
				)
				return c.Send("🔍", markup)
			}
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

	// Voice message handler for HR bot
	bot.Handle(tele.OnVoice, func(c tele.Context) error {
		sender := c.Sender()
		voice := c.Message().Voice

		state, _ := hrBotSvc.GetBotState(ctx, sender.ID)
		if state != nil {
			if state.State == domain.HRBotStatePostingVacancy || state.State == domain.HRBotStateAddingVacancyInfo || state.State == domain.HRBotStateCollectingCompanyData || state.State == domain.HRBotStateEditingPublishedVacancy {
				reader, err := bot.File(&voice.File)
				if err != nil {
					lang := langOrDefault(state.Data["language"])
					return c.Send(hrMsgDownloadFailed[lang])
				}
				fileData, _ := io.ReadAll(reader)
				reader.Close()
				mimeType := voice.MIME
				if mimeType == "" {
					mimeType = "audio/ogg"
				}
				if state.State == domain.HRBotStateCollectingCompanyData {
					return hb.handleCompanyDataInput(ctx, c, state, "voice", "", fileData, mimeType)
				}
				if state.State == domain.HRBotStateEditingPublishedVacancy {
					return hb.handleEditPublishedVacancy(ctx, c, state, "voice", "", fileData, mimeType)
				}
				return hb.handleVacancyInput(ctx, c, state, "voice", "", fileData, mimeType)
			}
		}

		lang := hb.resolveHRLang(ctx, sender)
		return c.Send(hrMsgUseMenu[lang], hrMenu(lang))
	})

	// Document handler for HR bot
	bot.Handle(tele.OnDocument, func(c tele.Context) error {
		sender := c.Sender()
		doc := c.Message().Document

		state, _ := hrBotSvc.GetBotState(ctx, sender.ID)
		if state != nil {
			if state.State == domain.HRBotStatePostingVacancy || state.State == domain.HRBotStateAddingVacancyInfo || state.State == domain.HRBotStateCollectingCompanyData || state.State == domain.HRBotStateEditingPublishedVacancy {
				lang := langOrDefault(state.Data["language"])
				if !isAllowedMIME(doc.MIME) {
					return c.Send(hrMsgUnsupportedFile[lang])
				}
				reader, err := bot.File(&doc.File)
				if err != nil {
					return c.Send(hrMsgDownloadFailed[lang])
				}
				fileData, _ := io.ReadAll(reader)
				reader.Close()
				if state.State == domain.HRBotStateCollectingCompanyData {
					return hb.handleCompanyDataInput(ctx, c, state, "file", "", fileData, doc.MIME)
				}
				if state.State == domain.HRBotStateEditingPublishedVacancy {
					return hb.handleEditPublishedVacancy(ctx, c, state, "file", "", fileData, doc.MIME)
				}
				return hb.handleVacancyInput(ctx, c, state, "file", "", fileData, doc.MIME)
			}
		}

		lang := hb.resolveHRLang(ctx, sender)
		return c.Send(hrMsgUseMenu[lang], hrMenu(lang))
	})

	// Photo handler for HR bot
	bot.Handle(tele.OnPhoto, func(c tele.Context) error {
		sender := c.Sender()
		photo := c.Message().Photo

		state, _ := hrBotSvc.GetBotState(ctx, sender.ID)
		if state == nil {
			lang := hb.resolveHRLang(ctx, sender)
			return c.Send(hrMsgUseMenu[lang], hrMenu(lang))
		}

		switch state.State {
		case domain.HRBotStateCollectingCompanyData:
			reader, err := bot.File(&photo.File)
			if err != nil {
				return c.Send(hrMsgDownloadFailed[langOrDefault(state.Data["language"])])
			}
			fileData, _ := io.ReadAll(reader)
			reader.Close()
			return hb.handleCompanyDataInput(ctx, c, state, "file", "", fileData, "image/jpeg")

		case domain.HRBotStateCollectingCompanyLogo:
			return hb.handleCompanyLogoUpload(ctx, c, state, photo)

		case domain.HRBotStateEditingPublishedVacancy:
			reader, err := bot.File(&photo.File)
			if err != nil {
				return c.Send(hrMsgDownloadFailed[langOrDefault(state.Data["language"])])
			}
			fileData, _ := io.ReadAll(reader)
			reader.Close()
			return hb.handleEditPublishedVacancy(ctx, c, state, "file", "", fileData, "image/jpeg")

		case domain.HRBotStatePostingVacancy, domain.HRBotStateAddingVacancyInfo:
			reader, err := bot.File(&photo.File)
			if err != nil {
				return c.Send(hrMsgDownloadFailed[langOrDefault(state.Data["language"])])
			}
			fileData, _ := io.ReadAll(reader)
			reader.Close()
			return hb.handleVacancyInput(ctx, c, state, "file", "", fileData, "image/jpeg")

		default:
			lang := hb.resolveHRLang(ctx, sender)
			return c.Send(hrMsgUseMenu[lang], hrMenu(lang))
		}
	})
}

func (hb *HRBot) handleCompanyDataInput(ctx context.Context, c tele.Context, state *domain.BotState, inputType, text string, fileData []byte, mimeType string) error {
	sender := c.Sender()
	lang := langOrDefault(state.Data["language"])
	hrIDStr := state.Data["hr_id"]

	// Send appropriate "processing" message
	waitMsg, _ := c.Bot().Send(c.Recipient(), hrMsgParsingCompany[lang])

	var parsed *gemini.ParsedCompanyFull
	var err error

	// If we already have a draft, merge additional info via Gemini
	existingDraft, _ := hb.hrBotSvc.GetCompanyDraft(ctx, sender.ID)

	if inputType == "text" {
		parsed, err = hb.hrBotSvc.ParseCompanyFromText(ctx, text)
	} else if len(fileData) > 0 {
		parsed, err = hb.hrBotSvc.ParseCompanyFromFile(ctx, fileData, mimeType)
	}

	if waitMsg != nil {
		_ = c.Bot().Delete(waitMsg)
	}

	if err != nil {
		log.Printf("hr parse company data error for %d: %v", sender.ID, err)
		_ = hb.hrBotSvc.ClearState(ctx, sender.ID)
		return c.Send(hrMsgCompanyFailed[lang], hrMenu(lang))
	}

	// Convert to domain.CompanyData
	draft := hb.hrBotSvc.ConvertParsedToCompanyData(parsed)

	// Preserve logo from existing draft if present
	if existingDraft != nil && existingDraft.LogoURL != "" {
		draft.LogoURL = existingDraft.LogoURL
	}

	// Save draft to Redis (not Postgres)
	if err := hb.hrBotSvc.SaveCompanyDraft(ctx, sender.ID, draft); err != nil {
		log.Printf("hr save company draft error for %d: %v", sender.ID, err)
		_ = hb.hrBotSvc.ClearState(ctx, sender.ID)
		return c.Send(hrMsgCompanyFailed[lang], hrMenu(lang))
	}

	// Set state to company review
	if err := hb.hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStateCompanyReview, map[string]string{"language": lang, "hr_id": hrIDStr}); err != nil {
		log.Printf("hr set company review state error for %d: %v", sender.ID, err)
	}

	return hb.sendCompanyProfile(c, draft, lang)
}

func (hb *HRBot) handleCompanyLogoUpload(ctx context.Context, c tele.Context, state *domain.BotState, photo *tele.Photo) error {
	sender := c.Sender()
	lang := langOrDefault(state.Data["language"])
	hrIDStr := state.Data["hr_id"]

	// Download photo from Telegram
	reader, err := hb.bot.File(&photo.File)
	if err != nil {
		return c.Send(hrMsgDownloadFailed[lang])
	}
	fileData, _ := io.ReadAll(reader)
	reader.Close()

	// Upload to MinIO
	result, err := hb.storageSvc.UploadCompanyLogo(ctx, fileData, "image/jpeg")
	if err != nil {
		log.Printf("hr upload company logo error for %d: %v", sender.ID, err)
		return c.Send(hrMsgCompanyFailed[lang])
	}

	// Update draft in Redis with logo URL
	draft, err := hb.hrBotSvc.GetCompanyDraft(ctx, sender.ID)
	if err != nil || draft == nil {
		return c.Send(hrMsgError[lang], hrMenu(lang))
	}
	draft.LogoURL = result.URL
	if err := hb.hrBotSvc.SaveCompanyDraft(ctx, sender.ID, draft); err != nil {
		log.Printf("hr save company draft with logo error for %d: %v", sender.ID, err)
	}

	// Save to Postgres
	hrID, parseErr := uuid.Parse(hrIDStr)
	if parseErr != nil {
		return c.Send(hrMsgError[lang], hrMenu(lang))
	}
	if _, err := hb.hrBotSvc.SaveCompanyDataFromDraft(ctx, hrID, draft); err != nil {
		log.Printf("hr save company data after logo error for %d: %v", sender.ID, err)
	}

	// Set state back to company review
	if err := hb.hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStateCompanyReview, map[string]string{"language": lang, "hr_id": hrIDStr}); err != nil {
		log.Printf("hr set company review state error for %d: %v", sender.ID, err)
	}

	return hb.sendCompanyProfile(c, draft, lang)
}

func (hb *HRBot) sendCompanyProfile(c tele.Context, draft *domain.CompanyData, lang string) error {
	msg := formatCompanyProfile(draft, lang)

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data(hrBtnCreateVacancyFromCompany[lang], "hr_company_create_vacancy")),
		markup.Row(markup.Data(hrBtnEditCompanyInfo[lang], "hr_company_edit")),
		markup.Row(markup.Data(hrBtnAddLogo[lang], "hr_company_logo")),
	)

	if draft.LogoURL != "" {
		photo := &tele.Photo{File: tele.FromURL(draft.LogoURL), Caption: msg}
		return c.Send(photo, markup)
	}
	return c.Send(msg, markup)
}

func formatCompanyProfile(draft *domain.CompanyData, lang string) string {
	// Find the text for the requested language
	var t *domain.CompanyDataText
	for i := range draft.Texts {
		if draft.Texts[i].Lang == lang {
			t = &draft.Texts[i]
			break
		}
	}
	if t == nil && len(draft.Texts) > 0 {
		t = &draft.Texts[0]
	}

	var b strings.Builder
	b.WriteString("✅ ")
	switch lang {
	case "ru":
		b.WriteString("Профиль компании создан.")
	case "uz":
		b.WriteString("Kompaniya profili yaratildi.")
	default:
		b.WriteString("Company profile created.")
	}

	if t != nil {
		if t.Name != "" {
			b.WriteString("\n\n")
			switch lang {
			case "ru":
				b.WriteString("Название: ")
			case "uz":
				b.WriteString("Nomi: ")
			default:
				b.WriteString("Name: ")
			}
			b.WriteString(t.Name)
		}
		if t.ActivityType != "" {
			b.WriteString("\n")
			switch lang {
			case "ru":
				b.WriteString("Сфера: ")
			case "uz":
				b.WriteString("Soha: ")
			default:
				b.WriteString("Field: ")
			}
			b.WriteString(t.ActivityType)
		}
	}

	if draft.Address != "" {
		b.WriteString("\n")
		switch lang {
		case "ru":
			b.WriteString("Город: ")
		case "uz":
			b.WriteString("Shahar: ")
		default:
			b.WriteString("City: ")
		}
		b.WriteString(draft.Address)
	}

	if t != nil && t.About != "" {
		b.WriteString("\n")
		switch lang {
		case "ru":
			b.WriteString("О компании: ")
		case "uz":
			b.WriteString("Kompaniya haqida: ")
		default:
			b.WriteString("About: ")
		}
		b.WriteString(t.About)
	}

	if draft.EmployeeCount > 0 {
		b.WriteString(fmt.Sprintf("\n— %d ", draft.EmployeeCount))
		switch lang {
		case "ru":
			b.WriteString("сотрудников")
		case "uz":
			b.WriteString("xodim")
		default:
			b.WriteString("employees")
		}
	}

	if draft.LogoURL != "" {
		b.WriteString("\n📎 ")
		b.WriteString(hrMsgLogoSaved[lang])
	}

	b.WriteString(hrMsgCompanyAutoAdd[lang])

	return b.String()
}

func (hb *HRBot) handleVacancyInput(ctx context.Context, c tele.Context, state *domain.BotState, inputType, text string, fileData []byte, mimeType string) error {
	sender := c.Sender()
	lang := langOrDefault(state.Data["language"])
	hrIDStr := state.Data["hr_id"]
	addAttempts := state.Data["add_attempts"]
	isAdding := state.State == domain.HRBotStateAddingVacancyInfo

	// Send appropriate "processing" message
	var waitMsg *tele.Message
	switch inputType {
	case "text":
		waitMsg, _ = c.Bot().Send(c.Recipient(), hrMsgParsingText[lang])
	case "voice":
		waitMsg, _ = c.Bot().Send(c.Recipient(), hrMsgParsingVoice[lang])
	case "file":
		waitMsg, _ = c.Bot().Send(c.Recipient(), hrMsgParsingFile[lang])
	}

	var parsed *gemini.ParsedVacancyFull
	var err error

	if isAdding {
		// Merge additional info with existing draft
		draft, _ := hb.hrBotSvc.GetVacancyDraft(ctx, sender.ID)
		if draft != nil {
			existingBytes, _ := json.Marshal(draft)
			existingJSON := string(existingBytes)

			var newInfo string
			if inputType == "text" {
				newInfo = text
			} else if len(fileData) > 0 {
				// For voice/file, first parse the file, then merge
				fileParsed, fileErr := hb.hrBotSvc.ParseVacancyFromFile(ctx, fileData, mimeType)
				if fileErr == nil {
					b, _ := json.Marshal(fileParsed)
					newInfo = string(b)
				}
			}

			if newInfo != "" {
				parsed, err = hb.hrBotSvc.MergeVacancy(ctx, existingJSON, newInfo)
			}
		}
	} else {
		// First-time parse
		if inputType == "text" {
			parsed, err = hb.hrBotSvc.ParseVacancyFromText(ctx, text)
		} else if len(fileData) > 0 {
			parsed, err = hb.hrBotSvc.ParseVacancyFromFile(ctx, fileData, mimeType)
		}
	}

	if waitMsg != nil {
		_ = c.Bot().Delete(waitMsg)
	}

	if err != nil {
		log.Printf("hr parse vacancy error for %d: %v", sender.ID, err)
		return c.Send(hrMsgVacancyFailed[lang], hrMenu(lang))
	}

	// Save draft to Redis
	if err := hb.hrBotSvc.SaveVacancyDraft(ctx, sender.ID, parsed); err != nil {
		log.Printf("hr save vacancy draft error for %d: %v", sender.ID, err)
	}

	// Set review state
	if err := hb.hrBotSvc.SetState(ctx, sender.ID, domain.HRBotStateVacancyReview, map[string]string{
		"language":     lang,
		"hr_id":        hrIDStr,
		"add_attempts": addAttempts,
	}); err != nil {
		log.Printf("hr set review state error for %d: %v", sender.ID, err)
	}

	// Format and send the vacancy summary + missing fields + action buttons
	if parsed == nil {
		log.Printf("hr parse vacancy result is nil for %d", sender.ID)
		return c.Send(hrMsgVacancyFailed[lang], hrMenu(lang))
	}

	return hb.sendVacancyReview(c, parsed, lang)
}

func (hb *HRBot) handleEditPublishedVacancy(ctx context.Context, c tele.Context, state *domain.BotState, inputType, text string, fileData []byte, mimeType string) error {
	sender := c.Sender()
	lang := langOrDefault(state.Data["language"])
	vacancyIDStr := state.Data["vacancy_id"]

	vacancyID, err := uuid.Parse(vacancyIDStr)
	if err != nil {
		_ = hb.hrBotSvc.ClearState(ctx, sender.ID)
		return c.Send(hrMsgError[lang], hrMenu(lang))
	}

	// Send processing message
	switch inputType {
	case "text":
		_ = c.Send(hrMsgParsingText[lang])
	case "voice":
		_ = c.Send(hrMsgParsingVoice[lang])
	default:
		_ = c.Send(hrMsgParsingFile[lang])
	}

	// Load existing vacancy, convert to JSON for merge
	existing, err := hb.hrBotSvc.GetVacancy(ctx, vacancyID)
	if err != nil {
		_ = hb.hrBotSvc.ClearState(ctx, sender.ID)
		return c.Send(hrMsgError[lang], hrMenu(lang))
	}
	existingJSON := vacancyToMergeJSON(existing, lang)

	// Parse user input
	var newInfo string
	if inputType == "text" {
		newInfo = text
	} else if len(fileData) > 0 {
		fileParsed, fileErr := hb.hrBotSvc.ParseVacancyFromFile(ctx, fileData, mimeType)
		if fileErr == nil {
			b, _ := json.Marshal(fileParsed)
			newInfo = string(b)
		}
	}

	if newInfo == "" {
		_ = hb.hrBotSvc.ClearState(ctx, sender.ID)
		return c.Send(hrMsgVacancyFailed[lang], hrMenu(lang))
	}

	// Merge via Gemini
	merged, err := hb.hrBotSvc.MergeVacancy(ctx, existingJSON, newInfo)
	if err != nil {
		log.Printf("hr merge published vacancy error for %d: %v", sender.ID, err)
		_ = hb.hrBotSvc.ClearState(ctx, sender.ID)
		return c.Send(hrMsgVacancyFailed[lang], hrMenu(lang))
	}

	// Update in Postgres
	result, err := hb.hrBotSvc.UpdatePublishedVacancy(ctx, vacancyID, merged)
	if err != nil {
		log.Printf("hr update published vacancy error for %d: %v", sender.ID, err)
		_ = hb.hrBotSvc.ClearState(ctx, sender.ID)
		return c.Send(hrMsgVacancyFailed[lang], hrMenu(lang))
	}

	_ = hb.hrBotSvc.ClearState(ctx, sender.ID)

	// Show published message with 4 buttons
	var skills []string
	for _, sk := range result.Skills {
		skills = append(skills, sk.Name)
	}
	matchCount := hb.hrBotSvc.CountMatchingCandidates(ctx, vacancyTitle(result, "en"), skills)

	updatedMsg := map[string]string{
		"en": "✅ Vacancy updated!\n\n",
		"ru": "✅ Вакансия обновлена!\n\n",
		"uz": "✅ Vakansiya yangilandi!\n\n",
	}

	msg := updatedMsg[lang] + buildVacancyCreatedMessage(result, lang, matchCount)
	return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: vacancyPublishedMenu(lang, vacancyID.String(), hb.webAppURL)})
}

// vacancyToMergeJSON converts existing vacancy data to JSON for Gemini merge.
func vacancyToMergeJSON(v *application.VacancyWithDetails, lang string) string {
	data := map[string]any{
		"format":          v.Vacancy.Format,
		"schedule":        v.Vacancy.Schedule,
		"salary_min":      v.Vacancy.SalaryMin,
		"salary_max":      v.Vacancy.SalaryMax,
		"salary_currency": v.Vacancy.SalaryCurrency,
		"experience_min":  v.Vacancy.ExperienceMin,
		"experience_max":  v.Vacancy.ExperienceMax,
		"address":         v.Vacancy.Address,
		"phone":           v.Vacancy.Phone,
		"telegram":        v.Vacancy.Telegram,
		"email":           v.Vacancy.Email,
	}
	var skillNames []string
	for _, sk := range v.Skills {
		skillNames = append(skillNames, sk.Name)
	}
	data["skills"] = skillNames

	fields := map[string]map[string]string{}
	for _, t := range v.Texts {
		for _, field := range []struct{ key, val string }{
			{"title", t.Title},
			{"description", t.Description},
			{"responsibilities", t.Responsibilities},
			{"requirements", t.Requirements},
			{"benefits", t.Benefits},
		} {
			if field.val != "" {
				if fields[field.key] == nil {
					fields[field.key] = map[string]string{}
				}
				fields[field.key][t.Lang] = field.val
			}
		}
	}
	data["fields"] = fields

	b, _ := json.Marshal(data)
	return string(b)
}

func (hb *HRBot) sendVacancyReview(c tele.Context, draft *gemini.ParsedVacancyFull, lang string) error {
	var sb strings.Builder

	// Header
	yourVacancy := map[string]string{"en": "Your vacancy", "ru": "Ваша вакансия", "uz": "Vakansiyangiz"}
	sb.WriteString(fmt.Sprintf("*%s*\n\n", yourVacancy[lang]))

	// Title (bold)
	title := draftField(draft, "title", lang)
	if title != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", title))
	}

	// Description / About the company
	desc := draftField(draft, "description", lang)
	if desc != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", desc))
	}

	// Responsibilities
	resp := draftField(draft, "responsibilities", lang)
	if resp != "" {
		respLabel := map[string]string{"en": "Responsibilities:", "ru": "Обязанности:", "uz": "Vazifalar:"}
		sb.WriteString(fmt.Sprintf("*%s*\n%s\n\n", respLabel[lang], formatAsBullets(resp)))
	}

	// Requirements
	req := draftField(draft, "requirements", lang)
	if req != "" {
		reqLabel := map[string]string{"en": "Requirements:", "ru": "Требования:", "uz": "Talablar:"}
		sb.WriteString(fmt.Sprintf("*%s*\n%s\n\n", reqLabel[lang], formatAsBullets(req)))
	}

	// Benefits / Conditions
	ben := draftField(draft, "benefits", lang)
	if ben != "" {
		benLabel := map[string]string{"en": "We offer:", "ru": "Условия:", "uz": "Sharoitlar:"}
		sb.WriteString(fmt.Sprintf("*%s*\n%s\n\n", benLabel[lang], formatAsBullets(ben)))
	}

	// Details line: format, address, experience, salary
	if draft.Format != "" {
		formatLabel := map[string]string{"en": "Work format: ", "ru": "Формат работы: ", "uz": "Ish formati: "}
		formatNames := map[string]map[string]string{
			"office": {"en": "Office", "ru": "Офис", "uz": "Ofis"},
			"remote": {"en": "Remote", "ru": "Удалёнка", "uz": "Masofaviy"},
			"hybrid": {"en": "Hybrid", "ru": "Гибрид", "uz": "Gibrid"},
		}
		if names, ok := formatNames[draft.Format]; ok {
			sb.WriteString(fmt.Sprintf("%s%s\n", formatLabel[lang], names[lang]))
		}
	}
	if draft.Address != "" {
		addrLabel := map[string]string{"en": "Location: ", "ru": "Адрес: ", "uz": "Manzil: "}
		sb.WriteString(fmt.Sprintf("%s%s\n", addrLabel[lang], draft.Address))
	}
	if draft.SalaryMin > 0 || draft.SalaryMax > 0 {
		salaryLabel := map[string]string{"en": "Salary: ", "ru": "Зарплата: ", "uz": "Maosh: "}
		minStr := formatNumber(int64(draft.SalaryMin))
		maxStr := formatNumber(int64(draft.SalaryMax))
		sb.WriteString(fmt.Sprintf("💸%s%s – %s %s\n", salaryLabel[lang], minStr, maxStr, draft.SalaryCurrency))
	}
	if draft.ExperienceMin > 0 || draft.ExperienceMax > 0 {
		expYears := map[string]string{"en": "years", "ru": "лет", "uz": "yil"}
		expLabel := map[string]string{"en": "Experience: ", "ru": "Опыт: ", "uz": "Tajriba: "}
		if draft.ExperienceMin > 0 && draft.ExperienceMax > 0 {
			sb.WriteString(fmt.Sprintf("%s%d–%d %s\n", expLabel[lang], draft.ExperienceMin, draft.ExperienceMax, expYears[lang]))
		} else if draft.ExperienceMin > 0 {
			sb.WriteString(fmt.Sprintf("%s%d+ %s\n", expLabel[lang], draft.ExperienceMin, expYears[lang]))
		}
	}

	if len(draft.Skills) > 0 {
		sb.WriteString(fmt.Sprintf("\n🛠 %s\n", strings.Join(draft.Skills, ", ")))
	}

	// Contacts
	var contacts []string
	if draft.Phone != "" {
		contacts = append(contacts, fmt.Sprintf("📞 %s", draft.Phone))
	}
	if draft.Telegram != "" {
		contacts = append(contacts, fmt.Sprintf("✈️ %s", draft.Telegram))
	}
	if draft.Email != "" {
		contacts = append(contacts, fmt.Sprintf("📧 %s", draft.Email))
	}
	if len(contacts) > 0 {
		sb.WriteString("\n")
		sb.WriteString(strings.Join(contacts, "  "))
		sb.WriteString("\n")
	}

	// Missing fields
	var missing []string
	if draft.SalaryMin == 0 && draft.SalaryMax == 0 {
		missing = append(missing, hrMsgMissingSalary[lang])
	}
	if draft.Format == "" {
		missing = append(missing, hrMsgMissingFormat[lang])
	}
	if draftField(draft, "responsibilities", lang) == "" {
		missing = append(missing, hrMsgMissingResponsibilities[lang])
	}
	if draftField(draft, "requirements", lang) == "" {
		missing = append(missing, hrMsgMissingRequirements[lang])
	}
	if draft.ExperienceMin == 0 && draft.ExperienceMax == 0 {
		missing = append(missing, hrMsgMissingExperience[lang])
	}

	if len(missing) > 0 {
		missingNow := map[string]string{"en": "Currently missing:", "ru": "Сейчас не хватает:", "uz": "Hozircha yetishmayapti:"}
		sb.WriteString(fmt.Sprintf("\n❗️ %s\n", missingNow[lang]))
		for _, m := range missing {
			sb.WriteString(fmt.Sprintf("• %s\n", m))
		}
	}

	sb.WriteString("\n")
	sb.WriteString(hrMsgHowToContinue[lang])

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data(hrMsgBtnAddMissingInfo[lang], "hr_vac_add_info")),
		markup.Row(markup.Data(hrMsgBtnContinueCurrent[lang], "hr_vac_continue")),
		markup.Row(markup.Data(hrMsgBtnCreateDescription[lang], "hr_vac_create_desc")),
	)

	return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: markup})
}

// sendVacancyConfirmation shows the vacancy preview with two buttons: create and add/change info.
func (hb *HRBot) sendVacancyConfirmation(c tele.Context, draft *gemini.ParsedVacancyFull, lang string) error {
	var sb strings.Builder

	header := map[string]string{
		"en": "📋 *Your vacancy:*\n\n",
		"ru": "📋 *Ваша вакансия:*\n\n",
		"uz": "📋 *Sizning vakansiyangiz:*\n\n",
	}
	sb.WriteString(header[lang])

	// Title
	if title := draftField(draft, "title", lang); title != "" {
		sb.WriteString(fmt.Sprintf("*%s*\n", title))
	}

	// Salary
	if draft.SalaryMin > 0 || draft.SalaryMax > 0 {
		sb.WriteString(fmt.Sprintf("💰 %s – %s %s\n", formatNumber(int64(draft.SalaryMin)), formatNumber(int64(draft.SalaryMax)), draft.SalaryCurrency))
	}

	// Format
	if draft.Format != "" {
		formatNames := map[string]map[string]string{
			"office": {"en": "Office", "ru": "Офис", "uz": "Ofis"},
			"remote": {"en": "Remote", "ru": "Удалёнка", "uz": "Masofaviy"},
			"hybrid": {"en": "Hybrid", "ru": "Гибрид", "uz": "Gibrid"},
		}
		if names, ok := formatNames[draft.Format]; ok {
			sb.WriteString(fmt.Sprintf("📍 %s\n", names[lang]))
		}
	}

	// Experience
	if draft.ExperienceMin > 0 || draft.ExperienceMax > 0 {
		expLabel := map[string]string{"en": "Experience", "ru": "Опыт", "uz": "Tajriba"}
		sb.WriteString(fmt.Sprintf("📅 %s: %d–%d\n", expLabel[lang], draft.ExperienceMin, draft.ExperienceMax))
	}

	// Description
	if desc := draftField(draft, "description", lang); desc != "" {
		sb.WriteString(fmt.Sprintf("\n%s\n", desc))
	}

	// Responsibilities
	if resp := draftField(draft, "responsibilities", lang); resp != "" {
		respLabel := map[string]string{"en": "Responsibilities", "ru": "Обязанности", "uz": "Vazifalar"}
		sb.WriteString(fmt.Sprintf("\n*%s:*\n%s\n", respLabel[lang], resp))
	}

	// Requirements
	if req := draftField(draft, "requirements", lang); req != "" {
		reqLabel := map[string]string{"en": "Requirements", "ru": "Требования", "uz": "Talablar"}
		sb.WriteString(fmt.Sprintf("\n*%s:*\n%s\n", reqLabel[lang], req))
	}

	// Conditions
	if cond := draftField(draft, "conditions", lang); cond != "" {
		condLabel := map[string]string{"en": "Conditions", "ru": "Условия", "uz": "Shartlar"}
		sb.WriteString(fmt.Sprintf("\n*%s:*\n%s\n", condLabel[lang], cond))
	}

	// Skills
	if len(draft.Skills) > 0 {
		sb.WriteString(fmt.Sprintf("\n🛠 %s\n", strings.Join(draft.Skills, ", ")))
	}

	// Contacts
	var contacts []string
	if draft.Phone != "" {
		contacts = append(contacts, fmt.Sprintf("📞 %s", draft.Phone))
	}
	if draft.Email != "" {
		contacts = append(contacts, fmt.Sprintf("📧 %s", draft.Email))
	}
	if len(contacts) > 0 {
		sb.WriteString("\n")
		sb.WriteString(strings.Join(contacts, "  "))
		sb.WriteString("\n")
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data(hrMsgBtnConfirmCreate[lang], "hr_vac_confirm")),
		markup.Row(markup.Data(hrMsgBtnAddMoreInfo[lang], "hr_vac_add_info")),
	)

	return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: markup})
}

// buildAddInfoMessage shows the current vacancy content, missing fields, and asks for more info.
func (hb *HRBot) buildAddInfoMessage(draft *gemini.ParsedVacancyFull, lang string) string {
	var sb strings.Builder

	// Current data header
	currentLabel := map[string]string{
		"en": "📋 *Current vacancy data:*\n\n",
		"ru": "📋 *Текущие данные вакансии:*\n\n",
		"uz": "📋 *Joriy vakansiya ma'lumotlari:*\n\n",
	}
	sb.WriteString(currentLabel[lang])

	// Title
	if title := draftField(draft, "title", lang); title != "" {
		sb.WriteString(fmt.Sprintf("*%s*\n", title))
	}

	// Key details in compact form
	if draft.SalaryMin > 0 || draft.SalaryMax > 0 {
		sb.WriteString(fmt.Sprintf("💰 %s – %s %s\n", formatNumber(int64(draft.SalaryMin)), formatNumber(int64(draft.SalaryMax)), draft.SalaryCurrency))
	}
	if draft.Format != "" {
		formatNames := map[string]map[string]string{
			"office": {"en": "Office", "ru": "Офис", "uz": "Ofis"},
			"remote": {"en": "Remote", "ru": "Удалёнка", "uz": "Masofaviy"},
			"hybrid": {"en": "Hybrid", "ru": "Гибрид", "uz": "Gibrid"},
		}
		if names, ok := formatNames[draft.Format]; ok {
			sb.WriteString(fmt.Sprintf("📍 %s\n", names[lang]))
		}
	}
	if draft.ExperienceMin > 0 || draft.ExperienceMax > 0 {
		expYears := map[string]string{"en": "years", "ru": "лет", "uz": "yil"}
		if draft.ExperienceMin > 0 && draft.ExperienceMax > 0 {
			sb.WriteString(fmt.Sprintf("💼 %d–%d %s\n", draft.ExperienceMin, draft.ExperienceMax, expYears[lang]))
		} else if draft.ExperienceMin > 0 {
			sb.WriteString(fmt.Sprintf("💼 %d+ %s\n", draft.ExperienceMin, expYears[lang]))
		}
	}
	if len(draft.Skills) > 0 {
		sb.WriteString(fmt.Sprintf("🛠 %s\n", strings.Join(draft.Skills, ", ")))
	}
	sb.WriteString("\n")

	// Missing fields
	var missing []string
	if draft.SalaryMin == 0 && draft.SalaryMax == 0 {
		missing = append(missing, hrMsgMissingSalary[lang])
	}
	if draft.Format == "" {
		missing = append(missing, hrMsgMissingFormat[lang])
	}
	if draftField(draft, "responsibilities", lang) == "" {
		missing = append(missing, hrMsgMissingResponsibilities[lang])
	}
	if draftField(draft, "requirements", lang) == "" {
		missing = append(missing, hrMsgMissingRequirements[lang])
	}
	if draft.ExperienceMin == 0 && draft.ExperienceMax == 0 {
		missing = append(missing, hrMsgMissingExperience[lang])
	}

	if len(missing) > 0 {
		missingLabel := map[string]string{
			"en": "⚠️ Send the missing information:",
			"ru": "⚠️ Отправьте недостающую информацию:",
			"uz": "⚠️ Yetishmayotgan ma'lumotlarni yuboring:",
		}
		sb.WriteString(fmt.Sprintf("%s\n", missingLabel[lang]))
		for _, m := range missing {
			sb.WriteString(fmt.Sprintf("• %s\n", m))
		}
	} else {
		sb.WriteString(hrMsgSendAdditionalInfo[lang])
	}

	sb.WriteString("\n")

	howToSend := map[string]string{
		"en": "You can:\n✍️ write\n🎤 send a voice message\n📎 attach a file\n\n⚡ The more detail — the more accurate the match.",
		"ru": "Вы можете:\n✍️ написать\n🎤 отправить голосовое\n📎 прикрепить файл\n\n⚡ Чем подробнее — тем точнее подбор.",
		"uz": "Siz quyidagilarni yuborishingiz mumkin:\n✍️ yozish\n🎤 ovozli xabar yuborish\n📎 fayl biriktirish\n\n⚡ Qanchalik batafsil — shunchalik aniq tanlov.",
	}
	sb.WriteString(howToSend[lang])

	return sb.String()
}

// draftField returns the text field value for the given language, falling back to English.
func draftField(draft *gemini.ParsedVacancyFull, field, lang string) string {
	if draft == nil || draft.Fields == nil {
		return ""
	}
	if fields, ok := draft.Fields[field]; ok {
		if v := fields[lang]; v != "" {
			return v
		}
		return fields["en"]
	}
	return ""
}

// formatAsBullets takes a text block and formats each line/sentence as a bullet point.
// If the text already has bullet points or newlines, it preserves them.
func formatAsBullets(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Strip existing bullet markers
		line = strings.TrimPrefix(line, "• ")
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		result = append(result, "• "+line)
	}
	return strings.Join(result, "\n")
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

	applicantsLabel := map[string]string{"en": "applicants", "ru": "откликов", "uz": "ariza"}
	newLabel := map[string]string{"en": "new", "ru": "новых", "uz": "yangi"}

	var sb strings.Builder
	for i, v := range vacancies {
		title := vacancyTitle(&v, lang)
		minStr := formatNumber(int64(v.Vacancy.SalaryMin))
		maxStr := formatNumber(int64(v.Vacancy.SalaryMax))

		statsStr := ""
		total, unseen, err := hb.hrBotSvc.GetVacancyStats(ctx, v.Vacancy.ID)
		if err == nil && total > 0 {
			if unseen > 0 {
				statsStr = fmt.Sprintf(" | 👥 %d %s (%d %s)", total, applicantsLabel[lang], unseen, newLabel[lang])
			} else {
				statsStr = fmt.Sprintf(" | 👥 %d %s", total, applicantsLabel[lang])
			}
		}

		sb.WriteString(fmt.Sprintf("%d. **%s**\n   %s – %s %s | %s%s\n\n", i+1, title, minStr, maxStr, v.Vacancy.SalaryCurrency, v.Vacancy.Status, statsStr))
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

	waitMsg, _ := c.Bot().Send(c.Recipient(), hrMsgSearching[lang])

	results, err := hb.hrBotSvc.SearchCandidates(ctx, c.Text())
	if waitMsg != nil {
		_ = c.Bot().Delete(waitMsg)
	}
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
		markup.Row(tele.Btn{Text: hrMenuBtnCreateVacancy[lang]}, tele.Btn{Text: hrMenuBtnActiveVacancies[lang]}),
		markup.Row(tele.Btn{Text: hrMenuBtnFindCandidates[lang]}, tele.Btn{Text: hrMenuBtnChangeLang[lang]}),
	)
	return markup
}

func (hb *HRBot) hrInlineMenu(lang string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	rows := []tele.Row{
		markup.Row(markup.Data(hrMenuBtnCreateVacancy[lang], "hr_menu", "create_vacancy")),
		markup.Row(markup.Data(hrMenuBtnActiveVacancies[lang], "hr_menu", "active_vacancies")),
	}
	if hb.webAppURL != "" {
		rows = append(rows, markup.Row(markup.WebApp(hrMenuBtnFindCandidates[lang], &tele.WebApp{URL: hb.webAppURL})))
	} else {
		rows = append(rows, markup.Row(markup.Data(hrMenuBtnFindCandidates[lang], "hr_menu", "find_candidates")))
	}
	rows = append(rows, markup.Row(markup.Data(hrMenuBtnChangeLang[lang], "hr_menu", "change_lang")))
	markup.Inline(rows...)
	return markup
}

var hrBtnShowCandidates = map[string]string{
	"en": "👀 Show candidates",
	"ru": "👀 Показать кандидатов",
	"uz": "👀 Nomzodlarni ko'rsatish",
}
var hrBtnViewVacancy = map[string]string{
	"en": "📄 View vacancy",
	"ru": "📄 Посмотреть вакансию",
	"uz": "📄 Vakansiyani ko'rish",
}
var hrBtnEditVacancy = map[string]string{
	"en": "🔄 Add or edit info",
	"ru": "🔄 Добавить или изменить информацию",
	"uz": "🔄 Ma'lumot qo'shish yoki o'zgartirish",
}
var hrBtnStopPublication = map[string]string{
	"en": "⏹️ Stop publication",
	"ru": "⏹️ Остановить публикацию",
	"uz": "⏹️ E'lonni to'xtatish",
}

var hrBtnActivateVacancy = map[string]string{
	"en": "▶️ Activate",
	"ru": "▶️ Активировать",
	"uz": "▶️ Faollashtirish",
}

func vacancyPublishedMenu(lang, vacancyID, webAppURL string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row

	if webAppURL != "" {
		rows = append(rows,
			markup.Row(markup.WebApp(hrBtnShowCandidates[lang], &tele.WebApp{
				URL: webAppURL + "/?vacancy_id=" + vacancyID,
			})),
			markup.Row(markup.WebApp(hrBtnViewVacancy[lang], &tele.WebApp{
				URL: webAppURL + "/vacancies/" + vacancyID,
			})),
		)
	} else {
		rows = append(rows,
			markup.Row(markup.Data(hrBtnShowCandidates[lang], "hr_pub_candidates", vacancyID)),
			markup.Row(markup.Data(hrBtnViewVacancy[lang], "hr_pub_view", vacancyID)),
		)
	}
	rows = append(rows,
		markup.Row(markup.Data(hrBtnEditVacancy[lang], "hr_pub_edit", vacancyID)),
		markup.Row(markup.Data(hrBtnStopPublication[lang], "hr_pub_stop", vacancyID)),
	)
	markup.Inline(rows...)
	return markup
}

func buildVacancyCreatedMessage(v *application.VacancyWithDetails, lang string, matchingCount int) string {
	date := v.Vacancy.CreatedAt.Format("02.01.2006")
	shortID := v.Vacancy.ID.String()[:8]

	switch lang {
	case "ru":
		msg := fmt.Sprintf("✅ *Вакансия опубликована!*\n\n"+
			"📌 ID вакансии: #%s\n"+
			"📅 Дата публикации: %s\n"+
			"👁 Статус: Активна\n\n"+
			"Первые отклики ожидаются в течение 2–24 часов.",
			shortID, date)
		if matchingCount > 0 {
			msg += fmt.Sprintf("\n\n📊 В базе предварительно подходят: *%d кандидатов*.", matchingCount)
			msg += "\n\nЯ могу показать ТОП-кандидатов с наивысшим %% совместимости (80%%+)."
		}
		return msg
	case "uz":
		msg := fmt.Sprintf("✅ *Vakansiya e'lon qilindi!*\n\n"+
			"📌 Vakansiya ID: #%s\n"+
			"📅 E'lon sanasi: %s\n"+
			"👁 Holat: Faol\n\n"+
			"Birinchi javoblar 2–24 soat ichida kutilmoqda.",
			shortID, date)
		if matchingCount > 0 {
			msg += fmt.Sprintf("\n\n📊 Bazada taxminan *%d nomzod* mos keladi.", matchingCount)
			msg += "\n\nEng yuqori moslik foiziga ega (80%%+) TOP-nomzodlarni ko'rsata olaman."
		}
		return msg
	default:
		msg := fmt.Sprintf("✅ *Vacancy published!*\n\n"+
			"📌 Vacancy ID: #%s\n"+
			"📅 Published: %s\n"+
			"👁 Status: Active\n\n"+
			"First responses are expected within 2–24 hours.",
			shortID, date)
		if matchingCount > 0 {
			msg += fmt.Sprintf("\n\n📊 Preliminary matching candidates: *%d*.", matchingCount)
			msg += "\n\nI can show TOP candidates with the highest compatibility (80%%+)."
		}
		return msg
	}
}

func buildVacancyStoppedMessage(shortID, lang string) string {
	switch lang {
	case "ru":
		return fmt.Sprintf("⏸ *Публикация вакансии приостановлена.*\n\n"+
			"📌 ID вакансии: #%s\n"+
			"🔴 Статус: Неактивна\n\n"+
			"Вакансия больше не отображается кандидатам и исключена из поиска.\n\n"+
			"— Все полученные отклики сохранены.\n"+
			"— Вы можете вернуться к подбору в любое время.", shortID)
	case "uz":
		return fmt.Sprintf("⏸ *Vakansiya e'loni to'xtatildi.*\n\n"+
			"📌 Vakansiya ID: #%s\n"+
			"🔴 Holat: Nofaol\n\n"+
			"Vakansiya nomzodlarga ko'rinmaydi va qidiruvdan chiqarildi.\n\n"+
			"— Barcha olingan javoblar saqlangan.\n"+
			"— Istalgan vaqtda tanlovga qaytishingiz mumkin.", shortID)
	default:
		return fmt.Sprintf("⏸ *Vacancy publication paused.*\n\n"+
			"📌 Vacancy ID: #%s\n"+
			"🔴 Status: Inactive\n\n"+
			"The vacancy is no longer visible to candidates and excluded from search.\n\n"+
			"— All received responses are saved.\n"+
			"— You can return to recruitment at any time.", shortID)
	}
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

// resolveHRLang returns the language for the HR: saved language → telegram language → "en".
func (hb *HRBot) resolveHRLang(ctx context.Context, sender *tele.User) string {
	tgID := strconv.FormatInt(sender.ID, 10)
	hr, err := hb.hrBotSvc.GetHRByTelegramID(ctx, tgID)
	if err == nil && hr.Language != "" {
		return langOrDefault(hr.Language)
	}
	return detectLang(sender.LanguageCode)
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
