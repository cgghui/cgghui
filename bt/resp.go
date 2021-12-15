package bt

import "encoding/json"

type ResponseBase struct {
	Status bool   `json:"status"`
	Msg    string `json:"msg"`
}

type ResponseDirs struct {
	Dir         []string `json:"DIR"`
	Files       []string `json:"FILES"`
	FileRecycle bool     `json:"FILE_RECYCLE"`
	Page        string   `json:"PAGE"`
	Path        string   `json:"PATH"`
}

type ResponseTableDatabases struct {
	Where string     `json:"where"`
	Page  string     `json:"page"`
	Data  []Database `json:"data"`
}

type ResponseTableSites struct {
	Where string `json:"where"`
	Page  string `json:"page"`
	Data  []Site `json:"data"`
}

type ResponseTableDomain struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	PID     uint   `json:"pid"`
	Port    uint   `json:"port"`
	AddTime string `json:"addtime"`
}

type ResponseGetFileBody struct {
	ResponseBase
	Data     string `json:"data"`
	Encoding string `json:"encoding"`
}

type ResponsePlugin struct {
	Name    string `json:"name"`
	Setup   bool   `json:"setup"`
	OS      string `json:"os"`
	Title   string `json:"title"`
	Status  bool   `json:"status"`
	Version string `json:"version"`
}

type ResponseNetWork struct {
	CPU           []interface{}  `json:"cpu"`
	SiteTotal     int            `json:"site_total"`
	DatabaseTotal int            `json:"database_total"`
	FtpTotal      int            `json:"ftp_total"`
	SystemName    string         `json:"system"`
	BtVersion     string         `json:"version"`
	Disk          []ResponseDisk `json:"disk"`
	Mem           ResponseMem    `json:"mem"`
}

type ResponseDisk struct {
	FileSystem string   `json:"filesystem"`
	Inodes     []string `json:"inodes"`
	Path       string   `json:"path"`
	Size       []string `json:"size"`
	Type       string   `json:"type"`
}

type ResponseMem struct {
	Buffers  int `json:"memBuffers"`
	Cached   int `json:"memCached"`
	Free     int `json:"memFree"`
	RealUsed int `json:"memRealUsed"`
	Total    int `json:"memTotal"`
}

type ResponseCreateSite struct {
	ResponseBase
	SiteStatus     bool   `json:"siteStatus"`
	SiteId         int    `json:"siteId"`
	FtpStatus      bool   `json:"ftpStatus"`
	DatabaseUser   string `json:"databaseUser"`
	DatabaseStatus bool   `json:"databaseStatus"`
	DatabasePass   string `json:"databasePass"`
}

type ResponseDeleteSite struct {
	ResponseBase
	Success []string `json:"success"`
}

type ResponseTask struct {
	ID    uint            `json:"id"`
	Type  string          `json:"type"`
	Name  string          `json:"name"`
	Shell string          `json:"shell"`
	Other json.RawMessage `json:"other"`
}

type ResponseTaskOther struct {
	DFile    string `json:"dfile"`
	Password string `json:"password"`
}

func (t ResponseTask) ParseOther() (ResponseTaskOther, error) {
	var (
		ret ResponseTaskOther
		tmp string
		err error
	)
	if err = json.Unmarshal(t.Other, &tmp); err != nil {
		return ret, err
	}
	if err = json.Unmarshal([]byte(tmp), &ret); err != nil {
		return ret, err
	}
	return ret, nil
}
