package main

import (
	"./term"
	"bufio"
	"bytes"
	_ "crypto/sha512"
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb"
	"io"
	flag "launchpad.net/gnuflag"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

var VERSION string
var MINVERSION string
var OAUTH_CLIENT_ID string
var OAUTH_CLIENT_SECRET string

var OAUTH_REDIRECT_URI = "http://www.rhythmandpoetry.net/mixcloud_code.php"
var API_URL = "https://api.mixcloud.com/upload/?access_token="
var ACCESS_TOKEN_URL = "https://www.mixcloud.com/oauth/access_token?client_id=" + OAUTH_CLIENT_ID + "&redirect_uri=" + OAUTH_REDIRECT_URI + "&client_secret=" + OAUTH_CLIENT_SECRET + "&code=%s"
var CONFIG_FILE = "config.json"
var CONFIG_FILE_PATH = ""

var TRACKLIST_OUTPUT_FORMAT = "%d. %s-%s\n"

var configuration = Configuration{}

var aboutFlag = flag.Bool("about", false, "About the application")
var configFlag = flag.Bool("config", false, "Configure the application")
var fileFlag = flag.String("file", "", "The mp3 file to upload to mixcloud")
var coverFlag = flag.String("cover", "", "The image file to upload to mixcloud as the cover")
var trackListFlag = flag.String("tracklist", "", "A file containing a VirtualDJ Tracklist for the cloudcast")

var STD_OUT = bufio.NewWriter(os.Stdout)
var STD_ERR = bufio.NewWriter(os.Stderr)
var STD_IN = bufio.NewReader(os.Stdin)

type Configuration struct {
	ACCESS_TOKEN string
	DEFAULT_TAGS string
}

type Track struct {
	Artist   string
	Song     string
	Duration int
	Cover		 string
}

func showWelcomeMessage() {
	OutputMessage(term.Green + "Mixcloud CLI Uploader v" + VERSION + term.Reset + "\n\n")
}

func showAboutMessage() {
	OutputMessage(fmt.Sprintf("Build Number: %s\n", MINVERSION))
	OutputMessage("Created by: Greg Tangey (http://ignite.digitalignition.net/)\n")
	OutputMessage("Website: http://www.rhythmandpoetry.net/\n")
}

func createConfig() {
	OutputMessage("Creating Configuration File...\n")
	OutputMessage("Please visit the URL below\n\nhttps://www.mixcloud.com/oauth/authorize?client_id=z3CWHgULyawutvpcD3&redirect_uri=http://www.rhythmandpoetry.net/mixcloud_code.php\n")

	OutputMessage("Enter the provided code: ")
	code, err := STD_IN.ReadString('\n')
	if err != nil {
		OutputError("Code Error.")
		os.Exit(2)
	}

	code = strings.TrimSpace(code)
	access_token := fetchAccessCode(code)
	if access_token == "" {
		OutputError("Error fetching access token")
		os.Exit(2)
	} else {
		configuration.ACCESS_TOKEN = access_token
	}

	OutputMessage("Enter default tags (comma separated): ")
	tags, err := STD_IN.ReadString('\n')
	if err != nil {
		OutputError("Incorrect tag format.")
		os.Exit(2)
	} else {
		configuration.DEFAULT_TAGS = strings.TrimSpace(tags)
	}

	saveConfig()
}

func fetchAccessCode(code string) string {
	url := fmt.Sprintf(ACCESS_TOKEN_URL, code)

	request, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		OutputError("Error request Access Code")
		os.Exit(2)
	}

	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, doError := client.Do(request)
	if doError != nil {
		OutputError("Error fetching Access Code: " + doError.Error())
		os.Exit(2)
	}

	var jsonResponse map[string]interface{}
	jsonError := json.NewDecoder(resp.Body).Decode(&jsonResponse)
	resp.Body.Close()
	if jsonError != nil {
		OutputError("Error decoding response from API - " + jsonError.Error())
		os.Exit(2)
	}

	var access_token = ""

	if jsonResponse["access_token"] != nil {
		access_token = jsonResponse["access_token"].(string)
	}

	return access_token
}

