package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Структура информации хранимой в БД о каждом пользователе
type dg_user struct {
	id       int
	usr_name string
	usr_cat  string
}

// ?
type notice struct {
	id   int
	text string
	cat  string
	hour int
	min  int
	date int
}

// Функция создает событие с заданными параметрами:
// cdesc - Описание параметров запуска для функции Cron
// textGlob - Сообщение переданное пользователем в напоминание
// cat - Категория которой адресовано напоминание
// ifAlw - Флаг определяющий, однократно ли воспроизведется функция или будет воспроизводиться с указанной периодичностью
// senders - Карта хранящая список чатов в виде ID чата - все координаты чата
// b - сущность бота передана в фукцию для осуществления возможности отправки напоминаний
func makeCron(cdesc, textGlob, cat string, ifAlw bool, senders map[int64]*tb.Chat, b *tb.Bot) *cron.Cron {

	var userList string

	if cat == "all" {
		userList = selectAll()
	} else {
		userList = selectDB(cat)
	}
	c := cron.New()
	c.AddFunc(cdesc, func() {
		for key := range senders {
			uresp := "Уважаемые " + userList + "\nНапоминаю:\n" + textGlob
			b.Send(senders[key], uresp)

		}
		if ifAlw == false {
			c.Stop()
		}
	})

	e := c.Entries()

	log.Println("cron: ", e)

	return c
}

// Поиск всех пользователей из таблицы
// возвращает форматированный список
func selectAll() string {
	connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select distinct(usr_name) from dg_users")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	dgUsers := []dg_user{}

	for rows.Next() {
		p := dg_user{}
		err := rows.Scan(&p.usr_cat)
		if err != nil {
			fmt.Println(err)
			continue
		}
		dgUsers = append(dgUsers, p)
	}

	prodResp := ""

	// Сохранит список пользователей в переменную prodResp в формате "userName1, userName2"
	for i := 0; i <= len(dgUsers)-1; i++ {
		prodResp += dgUsers[i].usr_cat + ", "
	}

	return prodResp
}

func selectDB(cg string) string {
	connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select * from dg_users where usr_cat = '" + cg + "'")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	dgUsers := []dg_user{}

	for rows.Next() {
		p := dg_user{}
		err := rows.Scan(&p.id, &p.usr_name, &p.usr_cat)
		if err != nil {
			fmt.Println(err)
			continue
		}
		dgUsers = append(dgUsers, p)
	}

	var prodResp string

	// Сохранит список пользователей в переменную prodResp в формате "userName1, userName2"
	for i := 0; i <= len(dgUsers)-1; i++ {
		prodResp += dgUsers[i].usr_name + ", "
	}

	return prodResp
}

