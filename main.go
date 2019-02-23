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
	"encoding/json"

	"github.com/bogem/id3v2"
)

type config struct {
	AlbumPrefix string `json:"album_prefix"`
}

func showUsage() {
	fmt.Println()

	fmt.Println("'scribe' gets and applies meta-data from Deezer to MP3 files.")

	fmt.Println("\nUsage:")
	fmt.Println("\tscribe.exe [path to album]")
	fmt.Println("\t(Ensure your folders are named after the albums they contain!)")

	fmt.Println()
}

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

// Load scribe config file
func loadConfig() config {
	const defaultConfig = `{
	"album_prefix": ""
}`

	configFile := "scribe_config.json"

	// If the config file doesn't exist, make one and write a default config to it
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		newFile, err := os.Create(configFile)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		defer newFile.Close()

		_, err = newFile.Write([]byte(defaultConfig))
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	configText, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("Failed to load config file: ", err)
		os.Exit(1)
	}

	var config config
	err = json.Unmarshal(configText, &config)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return config
}

func main() {
	// Check args and show usage
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	config := loadConfig()

	albumName := strings.TrimPrefix(filepath.Base(os.Args[1]), config.AlbumPrefix)

	fmt.Println("\n\tWorking...\n")

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	albumID, err := searchAlbum(client, albumName)
	if err != nil {
		log.Fatal(err)
		return
	}

	album, err := getAlbum(client, strconv.Itoa(albumID))
	if err != nil {
		log.Fatal(err)
		return
	}

	formatAlbumData(album)

	fmt.Println("Album name: " + album.Title)
	fmt.Println("Artist: " + album.Artist.Name)
	fmt.Println("Released: " + album.Released)

	for i, track := range album.Tracks.Data {
		fmt.Println(strconv.Itoa(i) + ". " + track.Title)
	}

	albumDir, err := ioutil.ReadDir(os.Args[1])
	if err != nil {
		log.Fatal(err)
		return
	}

	for i, file := range albumDir {
		if file.IsDir() {
			continue
		}

		fileName := filepath.Join(os.Args[1], file.Name())
		fileExt  := filepath.Ext(fileName)

		if fileExt == ".mp3" {
			writeTagMP3(fileName, i, album)
		} else {
			log.Fatal(fileExt + " not supported right now.")
			return
		}

		if err != nil {
			return
		}
	}
}