func saveConfig() {
	file, error := os.Create(CONFIG_FILE)
	defer file.Close()
	if error != nil {
		OutputError(fmt.Sprintf("Unable to save configuration file conf.json - ", error))
		os.Exit(2)
	}

	encoder := json.NewEncoder(file)

	err := encoder.Encode(&configuration)
	if err != nil {
		OutputError(fmt.Sprintf("Error writing to config file: %s", err))
		os.Exit(2)
	} else {
		OutputMessage(term.Green + "Configuration saved." + term.Reset + "\n")
	}
}

func loadConfig() {
	file, error := os.Open(CONFIG_FILE)
	defer file.Close()

	if error != nil {
		//Config file doesn't exist, create
		createConfig()
	} else {
		decoder := json.NewDecoder(file)

		err := decoder.Decode(&configuration)
		if err != nil {
			fmt.Println("Error reading config file: ", err)
			os.Exit(2)
		}
	}

	if configuration.ACCESS_TOKEN == "" {
		OutputError("Access Token configuration missing.")
		os.Exit(2)
	}
}

func setupApp() {
	usr, _ := user.Current()
	CONFIG_FILE_PATH = filepath.Join(usr.HomeDir, ".mixcloud")
	CONFIG_FILE = filepath.Join(CONFIG_FILE_PATH, CONFIG_FILE)

	if _, err := os.Stat(CONFIG_FILE_PATH); os.IsNotExist(err) {
		os.Mkdir(CONFIG_FILE_PATH, 0700)
	}
}

func main() {
	flag.Parse(true)

	showWelcomeMessage()
	if *aboutFlag == true {
		showAboutMessage()
		os.Exit(0)
	}

	setupApp()
	loadConfig()

	var tracklist []Track

	if *configFlag == true {
		createConfig()
	}

	if *trackListFlag != "" {
		tracklist = parseVirtualDJTrackList(trackListFlag)
	}

	if *fileFlag == "" {
		OutputError("You must pass a file to upload, use --file or see --help.\n Exiting.")
		os.Exit(2)
	}

	OutputMessage("Enter a name for the cloudcast: ")
	cast_name, err := STD_IN.ReadString('\n')
	if err != nil {
		OutputError("Incorrect name.")
		os.Exit(2)
	}

	OutputMessage("Enter a description: ")
	cast_desc, err := STD_IN.ReadString('\n')
	if err != nil {
		OutputError("Incorrect description.")
		os.Exit(2)
	}

	OutputMessage(fmt.Sprintf("Enter tags (comma separated) [%s]: ", configuration.DEFAULT_TAGS))
	cast_tags, err := STD_IN.ReadString('\n')
	if err != nil {
		OutputError("Incorrect tag format.")
		os.Exit(2)
	}

	if cast_tags == "" || cast_tags == "\n" {
		cast_tags = configuration.DEFAULT_TAGS
	}
	tags_arr := strings.Split(cast_tags, ",")

	request, bar := HttpUploadRequest(*fileFlag, *coverFlag, cast_name, cast_desc, tags_arr, tracklist)
	bar.Empty = term.Red + "-" + term.Reset
	bar.Current = term.Green + "=" + term.Reset
	client := &http.Client{}
	OutputMessage("\n\n")
	STD_OUT.Flush()
	bar.Start()
	resp, err := client.Do(request)
	if err != nil {
		OutputError("Error: " + err.Error())
		os.Exit(2)
	}
	bar.Finish()

	var jsonResponse map[string]interface{}
	error := json.NewDecoder(resp.Body).Decode(&jsonResponse)
	resp.Body.Close()
	if error != nil {
		OutputError("Error decoding response from API - " + error.Error())
		os.Exit(2)
	}

	if handleJSONResponse(jsonResponse) {
		printTracklist(tracklist)
	} else {
		os.Exit(2)
	}
}

func printTracklist(tracklist []Track) {
	OutputMessage("Tracklist\n")
	for i, track := range tracklist {
		OutputMessage(fmt.Sprintf(TRACKLIST_OUTPUT_FORMAT,i+1,track.Artist,track.Song))
	}
}