func main() {

	senders := make(map[int64]*tb.Chat)

	var crn []*cron.Cron
	var notice []string

	botToken := "1188263369:AAEHda-pECm1HfI_TOlREfB59pruZtidKOg"
	b, err := tb.NewBot(tb.Settings{
		Token:  botToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return
	}

	log.Printf("Authorized on account posivnoy_bot")

//	c := cron.New()
//	c.AddFunc("CRON_TZ=Europe/Moscow 0 18 * * 1-5", func() {
//		auth := smtp.PlainAuth("", "mama210897@gmail.com", "mama23740740089334801021578749", "smtp.gmail.com")
//
//		// Connect to the server, authenticate, set the sender and recipient,
//		// and send the email all in one step.
//		to := []string{"schernetsov@fil-it.ru"}
//		msg := []byte("To: schernetsov@fil-it.ru\r\n" +
//			"Subject: discount Gophers!\r\n" +
//			"\r\n" +
//			"This is the email body.\r\n")
//		err := smtp.SendMail("smtp.gmail.com:587", auth, "mama210897@gmail.com", to, msg)
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Println("Remaind sended")
//	})
//	c.Start()

	// Задаем перечень кнопок

	day := tb.InlineButton{
		Unique: "DAY",
		Text:   "⌚ Ближайшие сутки",
	}

	chooseDate := tb.InlineButton{
		Unique: "CD",
		Text:   "📆 Выбрать дату",
	}

	always := tb.InlineButton{
		Unique: "AL",
		Text:   "♾️ Постоянные",
	}

	everyWeek := tb.InlineButton{
		Unique: "EW",
		Text:   "📅 Каждую неделю",
	}

	everyMonth := tb.InlineButton{
		Unique: "EM",
		Text:   "🗓️ Каждый месяц",
	}

	everyDay := tb.InlineButton{
		Unique: "ED",
		Text:   "🌀 Ежедневно",
	}

	// Собираем все кнопки в группы

	mainInline := [][]tb.InlineButton{
		[]tb.InlineButton{day, chooseDate},
		[]tb.InlineButton{always},
	}

	perInline := [][]tb.InlineButton{
		[]tb.InlineButton{everyWeek, everyMonth},
		[]tb.InlineButton{everyDay},
	}

	b.Handle("/start", func(m *tb.Message) {
		log.Println(m.Sender.Username, " id = ", m.Sender.ID, ": start")
		uresp := "Привет! 🗳️\nЯ помогу обратиться к определенной группе людей в беседе\nУ меня есть несколько комманд для разных групп:\n\n/all\n\n/an\n\n/dev\n\nЕсли необходимая группа не входит в список существующих команд можно воспользоваться конструкцией из примера ниже: \n\n/ng designers Привет 👋\n\nЕсли необходимо узнать список категорий можно воспользоваться командой:\n/show_cat\n\nДля настройки напоминаний можно воспользоваться коммандой:\n/menu"
		b.Send(m.Chat, uresp)

		// Собираем список чатов где была выполнена команда /start
		senders[m.Chat.ID] = m.Chat

		// Меню для настройки напоминаний
		b.Handle("/menu", func(m *tb.Message) {

			log.Println(m.Sender.Username, ": menu")

			uresp := "✉️ Меню ✉️"
			b.Send(m.Chat, uresp, &tb.ReplyMarkup{
				InlineKeyboard: mainInline,
			})

		})

		b.Handle(&day, func(c *tb.Callback) {

			log.Println(c.Sender.Username, ": menu")

			connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			rows, err := db.Query("select distinct(usr_cat) from dg_users")
			if err != nil {
				panic(err)
			}
			defer rows.Close()
			dgUsers := []dg_user{}

			for rows.Next() {
				p := dg_user{}
				err := rows.Scan(&p.usr_cat)
				if err != nil {
					fmt.Println(err)
					continue
				}
				dgUsers = append(dgUsers, p)
			}

			uresp := "Выбери категорию:\n\n/c_all\n\n"

			for i := 0; i <= len(dgUsers)-1; i++ {
				uresp += "/c_" + dgUsers[i].usr_cat + "\n\n"
			}
			b.Edit(c.Message, uresp)

			b.Handle(tb.OnText, func(m *tb.Message) {

				log.Println(m.Sender.Username, ": ", m.Text)

				usCat := m.Text[3:]

				log.Println(m.Sender.Username, ": ")

				uresp := "Укажите время оповещения.\nФормат: hh:mm"
				b.Send(m.Chat, uresp)

				b.Handle(tb.OnText, func(m *tb.Message) {

					log.Println(m.Sender.Username, ": ", m.Text)

					hourMin := strings.Split(m.Text, ":")

					hour := hourMin[0]
					min := hourMin[1]

					log.Println(m.Sender.Username, ": ")

					uresp := "Укажите текст оповещения:"
					b.Send(m.Chat, uresp)

					b.Handle(tb.OnText, func(m *tb.Message) {

						cp := "CRON_TZ=Europe/Moscow " + min + " " + hour + " * * *"

						crn = append(crn, makeCron(cp, m.Text, usCat, true, senders, b))
						notice = append(notice, m.Text)

						crn[len(crn)-1].Start()

						uresp := "Напоминание создано"
						b.Send(m.Chat, uresp)

						b.Handle(tb.OnText, func(m *tb.Message) {})

					})

				})

			})

		})

		b.Handle(&chooseDate, func(c *tb.Callback) {

			log.Println(c.Sender.Username, ": menu")

			connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			rows, err := db.Query("select distinct(usr_cat) from dg_users")
			if err != nil {
				panic(err)
			}
			defer rows.Close()
			dgUsers := []dg_user{}

			for rows.Next() {
				p := dg_user{}
				err := rows.Scan(&p.usr_cat)
				if err != nil {
					fmt.Println(err)
					continue
				}
				dgUsers = append(dgUsers, p)
			}

			uresp := "Выбери категорию:\n\n/c_all\n\n"

			for i := 0; i <= len(dgUsers)-1; i++ {
				uresp += "/c_" + dgUsers[i].usr_cat + "\n\n"
			}
			b.Edit(c.Message, uresp)

			b.Handle(tb.OnText, func(m *tb.Message) {

				usCat := m.Text[3:]

				log.Println(m.Sender.Username, ": ")

				uresp := "Укажите дату оповещения.\nФормат: DD"
				b.Send(m.Chat, uresp)

				b.Handle(tb.OnText, func(m *tb.Message) {

					log.Println(m.Sender.Username, ": ", m.Text)

					date := m.Text

					uresp := "Укажите время оповещения.\nФормат: hh:mm"
					b.Send(m.Chat, uresp)

					b.Handle(tb.OnText, func(m *tb.Message) {

						log.Println(m.Sender.Username, ": ", m.Text)

						hourMin := strings.Split(m.Text, ":")

						hour := hourMin[0]
						min := hourMin[1]

						log.Println(m.Sender.Username, ": ")

						uresp := "Укажите текст оповещения:"
						b.Send(m.Chat, uresp)

						b.Handle(tb.OnText, func(m *tb.Message) {

							cp := "CRON_TZ=Europe/Moscow " + min + " " + hour + " " + date + " * *"

							crn = append(crn, makeCron(cp, m.Text, usCat, true, senders, b))
							notice = append(notice, m.Text)

							crn[len(crn)-1].Start()

							uresp := "Напоминание создано"
							b.Send(m.Chat, uresp)

							b.Handle(tb.OnText, func(m *tb.Message) {})

						})

					})

				})
			})

		})

		b.Handle(&always, func(c *tb.Callback) {

			log.Println(c.Sender.Username, ": menu")

			connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			rows, err := db.Query("select distinct(usr_cat) from dg_users")
			if err != nil {
				panic(err)
			}
			defer rows.Close()
			dgUsers := []dg_user{}

			for rows.Next() {
				p := dg_user{}
				err := rows.Scan(&p.usr_cat)
				if err != nil {
					fmt.Println(err)
					continue
				}
				dgUsers = append(dgUsers, p)
			}

			uresp := "Выбери категорию:\n\n/c_all\n\n"

			for i := 0; i <= len(dgUsers)-1; i++ {
				uresp += "/c_" + dgUsers[i].usr_cat + "\n\n"
			}
			b.Edit(c.Message, uresp)

			b.Handle(tb.OnText, func(m *tb.Message) {

				log.Println(m.Sender.Username, ": ", m.Text)

				usCat := m.Text[3:]

				log.Println(m.Sender.Username, ": ")

				uresp := "Выбери периодичность"
				b.Send(m.Chat, uresp, &tb.ReplyMarkup{
					InlineKeyboard: perInline,
				})

				b.Handle(&everyWeek, func(c *tb.Callback) {

					log.Println(c.Sender.Username, ": ", m.Text)

					uresp := "Укажи номер дня недели.\nЕсли дней несколько напиши номера через запятую без пробелов)"
					b.Edit(c.Message, uresp)

					b.Handle(tb.OnText, func(m *tb.Message) {

						log.Println(m.Sender.Username, ": ", m.Text)

						weekDay := m.Text

						uresp := "Укажите время оповещения.\nФормат: hh:mm"
						b.Send(m.Chat, uresp)

						b.Handle(tb.OnText, func(m *tb.Message) {

							log.Println(m.Sender.Username, ": ", m.Text)

							hourMin := strings.Split(m.Text, ":")

							hour := hourMin[0]
							min := hourMin[1]

							log.Println(m.Sender.Username, ": ")

							uresp := "Укажите текст оповещения:"
							b.Send(m.Chat, uresp)

							b.Handle(tb.OnText, func(m *tb.Message) {

								cp := "CRON_TZ=Europe/Moscow " + min + " " + hour + " * * " + weekDay

								crn = append(crn, makeCron(cp, m.Text, usCat, true, senders, b))
								notice = append(notice, m.Text)

								crn[len(crn)-1].Start()

								uresp := "Напоминание создано"
								b.Send(m.Chat, uresp)

								b.Handle(tb.OnText, func(m *tb.Message) {})

							})

						})

					})

				})

				b.Handle(&everyMonth, func(c *tb.Callback) {

					log.Println(c.Sender.Username, ": ", m.Text)

					uresp := "Укажи число.\nЕсли дней несколько напиши числа через запятую без пробелов)"
					b.Edit(c.Message, uresp)

					b.Handle(tb.OnText, func(m *tb.Message) {

						log.Println(m.Sender.Username, ": ", m.Text)

						date := m.Text

						uresp := "Укажите время оповещения.\nФормат: hh:mm"
						b.Send(m.Chat, uresp)

						b.Handle(tb.OnText, func(m *tb.Message) {

							log.Println(m.Sender.Username, ": ", m.Text)

							hourMin := strings.Split(m.Text, ":")

							hour := hourMin[0]
							min := hourMin[1]

							log.Println(m.Sender.Username, ": ")

							uresp := "Укажите текст оповещения:"
							b.Send(m.Chat, uresp)

							b.Handle(tb.OnText, func(m *tb.Message) {

								cp := "CRON_TZ=Europe/Moscow " + min + " " + hour + " * " + date + " *"

								crn = append(crn, makeCron(cp, m.Text, usCat, true, senders, b))
								notice = append(notice, m.Text)

								crn[len(crn)-1].Start()

								uresp := "Напоминание создано"
								b.Send(m.Chat, uresp)

								b.Handle(tb.OnText, func(m *tb.Message) {})

							})

						})

					})

				})

				b.Handle(&everyDay, func(c *tb.Callback) {

					log.Println(m.Sender.Username, ": ", m.Text)

					uresp := "Укажите время оповещения.\nФормат: hh:mm"
					b.Send(m.Chat, uresp)

					b.Handle(tb.OnText, func(m *tb.Message) {

						log.Println(m.Sender.Username, ": ", m.Text)

						hourMin := strings.Split(m.Text, ":")

						hour := hourMin[0]
						min := hourMin[1]

						log.Println(m.Sender.Username, ": ")

						uresp := "Укажите текст оповещения:"
						b.Send(m.Chat, uresp)

						b.Handle(tb.OnText, func(m *tb.Message) {

							cp := "CRON_TZ=Europe/Moscow " + min + " " + hour + " * * *"

							crn = append(crn, makeCron(cp, m.Text, usCat, true, senders, b))
							notice = append(notice, m.Text)

							crn[len(crn)-1].Start()

							uresp := "Напоминание создано"
							b.Send(m.Chat, uresp)

							b.Handle(tb.OnText, func(m *tb.Message) {})

						})

					})

				})

			})

		})

		// Комманды вызова разных групп

		b.Handle("/an", func(m *tb.Message) {

			log.Println(m.Sender.Username, ": analysts")

			userList := selectDB("an")

			log.Println(m.Sender.Username, ": ", userList)

			uresp := "Дорогие, " + userList + "сообщение выше для вас ✉️"
			b.Send(m.Chat, uresp)

		})

		b.Handle("/dev", func(m *tb.Message) {

			log.Println(m.Sender.Username, ": dev")

			userList := selectDB("dev")

			log.Println(m.Sender.Username, ": ", userList)

			uresp := "Дорогие, " + userList + "сообщение выше для вас ✉️"
			b.Send(m.Chat, uresp)

		})

		b.Handle("/ng", func(m *tb.Message) {

			ng := strings.Split(m.Text, " ")

			if len(ng) == 1 {
				uresp := "Группа не указана"
				b.Send(m.Chat, uresp)
			} else {

				log.Println(m.Sender.Username, ": ng: "+ng[1])

				userList := selectDB(ng[1])

				log.Println(m.Sender.Username, ": ", userList)

				uresp := ""

				if userList == "" {
					uresp = "Группа " + ng[1] + " не найдена("
				} else {
					uresp = "Дорогие, " + userList + "сообщение выше для вас ✉️"
				}

				b.Send(m.Chat, uresp)
			}

		})

		b.Handle("/all", func(m *tb.Message) {

			connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			rows, err := db.Query("select distinct(usr_name) from dg_users")
			if err != nil {
				panic(err)
			}
			defer rows.Close()
			dgUsers := []dg_user{}

			for rows.Next() {
				p := dg_user{}
				err := rows.Scan(&p.usr_cat)
				if err != nil {
					fmt.Println(err)
					continue
				}
				dgUsers = append(dgUsers, p)
			}

			prodResp := ""

			for i := 0; i <= len(dgUsers)-1; i++ {
				prodResp += dgUsers[i].usr_cat + ", "
			}

			uresp = "Дорогие, " + prodResp + "сообщение выше для вас ✉️"

			b.Send(m.Chat, uresp)

		})

		// Показать все существующие в базе категории
		b.Handle("/show_cat", func(m *tb.Message) {

			connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			rows, err := db.Query("select distinct(usr_cat) from dg_users")
			if err != nil {
				panic(err)
			}
			defer rows.Close()
			dgUsers := []dg_user{}

			for rows.Next() {
				p := dg_user{}
				err := rows.Scan(&p.usr_cat)
				if err != nil {
					fmt.Println(err)
					continue
				}
				dgUsers = append(dgUsers, p)
			}

			prodResp := "Categories:\n"

			for i := 0; i <= len(dgUsers)-1; i++ {
				prodResp += "⭕ " + dgUsers[i].usr_cat + "\n"
			}

			uresp := prodResp

			b.Send(m.Chat, uresp)

		})

		b.Handle("/show_cron", func(m *tb.Message) {

			log.Println(m.Sender.Username, ": remove cron")

			uresp := "Весь перечень напоминаний:\n\n"

			for i := range notice {
				uresp += strconv.Itoa(i) + " " + notice[i] + "\n"
			}

			b.Send(m.Chat, uresp)

		})

		b.Handle("/cron", func(m *tb.Message) {

			input := strings.Split(m.Text, " ")

			num, err := strconv.Atoi(input[1])
			if err != nil {
				log.Println("Error conv")
			}

			crn[num].Remove(1)

			copy(crn[num:], crn[num+1:])
			crn = crn[:len(crn)-1]

			copy(notice[num:], notice[num+1:])
			notice = notice[:len(notice)-1]

			log.Println(m.Sender.Username, ": remove cron")

			uresp := "Напоминание " + m.Text + " удалено!"
			b.Send(m.Chat, uresp)

		})

		b.Handle("/help", func(m *tb.Message) {

			log.Println(m.Sender.Username, ": help")
			uresp := "Привет! 🗳️\nЯ помогу обратиться к определенной группе людей в беседе\nУ меня есть несколько комманд для разных групп:\n\n/an\n\n/dev\n\nЕсли необходимая группа не входит в список существующих команд можно воспользоваться конструкцией из примера ниже: \n\n/ng designers Привет 👋\n\nЕсли необходимо узнать список категорий можно воспользоваться командой:\n/show_cat\n\nДля настройки напоминаний можно воспользоваться коммандой:\n/menu\n\nЧтобы посмотреть список напоминаний используй команду:\n/show_cron\n\nДля удаления напоминания необходимо вызвать комманду /cron и через пробел указать номер напоминания, например:\n/cron 0"
			b.Send(m.Chat, uresp)

		})

	})

	b.Start()
}
