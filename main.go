package main

import (
	"net/http"
	"path/filepath"
	"io/ioutil"
	"os"
	"fmt"
	"strconv"
	"time"
	"log"

	"github.com/bogem/id3v2"
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

	err = writeTags(album)
	if err != nil {
		return
	}
}

func writeTags(album DeezerAlbum) error {
	albumDir, err := ioutil.ReadDir(os.Args[1])
	if err != nil {
		log.Fatal(err)
		return err
	}

	for i, file := range albumDir {
		if file.IsDir() {
			continue
		}

		fileName := filepath.Join(os.Args[1], file.Name())
		fmt.Println(fileName)

		mp3, err := id3v2.Open(fileName, id3v2.Options{
			Parse: false,
		})
		defer mp3.Close()

		if err != nil {
			log.Fatal(err)
			return err
		}

		trackName := album.Tracks.Data[i].Title

		mp3.SetTitle(trackName)
		mp3.SetAlbum(album.Title)
		mp3.SetArtist(album.Artist.Name)
		mp3.SetYear(album.Released)
		mp3.SetGenre(album.Genres.Data[0].Name)

		mp3.Save()
	}

	return nil
}

func showUsage() {
	fmt.Println()

	fmt.Println("'scribe' gets and applies meta-data for MP3 files from Genius.")

	fmt.Println("\nUsage:")
	fmt.Println("\tscribe.exe [path to album]")
	fmt.Println("\t(Ensure your folders are named after the albums they contain!)")

	fmt.Println()
}
