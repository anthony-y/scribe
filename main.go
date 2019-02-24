package main

import (
	"strings"
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

func formatAlbumData(album *deezerAlbum) {
	// Change to release year instead of release date
	album.Released = album.Released[:4]

	// Make artist title case (sometimes Deezer stores artists in all upper or lower case)
	album.Artist.Name = strings.Title(strings.ToLower(album.Artist.Name))
}

// Write meta-data to an mp3 file
func writeTagMP3(fileName string, trackNum int, album *deezerAlbum) {
	mp3, err := id3v2.Open(fileName, id3v2.Options{
		Parse: false,
	})
	defer mp3.Close()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	trackName := album.Tracks.Data[trackNum].Title

	mp3.SetTitle(trackName)
	mp3.SetAlbum(album.Title)
	mp3.SetArtist(album.Artist.Name)
	mp3.SetYear(album.Released)
	mp3.SetGenre(album.Genres.Data[0].Name)

	mp3.Save()
}

// Get meta-data and tag an individual album
func tagAlbum(path string) {
	albumName := filepath.Base(path)

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	albumID, err := searchAlbum(client, albumName)
	if err != nil {
		log.Fatal(err)
	}

	album, err := getAlbum(client, strconv.Itoa(albumID))
	if err != nil {
		log.Fatal(err)
	}

	formatAlbumData(album)

	albumDir, err := ioutil.ReadDir(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if len(albumDir) != len(album.Tracks.Data) {
		log.Fatal("No Deezer results for " + albumName)
	}

	for i, file := range albumDir {
		if file.IsDir() {
			continue
		}

		fileName := filepath.Join(os.Args[1], file.Name())
		fileExt  := filepath.Ext(fileName)

		switch fileExt {
		case ".mp3": writeTagMP3(fileName, i, album)
		default:
			log.Fatal(fileExt + " not supported right now.")
		}

		if err != nil {
			return
		}
	}
}

func main() {
	// Check args and show usage
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	fmt.Println("\n\tWorking...\n")

	tagAlbum(os.Args[1])
}

func showUsage() {
	fmt.Println()

	fmt.Println("'scribe' gets and applies meta-data from Deezer to MP3 files.")

	fmt.Println("\nUsage:")
	fmt.Println("\tscribe.exe [path to album]")
	fmt.Println("\t(Ensure your folders are named after the albums they contain!)")

	fmt.Println()
}
