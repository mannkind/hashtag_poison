package main

import (
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/spf13/viper"
)

const TRENDLEN int8 = 50
const USA int64 = 23424977
const HASHTAG string = "#"
const EMPTY string = ""
const DOTDOTDOT string = "…"
const REPLACE string = "~1REPLACEMENTHASHTAG1~"
const SLEEP_SHORT int = 3
const SLEEP_LONG int = 17

type TwitterConfig struct {
	Accounts []TwitterAccount
}

type TwitterAccount struct {
	ConsumerKey    string
	ConsumerSecret string
	OAuthToken     string
	OAuthSecret    string
	Name           string
}

type TwitterTweet struct {
	Hashtag string
	Text    string
}

// A place to hold known tweets, in order to avoid posting duplicate tweets
var knownTweets map[int64]bool = make(map[int64]bool)

func main() {
	config := TwitterConfig{}

	log.Printf("Reading config.yaml")
	viper.SetConfigFile("config.yaml")
	viper.ReadInConfig()
	viper.Unmarshal(&config)
	log.Printf("Done reading config.yaml")

	for true {
		for _, account := range config.Accounts {
			log.Printf("Poisoning trends from %s", account.Name)
			poisonTrends(account)

			r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
			sleep := time.Duration(r1.Intn(SLEEP_SHORT)) * time.Minute

			log.Printf("Done poisoning trends from %s", account.Name)
			log.Printf("Sleeping %s before next account", sleep)
			time.Sleep(sleep)
		}

		r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
		sleep := time.Duration(r1.Intn(SLEEP_LONG)) * time.Minute

		log.Printf("Done poisoning from all accounts")
		log.Printf("Sleeping %s before restarting", sleep)
		time.Sleep(sleep)
	}
}

// poisonTrends - Find trends, find tweets, truffleShuffle, post nonsense
func poisonTrends(account TwitterAccount) {
	config := oauth1.NewConfig(account.ConsumerKey, account.ConsumerSecret)
	token := oauth1.NewToken(account.OAuthToken, account.OAuthSecret)
	client := twitter.NewClient(config.Client(oauth1.NoContext, token))
	trends := findTrends(client)

	poison(client, trends)
}

/// findTrends - Get the current trends, shuffling the values returned by Twitter
func findTrends(client *twitter.Client) []string {
	trends := make([]string, 0, TRENDLEN)
	trendLists, _, _ := client.Trends.Place(USA, nil)
	for _, trendList := range trendLists {
		for _, trend := range trendList.Trends {
			// Let's only take trending topics with hashtags
			name := strings.TrimSpace(trend.Name)
			if !strings.HasPrefix(name, HASHTAG) || name == EMPTY {
				continue
			}

			trends = append(trends, name)
		}
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(trends) - 1; i > 0; i-- {
		j := random.Intn(i + 1)
		trends[i], trends[j] = trends[j], trends[i]
	}

	return trends
}

// poison - Poison the trends on Twitter
func poison(client *twitter.Client, trends []string) {
	trendTweetMap := make(map[string]TwitterTweet)
	for _, trend := range trends {
		tweet := searchTweets(client, trend)
		trendTweetMap[trend] = tweet
	}

	// Shuffle the map
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := make([]string, 0, len(trendTweetMap))
	for k := range trendTweetMap {
		keys = append(keys, k)
	}

	for i := len(keys) - 1; i > 0; i-- {
		j := random.Intn(i + 1)
		trendTweetMap[keys[i]], trendTweetMap[keys[j]] = trendTweetMap[keys[j]], trendTweetMap[keys[i]]
	}

	// Iterate, manipulate, and tweet
	for trend, tweet := range trendTweetMap {
		log.Printf("    Poisoning trend: %s", trend)

		// Skip the tweet if it already contains the trend
		if strings.Contains(tweet.Text, trend) {
			continue
		}

		// Manipulate the text of the tweet
		status := tweet.Text

		// Replace the trend w/a placeholder
		status = strings.Replace(status, tweet.Hashtag, REPLACE, -1)

		// Remove mentions, other hashtags, etc
		re1 := regexp.MustCompile("\\B[@#]\\S+\\b\\s?")
		re2 := regexp.MustCompile("RT : ")
		status = re1.ReplaceAllString(status, EMPTY)
		status = re2.ReplaceAllString(status, EMPTY)

		// Replace the placeholder w/a new trend
		status = strings.Replace(status, REPLACE, trend, -1)

		// Post a new tweet
		postTweet(client, status)

		log.Printf("       - Original Tweet: %s", tweet.Text)
		log.Printf("       - New Tweet: %s", status)

		r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
		sleep := time.Duration(r1.Intn(SLEEP_SHORT)) * time.Minute

		log.Printf("    Done poisoning trend: %s", trend)
		log.Printf("    Sleeping %s before next trend", sleep)
		time.Sleep(sleep)
	}
}

// searchTweets - Search for new tweets for a given trend
func searchTweets(client *twitter.Client, trend string) TwitterTweet {
	tweet := TwitterTweet{
		Hashtag: trend,
	}

	search, _, _ := client.Search.Tweets(&twitter.SearchTweetParams{
		Query: trend,
	})

	// Find a tweet for the trend, but don't use known tweets from the cycle
	for _, status := range search.Statuses {
		if _, ok := knownTweets[status.ID]; !ok &&
			!strings.HasSuffix(status.Text, DOTDOTDOT) &&
			strings.Contains(status.Text, trend) {

			knownTweets[status.ID] = true
			tweet.Text = status.Text
			break
		}
	}

	return tweet
}

/// postTweet - Post the new status update to Twitter
func postTweet(client *twitter.Client, status string) {
	client.Statuses.Update(status, nil)
}
