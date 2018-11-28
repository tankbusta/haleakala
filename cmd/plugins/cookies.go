package plugins

import (
	"math/rand"
	"time"

	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/muxer"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

var cookies = []string{
	"I think you need a :cookie:", "Have a :cookie:",
	"Have two :cookie:", "Sorry, I am out of cookies! Oh wait, found one! :cookie:",
	"C is for :cookie:", "https://www.youtube.com/watch?v=Ye8mB6VsUHw",
	"https://www.youtube.com/watch?v=-qTIGg3I5y8", "I think you had enough!",
	"Okay, but only one more :cookie:!", "I like cookies too! :thumbsup:",
	"C O O K I E", ":cookie: :cookie: :cookie: :cookie: :cookie: :cookie: :cookie: :cookie:",
	"Cafe?", "Beer?", "Cake? :cake: ", "One cookie for you. One cookie. I said one! :cookie: ",
	"Okay... I'm running to the store and getting some new cookies! BRB",
	"Here is a cookie! But you must promise to not give it to PK...",
	"There you go! :cookie: Do you want some !whiskey to your cookie?",
	"All you eat is cookies... ", "All your cookies are belong to us",
	"https://img.clipartfest.com/bfc51976969548980749767cd0a684b2_cookie-monster-as-a-baby-cookie-monster-clipart-baby_236-236.jpeg",
	"http://rack.0.mshcdn.com/media/ZgkyMDEzLzEwLzA3L2JmL0Nvb2tpZU1vbnN0LmE4NjZlLmpwZwpwCXRodW1iCTk1MHg1MzQjCmUJanBn/19941105/9eb/CookieMonster.jpg",
	"http://orig08.deviantart.net/357b/f/2011/235/d/8/cute_cookie_x3_by_lanahx3-d47lt9o.jpg",
	"You want cookie? Yes no spain?",
	"Please wait, while we process your request...", "Free Cookies for everyone! :cookie: :cookie: :cookie: :cookie: :cookie:",
	"Omnomnomnomnom... :yum: You want a :cookie: too?", "Sorry, but Deathwhisper ate all my cookies :(",
	"Are you sure?", "The cookie is a lie!", "Ofcourse! Here is a :cookie: for you!",
	"Share your :cookie: with a friend!", "Cookie? :cookie:", "NO!",
	"http://i4.manchestereveningnews.co.uk/incoming/article10580003.ece/ALTERNATES/s615/JS47622759.jpg",
	"Waiting on a cookie delivery... :truck: ", "Nobody ever gives me cookies :(",
	"Omnomnomnom sorry, that was the last one! :cry: ",
	"https://cdn75.picsart.com/185646673001202.png?r1024x1024",
	"https://www.youtube.com/watch?v=o41k-faChfA",
	"Are you feeling crummy? :cookie:", "Why do we cook bacon and bake cookies? Have a :cookie: and have some bacon!",
	"You want my :cookie:?", "I give you :cookie: ",
}

type CookiePlugin struct{}

func (s CookiePlugin) OnMessage(ctx *muxer.Context) {
	ctx.Send(cookies[r.Intn(len(cookies))])
}

func (s CookiePlugin) InstallRoute(f haleakala.InstallFunc) error {
	return f(s.Name(), "Want a cookie?", s.OnMessage)
}

func (s CookiePlugin) Name() string {
	return "cookies"
}
