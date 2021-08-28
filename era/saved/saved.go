package saved

import (
	"context"
	"sort"

	"github.com/zmb3/spotify/v2"
)

func GetSavedTracks(ctx context.Context, client *spotify.Client) ([]spotify.SavedTrack, error) {
	savedTracks := []spotify.SavedTrack{}
	trackQuery, err := client.CurrentUsersTracks(context.Background())
	if err != nil {
		return savedTracks, err
	}
	// add first items
	savedTracks = append(savedTracks, trackQuery.Tracks...)
	// continue query
	for page := 1; ; page++ {
		err = client.NextPage(context.Background(), trackQuery)
		if err == spotify.ErrNoMorePages {
			savedTracks = append(savedTracks, trackQuery.Tracks...)
			break
		}
		if err != nil {
			return savedTracks, err
		}
		savedTracks = append(savedTracks, trackQuery.Tracks...)
	}
	return savedTracks, nil
}

func GroupTracksByEra(savedTracks []spotify.SavedTrack) map[int][]spotify.ID {
	eraMaps := map[int][]spotify.ID{}
	for _, track := range savedTracks {
		year, _, _ := track.Album.ReleaseDateTime().Date()
		// divide track by 10 to elimate last number
		// eg 1990/10 = 199 so all year start with 199 will be grouped
		// eg 2010/10 = 201
		yearDivided := year / 10
		eraMaps[yearDivided] = append(eraMaps[yearDivided], track.ID)
	}
	return eraMaps
}

func GetSortedEra(tracksEraMap map[int][]spotify.ID) []int {
	keys := make([]int, 0, len(tracksEraMap))
	for k := range tracksEraMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
