package bt

import (
	"encoding/json"
	"github.com/cgghui/cgghui"
	"strings"
)

type Site struct {
	ID            uint            `json:"id"`
	AddTime       string          `json:"addtime"`
	EndTime       string          `json:"edate"`
	BackupCount   int             `json:"backup_count"`
	BindDomainNum int             `json:"domain"`
	Name          string          `json:"name"`
	Path          string          `json:"path"`
	PHPVersion    string          `json:"php_version"`
	SSL           json.RawMessage `json:"ssl"`
	Status        string          `json:"status"`
	PS            string          `json:"ps"`
}

// CheckStatus 如果站点为开启状态 返回true
func (s *Site) CheckStatus() bool {
	return s.Status == "1"
}

// CheckStatic 如果站点为纯静态 返回true
func (s *Site) CheckStatic() bool {
	return s.PHPVersion == "静态"
}

const (
	PHPVersion72 = "72"
	PHPVersion00 = "00"
	DBNameMySQL  = "MySQL"
	BoolFalse    = "false"
	ListenPort80 = "80"
)

type CreateSiteConfig struct {
	BindDomain string
	PHPVersion string
	Port       string
	dbSoftName string
}

func (c *CreateSiteConfig) SetBindDomain(v string) {
	c.BindDomain = v
}

func (c *CreateSiteConfig) SetPHPVersion(v string) {
	c.PHPVersion = v
}

func (c *CreateSiteConfig) EnableSQL(name string) {
	if name == "" {
		c.dbSoftName = BoolFalse
		return
	}
	c.dbSoftName = name
}

func (c *CreateSiteConfig) GetNote() string {
	return strings.ReplaceAll(c.BindDomain, ".", "_")
}

func (c *CreateSiteConfig) SitePath() string {
	return SiteRoot + "/" + c.GetNote()
}

func (c *CreateSiteConfig) GeneratePassword(length int) string {
	return GeneratePassword(length)
}

func (c *CreateSiteConfig) GetDatabaseUsername() string {
	alias := strings.ReplaceAll(c.GetNote(), ".", "_")
	s := len(alias)
	if s > 16 {
		return alias[:16]
	}
	return alias
}

func GeneratePassword(length int) string {
	return cgghui.RandomString(length, "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ@")
}
