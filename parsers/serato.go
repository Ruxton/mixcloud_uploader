package parsers

/**
 * The Serato parser supports CSV files files from Serato's history export.
 * Track timings are measure from start_time
 */

import (
	"encoding/csv"
	"fmt"
	"github.com/ruxton/mixcloud/mixcloud"
	"github.com/ruxton/term"
	"os"
	"time"
)

func ParseSeratoTrackList(tracklist *string) []mixcloud.Track {
	var list []mixcloud.Track

	fin, err := os.Open(*tracklist)
	if err != nil {
		term.OutputError(fmt.Sprintf("Error loading %s - %s\n", *tracklist, err))
		os.Exit(2)
	}
	defer fin.Close()

	reader := csv.NewReader(fin)
	reader.FieldsPerRecord = 9

	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		term.OutputError(fmt.Sprintf("Error processing Serato CSV - %s\n", err))
		os.Exit(2)
	}

	var last_track_time_str string = ""

	for i, row := range rawCSVdata {
		// name,artist,bpm,start time,end time,playtime,deck,notes,album
		track, artist, start := row[0], row[1], row[3]

		// DON'T parse the first 2 rows, they're a row heading and information for
		// the entire mix
		if i > 1 {
			thistrack := new(mixcloud.Track)

			thistrack.Artist = artist
			thistrack.Song = track

			last_time, _ := time.Parse("15:04:05 PM", last_track_time_str)
			track_time, err := time.Parse("15:04:05 PM", start)
			if err != nil {
				term.OutputError("Unable to parse time." + err.Error())
				os.Exit(2)
			}

			if last_track_time_str != "" {
				duration := track_time.Sub(last_time)
				thistrack.Duration = int(duration.Seconds())
			} else {

			}
			last_track_time_str = start

			list = append(list, *thistrack)
		}
	}

	return list

}
