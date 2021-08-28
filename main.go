package main

import (
	"context"
	"fmt"
	"log"
	"spotifyera/era/auth"
	"spotifyera/era/playlist"
	"spotifyera/era/saved"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
)

const (
	playlistNameFormat = "spotifyera - %s - (%d)"
	playlistDescFormat = "auto generated playlist based on release year from your liked song, this is era: %s"
)

func main() {

	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// get client by start auth process
	client, err := auth.StartAuthProcess()
	if err != nil {
		log.Fatal(err)
	}

	user, err := auth.GetUser(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	// get user saved tracks
	log.Println("fetching user's saved tracks")
	savedTracks, err := saved.GetSavedTracks(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	// map tracks to it's era
	tracksEraMap := saved.GroupTracksByEra(savedTracks)
	sortedEraKeys := saved.GetSortedEra(tracksEraMap)

	// get user's playlists
	log.Println("fetching user's playlists")
	userPlaylists, err := playlist.GetUserPlaylists(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	// map playlist name to id
	playlistNameMap := playlist.PlaylistToMapIDName(userPlaylists)

	// TODO: go routine each era
	for _, era := range sortedEraKeys {
		tracks := tracksEraMap[era]
		var (
			eraClean       = fmt.Sprintf("%d0s", era)
			playlistName   = fmt.Sprintf(playlistNameFormat, eraClean, len(tracks))
			playl          = new(spotify.FullPlaylist)
			playlistExists = false
		)

		// check if playlist with same name exists
		if id, ok := playlistNameMap[playlistName]; ok {
			playlistExists = true
			log.Println("playlist: ", playlistName, " exists, fetching...")
			playl, err = playlist.GetPlaylist(ctx, client, id)
			if err != nil {
				log.Fatal("err on get playlist: ", playlistName, ": ", err)
			}
		}

		if !playlistExists {
			log.Println("playlist: ", playlistName, " doesn't exists, creating...")
			desc := fmt.Sprintf(playlistDescFormat, eraClean)
			playl, err = playlist.CreatePlaylist(ctx, client, user.ID, playlistName, desc, false)
			if err != nil {
				log.Fatal("err on create playlist: ", playlistName, ": ", err)
			}
		}

		log.Println("fetching playlist's tracks")
		playlistTracks, err := playlist.GetPlaylistTracks(ctx, client, playl)
		if err != nil {
			log.Fatal("err on fetch playlist tracks: ", playlistName, ": ", err)
		}

		// remove duplicate tracks before insert it
		// TODO: remove tracks from insert instead of remove existing
		// this was for cleaning up existing playlist that has duplicate song
		playlistTracksBlankMap := playlist.PlaylistTrackBlankMap(playlistTracks)
		duplicateTracks := playlist.FindDuplicaTracks(tracks, playlistTracksBlankMap)
		if len(duplicateTracks) > 0 {
			log.Println("deleting playlist's duplicate tracks")
			err = playlist.DeleteTracksFromPlaylist(ctx, client, duplicateTracks, playl.ID)
			if err != nil {
				log.Fatal("err on delete duplicate tracks on playlist: ", playlistName, ": ", err)
			}
		}

		log.Println("adding tracks to playlist: ", playlistName)
		err = playlist.AddTracksToPlaylist(ctx, client, tracks, playl.ID)
		if err != nil {
			log.Fatal("err on adding tracks to playlist: ", playlistName, ": ", err)
		}
	}
}
