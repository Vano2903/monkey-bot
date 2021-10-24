package main

import (
	"context"
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
	PersonalBest PB                 `bson:"personalBest, omitempty"json:"personalBest, omitempty"` //personal best
}

type PB struct {
	Time  Time  `bson:"time, omitempty"json:"time, omitempty"`   //personal bests about time
	Words Words `bson:"words, omitempty"json:"words, omitempty"` //personal bests about words
}

type Time struct {
	T15  []Stats `bson:"15, omitempty"json:"15, omitempty"`   //15 sec timer
	T30  []Stats `bson:"30, omitempty"json:"30, omitempty"`   //30 sec timer
	T60  []Stats `bson:"60, omitempty"json:"60, omitempty"`   //60 sec timer
	T120 []Stats `bson:"120, omitempty"json:"120, omitempty"` //120 sec timer
}

type Words struct {
	W10  []Stats `bson:"10, omitempty"json:"10, omitempty"`   //10 words
	W25  []Stats `bson:"25, omitempty"json:"25, omitempty"`   //25 words
	W50  []Stats `bson:"50, omitempty"json:"50, omitempty"`   //50 words
	W100 []Stats `bson:"100, omitempty"json:"100, omitempty"` //100 words
}

type Stats struct {
	Accuracy float64 `bson:"acc, omitempty"json:"acc, omitempty"`           //accuracy of the run
	Wpm      float64 `bson:"wpm, omitempty"json:"wpm, omitempty"`           //wpm of the run
	Language string  `bson:"language, omitempty"json:"language, omitempty"` //language of the run
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

	g, _ := s.Guild(guilds[0].ID)
	for _, role := range g.Roles {
		if role.Name == "typer" {
			err = s.GuildMemberRoleAdd(guilds[0].ID, u.UserID, role.ID)
			return err
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

	g, _ := s.Guild(guilds[0].ID)
	for _, role := range g.Roles {
		if role.Name == "typer" {
			err = s.GuildMemberRoleRemove(guilds[0].ID, u.UserID, role.ID)
			return err
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

func (u User) UpdatePersonalBest() error {
	update := bson.M{"personalBest": u.PersonalBest}
	_, err := collectionUser.UpdateOne(
		ctxUser,
		bson.M{"userID": u.UserID},
		bson.D{
			{"$set", update},
		},
	)
	return err
}

func (u User) UpdateTokens() error {
	update := bson.M{"refreshToken": u.RefreshToken, "idToken": u.IDToken}
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

//do request to monkey type's api and get/update the user's personal bests
func (u *User) GetPersonaBest() error {
	var err error
	u.PersonalBest, err = GetPersonaBest(u.IDToken)
	if err != nil {
		u.IDToken, u.RefreshToken, err = GetNewAccessToken(u.RefreshToken)
		if err != nil {
			u.IDToken, u.RefreshToken, err = Login(u.Email, u.Password)
			if err != nil {
				return err
			} else {
				u.UpdateTokens()
				u.PersonalBest, err = GetPersonaBest(u.IDToken)
				if err != nil {
					return err
				}
				u.UpdatePersonalBest()
			}
		} else {
			u.UpdateTokens()
			u.PersonalBest, err = GetPersonaBest(u.IDToken)
			if err != nil {
				return err
			}
			u.UpdatePersonalBest()
		}
	}
	u.UpdatePersonalBest()
	return nil
}

//given discord id will return the user saved
func GetUser(userID string) (User, error) {
	query := bson.M{"userID": userID}
	cur, err := collectionUser.Find(ctxUser, query)
	if err != nil {
		return User{}, err
	}
	defer cur.Close(ctxUser)
	var userFound []User

	//convert cur in []User
	if err = cur.All(context.TODO(), &userFound); err != nil {
		return User{}, err
	}
	return userFound[0], nil
}

//retrun a slice with all the typers
func GetAllTypers() ([]User, error) {
	cur, err := collectionUser.Find(ctxUser, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctxUser)
	var usersFound []User

	//convert cur in []User
	if err = cur.All(context.TODO(), &usersFound); err != nil {
		return nil, err
	}
	return usersFound, nil
}
