package plugins

import (
	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/muxer"
)

var beer = []string{
	"https://i.pinimg.com/originals/e7/b2/a1/e7b2a10cfc5391cfa74152d798e224aa.jpg",
	"Beer is proof that God loves us and wants us to be happy. :beer:",
	"Milk is for babies. When you grow up you have to drink beer. :beer:",
	"Yes, :beer: for everyone! :beers:",
	"An Irishman is the only man in the world who will step over the bodies of a dozen naked women to get to a bottle of stout.",
	"Beer! The cause and solution to all of life's problems. ",
	"A woman is like beer. They look good, they smell good, and you'd step over your own mother just to get one! ",
	"Me no function beer well without.",
	"https://bierologie.de/wp-content/uploads/2015/04/glass-of-beer.png",
	"Would you not rather have a :cookie:?",
	"http://pngimg.com/upload/beer_PNG2346.png",
	"http://i.telegraph.co.uk/multimedia/archive/02326/obama_2326627b.jpg",
	"https://thelistlove.files.wordpress.com/2014/03/47.jpg",
	"Beer Pong is a sport, right guys?",
	"B E E R! :beers: Everyone drink! :beer: ",
	"https://www.youtube.com/watch?v=25NQqK4E5vk",
	"https://www.youtube.com/watch?v=QghICtdvNH4",
}

type BeerPlugin struct{}

func (s BeerPlugin) OnMessage(ctx *muxer.Context) {
	ctx.Send(beer[r.Intn(len(beer))])
}

func (s BeerPlugin) InstallRoute(f haleakala.InstallFunc) error {
	return f(s.Name(), "Mmmm beeer", s.OnMessage)
}

func (s BeerPlugin) Name() string {
	return "beer"
}
