package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var gBot *tgbotapi.BotAPI
var gToken string
var gChatId int64

var gUsersInChat Users

var gUsefulActivities = Activities{
	// Саморазвитие
	{"yoga", "Йога (15 минут)", 1},
	{"meditation", "Медитация (15 минут)", 1},
	{"language", "Изучение иностранного языка (15 минут)", 1},
	{"swimming", "Плавание (15 минут)", 1},
	{"walk", "Прогулка пешком (15 минут)", 1},
	{"chores", "Домашние дела", 1},

	//Работка типо тут
	{"work_learning", "Изучение рабочих материалов (15 минут)", 1},
	{"portfolio_work", "Работа над портфолио-проектом (15 минут)", 1},
	{"resume_edit", "Редактирование резюме (15 минут)", 1},

	// шото креативноре
	{"creative", "Творческое созидание (15 минут)", 1},
	{"reading", "Чтение художественной литературы (15 минут)", 1},
}

var gRewards = Activities{
	// развлекаловка
	{"watch_series", "Просмотр сериала (1 серия)", 10},
	{"watch_movie", "Просмотр фильма (1 предмет)", 30},
	{"social_nets", "Просмотр социальных сетей (30 минут)", 10},

	// жратва
	{"eat_sweets", "300 ккал сладостей", 60},
}

type User struct {
	id    int64
	name  string
	coins uint16
}
type Users []*User

type Activity struct {
	code, name string
	coins      uint16
}
type Activities []*Activity

func init() {
	// Раскомментируйте и обновите значение токена, чтобы задать переменную окружения для токена Telegram-бота, предоставленного BotFather.
	// Удалите эту строку после установки env var. Храните токен в недоступном для общественности месте!
	//_ = os.Setenv(ИМЯ_ТОКЕНА В_OS, "ВСТАВЬ_ СВОЙ_ТОКЕН")
	_ = os.Setenv(TOKEN_NAME_IN_OS, "6484222809:AAF9tAiq9PPNwLlSDn8QxSBCr-8Th4rY_XE")
	if gToken = os.Getenv(TOKEN_NAME_IN_OS); gToken == "" {
		panic(fmt.Errorf(`не удалось загрузить переменную окружения "%s"`, TOKEN_NAME_IN_OS))
	}

	var err error
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	}
	gBot.Debug = true
}

func isStartMessage(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/start"
}

func isCallbackQuery(update *tgbotapi.Update) bool {
	return update.CallbackQuery != nil && update.CallbackQuery.Data != ""
}

func delay(seconds uint8) {
	time.Sleep(time.Second * time.Duration(seconds))
}

func sendStringMessage(msg string) {
	gBot.Send(tgbotapi.NewMessage(gChatId, msg))
}

func sendMessageWithDelay(delayInSec uint8, message string) {
	sendStringMessage(message)
	delay(delayInSec)
}

func printIntro(_ *tgbotapi.Update) {
	sendMessageWithDelay(2, "Привет"+EMOJI_SUNGLASSES)
	sendMessageWithDelay(7, "Существует множество полезных действий, выполняя которые регулярно, мы улучшаем качество своей жизни. Однако часто бывает веселее, проще или вкуснее сделать что-то вредное. Не так ли?")
	sendMessageWithDelay(7, "С большей вероятностью мы предпочтем смотреть короткометражки на YouTube вместо урока английского, покупать M&M's вместо овощей или валяться в постели вместо занятий йогой.")
	sendMessageWithDelay(1, EMOJI_SAD)
	sendMessageWithDelay(10, "Каждый из нас играл хотя бы в одну игру, где нужно прокачивать персонажа, делая его сильнее, умнее или красивее. Это приятно, потому что каждое действие приносит результат. Однако в реальной жизни систематические действия со временем становятся заметны. Давайте изменим это, не так ли?")
	sendMessageWithDelay(1, EMOJI_SMILE)
	sendMessageWithDelay(14, `Перед вами две таблицы: "Полезные действия" и "Награды". В первой таблице перечислены простые короткие действия, и за выполнение каждого из них вы заработаете указанное количество монет. Во второй таблице вы увидите список действий, которые вы можете выполнять только после оплаты монетами, заработанными на предыдущем шаге.`)
	sendMessageWithDelay(1, EMOJI_COIN)
	sendMessageWithDelay(10, `Например, вы проводите полчаса, занимаясь йогой, за что получаете 2 монеты. После этого у вас есть 2 часа занятий программированием, за которые вы получаете 8 монет. Теперь вы можете посмотреть 1 серию "Интернов" и выйти в ноль. Это так просто!`)
	sendMessageWithDelay(6, `Отмечайте выполненные полезные действия, чтобы не потерять свои монеты. И не забудьте "купить" награду, прежде чем приступить к ее выполнению.`)
}

