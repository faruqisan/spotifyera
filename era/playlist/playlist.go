package playlist

import (
	"context"
	"math"

	"github.com/zmb3/spotify/v2"
)

func GetUserPlaylists(ctx context.Context, client *spotify.Client) ([]spotify.SimplePlaylist, error) {
	userPlaylists := []spotify.SimplePlaylist{}
	playlistQuery, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return userPlaylists, err
	}
	userPlaylists = append(userPlaylists, playlistQuery.Playlists...)
	for page := 1; ; page++ {
		err = client.NextPage(ctx, playlistQuery)
		if err == spotify.ErrNoMorePages {
			userPlaylists = append(userPlaylists, playlistQuery.Playlists...)
			break
		}
		if err != nil {
			return userPlaylists, err
		}
		userPlaylists = append(userPlaylists, playlistQuery.Playlists...)
	}
	return userPlaylists, nil
}

func PlaylistToMapIDName(userPlaylists []spotify.SimplePlaylist) map[string]spotify.ID {
	// create a map of playlist name to it id
	playlistNameToID := make(map[string]spotify.ID, len(userPlaylists))
	for _, userPlaylist := range userPlaylists {
		playlistNameToID[userPlaylist.Name] = userPlaylist.ID
	}
	return playlistNameToID
}

func GetPlaylist(ctx context.Context, client *spotify.Client, id spotify.ID) (*spotify.FullPlaylist, error) {
	return client.GetPlaylist(ctx, id)
}

func CreatePlaylist(ctx context.Context, client *spotify.Client, userID string, name string, description string, isPublic bool) (*spotify.FullPlaylist, error) {
	return client.CreatePlaylistForUser(ctx, userID, name, description, isPublic, false)
}

func GetPlaylistTracks(ctx context.Context, client *spotify.Client, playlist *spotify.FullPlaylist) ([]spotify.PlaylistTrack, error) {
	playlistTracks := []spotify.PlaylistTrack{}
	playlistTracksQuery := playlist.Tracks

	playlistTracks = append(playlistTracks, playlistTracksQuery.Tracks...)
	for page := 1; ; page++ {
		err := client.NextPage(ctx, &playlistTracksQuery)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			return playlistTracks, err
		}
		playlistTracks = append(playlistTracks, playlistTracksQuery.Tracks...)
	}
	return playlistTracks, nil
}

func PlaylistTrackBlankMap(tracks []spotify.PlaylistTrack) map[spotify.ID]struct{} {
	playlistTracksMap := make(map[spotify.ID]struct{}, len(tracks))
	// build the map
	for _, ptrack := range tracks {
		playlistTracksMap[ptrack.Track.ID] = struct{}{}
	}
	return playlistTracksMap
}

func FindDuplicaTracks(tracks []spotify.ID, trackBlankMap map[spotify.ID]struct{}) []spotify.ID {
	duplicateTracks := []spotify.ID{}
	for _, track := range tracks {
		if _, ok := trackBlankMap[track]; ok {
			duplicateTracks = append(duplicateTracks, track)
		}
	}
	return duplicateTracks
}

func DeleteTracksFromPlaylist(ctx context.Context, client *spotify.Client, tracks []spotify.ID, playlist spotify.ID) error {
	le := len(tracks)
	maxBatchSize := 100
	batchAmount := int(math.Ceil(float64(le / maxBatchSize)))
	if le == maxBatchSize {
		batchAmount -= 1
	}
	skip := 0
	for i := 0; i <= batchAmount; i++ {
		// batches
		lowerBound := skip
		upperBound := skip + maxBatchSize
		if upperBound > le {
			upperBound = le
		}

		batchDeleteTracks := tracks[lowerBound:upperBound]
		// removing duplicate track for playlist
		_, err := client.RemoveTracksFromPlaylist(ctx, playlist, batchDeleteTracks...)
		if err != nil {
			return err
		}
		skip += maxBatchSize
	}
	return nil
}

func AddTracksToPlaylist(ctx context.Context, client *spotify.Client, tracks []spotify.ID, playlist spotify.ID) error {
	le := len(tracks)
	maxBatchSize := 100
	batchAmount := int(math.Ceil(float64(le / maxBatchSize)))
	if le == maxBatchSize {
		batchAmount -= 1
	}
	skip := 0
	for i := 0; i <= batchAmount; i++ {
		// batches
		lowerBound := skip
		upperBound := skip + maxBatchSize
		if upperBound > le {
			upperBound = le
		}

		batchTracks := tracks[lowerBound:upperBound]
		_, err := client.AddTracksToPlaylist(ctx, playlist, batchTracks...)
		if err != nil {
			return err
		}
		skip += maxBatchSize
	}
	return nil
}

func UnfollowPlaylist(ctx context.Context, client *spotify.Client, playlist spotify.ID) error {
	return client.UnfollowPlaylist(ctx, playlist)
}
