package bt

type Database struct {
	ID          uint   `json:"id"`
	PID         uint   `json:"pid"`
	AddTime     string `json:"addtime"`
	Accept      string `json:"accept"`
	BackupCount int    `json:"backup_count"`
	Name        string `json:"name"`
	Password    string `json:"password"`
	Username    string `json:"username"`
	PS          string `json:"ps"`
}
