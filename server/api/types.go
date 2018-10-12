package api

import (
	"github.com/dgraph-io/badger"
	"ill.fi/superpower/server/lib"
	"os"
	"strconv"
	"strings"
	"time"
)

type DBHandler struct {
	Users, Files, Invites *badger.DB
}

func (handler DBHandler) Close() {
	handler.Invites.Close()
	handler.Users.Close()
	handler.Files.Close()
}

type Invite struct {
	AuthorizedEmail string
	From            int // id
}

type Response struct {
	Error    bool        `json:"error"`
	Response interface{} `json:"response"`
}

func StringResponse(err bool, resp string) Response {
	return Response{
		Error:    err,
		Response: resp,
	}
}

type Account struct {
	ID            int       `json:"id"`
	Token         string    `json:"api_token"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"name"`
	PasswordHash  string    `json:"password_hash"`
	JoinDate      time.Time `json:"join_date"`
	Invites       []int     `json:"invites"`
	Administrator bool      `json:"admin"`
}

type UploadType int

const (
	CodePaste UploadType = iota // code syntax highlighting & formatting
	TextPaste
	ImageUpload // optional compression
	FileUpload
)

type Upload struct {
	Type     UploadType `json:"type"`
	Size     int        `json:"size"`
	Location *os.File   `json:"local_location"`
	Author   int        `json:"author_id"`
}

func SerializeAccount(a Account) string {
	// admin, email, display, joindate, pwhash, token, invites
	return strings.Join([]string{strconv.FormatBool(a.Administrator), a.Email, a.DisplayName, strconv.FormatInt(a.JoinDate.Unix(), 10), string(a.PasswordHash), a.Token, transformInvites(a.Invites)}, ",")
}

func parseInvites(s string) []int {
	inv := strings.Split(s, ",")
	var i []int
	if len(s) == 0 {
		return i
	}
	for _, x := range inv {
		v, e := strconv.Atoi(x)
		lib.NOPCheck(e)
		i = append(i, v)
	}
	return i
}

func transformInvites(invs []int) string {
	if len(invs) == 0 {
		return ","
	}
	buf := ""
	for ind, i := range invs {
		buf += string(i)
		if ind != len(invs)-1 {
			buf += ","
		}
	}
	return buf
}

func ParseAccount(a Account, data string) Account {
	v := strings.Split(data, ",")
	jd, e := strconv.Atoi(v[3])
	lib.NOPCheck(e)
	ad, e := strconv.ParseBool(v[0])
	lib.NOPCheck(e)
	n := Account{
		Administrator: ad,
		ID:            a.ID,
		Email:         v[1],
		DisplayName:   v[2],
		JoinDate:      time.Unix(int64(jd), 0),
		PasswordHash:  v[4],
		Token:         v[5],
		Invites:       parseInvites(v[6]),
	}
	return n
}
