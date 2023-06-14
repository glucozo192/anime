package services

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glu/shopvui/internal/entities"
	"github.com/glu/shopvui/internal/golibs/database"
	"github.com/glu/shopvui/internal/models"
	"github.com/glu/shopvui/util"
	"github.com/google/uuid"
)

func (s *Server) loginUser(ctx *gin.Context) {
	var req models.LoginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, "cannot bin JSON")
		return
	}
	email := database.Text(req.Email)
	user, err := s.UserRepo.GetUser(ctx, s.db, email)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, "User not found")
			return
		}
		ctx.JSON(http.StatusInternalServerError, "User not found")
		return
	}
	if !user.Active.Bool {
		ctx.JSON(http.StatusInternalServerError, "user war detected")
	}

	err = util.CheckPassword(req.Password, user.Password.String)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, err)
		return
	}

	accessToken, _, err := s.tokenMaker.CreateToken(
		user.ID.String,
		s.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "error")
		return
	}

	// refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
	// 	user.Username,
	// 	server.config.RefreshTokenDuration,
	// )
	// if err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, "error")
	// 	return
	// }

	rsp := models.LoginUserResponse{
		AccessToken: accessToken,
		//AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		//RefreshToken:          refreshToken,
		//RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User: models.UserResponse{
			Email:     user.Email.String,
			FirstName: user.FirstName.String,
			LastName:  user.LastName.String,
			CreatedAt: user.InsertedAt.Time,
			UpdatedAt: user.InsertedAt.Time,
		},
	}
	ctx.JSON(http.StatusOK, rsp)
}

func toUser(req models.CreateUserRequest) (*entities.User, error) {
	hashPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return &entities.User{}, err
	}
	return &entities.User{
		ID:        database.Text(uuid.NewString()),
		Email:     database.Text(req.Email),
		FirstName: database.Text(req.FirstName),
		LastName:  database.Text(req.LastName),
		Password:  database.Text(hashPassword),
	}, nil
}

func (s *Server) register(ctx *gin.Context) {
	var req models.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	email := database.Text(req.Email)
	_, err := s.UserRepo.GetUser(ctx, s.db, email)
	if err == nil {
		ctx.JSON(http.StatusInternalServerError, "User already exists")
		return
	}
	u, err := toUser(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := s.UserRepo.CreateUser(ctx, s.db, u)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("can't create user: %v", err))
		return
	}
	rsp := models.UserResponse{
		ID:        user.ID.String,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
		Email:     user.Email.String,
		CreatedAt: user.InsertedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}
	ctx.JSON(http.StatusOK, rsp)
}

func toRole(req models.AddRoleRequest) *entities.Role {
	return &entities.Role{
		ID:   database.Text(uuid.NewString()),
		Name: database.Text(req.Name),
	}
}

// func (s *Server) addRole(ctx *gin.Context) {
// 	var req models.AddRoleRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}
// 	roles := toRole(req)
// 	err := s.UserRepo.AddRoles(ctx, s.db, roles)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("can't init roles: %v", err))
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, "success")
// }