func parseVirtualDJTrackList(tracklist *string) []Track {
	var list []Track

	fin, err := os.Open(*tracklist)
	if err != nil {
		fmt.Fprintf(os.Stderr, "The file %s does not exist!\n", tracklist)
		return nil
	}
	defer fin.Close()

	bufReader := bufio.NewReader(fin)
	var last_track_time_str string = ""

	for line, _, err := bufReader.ReadLine(); err != io.EOF; line, _, err = bufReader.ReadLine() {
		data := strings.Split(string(line), " : ")
		tracktimestr, track := data[0], data[1]

		thistrack := new(Track)

		trackdata := strings.SplitN(string(track), " - ", 2)
		thistrack.Artist = trackdata[0]
		thistrack.Song = trackdata[1]

		last_time, _ := time.Parse("15:04", last_track_time_str)
		track_time, err := time.Parse("15:04", tracktimestr)
		if err != nil {
			OutputError("Unable to parse time." + err.Error())
			os.Exit(2)
		}

		if last_track_time_str != "" {
			duration := track_time.Sub(last_time)
			thistrack.Duration = int(duration.Seconds())
		}
		last_track_time_str = tracktimestr

		list = append(list, *thistrack)

		// if !isPrefix {
		//   fmt.Printf("Lines: %s (error %v)\n", string(bytes), err)
		//   bytes = bytes[:0]
		// }

	}

	return list
}

<<<<<<< HEAD
func handleJSONResponse(jsonResponse map[string]interface{}) bool {
=======
func handleJSONResponse(jsonResponse map[string]interface{}) Boolean {
>>>>>>> 5ba6bdc92303728b3523ef36c473a2168ac7843a
	if error := jsonResponse["error"]; error != nil {
		OutputError(error.(map[string]interface{})["message"].(string))
		return false
	} else {
		OutputMessage(term.Green + "Sucessfully uploaded file" + term.Reset + "\n")
		path := jsonResponse["result"].(map[string]interface{})["key"].(string)
		OutputMessage(term.Green + "https://mixcloud.com" + path + "edit" + term.Reset + "\n")
		return true
	}
}

func OutputError(message string) {
	STD_ERR.WriteString(term.Bold + term.Red + message + term.Reset + "\n")
	STD_ERR.Flush()
}

func OutputMessage(message string) {
	STD_OUT.WriteString(message)
	STD_OUT.Flush()
}

func loadFileToWriter(file string, key string, writer *multipart.Writer) {
	f, err := os.Open(file)
	if err != nil {
		OutputError("Error opening file " + file + "\n")
		os.Exit(2)
	}
	defer f.Close()

	fw, err := writer.CreateFormFile(key, file)
	if err != nil {
		OutputError("Error reading file " + file + "\n")
		os.Exit(2)
	}

	if _, err = io.Copy(fw, f); err != nil {
		OutputError("Error opening file " + file + " to buffer\n")
		os.Exit(2)
	}
}

func HttpUploadRequest(file string, cover string, name string, desc string, tag_list []string, tracklist []Track) (*http.Request, *pb.ProgressBar) {
	b := &bytes.Buffer{}
	writer := multipart.NewWriter(b)

	// Add information name/description
	writer.WriteField("name", name)
	writer.WriteField("description", desc)

	// Add tags
	for i, tag := range tag_list {
		field_name := fmt.Sprintf("tags-%d-tag", i)
		writer.WriteField(field_name, tag)
	}

	// Add tracklist
	if tracklist != nil {
		var total_duration int = 0

		for i, track := range tracklist {
			artist_field_name := fmt.Sprintf("sections-%d-artist", i)
			song_field_name := fmt.Sprintf("sections-%d-song", i)
			duration_field_name := fmt.Sprintf("sections-%d-start_time", i)

			total_duration += track.Duration

			writer.WriteField(artist_field_name, track.Artist)
			writer.WriteField(song_field_name, track.Song)
			writer.WriteField(duration_field_name, fmt.Sprintf("%d", total_duration))
		}
	}

	// Add MP3
	if file != "" {
		loadFileToWriter(file, "mp3", writer)
	}

	// Add cover image
	if cover != "" {
		loadFileToWriter(cover, "picture", writer)
	}

	writer.Close()

	url := API_URL + configuration.ACCESS_TOKEN

	var bar = pb.New(b.Len()).SetUnits(pb.U_BYTES)
	reader := bar.NewProxyReader(b)

	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		OutputError("Error building request")
		os.Exit(2)
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())

	return request, bar
}
