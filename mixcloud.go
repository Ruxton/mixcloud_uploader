package main

import (
	"bytes"
	_ "crypto/sha512"
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/ruxton/mixcloud/mixcloud"
	"github.com/ruxton/mixcloud/parsers"
	"github.com/ruxton/mixcloud/versions"
	"github.com/ruxton/term"
	"io"
	flag "github.com/juju/gnuflag"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

var API_URL = "https://api.mixcloud.com/upload/?access_token="
var EDIT_URL = "https://api.mixcloud.com/upload/%s/%s/edit/?access_token=%s"
var CONFIG_FILE = "config.json"
var CONFIG_FILE_PATH = ""

var TRACKLIST_TRACK_OUTPUT_FORMAT = "%d. %s - %s [%d] (%s)\n"
var TRACKLIST_CHAPTER_OUTPUT_FORMAT = "%d. %s [%d] (%s)\n"

var configuration = mixcloud.Configuration{}

var aboutFlag = flag.Bool("about", false, "About the application")
var configFlag = flag.Bool("config", false, "Configure the application")
var fileFlag = flag.String("file", "", "The mp3 file to upload to mixcloud")
var coverFlag = flag.String("cover", "", "The image file to upload to mixcloud as the cover")
var trackListFlag = flag.String("tracklist", "", "A file containing a VirtualDJ Tracklist for the cloudcast")
var trackListTypeFlag = flag.String("tracklist-type", "virtualdj", "The tracklist type to parse (virtualdj,serato,traktor,reaper,cue)")
var editFlag = flag.String("edit", "", "The id of the track to edit")

func showWelcomeMessage() {
	term.OutputMessage(term.Green + "Mixcloud CLI Uploader v" + versions.VERSION + term.Reset + "\n\n")
}

func showAboutMessage() {
	term.OutputMessage(fmt.Sprintf("Build Number: %s\n", versions.MINVERSION))
	term.OutputMessage("Created by: Greg Tangey (http://ignite.digitalignition.net/)\n")
	term.OutputMessage("Website: http://www.rhythmandpoetry.net/\n")
}

func createConfig() {
	term.OutputMessage("Creating Configuration File...\n")
	term.OutputMessage("Please visit the URL below\n\nhttps://www.mixcloud.com/oauth/authorize?client_id="+mixcloud.OAUTH_CLIENT_ID+"&redirect_uri=http://www.rhythmandpoetry.net/mixcloud_code.php\n")

	term.OutputMessage("Enter the provided code: ")
	code, err := term.STD_IN.ReadString('\n')
	if err != nil {
		term.OutputError("Code Error.")
		os.Exit(2)
	}

	code = strings.TrimSpace(code)
	term.OutputMessage(code)
	access_token := mixcloud.FetchAccessCode(code)
	if access_token == "" {
		term.OutputError("Error fetching access token")
		os.Exit(2)
	} else {
		configuration.ACCESS_TOKEN = access_token
	}

	term.OutputMessage("Enter default tags (comma separated): ")
	tags, err := term.STD_IN.ReadString('\n')
	if err != nil {
		term.OutputError("Incorrect tag format.")
		os.Exit(2)
	} else {
		configuration.DEFAULT_TAGS = strings.TrimSpace(tags)
	}

	saveConfig()
}

func saveConfig() {
	file, error := os.Create(CONFIG_FILE)
	defer file.Close()
	if error != nil {
		term.OutputError(fmt.Sprintf("Unable to save configuration file conf.json - ", error))
		os.Exit(2)
	}

	encoder := json.NewEncoder(file)

	err := encoder.Encode(&configuration)
	if err != nil {
		term.OutputError(fmt.Sprintf("Error writing to config file: %s", err))
		os.Exit(2)
	} else {
		term.OutputMessage(term.Green + "Configuration saved." + term.Reset + "\n")
	}
}

func loadConfig() {
	file, error := os.Open(CONFIG_FILE)
	defer file.Close()

	if error != nil {
		//Config file doesn't exist, create
		term.OutputError(error.Error())
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
		term.OutputError("Access Token configuration missing.")
		os.Exit(2)
	}
}

func setupApp() {
	homedir := "/home/ruxton/"
	usr, err := user.Current()
	if err != nil {
		term.OutputError(fmt.Sprintf("Error fetching local user - %s", err.Error()))
	} else {
		homedir = usr.HomeDir
	}
	CONFIG_FILE_PATH = filepath.Join(homedir, ".mixcloud")
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

	mixcloud.CURRENT_USER = mixcloud.FetchMe(configuration.ACCESS_TOKEN)

	var tracklist []mixcloud.Track

	if *configFlag == true {
		createConfig()
	}

	if *trackListFlag != "" {
		if *trackListTypeFlag == "virtualdj" {
			tracklist = parsers.ParseVirtualDJTrackList(trackListFlag)
		} else if *trackListTypeFlag == "serato" {
			tracklist = parsers.ParseSeratoTrackList(trackListFlag)
    } else if *trackListTypeFlag == "reaper" {
      tracklist = parsers.ParseReaperMarkerList(trackListFlag)
		} else if *trackListTypeFlag == "traktor" {
			term.OutputError("Traktor tracklists are currently not supported.\n Exiting.")
			os.Exit(2)
		}
	}

	printTracklist(tracklist)

  if *editFlag != "" {
    editMix(*editFlag,*coverFlag,tracklist)
  } else {
    uploadNewMix(*fileFlag,*coverFlag,tracklist)
  }

}

func editMix(mix string, cover string, tracklist []mixcloud.Track) {

    if(len(tracklist) == 0) {
      term.OutputError("You must provide a tracklist when editing.")
      os.Exit(2)
    }

    b := &bytes.Buffer{}
  	writer := multipart.NewWriter(b)

    mixcloud.AddBasicDataToHTTPWriter(configuration, writer, tracklist, true)
    if cover != "" {
  		loadFileToWriter(cover, "picture", writer)
  	}

    url := fmt.Sprintf(EDIT_URL,mixcloud.CURRENT_USER.Username,mix,configuration.ACCESS_TOKEN)

    term.OutputMessage(url+"\n")

    request,bar := HttpEditRequest(url,b, writer)
    bar.Empty = term.Red + "-" + term.Reset
  	bar.Current = term.Green + "=" + term.Reset
  	client := &http.Client{}
  	term.OutputMessage("\n\n")
  	term.STD_OUT.Flush()

  	// Start uploading
  	bar.Start()
  	resp, err := client.Do(request)
  	if err != nil {
  		term.OutputError("Error: " + err.Error())
  		os.Exit(2)
  	}
  	bar.Finish()

    var Response *mixcloud.Response = new(mixcloud.Response)

    error := json.NewDecoder(resp.Body).Decode(&Response)
    resp.Body.Close()
    if error != nil {
      term.OutputError("Error decoding response from API - " + error.Error())
      os.Exit(2)
    }

    if handleJSONResponse(*Response) {
      printTracklist(tracklist)
    } else {
      os.Exit(2)
    }
}

func uploadNewMix(file string, cover string, tracklist []mixcloud.Track) {
  if file == "" {
		term.OutputError("You must pass a file to upload, use --file or see --help.\n Exiting.")
		os.Exit(2)
	}

	b := &bytes.Buffer{}
	writer := multipart.NewWriter(b)

	// Collect user input
	mixcloud.AddBasicDataToHTTPWriter(configuration, writer, tracklist, false)
	mixcloud.AddPremiumDataToHTTPWriter(writer)

	// Add MP3
	if file != "" {
		loadFileToWriter(file, "mp3", writer)
	}
	// Add cover image
	if cover != "" {
		loadFileToWriter(cover, "picture", writer)
	}
	writer.Close()

	// Prepare for uploading
	request, bar := HttpUploadRequest(b, writer)
	bar.Empty = term.Red + "-" + term.Reset
	bar.Current = term.Green + "=" + term.Reset
	client := &http.Client{}
	term.OutputMessage("\n\n")
	term.STD_OUT.Flush()

	// Start uploading
	bar.Start()
	resp, err := client.Do(request)
	if err != nil {
		term.OutputError("Error: " + err.Error())
		os.Exit(2)
	}
	bar.Finish()

	// Uploading finished
	var Response *mixcloud.Response = new(mixcloud.Response)
	error := json.NewDecoder(resp.Body).Decode(&Response)
	resp.Body.Close()
	if error != nil {
		term.OutputError("Error decoding response from API - " + error.Error())
		os.Exit(2)
	}

	if handleJSONResponse(*Response) {
		printTracklist(tracklist)
	} else {
		os.Exit(2)
	}
}

func printTracklist(tracklist []mixcloud.Track) {
	term.OutputMessage("Tracklist\n")
	total_start_time:=0
	for i, track := range tracklist {
    total_start_time = total_start_time+track.Duration
    message := ""

    if(track.Chapter != "") {
      message = fmt.Sprintf(TRACKLIST_CHAPTER_OUTPUT_FORMAT, i+1, track.Chapter, track.Duration, (time.Duration(total_start_time)*time.Second).String() )
    } else {
      message = fmt.Sprintf(TRACKLIST_TRACK_OUTPUT_FORMAT, i+1, track.Artist, track.Song, track.Duration, (time.Duration(total_start_time)*time.Second).String() )
    }

    term.OutputMessage(message)

	}
}

func handleJSONResponse(response mixcloud.Response) bool {
	if response.Error != nil {
		term.OutputError(response.Error.Message)
		fmt.Printf("%v", response.Details)
		return false
	} else if response.Result.Success {
		term.OutputMessage(term.Green + "Sucessfully uploaded file" + term.Reset + "\n")
		path := response.Result.Key
		term.OutputMessage(term.Green + "https://mixcloud.com" + path + "edit" + term.Reset + "\n")
		return true
	} else {
		term.OutputError("Error uploading, no success")
		fmt.Printf("%v", response)
		return false
	}
}

func loadFileToWriter(file string, key string, writer *multipart.Writer) {
	f, err := os.Open(file)
	if err != nil {
		term.OutputError("Error opening file " + file + "\n")
		os.Exit(2)
	}
	defer f.Close()

	fw, err := writer.CreateFormFile(key, file)
	if err != nil {
		term.OutputError("Error reading file " + file + "\n")
		os.Exit(2)
	}

	if _, err = io.Copy(fw, f); err != nil {
		term.OutputError("Error opening file " + file + " to buffer\n")
		os.Exit(2)
	}
}

func HttpEditRequest(url string,b *bytes.Buffer, writer *multipart.Writer) (*http.Request, *pb.ProgressBar) {

	var bar = pb.New(b.Len()).SetUnits(pb.U_BYTES)
	reader := bar.NewProxyReader(b)

	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		term.OutputError("Error building request")
		os.Exit(2)
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())

	return request, bar
}

func HttpUploadRequest(b *bytes.Buffer, writer *multipart.Writer) (*http.Request, *pb.ProgressBar) {

	url := API_URL + configuration.ACCESS_TOKEN

	var bar = pb.New(b.Len()).SetUnits(pb.U_BYTES)
	reader := bar.NewProxyReader(b)

	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		term.OutputError("Error building request")
		os.Exit(2)
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())

	return request, bar
}
