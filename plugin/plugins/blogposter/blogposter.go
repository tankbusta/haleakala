package blogposter

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tankbusta/haleakala/plugin"
	"github.com/uptrace/bun"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// BlogChannelID is the discord channel ID where blog posts will end up
//
// TODO: Make this configurable
const BlogChannelID = "896809463141511212"

var (
	printer            *message.Printer
	checkForBlogsEvery = time.Hour * 1
)

func init() {
	message.Set(language.English, "%d minutes",
		plural.Selectf(1, "%d",
			plural.One, "%d minute",
			plural.Other, "%d minutes",
		))
	printer = message.NewPrinter(language.English)

	plug := &mandiantBlogP{}
	plugin.Register(plug.Name(), plug)
}

type mandiantBlogP struct {
	plugin.BasePlugin
	db       *bun.DB
	ds       *discordgo.Session
	shutdown chan struct{}
	done     chan error
}

func (mandiantBlogP) Name() string { return "mandiant_blog_poster" }

func (s *mandiantBlogP) OnLoad(ds *discordgo.Session, db *bun.DB) error {
	s.db = db
	s.ds = ds

	// Shutdown is closed when we want to exist
	s.shutdown = make(chan struct{}, 1)
	// Done contains any error that fired when we were shutting down
	s.done = make(chan error, 1)

	go s.checkForPosts()

	return nil
}

func (s *mandiantBlogP) OnShutdown(_ *discordgo.Session) error {
	close(s.shutdown)
	return <-s.done
}

func (s *mandiantBlogP) checkForPosts() {
	// All good code starts off with a bit of a hack amirite?
	ticker := time.NewTicker(checkForBlogsEvery)
	defer ticker.Stop()

CheckerLoop:
	for {
		select {
		case <-s.shutdown:
			break CheckerLoop
		case <-ticker.C:
			log.Info().Msg("Checking for new blog posts...")
			// plugin.DiscordLogEvent(s.ds, "Checking for new blog posts...")
			webPageNode, err := getBlogContent()
			if err != nil {
				plugin.DiscordLogEvent(s.ds, "Failed checking for new blog posts: %s", err.Error())
				continue
			}

			posts, err := GetMandiantBlogs(webPageNode)
			if err != nil {
				plugin.DiscordLogEvent(s.ds, "Failed parsing blog posts: %s", err.Error())
				continue
			}

		CheckForNewPostLoop:
			for _, post := range posts {
				// Could probably do a much better query here...
				count, err := s.db.NewSelect().Model((*BlogPost)(nil)).Where("url = ?", post.URL).Limit(1).Count(context.Background())
				if err != nil {
					plugin.DiscordLogEvent(s.ds, "Failed checking if blog post was already posted: %s", err.Error())
					continue
				}

				switch count {
				case 0:
					// New post!
					bpme := blogPostToEmbed(post)
					if _, err := s.ds.ChannelMessageSendEmbed(BlogChannelID, bpme); err != nil {
						log.Error().
							Err(err).
							Str("channel", BlogChannelID).
							Msg("Failed sending new blog post to channel")
					}
				case 1:
					// Already posted and because we're sorting by date, we're done
					break CheckForNewPostLoop
				}
			}

			log.Info().Msg("Done checking for new blog posts")
			// plugin.DiscordLogEvent(s.ds, "Discovered %d blog posts", len(posts))
		}
	}

	s.done <- nil
	close(s.done)
}

func blogPostToEmbed(bp BlogPost) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		URL:         bp.URL,
		Title:       "New Blog!",
		Description: bp.Title,
		Color:       0xd80e0b,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Length",
				Value: printer.Sprintf("%d minutes", bp.TimeToRead),
			},
		},
	}
}
