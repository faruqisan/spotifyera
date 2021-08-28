package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"

var (
	auth = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(
		spotifyauth.ScopeUserReadPrivate,
		spotifyauth.ScopePlaylistModifyPrivate,
		spotifyauth.ScopePlaylistReadPrivate,
		spotifyauth.ScopeUserLibraryRead,
	))
	ch = make(chan *spotify.Client)
	// TODO: handle map of user state currently use 1 time generate
	state = uuid.New().String()
)

func StartAuthProcess() (*spotify.Client, error) {
	// first start an HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", completeAuth)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("http server started")
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	// shutdown auth server
	err := server.Shutdown(ctxShutDown)
	if err != nil && err != http.ErrServerClosed {
		return nil, err
	}
	log.Println("auth completed, http server now closed")
	return client, nil
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))
	fmt.Fprintf(w, "Login Completed! token: %+v", tok)
	ch <- client
}

func GetUser(ctx context.Context, client *spotify.Client) (*spotify.PrivateUser, error) {
	return client.CurrentUser(ctx)
}
