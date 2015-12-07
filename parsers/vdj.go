package parsers

import (
	"bufio"
	"fmt"
	"github.com/ruxton/mixcloud/mixcloud"
	"github.com/ruxton/term"
	"io"
	"os"
	"strings"
	"time"
)

func ParseVirtualDJTrackList(tracklist *string) []mixcloud.Track {
	var list []mixcloud.Track

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

		thistrack := new(mixcloud.Track)

		var trackdata []string = strings.SplitN(string(track), " - ", 2)

		if len(trackdata) != 2 {
			term.OutputError("Error parsing track " + string(track) + " at " + tracktimestr)
			term.OutputMessage("Please enter an artist for this track: ")
			artist, err := term.STD_IN.ReadString('\n')
			if err != nil {
				term.OutputError("Incorrect artist entry.")
				os.Exit(2)
			}
			term.OutputMessage("Please enter a name for this track: ")
			track, err := term.STD_IN.ReadString('\n')
			if err != nil {
				term.OutputError("Incorrect track name entry.")
				os.Exit(2)
			}

			trackdata = []string{artist, track}
		}

		thistrack.Artist = trackdata[0]
		thistrack.Song = trackdata[1]

		last_time, _ := time.Parse("15:04", last_track_time_str)
		track_time, err := time.Parse("15:04", tracktimestr)
		if err != nil {
			term.OutputError("Unable to parse time." + err.Error())
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
