package api

import (
	"encoding/json"
	"strconv"
	"time"
)

type SingleMediaListItem struct {
	Name            string    `json:"n"`
	CreatedAt       time.Time `json:"cre"`
	ModifiedAt      time.Time `json:"mod"`
	Size            uint64    `json:"s"`
	LowResVideoSize int64     `json:"glrv"` // Low resolution video size
	LowResFileSize  int64     `json:"ls"`   //Low resolution file size. -1 if there is no LRV file
}

type MediaInfo struct {
	Directory string                `json:"d"`
	FileList  []SingleMediaListItem `json:"fs"`
}

type MediaList struct {
	MediaListID string      `json:"id"` // media list identifier
	Media       []MediaInfo `json:"media"`
}

func (i *SingleMediaListItem) UnmarshalJSON(data []byte) error {
	type Alias SingleMediaListItem

	aux := &struct {
		Cre  string `json:"cre"`
		Mod  string `json:"mod"`
		S    string `json:"s"`
		Glrv string `json:"glrv"`
		Ls   string `json:"ls"`
		*Alias
	}{
		Alias: (*Alias)(i),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var parseErr error
	if i.Size, parseErr = strconv.ParseUint(aux.S, 10, 64); parseErr != nil {
		return parseErr
	}
	if cre, err := strconv.ParseInt(aux.Cre, 10, 64); err == nil {
		i.CreatedAt = time.Unix(cre, 0)
	}
	if mod, err := strconv.ParseInt(aux.Mod, 10, 64); err == nil {
		i.ModifiedAt = time.Unix(mod, 0)
	}
	if len(aux.Ls) > 0 {
		if i.LowResFileSize, parseErr = strconv.ParseInt(aux.Ls, 10, 64); parseErr != nil {
			return parseErr
		}
	}

	if len(aux.Glrv) > 0 {
		if i.LowResVideoSize, parseErr = strconv.ParseInt(aux.Glrv, 10, 64); parseErr != nil {
			return parseErr
		}
	}

	return nil
}
