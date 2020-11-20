package main

import (
	"fmt"
	"log"
	"time"

	//"strconv"
	//"net/http"
	"database/sql"

	_ "github.com/lib/pq"
	tb "gopkg.in/tucnak/telebot.v2"
)

type dg_user struct {
	id       int
	usr_name string
	usr_cat  string
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

	prodResp := cg + ":\n"

	for i := 0; i <= len(dgUsers)-1; i++ {
		prodResp += "⭕ " + dgUsers[i].usr_name + "\n"
	}

	return prodResp
}

func main() {

	catGlobal := "analysts"

	botToken := "1375735250:AAHnJv-xbLQS95DnEwmMcGFTaWdnsfzRYbg"
	b, err := tb.NewBot(tb.Settings{
		Token:  botToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return
	}

	log.Printf("Authorized on account posivnoyAdmin_bot")

	dev := tb.InlineButton{
		Unique: "DEV",
		Text:   "💻 Dev",
	}

	analysts := tb.InlineButton{
		Unique: "AN",
		Text:   "📈 Analysts",
	}

	newCat := tb.InlineButton{
		Unique: "NC",
		Text:   "➕ New Category",
	}

	allUsers := tb.InlineButton{
		Unique: "AU",
		Text:   "🧾 All Users",
	}

	allCat := tb.InlineButton{
		Unique: "AC",
		Text:   "🧾 All Categories",
	}

	handSelect := tb.InlineButton{
		Unique: "HS",
		Text:   "🔍 Hand Select",
	}

	deleteProd := tb.InlineButton{
		Unique: "DP",
		Text:   "❌ Delete",
	}

	deleteLast := tb.InlineButton{
		Unique: "DL",
		Text:   "❌ Delete Last",
	}

	// Collect buttons on group

	mainInline := [][]tb.InlineButton{
		[]tb.InlineButton{analysts, dev},
		[]tb.InlineButton{allUsers, allCat},
		[]tb.InlineButton{newCat, handSelect},
		[]tb.InlineButton{deleteLast},
	}

	b.Handle("/start", func(m *tb.Message) {
		log.Println(m.Sender.Username, " id = ", m.Sender.ID, ": start")
		uresp := "Привет! 🗳️\nЯ помогу настроить список сотрудников для каждой группы!"
		b.Send(m.Chat, uresp, &tb.ReplyMarkup{
			InlineKeyboard: mainInline,
		})

		b.Handle(tb.OnText, func(m *tb.Message) {

			if catGlobal == "new_cat" {

				catGlobal = m.Text
				uresp := "Поочередно введите список пользователей для группы " + catGlobal
				b.Send(m.Chat, uresp, &tb.ReplyMarkup{
					InlineKeyboard: mainInline,
				})

			} else if catGlobal == "hand_select" {

				catGlobal = m.Text
				uresp := selectDB(catGlobal)
				b.Send(m.Chat, uresp, &tb.ReplyMarkup{
					InlineKeyboard: mainInline,
				})

			} else {

				log.Println(m.Sender.Username, ": added '", m.Text, "' to ", catGlobal)

				connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
				db, err := sql.Open("postgres", connStr)
				if err != nil {
					panic(err)
				}
				defer db.Close()

				result, err := db.Exec("insert into dg_users (usr_name, usr_cat) values ($1, $2)",
					m.Text, catGlobal)
				if err != nil {
					panic(err)
				}

				log.Println(result.RowsAffected()) // количество добавленных строк

				uresp := selectDB(catGlobal)
				b.Send(m.Chat, uresp, &tb.ReplyMarkup{
					InlineKeyboard: mainInline,
				})
			}

		})

		b.Handle(&analysts, func(c *tb.Callback) {

			log.Println(c.Sender.Username, ": show analysts")

			catGlobal = "an"

			uresp := selectDB(catGlobal)

			b.Edit(c.Message, uresp, &tb.ReplyMarkup{
				InlineKeyboard: mainInline,
			})

		})

		b.Handle(&dev, func(c *tb.Callback) {

			log.Println(c.Sender.Username, ": show developers ")

			catGlobal = "dev"

			uresp := selectDB(catGlobal)

			b.Edit(c.Message, uresp, &tb.ReplyMarkup{
				InlineKeyboard: mainInline,
			})

		})

		b.Handle(&allUsers, func(c *tb.Callback) {

			connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			rows, err := db.Query("select * from dg_users order by usr_cat")
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

			prodResp := "All Users:\n"

			for i := 0; i <= len(dgUsers)-1; i++ {
				prodResp += "⭕ " + dgUsers[i].usr_cat + " " + dgUsers[i].usr_name + " " + "\n"
			}

			uresp := prodResp

			b.Edit(c.Message, uresp, &tb.ReplyMarkup{
				InlineKeyboard: mainInline,
			})

		})

		b.Handle(&allCat, func(c *tb.Callback) {

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

			b.Edit(c.Message, uresp, &tb.ReplyMarkup{
				InlineKeyboard: mainInline,
			})

		})

		b.Handle(&newCat, func(c *tb.Callback) {

			log.Println(c.Sender.Username, ": set new category")

			catGlobal = "new_cat"

			uresp := "Ввведите новую категорию:"
			b.Edit(c.Message, uresp)

		})

		b.Handle(&handSelect, func(c *tb.Callback) {

			log.Println(m.Sender.Username, ": hand select")

			catGlobal = "hand_select"

			uresp := "Ввведите искомую категорию:"
			b.Edit(c.Message, uresp)

		})

		b.Handle(&deleteProd, func(c *tb.Callback) {

			log.Println(c.Sender.Username, ": delet all in category ", catGlobal)

			connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			result, err := db.Exec("delete from dg_users where usr_cat = '" + catGlobal + "'")
			if err != nil {
				panic(err)
			}

			log.Println(result.RowsAffected()) // количество добавленных строк

			uresp := "Список сотрудников очищен"
			b.Edit(c.Message, uresp, &tb.ReplyMarkup{
				InlineKeyboard: mainInline,
			})

		})

		b.Handle(&deleteLast, func(c *tb.Callback) {

			connStr := "user=postgres password=qwe123 dbname=digital_users sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			result, err := db.Exec("delete from dg_users where id = (select max(id) from dg_users where usr_cat = '" + catGlobal + "')")
			if err != nil {
				panic(err)
			}

			log.Println(c.Sender.Username, ": delete last row where user category = ", catGlobal)

			log.Println(result.RowsAffected()) // количество добавленных строк

			uresp := selectDB(catGlobal)
			b.Edit(c.Message, uresp, &tb.ReplyMarkup{
				InlineKeyboard: mainInline,
			})
		})

	})

	b.Start()
}
