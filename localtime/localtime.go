package localtime

import "time"

var loc = time.FixedZone("Asia/Tokyo", 9*60*60)

// Date : LocalDate in Asia/Tokyo
func Date(year int, month time.Month, day int, hour int, minute int, second int) time.Time {
	return time.Date(year, month, day, hour, minute, second, 0, loc)
}

// Parse : Wrapper of time.ParseInLocation
func Parse(date string, layout string) (time.Time, error) {
	jst, err := time.ParseInLocation(layout, date, loc)
	if err != nil {
		return time.Time{}, err
	}
	return jst, nil
}

// Today : Wrapper of time.Now().In(loc)
func Today() time.Time {
	return time.Now().In(loc)
}
