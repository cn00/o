package utils

import (
	"time"
)

func CheckFromToDateFormat(fromDate string, toDate string) (error, string) {
	var layout = "2006-01-02 15:04:05"
	var errMsg string = ""
	if fromDate != "" {
		_, err := time.Parse(layout, fromDate)
		if err != nil {
			errMsg = "From Date " + fromDate + " is illegal date format. Example of valid format is 2017-01-02 12:59:40"
			return err, errMsg
		}
	}

	if toDate != "" {
		_, err := time.Parse(layout, toDate)
		if err != nil {
			errMsg = "To Date " + toDate + " is illegal date format. Example of valid format is 2017-01-02 12:59:40"
			return err, errMsg
		}
	}
	return nil, errMsg
}

func GetDate(layout string) string {
	day := time.Now()
	return day.Format(layout)
}
