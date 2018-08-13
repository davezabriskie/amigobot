package fantano

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/ryanmiville/amigobot"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

//Handler handles the ?remindme [duration] command
type Handler struct {
}

//Command is the trigger for the remindme handler
func (h *Handler) Command() string {
	return "?fantano "
}

//TODO parse out a query term
//TODO look into how to hyperlink, otherwise think about how to style
//TODO refactor to no be horrible
//TODO figure out how to safeguard api key
//Handle parses the ?remindme message and notifies the user
func (h *Handler) Handle(s amigobot.Session, m *discordgo.MessageCreate) {
	queryTerm := "Death Grips album review -podcast -weekly -top"
	videos := []YoutubeVideoData{}

	var (
		query      = flag.String("query", queryTerm, "Search term")
		maxResults = flag.Int64("max-results", 5, "Max YouTube results")
	)

	developerKey := "<<SOME_KEY>>"

	flag.Parse()

	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(*query).
		ChannelId("UCt7fwAhXDy3oNFTAzF2o8Pw").
		Order("date").
		MaxResults(*maxResults)
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error making search API call: %v", err)
	}

	// Iterate through each item and add it to the correct list.
	for _, item := range response.Items {
		videoMetadata := YoutubeVideoData{
			Id:          item.Id.VideoId,
			Title:       item.Snippet.Title,
			Description: GetDescription(item.Id.VideoId, service),
		}
		videos = append(videos, videoMetadata)
	}

	var buffer bytes.Buffer
	for _, video := range videos {
		rating := video.GetRating()
		videoMessage := fmt.Sprintf("[%s](<https://www.youtube.com/watch?v=%s>) : %s\n", video.Title, video.Id, rating)
		buffer.Write([]byte(videoMessage))
	}
	s.ChannelMessageSend(m.ChannelID, buffer.String())
}

func GetDescription(videoId string, service *youtube.Service) string {
	fmt.Print(videoId)
	call := service.Videos.List("snippet").Id(videoId).MaxResults(1)
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error making desc API call: %v", err)
	}
	if len(response.Items) == 0 {
		return ""
	}
	return response.Items[0].Snippet.Description
}

func (v *YoutubeVideoData) GetRating() string {
	var rating = regexp.MustCompile(`\n([0-9]{1,2}\/[0-9]{2})`)
	possibleRatings := rating.FindStringSubmatch(v.Description)
	if !(len(possibleRatings) > 0) {
		return "??/??"
	}
	return possibleRatings[len(possibleRatings)-1]
}
