package main

import (
	"github.com/joho/godotenv"

	core "github.com/willmroliver/plathbot/src/api_core"
	_ "github.com/willmroliver/plathbot/src/include"
)

func init() {
	godotenv.Load()
}

func main() {
	core.NewServer().Listen()
}
