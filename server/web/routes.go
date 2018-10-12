package web

import (
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"ill.fi/superpower/server/api"
	"net/http"
	"strconv"
	"strings"
	"time"
)

///////////////////
// USER HANDLERS //
///////////////////

func UserCreationHandler(ctx *gin.Context, db *api.DBHandler) {
	ctx.Header("Content-Type", "application/json")
	disp := ctx.PostForm("display_name")
	email := ctx.PostForm("email")
	pw := ctx.PostForm("pw_hash")
	inv := ctx.PostForm("invite")

	if api.IsInvited(email, db.Invites) || inv == "debug" {
		a := api.Account{
			ID:           api.GetLastID(db.Users),
			DisplayName:  disp,
			Email:        email,
			PasswordHash: pw,
			JoinDate:     time.Now(),
			Invites:      []int{},
		}
		t := api.GetAccountByEmail(email, db.Users)
		if t == nil {
			api.AddAccount(a, db.Users)
			ctx.JSON(http.StatusOK, api.Response{
				Error:    false,
				Response: fmt.Sprintf("user %s (%s) [%d] created", a.DisplayName, a.Email, a.ID),
			})
		} else {
			ctx.JSON(http.StatusBadRequest, api.Response{
				Error:    true,
				Response: fmt.Sprintf("email %s is already registered", email),
			})
		}
	} else {
		// not invited
		ctx.JSON(http.StatusBadRequest, api.Response{
			Error:    true,
			Response: fmt.Sprintf("email %s is not invited", email),
		})
	}
}

func ListUsersHandler(ctx *gin.Context, db *api.DBHandler) {
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, api.Response{
		Error:    false,
		Response: api.GetAccounts(db.Users),
	})
}

func UserInfoHandler(ctx *gin.Context, db *api.DBHandler) {
	ctx.Header("Content-Type", "application/json")
	ids := ctx.Param("id")
	id, e := strconv.Atoi(ids)
	if e != nil {
		ctx.JSON(http.StatusOK, api.StringResponse(true, "failed to parse id \""+ids+"\"."))
		return
	}
	acc := api.GetAccountByID(id, db.Users)
	if acc != nil {
		ctx.JSON(http.StatusOK, api.Response{
			Error:    false,
			Response: acc,
		})
	} else {
		ctx.JSON(http.StatusBadRequest, api.StringResponse(true, fmt.Sprintf("user with the id %d does not exist", id)))
	}
}

//////////
// auth //
//////////

func LoginHandler(ctx *gin.Context, db *api.DBHandler) {
	ctx.Header("Content-Type", "application/json")
	session := sessions.Default(ctx)
	email := ctx.PostForm("email")
	password := ctx.PostForm("pw_hash")

	if session.Get("user") != nil {
		ctx.JSON(http.StatusUnauthorized, api.Response{
			Error:    true,
			Response: "Already authenticated",
		})
	}

	if strings.TrimSpace(email) == "" {
		ctx.JSON(http.StatusUnauthorized, api.Response{
			Error:    true,
			Response: "Email may not be empty",
		})
		return
	}

	if strings.TrimSpace(password) == "" {
		ctx.JSON(http.StatusUnauthorized, api.Response{
			Error:    true,
			Response: "Password may not be empty",
		})
		return
	}

	acc := api.GetAccountByEmail(email, db.Users)

	if acc == nil {
		ctx.JSON(http.StatusUnauthorized, api.Response{
			Error:    true,
			Response: "Email not found",
		})
		return
	}

	if password != acc.PasswordHash {
		ctx.JSON(http.StatusUnauthorized, api.Response{
			Error:    true,
			Response: "Incorrect password",
		})
		return
	} else {
		session.Set("user", acc.ID)
		err := session.Save()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, api.Response{
				Error:    true,
				Response: "Failed to generate session token, " + err.Error(),
			})
		} else {
			ctx.JSON(http.StatusOK, api.Response{
				Error:    false,
				Response: "Successfully authenticated",
			})
		}
	}
}

func AuthenticationCheckHandler(ctx *gin.Context, db *api.DBHandler) {
	ctx.Header("Content-Type", "application/json")
	session := sessions.Default(ctx)
	user := session.Get("user")
	ctx.JSON(http.StatusOK, api.Response{
		Error:    false,
		Response: user != nil,
	})
}

func LogoutHandler(ctx *gin.Context, db *api.DBHandler) {
	ctx.Header("Content-Type", "application/json")
	session := sessions.Default(ctx)
	user := session.Get("user")
	if user == nil {
		ctx.JSON(http.StatusBadRequest, api.Response{
			Error:    true,
			Response: "Invalid session token",
		})
	} else {
		session.Delete("user")
		session.Save()
		ctx.JSON(http.StatusOK, api.Response{
			Error:    false,
			Response: "Successfully logged out",
		})
	}
}
