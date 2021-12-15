package bt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/cgghui/cgghui"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"

type Bt struct {
	Link        string
	reqLogin    *http.Request
	reqGetToken *http.Request
}

var SiteRoot = "/www/wwwroot"

func New(link string) *Bt {
	return &Bt{
		Link: strings.Trim(strings.TrimSpace(link), "/") + "/",
	}
}

type Option struct {
	Link     string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

func LoadOption(filePath string) (*Option, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var opt Option
	if err = json.Unmarshal(content, &opt); err != nil {
		return nil, err
	}
	return &opt, err
}

func (o *Option) New() *Bt {
	return New(o.Link)
}

func (o *Option) GetAddress() string {
	l, _ := url.Parse(o.Link)
	return strings.Split(l.Host, ":")[0]
}

type Session struct {
	bt          *Bt
	btToken     string
	cookieStr   string
	cookieToken string
	code        string
}

func (b *Bt) Login(code, username, password string) (*Session, error) {
	var err error
	if b.reqLogin == nil {
		u := cgghui.MD5(username)
		p := cgghui.MD5(cgghui.MD5(password) + "_bt.cn")
		s := strings.NewReader("username=" + u + "&password=" + p + "&code=")
		b.reqLogin, err = http.NewRequest(http.MethodPost, b.Link+"login", s)
		if err != nil {
			return nil, fmt.Errorf("登录宝塔失败 Error[1]: %v", err)
		}
		b.reqLogin.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		b.reqLogin.Header.Add("Origin", b.Link)
		b.reqLogin.Header.Add("Referer", b.Link+code+"/")
		b.reqLogin.Header.Add("User-Agent", UserAgent)
	}
	resp, err := http.DefaultClient.Do(b.reqLogin)
	if err != nil {
		return nil, fmt.Errorf("登录宝塔失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("登录宝塔失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseBase
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("登录宝塔失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if !ret.Status {
		return nil, fmt.Errorf("登录宝塔失败 Error[5]: %s", ret.Msg)
	}
	ckArr := make(map[string]string, 0)
	ckStr := make([]string, 0)
	for _, cookie := range resp.Cookies() {
		ckArr[cookie.Name] = cookie.Value
		ckStr = append(ckStr, cookie.Name+"="+cookie.Value)
	}
	ckArr["sites_path"] = SiteRoot
	ckArr["serverType"] = "nginx"
	ckArr["site_type"] = "-1"
	ckArr["order"] = "id%20desc"
	ckStr = append(ckStr, "sites_path="+SiteRoot)
	ckStr = append(ckStr, "serverType=nginx")
	ckStr = append(ckStr, "site_type=-1")
	ckStr = append(ckStr, "order=id%20desc")
	obj := &Session{
		bt:          b,
		cookieStr:   strings.Join(ckStr, "; "),
		cookieToken: ckArr["request_token"],
		btToken:     "",
		code:        code,
	}
	if err = obj.getToken(); err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *Session) GetAddress() string {
	l, _ := url.Parse(s.bt.Link)
	return strings.Split(l.Host, ":")[0]
}

func (s *Session) getToken() error {
	var err error
	if s.bt.reqGetToken == nil {
		s.bt.reqGetToken, err = http.NewRequest(http.MethodGet, s.bt.Link, nil)
		if err != nil {
			return fmt.Errorf("获取TOKEN失败 Error[1]: %v", err)
		}
		s.bt.reqGetToken.Header.Add("Referer", s.bt.Link+s.code+"/")
		s.bt.reqGetToken.Header.Add("Cookie", s.cookieStr)
		s.bt.reqGetToken.Header.Add("User-Agent", UserAgent)
	}
	resp, err := http.DefaultClient.Do(s.bt.reqGetToken)
	if err != nil {
		return fmt.Errorf("获取TOKEN失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("获取TOKEN失败 Error[3]: http code is %d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("TK宝塔失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	s.btToken = doc.Find("#request_token_head").AttrOr("token", "")
	return nil
}

func (s *Session) GetDir(path string) (*ResponseDirs, error) {
	body := url.Values{}
	body.Add("p", "1")
	body.Add("showRow", "2000")
	body.Add("path", path)
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"files?action=GetDir", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("获取目录列表 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取目录列表 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取目录列表 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseDirs
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("获取目录列表 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return &ret, nil
}

func (s *Session) ClearRecycle() (*ResponseBase, error) {
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"files?action=Close_Recycle_bin", nil)
	if err != nil {
		return nil, fmt.Errorf("清空回收站失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("清空回收站失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("清空回收站失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseBase
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("清空回收站失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return &ret, nil
}

func (s *Session) GetPlugin(ctx context.Context) ([]*ResponsePlugin, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.bt.Link+"plugin?action=get_index_list", nil)
	if err != nil {
		return nil, fmt.Errorf("获取服务器软件失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取服务器软件失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取服务器软件失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret []*ResponsePlugin
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("获取服务器软件失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return ret, nil
}

func (s *Session) GetStatus(ctx context.Context) (*ResponseNetWork, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.bt.Link+"system?action=GetNetWork", nil)
	if err != nil {
		return nil, fmt.Errorf("获取服务器状态失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取服务器状态失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取服务器状态失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseNetWork
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("获取服务器状态失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return &ret, nil
}

func (s *Session) CreateSite(conf *CreateSiteConfig) (*ResponseCreateSite, error) {
	ps := conf.GetNote()
	pwd := conf.GeneratePassword(27)
	body := url.Values{}
	body.Add("webname", `{"domain":"`+conf.BindDomain+`","domainlist":["*.`+conf.BindDomain+`"],"count":1}`)
	body.Add("type", "PHP")
	body.Add("port", conf.Port)
	body.Add("ps", ps)
	body.Add("path", SiteRoot+"/"+ps)
	body.Add("type_id", "0")
	body.Add("version", conf.PHPVersion)
	body.Add("ftp", "false")
	body.Add("sql", conf.dbSoftName)
	if conf.dbSoftName != "false" {
		body.Add("datauser", conf.GetDatabaseUsername())
		body.Add("datapassword", pwd)
	}
	body.Add("codeing", "utf8")
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"site?action=AddSite", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建站点失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("创建站点失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("创建站点失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseCreateSite
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("创建站点失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return &ret, nil
}

func (s *Session) CreateDatabase(dbName, dbUser, dbType, dbPass string) error {
	body := url.Values{}
	body.Add("name", dbName)
	body.Add("db_user", dbUser)
	body.Add("password", dbPass)
	body.Add("dtype", dbType)
	body.Add("dataAccess", "%")
	body.Add("address", "%")
	body.Add("ps", dbName)
	body.Add("codeing", "utf8")
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"database?action=AddDatabase", strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("创建数据库失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("创建数据库失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("创建数据库失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseBase
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return fmt.Errorf("创建数据库失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if !ret.Status {
		return fmt.Errorf("创建数据库失败 Error[5]: %s", ret.Msg)
	}
	return nil
}

func (s *Session) GetDatabaseList() (*ResponseTableDatabases, error) {
	body := url.Values{}
	body.Add("table", "databases")
	body.Add("tojs", "database.get_list")
	body.Add("limit", "10000")
	body.Add("p", "1")
	body.Add("search", "")
	body.Add("order", "id desc")
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"data?action=getData", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("获取数据库列表失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取数据库列表失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取数据库列表失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseTableDatabases
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("获取数据库列表失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return &ret, nil
}

func (s *Session) GetSiteList() (*ResponseTableSites, error) {
	body := url.Values{}
	body.Add("table", "sites")
	body.Add("type", "-1")
	body.Add("limit", "10000")
	body.Add("p", "1")
	body.Add("search", "")
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"data?action=getData", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("获取站点列表失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取站点列表失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取站点列表失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseTableSites
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("获取站点列表失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return &ret, nil
}

func (s *Session) GetSiteBindDomain(site *Site) ([]ResponseTableDomain, error) {
	body := url.Values{}
	body.Add("table", "domain")
	body.Add("list", "True")
	body.Add("search", strconv.FormatUint(uint64(site.ID), 10))
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"data?action=getData", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("获取站点绑定域名失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取站点绑定域名失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取站点绑定域名失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret []ResponseTableDomain
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("获取站点绑定域名失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return ret, nil
}

func (s *Session) GetSiteFileContent(path string) (*ResponseGetFileBody, error) {
	body := url.Values{}
	body.Add("path", path)
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"files?action=GetFileBody", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("获取文件内容失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取文件内容失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取文件内容失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseGetFileBody
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("获取文件内容失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return &ret, nil
}

// SetSiteFileContent 保存修改文件
// path 为文件的完整路径
func (s *Session) SetSiteFileContent(path string, r *ResponseGetFileBody) error {
	body := url.Values{}
	body.Add("data", r.Data)
	body.Add("encoding", "utf-8")
	body.Add("path", path)
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"files?action=SaveFileBody", strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("写入文件内容失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("写入文件内容失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("写入文件内容失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseBase
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return fmt.Errorf("获取文件内容失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if ret.Status == false {
		return fmt.Errorf("获取文件内容失败 Error[5]: %s", ret.Msg)
	}
	return nil
}

// UnZip 解压文件
// src 压缩文件路径 dst 解压到的文件夹
func (s *Session) UnZip(src, dst string) error {
	body := url.Values{}
	body.Add("sfile", src)
	body.Add("dfile", dst)
	body.Add("type", "zip")
	body.Add("coding", "UTF-8")
	body.Add("password", "")
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"files?action=UnZip", strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("解压文件内容失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("解压文件内容失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("解压文件内容失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret ResponseBase
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return fmt.Errorf("解压文件内容失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if !ret.Status {
		return fmt.Errorf("解压文件内容失败 Error[5]: %s", ret.Msg)
	}
	return nil
}

// GetTaskLists 获取任务列表
func (s *Session) GetTaskLists() ([]ResponseTask, error) {
	body := url.Values{}
	body.Add("status", "-3")
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"task?action=get_task_lists", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败 Error[2]: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取任务列表失败 Error[3]: http code is %d", resp.StatusCode)
	}
	var ret []ResponseTask
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("获取任务列表失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return ret, nil
}

// Upload 上传文件
func (s *Session) Upload(localFilePath, saveFilePath string, fp *os.File, isRemoveFile bool) error {
	if isRemoveFile {
		defer func() {
			_ = fp.Close()
			_ = os.Remove(localFilePath)
		}()
	}
	var (
		body bytes.Buffer
		ct   string
		err  error
	)
	if err = upload(filepath.Base(localFilePath), saveFilePath, fp, &body, &ct); err != nil {
		return fmt.Errorf("上传文件失败 Error[1]: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"files?action=upload", &body)
	if err != nil {
		return fmt.Errorf("上传文件失败 Error[2]: %v", err)
	}
	req.Header.Add("Content-Type", ct)
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("上传文件失败 Error[3]: %v", err)
	}
	var ret ResponseBase
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return fmt.Errorf("上传文件失败 Error[4]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if !ret.Status {
		return fmt.Errorf("上传文件失败 Error[5]: %v", ret.Msg)
	}
	return nil
}

func upload(name, saveFilePath string, fp *os.File, ret *bytes.Buffer, ct *string) error {
	fi, _ := fp.Stat()
	sendWriter := multipart.NewWriter(ret)
	_ = sendWriter.WriteField("f_path", saveFilePath)
	_ = sendWriter.WriteField("f_name", name)
	_ = sendWriter.WriteField("f_size", strconv.FormatInt(fi.Size(), 10))
	_ = sendWriter.WriteField("f_start", "0")
	fileWriter, err := sendWriter.CreateFormFile("blob", "blob")
	if err != nil {
		return err
	}
	if _, err = io.Copy(fileWriter, fp); err != nil {
		return err
	}
	*ct = sendWriter.FormDataContentType()
	_ = sendWriter.Close()
	return nil
}

func (s *Session) RestoreBackupDB(dbName, backupFilePath string) error {
	data := url.Values{}
	data.Add("file", backupFilePath)
	data.Add("name", dbName)
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"database?action=InputSql", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("还原数据库备份文件失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("还原数据库备份文件失败 Error[2]: %v", err)
	}
	var ret ResponseBase
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return fmt.Errorf("还原数据库备份文件失败 Error[3]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if !ret.Status {
		return fmt.Errorf("还原数据库备份文件失败 Error[5]: %s", ret.Msg)
	}
	return nil
}

// SiteDatabasePermission 将站点数据库更变为可远程操作
func (s *Session) SiteDatabasePermission(dbName string) error {
	body := url.Values{}
	body.Add("name", dbName)
	body.Add("dataAccess", "%")
	body.Add("access", "%")
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"database?action=SetDatabaseAccess", strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("修改数据库权限失败 %s Error[1]: %v", dbName, err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("修改数据库权限失败 %s Error[2]: %v", dbName, err)
	}
	var ret ResponseBase
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return fmt.Errorf("修改数据库权限失败 %s Error[3]: %v", dbName, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if !ret.Status {
		return fmt.Errorf("修改数据库权限失败 %s Error[4]: %v", dbName, err)
	}
	return nil
}

func (s *Session) SiteDelete(id []string) error {
	body := url.Values{}
	body.Add("database", "1")
	body.Add("path", "1")
	body.Add("ftp", "1")
	body.Add("sites_id", strings.Join(id, ","))
	req, err := http.NewRequest(http.MethodPost, s.bt.Link+"site?action=delete_website_multiple", strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("删除站点失败 Error[1]: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("x-http-token", s.btToken)
	req.Header.Add("x-cookie-token", s.cookieToken)
	req.Header.Add("Cookie", s.cookieStr)
	req.Header.Add("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("删除站点失败 Error[2]: %v", err)
	}
	var ret ResponseDeleteSite
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return fmt.Errorf("删除站点失败 Error[3]: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if !ret.Status {
		return fmt.Errorf("删除站点失败 Error[4]: %v", err)
	}
	return nil
}
