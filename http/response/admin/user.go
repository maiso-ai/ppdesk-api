package admin

import (
	"time"

	"github.com/lejianwen/rustdesk-api/v2/model"
)

type LoginPayload struct {
	Id         uint     `json:"id"`
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	Avatar     string   `json:"avatar"`
	Token      string   `json:"token"`
	RouteNames []string `json:"route_names"`
	Nickname   string   `json:"nickname"`
	GroupId    uint     `json:"group_id"`
	IsAdmin    *bool    `json:"is_admin"`
	Status     int      `json:"status"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

func (lp *LoginPayload) FromUser(user *model.User) {
	lp.Id = user.Id
	lp.Username = user.Username
	lp.Email = user.Email
	lp.Avatar = user.Avatar
	lp.Nickname = user.Nickname
	lp.GroupId = user.GroupId
	lp.IsAdmin = user.IsAdmin
	lp.Status = int(user.Status)
	lp.CreatedAt = time.Time(user.CreatedAt).Format("2006-01-02 15:04:05")
	lp.UpdatedAt = time.Time(user.UpdatedAt).Format("2006-01-02 15:04:05")
}

type UserOauthItem struct {
	Op     string `json:"op"`
	Status int    `json:"status"`
}
