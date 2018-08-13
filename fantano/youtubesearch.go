package fantano

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

type YoutubeSearch struct {
	DeveloperKey string
}

func (s *YoutubeSearch) getService() *youtube.Service {
	client := &http.Client{
		Transport: &transport.APIKey{Key: s.DeveloperKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}
	return service
}

func (s *YoutubeSearch) Search(channelId string, query string, batchSize int64) []YoutubeVideoData {
	queryTerm := fmt.Sprintf("\"%s\" album review", strings.TrimSpace(query))
	videos := []YoutubeVideoData{}

	service := s.getService()

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(queryTerm).
		ChannelId(channelId).
		Order("date").
		MaxResults(batchSize)

	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error making search API call: %v", err)
	}

	// Iterate through each item and add it to the correct list.
	for _, item := range response.Items {
		videoMetadata := YoutubeVideoData{
			Id:          item.Id.VideoId,
			Title:       item.Snippet.Title,
			Description: s.GetDescription(item.Id.VideoId),
		}
		videos = append(videos, videoMetadata)
	}

	return videos
}

func (s *YoutubeSearch) GetDescription(videoId string) string {
	service := s.getService()
	call := service.Videos.List("snippet").
		Id(videoId).
		MaxResults(1)
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error making desc API call: %v", err)
	}
	if len(response.Items) == 0 {
		return ""
	}
	return response.Items[0].Snippet.Description
}
