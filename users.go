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

//check if the user is already in the database (discord id)
func (u User) Exist() bool {
	//search in database
	result := collectionUser.FindOne(ctxUser, bson.M{"userID": u.UserID})

	return result.Err() == nil
}

//will add the @typer role to the user
func (u User) AddTyperRole(s *discordgo.Session) error {
	guilds, err := s.UserGuilds(100, "", "")
	if err != nil {
		return err
	}

	for _, guild := range guilds {
		g, _ := s.Guild(guild.ID)
		for _, role := range g.Roles {
			if role.Name == "typer" {
				err = s.GuildMemberRoleAdd(guild.ID, u.UserID, role.ID)
				return err
			}
		}
	}
	return errors.New("typer role not found")
}

//will remove the @typer role from the user
func (u User) RemoveTyperRole(s *discordgo.Session) error {
	guilds, err := s.UserGuilds(100, "", "")
	if err != nil {
		return err
	}

	for _, guild := range guilds {
		g, _ := s.Guild(guild.ID)
		for _, role := range g.Roles {
			if role.Name == "typer" {
				err = s.GuildMemberRoleRemove(guild.ID, u.UserID, role.ID)
				return err
			}
		}
	}
	return errors.New("typer role not found")
}

//return a string for tagging the user
func (u User) Mention(s *discordgo.Session) string {
	mention, _ := s.User(u.UserID)
	return mention.Mention()
}

//add the user to the database
func (u User) AddToDb() error {
	if u.Exist() {
		return errors.New("hai giá un account associato")
	}

	idToken, refreshToken, err := Login(u.Email, u.Password)
	if err != nil {
		return err
	}

	toInsert := struct {
		UserID       string `bson: "userID, omitempty"json: "userID, omitempty"`             //discord id
		Username     string `bson: "username, omitempty"json: "username, omitempty"`         //discord username
		Email        string `bson: "email, omitempty"json: "email, omitempty"`               //monkey type email
		Password     string `bson: "password, omitempty"json: "password, omitempty"`         //monkey type password
		IDToken      string `bson: "idToken, omitempty"json: "idToken, omitempty"`           //access token
		RefreshToken string `bson: "refreshToken, omitempty"json: "refreshToken, omitempty"` //refresh token
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

//update the user's info in the database
func (u User) UpdateUser() error {
	if !u.Exist() {
		return errors.New("l'utente non é registrato")
	}

	update := bson.M{"username": u.Username, "email": u.Email, "password": u.Password}
	_, err := collectionUser.UpdateOne(
		ctxUser,
		bson.M{"userID": u.UserID},
		bson.D{
			{"$set", update},
		},
	)

	return err
}

//remove the user from the db
func (u User) RemoveFromDB() error {
	_, err := collectionUser.DeleteOne(ctxUser, bson.M{"userID": u.UserID})
	if err != nil {
		return err
	}
	return nil
}