func getKeyboardRow(buttonText, buttonCode string) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonCode))
}

func askToPrintIntro() {
	msg := tgbotapi.NewMessage(gChatId, "Во вступительных сообщениях вы можете ознакомиться с назначением этого бота и правилами игры. Что вы думаете?")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		getKeyboardRow(BUTTON_TEXT_PRINT_INTRO, BUTTON_CODE_PRINT_INTRO),
		getKeyboardRow(BUTTON_TEXT_SKIP_INTRO, BUTTON_CODE_SKIP_INTRO),
	)
	gBot.Send(msg)
}

func showMenu() {
	msg := tgbotapi.NewMessage(gChatId, "Выберите один из вариантов:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		getKeyboardRow(BUTTON_TEXT_BALANCE, BUTTON_CODE_BALANCE),
		getKeyboardRow(BUTTON_TEXT_USEFUL_ACTIVITIES, BUTTON_CODE_USEFUL_ACTIVITIES),
		getKeyboardRow(BUTTON_TEXT_REWARDS, BUTTON_CODE_REWARDS),
	)
	gBot.Send(msg)
}

func showBalance(user *User) {
	msg := fmt.Sprintf("%s,в данный момент ваш кошелек пуст %s \nОтслеживайте полезные действия, чтобы зарабатывать монеты", user.name, EMOJI_DONT_KNOW)
	if coins := user.coins; coins > 0 {
		msg = fmt.Sprintf("%s, у тебя есть %d %s", user.name, coins, EMOJI_COIN)
	}
	sendStringMessage(msg)
	showMenu()
}

func callbackQueryFromIsMissing(update *tgbotapi.Update) bool {
	return update.CallbackQuery == nil || update.CallbackQuery.From == nil
}

func getUserFromUpdate(update *tgbotapi.Update) (user *User, found bool) {
	if callbackQueryFromIsMissing(update) {
		return
	}

	userId := update.CallbackQuery.From.ID
	for _, userInChat := range gUsersInChat {
		if userId == userInChat.id {
			return userInChat, true
		}
	}
	return
}

func storeUserFromUpdate(update *tgbotapi.Update) (user *User, found bool) {
	if callbackQueryFromIsMissing(update) {
		return
	}

	from := update.CallbackQuery.From
	user = &User{id: from.ID, name: strings.TrimSpace(from.FirstName + " " + from.LastName), coins: 0}
	gUsersInChat = append(gUsersInChat, user)
	return user, true
}

func showActivities(activities Activities, message string, isUseful bool) {
	activitiesButtonsRows := make([]([]tgbotapi.InlineKeyboardButton), 0, len(activities)+1)
	for _, activity := range activities {
		activityDescription := ""
		if isUseful {
			activityDescription = fmt.Sprintf("+ %d %s: %s", activity.coins, EMOJI_COIN, activity.name)
		} else {
			activityDescription = fmt.Sprintf("- %d %s: %s", activity.coins, EMOJI_COIN, activity.name)
		}
		activitiesButtonsRows = append(activitiesButtonsRows, getKeyboardRow(activityDescription, activity.code))
	}
	activitiesButtonsRows = append(activitiesButtonsRows, getKeyboardRow(BUTTON_TEXT_PRINT_MENU, BUTTON_CODE_PRINT_MENU))

	msg := tgbotapi.NewMessage(gChatId, message)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(activitiesButtonsRows...)
	gBot.Send(msg)
}

func showUsefulActivities() {
	showActivities(gUsefulActivities, "Отследите полезное действие или вернитесь в главное меню:", true)
}

func showRewards() {
	showActivities(gRewards, "Получите вознаграждение или вернитесь в главное меню:", false)
}

func findActivity(activities Activities, choiceCode string) (activity *Activity, found bool) {
	for _, activity := range activities {
		if choiceCode == activity.code {
			return activity, true
		}
	}
	return
}

