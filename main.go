package main

import (
	"net/http"
	"path/filepath"
	"os"
	"fmt"
	"strconv"
	"time"
)

func main() {
	// Check args and show usage
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	albumName := filepath.Base(os.Args[1])

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	fmt.Println("\n\tWorking...\n")

	albumID, err := searchAlbum(client, albumName)
	if err != nil {
		return
	}

	album, err := getAlbum(client, strconv.Itoa(albumID))
	if err != nil {
		return
	}

	fmt.Println("Album name: " + album.Title)
	fmt.Println("Artist: " + album.Artist.Name)
	fmt.Println("Released: " + album.Released)

	for i, track := range album.Tracks.Data {
		fmt.Println(strconv.Itoa(i) + ". " + track.Title)
	}
}

func showUsage() {
	fmt.Println()

	fmt.Println("'scribe' gets and applies meta-data for MP3 files from Genius.")

	fmt.Println("\nUsage:")
	fmt.Println("\tscribe.exe [path to album]")
	fmt.Println("\t(Ensure your folders are named after the albums they contain!)")

	fmt.Println()
}
