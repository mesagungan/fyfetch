//_ "github.com/joho/godotenv/autoload"

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/zmb3/spotify/v2"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"

var (
	auth  = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserReadCurrentlyPlaying, spotifyauth.ScopeUserTopRead))
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {
	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch

	ctx := context.Background()
	// use the client to make calls that require authorization
	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	playing, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		log.Fatal(err)
	}
	topArtist, err := client.CurrentUsersTopArtists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	topTracks, err := client.CurrentUsersTopTracks(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("username: ", user.DisplayName)

	fmt.Println("track artist: ", playing.Item.Artists[0].Name)
	fmt.Println("track album: ", playing.Item.Album.Name)
	fmt.Println("track name: ", playing.Item.Name)

	fmt.Println("top artist: ", topArtist.Artists[0].Name)
	fmt.Println("top track: ", topTracks.Tracks[0].Name)

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
	fmt.Fprintf(w, "Login Completed!")
	ch <- client
}
