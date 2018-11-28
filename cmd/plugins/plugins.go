package plugins

import "github.com/tankbusta/haleakala"

var DefaultPlugins = []haleakala.IBasicPlugin{
	CookiePlugin{},
	CurrentTimePlugin{},
	EVEStatusPlugin{},
	AboutPlugin{},
	WhiskeyPlugin{},
	BeerPlugin{},
}
