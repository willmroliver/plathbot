package main

import (
	"github.com/joho/godotenv"
	core "github.com/willmroliver/plathbot/src/api_core"
)

const (
	DonateLink string = "https://support.wwf.org.uk/"
	AdoptLink  string = "https://gifts.worldwildlife.org/gift-center/gifts/species-adoptions/duck-billed-platypus"
)

func main() {
	godotenv.Load()
	s := core.NewServer()
	s.Listen()
}
