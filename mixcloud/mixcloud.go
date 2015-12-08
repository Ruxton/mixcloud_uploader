package mixcloud

import (
	"encoding/json"
	"fmt"
	"github.com/ruxton/mixcloud/confirm"
	"github.com/ruxton/mixcloud/versions"
	"github.com/ruxton/term"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

var OAUTH_CLIENT_ID string
var OAUTH_CLIENT_SECRET string
var OAUTH_REDIRECT_URI = "http://www.rhythmandpoetry.net/mixcloud_code.php"

var API_ME_URL = "https://api.mixcloud.com/me?access_token="
var ACCESS_TOKEN_URL = "https://www.mixcloud.com/oauth/access_token?client_id=" + OAUTH_CLIENT_ID + "&redirect_uri=" + OAUTH_REDIRECT_URI + "&client_secret=" + OAUTH_CLIENT_SECRET + "&code=%s"

var CURRENT_USER User = User{}

func FetchMe(access_token string) User {

	term.OutputMessage(term.Green + "Fetching your user data.." + term.Reset + "\n")

	url := API_ME_URL + access_token

	request := build_http(url, "GET")

	client := http.Client{}
	resp, doError := client.Do(request)
	if doError != nil {
		term.OutputError("Error fetching your profile data: " + doError.Error())
		os.Exit(2)
	}

	var user User

	jsonError := json.NewDecoder(resp.Body).Decode(&user)
	resp.Body.Close()
	if jsonError != nil {
		term.OutputError("Error decoding response from API - " + jsonError.Error())
		os.Exit(2)
	}

	return user

}

func FetchAccessCode(code string) string {
	url := fmt.Sprintf(ACCESS_TOKEN_URL, code)

	request := build_http(url, "GET")

	client := &http.Client{}
	resp, doError := client.Do(request)
	if doError != nil {
		term.OutputError("Error fetching Access Code: " + doError.Error())
		os.Exit(2)
	}

	var jsonResponse map[string]interface{}
	jsonError := json.NewDecoder(resp.Body).Decode(&jsonResponse)
	resp.Body.Close()
	if jsonError != nil {
		term.OutputError("Error decoding response from API - " + jsonError.Error())
		os.Exit(2)
	}

	var access_token = ""

	if jsonResponse["access_token"] != nil {
		access_token = jsonResponse["access_token"].(string)
	}

	return access_token
}

func AddBasicDataToHTTPWriter(configuration Configuration, writer *multipart.Writer, tracklist []Track) {
	// Add information name/description

	name, desc, tag_list := getBasicInput(configuration)

	writer.WriteField("name", name)
	writer.WriteField("description", desc)

	// Add tags
	for i, tag := range tag_list {
		field_name := fmt.Sprintf("tags-%d-tag", i)
		writer.WriteField(field_name, tag)
	}

	// Add tracklist
	if tracklist != nil {
		// var total_duration int = 0
		var start_time int = 0

		for i, track := range tracklist {
			artist_field_name := fmt.Sprintf("sections-%d-artist", i)
			song_field_name := fmt.Sprintf("sections-%d-song", i)
			start_time_field_name := fmt.Sprintf("sections-%d-start_time", i)

			start_time += track.Duration

			writer.WriteField(start_time_field_name, fmt.Sprintf("%d", start_time))
			writer.WriteField(artist_field_name, track.Artist)
			writer.WriteField(song_field_name, track.Song)
		}
	}
}

func AddPremiumDataToHTTPWriter(writer *multipart.Writer) {

	// If you're not PRO, you can't do this, get out
	if !CURRENT_USER.IsPro {
		return
	}

	term.OutputMessage("\n" + term.Green + "Setting pro user attributes..." + term.Reset + "\n")

	publish_date, disable_comments, hide_stats, unlisted := getPremiumInput()

	if publish_date != "" {
		writer.WriteField("publish_date", publish_date)
	}
	if disable_comments {
		writer.WriteField("disable_comments", "1")
	}
	if hide_stats {
		writer.WriteField("hide_stats", "1")
	}
	if unlisted {
		writer.WriteField("unlisted", "1")
	}
}

func getBasicInput(configuration Configuration) (string, string, []string) {
	term.OutputMessage("Enter a name for the cloudcast: ")
	cast_name, err := term.STD_IN.ReadString('\n')
	if err != nil {
		term.OutputError("Incorrect name.")
		os.Exit(2)
	}

	term.OutputMessage("Enter a description: ")
	cast_desc, err := term.STD_IN.ReadString('\n')
	if err != nil {
		term.OutputError("Incorrect description.")
		os.Exit(2)
	}

	term.OutputMessage(fmt.Sprintf("Enter tags (comma separated) [%s]: ", configuration.DEFAULT_TAGS))
	cast_tags, err := term.STD_IN.ReadString('\n')
	if err != nil {
		term.OutputError("Incorrect tag format.")
		os.Exit(2)
	}

	if cast_tags == "" || cast_tags == "\n" {
		cast_tags = configuration.DEFAULT_TAGS
	}
	tags_arr := strings.Split(cast_tags, ",")

	return cast_name, cast_desc, tags_arr
}

func getPremiumInput() (string, bool, bool, bool) {
	disable_comments := false
	hide_stats := false
	unlisted := false
	publish_date := ""

	fmt.Printf("Disable comments? [y/n] ")
	if confirm.AskForConfirmation() {
		disable_comments = true
	}

	fmt.Printf("Hide statistics? [y/n] ")
	if confirm.AskForConfirmation() {
		hide_stats = true
	}

	fmt.Printf("Set to unlisted? [y/n] ")
	if confirm.AskForConfirmation() {
		unlisted = true
	}

	fmt.Printf("Set publish date? [y/n] ")
	if confirm.AskForConfirmation() {
		publish_date = publishDateInput()

	}

	return publish_date, disable_comments, hide_stats, unlisted
}

func parseDateInputToTime(dateIn string) time.Time {
	location, err := time.LoadLocation("Local")

	dateTime, err := time.ParseInLocation("02/01/2006 15:04", strings.TrimSpace(dateIn), location)

	if err != nil {
		term.OutputError("Incorrect date format  - " + err.Error())
		os.Exit(2)
	}

	return dateTime
}

func publishDateInput() string {
	current_time := time.Now().In(time.Local)
	zonename, offset := current_time.Zone()

	term.OutputMessage("Enter a publish date in " + zonename + " (" + fmt.Sprintf("%+d", offset/60/60) + " GMT) [DD/MM/YYYY HH:MM]: ")
	inPublishDate, err := term.STD_IN.ReadString('\n')
	if err != nil {
		term.OutputError("Incorrect publish date.")
		os.Exit(2)
	}

	publish_date := parseDateInputToTime(inPublishDate)

	if !publish_date.After(current_time) {
		term.OutputError("Date " + publish_date.Format(time.RFC1123) + " is not in the future")
		return publishDateInput()
	}

	return publish_date.UTC().Format(time.RFC3339)
}

func build_http(url string, request string) *http.Request {
	req, err := http.NewRequest(request, url, nil)
	if err != nil {
		term.OutputError(err.Error())
	}

	req.Header.Set("User-Agent", "Mixcloud CLI Uploader v"+versions.VERSION)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req
}
