package main

import (
	"github.com/willmroliver/plathbot/src/core"
)

const (
	DonateLink string = "https://support.wwf.org.uk/"
	AdoptLink  string = "https://gifts.worldwildlife.org/gift-center/gifts/species-adoptions/duck-billed-platypus"
)

func main() {
	s := core.NewServer()
	s.Listen()
}
