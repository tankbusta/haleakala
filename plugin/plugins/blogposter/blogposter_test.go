package blogposter

import (
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/require"
)

func TestInternal_blogPostToEmbed(t *testing.T) {
	t.Run("Minutes", func(subT *testing.T) {
		me := blogPostToEmbed(BlogPost{
			Title:      "FIN12: The Prolific Ransomware Intrusion Threat Actor That Has Aggressively Pursued Healthcare Targets",
			TimeToRead: 9,
			URL:        "https://www.mandiant.com/resources/fin12-ransomware-intrusion-actor-pursuing-healthcare-targets",
		})

		require.Equal(subT, "9 minutes", me.Fields[0].Value)
	})

	t.Run("Minute", func(subT *testing.T) {
		me := blogPostToEmbed(BlogPost{
			Title:      "FIN12: The Prolific Ransomware Intrusion Threat Actor That Has Aggressively Pursued Healthcare Targets",
			TimeToRead: 1,
			URL:        "https://www.mandiant.com/resources/fin12-ransomware-intrusion-actor-pursuing-healthcare-targets",
		})

		require.Equal(subT, "1 minute", me.Fields[0].Value)
	})
}

func TestBlogStuff(t *testing.T) {
	downloadedBlogPage := "./testdata/MBlogs.html"
	webPageNode, err := htmlquery.LoadDoc(downloadedBlogPage)
	require.Nil(t, err)

	blogs, err := GetMandiantBlogs(webPageNode)
	require.Nil(t, err)
	require.Len(t, blogs, 15)

	// Ensure the sort is good
	require.True(t, blogs[0].PostedOn.After(blogs[14].PostedOn))
}
