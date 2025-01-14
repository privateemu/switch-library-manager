package db

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type TitleAttributes struct {
	Id                string      `json:"id"`
	Name              string      `json:"name,omitempty"`
	Version           json.Number `json:"version,omitempty"`
	Region            string      `json:"region,omitempty"`
	ReleaseDate       int         `json:"releaseDate,omitempty"`
	ParsedReleaseDate string
	Publisher         string   `json:"publisher,omitempty"`
	IconUrl           string   `json:"iconUrl,omitempty"`
	Screenshots       []string `json:"screenshots,omitempty"`
	BannerUrl         string   `json:"bannerUrl,omitempty"`
	Description       string   `json:"description,omitempty"`
	Size              int      `json:"size,omitempty"`
}

type SwitchTitle struct {
	Attributes TitleAttributes
	Updates    map[int]string
	Dlc        map[string]TitleAttributes
}

type SwitchTitlesDB struct {
	TitlesMap map[string]*SwitchTitle
}

func CreateSwitchTitleDB(titlesFile, versionsFile io.Reader) (*SwitchTitlesDB, error) {
	//parse the titles objects
	var titles = map[string]TitleAttributes{}
	err := decodeToJsonObject(titlesFile, &titles)
	if err != nil {
		return nil, err
	}

	//parse the titles objects
	//titleID -> versionId-> release date
	var versions = map[string]map[int]string{}
	err = decodeToJsonObject(versionsFile, &versions)
	if err != nil {
		return nil, err
	}

	result := SwitchTitlesDB{TitlesMap: map[string]*SwitchTitle{}}
	for id, attr := range titles {
		id = strings.ToLower(id)

		//TitleAttributes id rules:
		//main TitleAttributes ends with 000
		//Updates ends with 800
		//Dlc adds 1 to 4th char starting from the right (always odd) and
		//have a running counter (starting with 001) in the 3 last chars
		switchTitle := &SwitchTitle{Dlc: map[string]TitleAttributes{}}
		idPrefix := id[0 : len(id)-3]
		if !(strings.HasSuffix(id, "000") || strings.HasSuffix(id, "800")) {
			intVar, _ := strconv.ParseUint(id[len(id)-4:len(id)-3], 16, 64)
			h := fmt.Sprintf("%x", intVar-1)
			idPrefix = id[0:len(id)-4] + h
		}

		if t, ok := result.TitlesMap[idPrefix]; ok {
			switchTitle = t
		}
		result.TitlesMap[idPrefix] = switchTitle

		// parse the release date to a date string
		prd := strconv.Itoa(attr.ReleaseDate)
		if len(prd) == 8 {
			attr.ParsedReleaseDate = prd[0:4] + "-" + prd[4:6] + "-" + prd[6:8]
		} else {
			attr.ParsedReleaseDate = prd
		}

		//process Updates
		if strings.HasSuffix(id, "800") {
			updates := versions[id[0:len(id)-3]+"000"]
			switchTitle.Updates = updates
			continue
		}

		//process main TitleAttributes
		if strings.HasSuffix(id, "000") {
			switchTitle.Attributes = attr
			continue
		}

		//not an update, and not main TitleAttributes, so treat it as a DLC
		switchTitle.Dlc[id] = attr

	}

	return &result, nil
}