func processUsefulActivity(activity *Activity, user *User) {
	errorMsg := ""
	if activity.coins == 0 {
		errorMsg = fmt.Sprintf(`деятельность "%s" не имеет определенной стоимости`, activity.name)
	} else if user.coins+activity.coins > MAX_USER_COINS {
		errorMsg = fmt.Sprintf("у вас не может быть больше, чем %d %s", MAX_USER_COINS, EMOJI_COIN)
	}

	resultMessage := ""
	if errorMsg != "" {
		resultMessage = fmt.Sprintf("%s, Мне очень жаль, но %s %s Ваш баланс остается неизменным.", user.name, errorMsg, EMOJI_SAD)
	} else {
		user.coins += activity.coins
		resultMessage = fmt.Sprintf(`%s, то "%s" действие завершено! %d %s был добавлен в ваш аккаунт. Продолжайте в том же духе! %s%sТеперь у вас есть %d %s`,
			user.name, activity.name, activity.coins, EMOJI_COIN, EMOJI_BICEPS, EMOJI_SUNGLASSES, user.coins, EMOJI_COIN)
	}
	sendStringMessage(resultMessage)
}

func processReward(activity *Activity, user *User) {
	errorMsg := ""
	if activity.coins == 0 {
		errorMsg = fmt.Sprintf(`награда "%s" не имеет определенной стоимости`, activity.name)
	} else if user.coins < activity.coins {
		errorMsg = fmt.Sprintf(`в настоящее время у вас есть %d %s. Вы не можете себе этого позволить "%s" для %d %s`, user.coins, EMOJI_COIN, activity.name, activity.coins, EMOJI_COIN)
	}

	resultMessage := ""
	if errorMsg != "" {
		resultMessage = fmt.Sprintf("%s,Мне очень жаль, но %s %s Ваш баланс остается неизменным, вознаграждение недоступно %s", user.name, errorMsg, EMOJI_SAD, EMOJI_DONT_KNOW)
	} else {
		user.coins -= activity.coins
		resultMessage = fmt.Sprintf(`%s, награда "%s" все оплачено, приступайте к работе! %d %s был списан с вашего счета. Теперь у вас есть %d %s`, user.name, activity.name, activity.coins, EMOJI_COIN, user.coins, EMOJI_COIN)
	}
	sendStringMessage(resultMessage)
}

func updateProcessing(update *tgbotapi.Update) {
	user, found := getUserFromUpdate(update)
	if !found {
		if user, found = storeUserFromUpdate(update); !found {
			sendStringMessage("Не удалось идентифицировать пользователя")
			return
		}
	}

	choiceCode := update.CallbackQuery.Data
	log.Printf("[%T] %s", time.Now(), choiceCode)

	switch choiceCode {
	case BUTTON_CODE_BALANCE:
		showBalance(user)
	case BUTTON_CODE_USEFUL_ACTIVITIES:
		showUsefulActivities()
	case BUTTON_CODE_REWARDS:
		showRewards()
	case BUTTON_CODE_PRINT_INTRO:
		printIntro(update)
		showMenu()
	case BUTTON_CODE_SKIP_INTRO:
		showMenu()
	case BUTTON_CODE_PRINT_MENU:
		showMenu()
	default:
		if usefulActivity, found := findActivity(gUsefulActivities, choiceCode); found {
			processUsefulActivity(usefulActivity, user)

			delay(2)
			showUsefulActivities()
			return
		}

		if reward, found := findActivity(gRewards, choiceCode); found {
			processReward(reward, user)

			delay(2)
			showRewards()
			return
		}

		log.Printf(`[%T] !!!!!!!!! ОШИБКА: Неизвестный код "%s"`, time.Now(), choiceCode)
		msg := fmt.Sprintf("%s, Извините, я не распознал код '%s' %s Пожалуйста, сообщите об этой ошибке моему господину Давиду.", user.name, choiceCode, EMOJI_SAD)
		sendStringMessage(msg)
	}
}

func main() {
	log.Printf("Авторизован в учетной записи %s", gBot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT

	for update := range gBot.GetUpdatesChan(updateConfig) {
		if isCallbackQuery(&update) {
			updateProcessing(&update)
		} else if isStartMessage(&update) {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			gChatId = update.Message.Chat.ID
			askToPrintIntro()
		}
	}
}
