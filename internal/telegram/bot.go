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

const welcomeNew = `👋 Hi and welcome!

This bot helps people and companies find each other faster using AI.

🔹 Looking for a job?
Upload your resume and get matched with relevant positions.

🔹 Hiring for your company?
Create vacancies and find the right candidates without endless searching.

Smart search, multiple languages, better matches.

🌍 Choose your language to get started:`

var welcomeBackUser = map[string]string{
	"en": "👋 Welcome back %s! Glad to see you again.\n\nWhat would you like to do today?\n\n🔹 Update your profile or resume\n🔹 Find new job opportunities\n\nJust choose an option from the menu 👇",
	"ru": "👋 С возвращением, %s! Рады снова вас видеть.\n\nЧем вы хотите заняться сегодня?\n\n🔹 Обновить профиль или резюме\n🔹 Найти новые вакансии\n\nВыберите нужный пункт в меню 👇",
	"uz": "👋 Qaytganingiz bilan, %s! Sizni yana ko'rib turganimizdan xursandmiz.\n\nBugun nimani qilmoqchisiz?\n\n🔹 Profil yoki rezyumeni yangilash\n🔹 Yangi ish imkoniyatlarini topish\n\nQuyidagi menyudan tanlang 👇",
}

var welcomeBackHR = map[string]string{
	"en": "👋 Welcome back %s! Glad to see you again.\n\nWhat would you like to do today?\n\n🔹 Create or manage vacancies\n🔹 Search for candidates\n\nJust choose an option from the menu 👇",
	"ru": "👋 С возвращением, %s! Рады снова вас видеть.\n\nЧем вы хотите заняться сегодня?\n\n🔹 Создать или управлять вакансиями\n🔹 Найти подходящих кандидатов\n\nВыберите нужный пункт в меню 👇",
	"uz": "👋 Qaytganingiz bilan, %s! Sizni yana ko'rib turganimizdan xursandmiz.\n\nBugun nimani qilmoqchisiz?\n\n🔹 Vakansiyalar yaratish yoki boshqarish\n🔹 Mos nomzodlarni topish\n\nQuyidagi menyudan tanlang 👇",
}

var chooseLangReminder = map[string]string{
	"en": "Please choose a language from the buttons above ☝️",
	"ru": "Пожалуйста, выберите язык, нажав на кнопку выше ☝️",
	"uz": "Iltimos, yuqoridagi tugmalardan tilni tanlang ☝️",
}

type Bot struct {
	bot    *tele.Bot
	botSvc *application.BotService
}

