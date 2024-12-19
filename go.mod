module github.com/willmroliver/plathbot

go 1.22.4

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.25.12
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.24 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace github.com/go-telegram-bot-api/telegram-bot-api/v5 => ./lib/telegram-bot-api
