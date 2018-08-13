package fantano

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ryanmiville/amigobot"
)

//Handler handles the ?remindme [duration] command
type Handler struct {
	Searcher YoutubeSearch
}

//Command is the trigger for the remindme handler
func (h *Handler) Command() string {
	return "?fantano"
}

//TODO look into how to hyperlink, otherwise think about how to style
//TODO figure out how to safeguard api key
//Handle parses the ?remindme message and notifies the user
func (h *Handler) Handle(s amigobot.Session, m *discordgo.MessageCreate) {
	msg := strings.TrimPrefix(m.Content, h.Command())
	videos := h.Searcher.Search("UCt7fwAhXDy3oNFTAzF2o8Pw", msg, 1)
	response := buildMessage(videos)
	s.ChannelMessageSend(m.ChannelID, response)
}

func buildMessage(videos []YoutubeVideoData) string {
	var buffer bytes.Buffer
	for _, video := range videos {
		rating := video.GetRating()
		videoMessage := fmt.Sprintf("[%s](<https://www.youtube.com/watch?v=%s>) : %s\n", video.Title, video.Id, rating)
		buffer.Write([]byte(videoMessage))
	}
	return buffer.String()
}

func (v *YoutubeVideoData) GetRating() string {
	var ratingRegex = regexp.MustCompile(`\n([0-9]{1,2}\/[0-9]{2})`)
	possibleRatings := ratingRegex.FindStringSubmatch(v.Description)
	if !(len(possibleRatings) > 0) {
		return "??/??"
	}
	rating := possibleRatings[len(possibleRatings)-1]
	if rating == "7/10" {
		rating = "DAMN/10"
	}
	return rating
}
