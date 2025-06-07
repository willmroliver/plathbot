module github.com/willmroliver/plathbot

go 1.24.3

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/joho/godotenv v1.5.1
	github.com/vartanbeno/go-reddit/v2 v2.0.1
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.25.12
)

require (
	github.com/golang/protobuf v1.2.0 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.24 // indirect
	golang.org/x/net v0.0.0-20190108225652-1e06a53dbb7e // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/appengine v1.4.0 // indirect
)

replace github.com/go-telegram-bot-api/telegram-bot-api/v5 => ./lib/telegram-bot-api
