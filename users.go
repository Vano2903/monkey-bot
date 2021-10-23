package main

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id, omitempty" json:"-"`                               //mongo id
	UserID       string             `bson:"userID, omitempty"json:"userID, omitempty"`             //discord id
	Username     string             `bson:"username, omitempty"json:"username, omitempty"`         //discord username
	Email        string             `bson:"email, omitempty"json:"email, omitempty"`               //monkey type email
	Password     string             `bson:"password, omitempty"json:"password, omitempty"`         //monkey type password
	IDToken      string             `bson:"idToken, omitempty"json:"idToken, omitempty"`           //access token
	RefreshToken string             `bson:"refreshToken, omitempty"json:"refreshToken, omitempty"` //refresh token
}

//check if the user is already in the database (email and password)
func (u User) Exist() bool {
	//search in database
	result := collectionUser.FindOne(ctxUser, bson.M{"email": u.Email, "password": u.Password})

	return result.Err() == nil
}

func (u User) Mention(s *discordgo.Session) string {
	mention, _ := s.User(u.UserID)
	return mention.Mention()
}

func (u User) AddToDb() error {
	if u.Exist() {
		return errors.New("Queste credenziali sono giá registrate")
	}

	idToken, refreshToken, err := Login(u.Email, u.Password)
	if err != nil {
		return err
	}

	toInsert := struct {
		UserID       string `bson:"userID, omitempty"json:"userID, omitempty"`             //discord id
		Username     string `bson:"username, omitempty"json:"username, omitempty"`         //discord username
		Email        string `bson:"email, omitempty"json:"email, omitempty"`               //monkey type email
		Password     string `bson:"password, omitempty"json:"password, omitempty"`         //monkey type password
		IDToken      string `bson:"idToken, omitempty"json:"idToken, omitempty"`           //access token
		RefreshToken string `bson:"refreshToken, omitempty"json:"refreshToken, omitempty"` //refresh token
	}{
		u.UserID,
		u.Username,
		u.Email,
		u.Password,
		idToken,
		refreshToken,
	}

	_, err = collectionUser.InsertOne(ctxUser, toInsert)
	return err
}