func NewBot(token string, botSvc *application.BotService) (*Bot, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	tb := &Bot{bot: b, botSvc: botSvc}
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
			return c.Send("Something went wrong. Please try again.")
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
			return c.Send(welcomeNew, markup)
		}

		if result.IsHR {
			lang := langOrDefault(result.HR.Language)
			return c.Send(fmt.Sprintf(welcomeBackHR[lang], result.HR.FirstName))
		}

		lang := langOrDefault(result.User.Language)
		return c.Send(fmt.Sprintf(welcomeBackUser[lang], result.User.FirstName))
	})

	// Language selection callback
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

		user, err := botSvc.HandleLanguageSelection(ctx, sender.ID, language, sender.FirstName, sender.LastName, sender.Username, photoData)
		if err != nil {
			log.Printf("language selection error for %d: %v", sender.ID, err)
			return c.Respond(&tele.CallbackResponse{Text: "Error occurred. Please try /start again."})
		}

		_ = c.Respond(&tele.CallbackResponse{})
		_ = c.Delete()

		lang := langOrDefault(user.Language)
		return c.Send(fmt.Sprintf(welcomeBackUser[lang], user.FirstName))
	})

	// Text message handler
	bot.Handle(tele.OnText, func(c tele.Context) error {
		sender := c.Sender()

		if isChoosingLanguage(ctx, botSvc, sender.ID) {
			return c.Send(chooseLangReminder["en"] + "\n" + chooseLangReminder["ru"] + "\n" + chooseLangReminder["uz"])
		}

		user, err := ensureUser(ctx, botSvc, sender)
		if err != nil {
			return c.Send("Something went wrong. Please try /start first.")
		}

		_ = c.Send("Parsing your resume... This may take a moment.")

		result, err := botSvc.HandleResumeText(ctx, user.ID, c.Text())
		if err != nil {
			log.Printf("parse resume text error for %s: %v", user.ID, err)
			return c.Send("Failed to parse your resume. Please try again.")
		}

		return c.Send("Your resume has been parsed successfully!\n\n" +
			"Source language: " + result.SourceLang + "\n" +
			"Profile fields: " + itoa(len(result.Fields)) + "\n" +
			"Experience items: " + itoa(len(result.Experience)) + "\n" +
			"Education items: " + itoa(len(result.Education)))
	})

	// Document handler
	bot.Handle(tele.OnDocument, func(c tele.Context) error {
		sender := c.Sender()
		doc := c.Message().Document

		if isChoosingLanguage(ctx, botSvc, sender.ID) {
			return c.Send(chooseLangReminder["en"] + "\n" + chooseLangReminder["ru"] + "\n" + chooseLangReminder["uz"])
		}

		user, err := ensureUser(ctx, botSvc, sender)
		if err != nil {
			return c.Send("Something went wrong. Please try /start first.")
		}

		mimeType := doc.MIME
		if !isAllowedMIME(mimeType) {
			return c.Send("Unsupported file type. Please send a PDF, image (PNG/JPG), or text file.")
		}

		reader, err := bot.File(&doc.File)
		if err != nil {
			log.Printf("download file error: %v", err)
			return c.Send("Failed to download your file. Please try again.")
		}
		defer reader.Close()

		fileData, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("read file error: %v", err)
			return c.Send("Failed to read your file. Please try again.")
		}

		_ = c.Send("Parsing your document... This may take a moment.")

		result, err := botSvc.HandleResumeFile(ctx, user.ID, fileData, mimeType)
		if err != nil {
			log.Printf("parse resume file error for %s: %v", user.ID, err)
			return c.Send("Failed to parse your document. Please try again.")
		}

		return c.Send("Your document has been parsed successfully!\n\n" +
			"Source language: " + result.SourceLang + "\n" +
			"Profile fields: " + itoa(len(result.Fields)) + "\n" +
			"Experience items: " + itoa(len(result.Experience)) + "\n" +
			"Education items: " + itoa(len(result.Education)))
	})

	// Photo handler
	bot.Handle(tele.OnPhoto, func(c tele.Context) error {
		sender := c.Sender()
		photo := c.Message().Photo

		if isChoosingLanguage(ctx, botSvc, sender.ID) {
			return c.Send(chooseLangReminder["en"] + "\n" + chooseLangReminder["ru"] + "\n" + chooseLangReminder["uz"])
		}

		user, err := ensureUser(ctx, botSvc, sender)
		if err != nil {
			return c.Send("Something went wrong. Please try /start first.")
		}

		reader, err := bot.File(&photo.File)
		if err != nil {
			log.Printf("download photo error: %v", err)
			return c.Send("Failed to download your photo. Please try again.")
		}
		defer reader.Close()

		fileData, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("read photo error: %v", err)
			return c.Send("Failed to read your photo. Please try again.")
		}

		_ = c.Send("Parsing your image... This may take a moment.")

		result, err := botSvc.HandleResumeFile(ctx, user.ID, fileData, "image/jpeg")
		if err != nil {
			log.Printf("parse resume photo error for %s: %v", user.ID, err)
			return c.Send("Failed to parse your image. Please try again.")
		}

		return c.Send("Your image has been parsed successfully!\n\n" +
			"Source language: " + result.SourceLang + "\n" +
			"Profile fields: " + itoa(len(result.Fields)) + "\n" +
			"Experience items: " + itoa(len(result.Experience)) + "\n" +
			"Education items: " + itoa(len(result.Education)))
	})
}

func isChoosingLanguage(ctx context.Context, botSvc *application.BotService, senderID int64) bool {
	state, err := botSvc.GetBotState(ctx, senderID)
	if err != nil || state == nil {
		return false
	}
	return state.State == domain.BotStateChoosingLanguage
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
	case "application/pdf", "image/png", "image/jpeg", "text/plain":
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
