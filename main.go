package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
)

type BotConf struct {
	UsersFilePath string `yaml: "usersFilePath"`
	Token         string `yaml:"token"`
	Prefix        string `yaml:"prefix"`
	BotID         string
}

type User struct {
	Username string `json:"username"`
	UserID   string `json:"userID"`
	Password string `json:"password"`
	Email    string `json:"email"`
	JWT      string `json:"jwt"`
}

var conf BotConf
var botID string

func init() {
	dat, err := ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("error in init function: %v", err.Error())
	}
	err = yaml.Unmarshal([]byte(dat), &conf)
	if err != nil {
		log.Fatalf("error unmarshalling config: %v", err)
	}

	if Exist(conf.UsersFilePath) {
		err = WriteFile(conf.UsersFilePath, []byte("[]"))
		if err != nil {
			log.Fatalf("error creating users json: %v", err)
		}
	}
}

func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return emailRegex.MatchString(e)
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, conf.Prefix) {
		if m.Author.ID == botID {
			return
		}

		elements := strings.Split(m.Content, " ")
		if elements[0] == conf.Prefix+"partecipate" {
			partecipateHandler(s, m, elements[1:])
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "codice sconosciuto, usa !help per sapere i codici che puoi usare")
		}
	}
}

//add the user's info to the file and request the jwt to monkeytype.com
func partecipateHandler(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	var u User
	messageID := m.ID
	err := s.ChannelMessageDelete(m.ChannelID, messageID)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("unable to delete the message: %v", err.Error()))
	}
	u.UserID = m.Author.ID
	u.Username = m.Author.Username
	if isEmailValid(params[1]) {
		u.Email = params[1]
	}
	u.Password = params[2]
	_, _ = s.ChannelMessageSend(m.ChannelID, "registrato correttamante")

}

//return all codes knows by the bot
func helpHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, _ = s.ChannelMessageSend(m.ChannelID, "```\nCiao sono HAIbot e sono un bot scritto da VanoUwU, il mio compito é automatizzare l'invio di codici hentai al database di HAI\n I codici disponibili sono:\n\t⸭ !add \"codice\" ⇒ aggiungi un codice al database di HAI con \"IsCollection\" settato a false di default\n\t⸭ !add \"codice\" \"bool\" ⇒ aggiungi un codice al database di HAI con \"IsCollection\" settato al valore inserito nel boleano e \"IsChecked\" a true\n\t⸭ !exist \"codice\" ⇒ controlla se un codice esiste o meno all'interno del'database\n\t⸭ !stats ⇒ statistiche degli hentai raccolti fino ad ora (i.e. num hentai raccolti, collezioni verificate, ...)\n\t⸭ !stat \"codice\" ⇒ visualizza tutti i parametri relativi al codice inserito\n\t⸭ !train \"codice\" bool ⇒ (COMING SOON) permette di far analizzare il codice assegnato dall'AI, il booleano deve verificare se il codice é una collezione o meno\n\t⸭ !verify \"codice\" bool ⇒ permette di verificare un codice che non é verificato mettendo il booleano a true se é una collezione e false se hentai \n\t⸭ !toVerify ⇒ ritorna la lista di tutti i codici degli hentai da verificare e il rispettivo link)\n\t⸭ !leaderboard ⇒ visualizza la leaderboard dei punti raccolti dagli utenti\n\t⸭ !points ⇒ visualizza il valore in punti di ogni azione```")
}

func main() {
	discord, err := discordgo.New("Bot " + conf.Token)
	if err != nil {
		log.Fatal(err)
	}

	u, err := discord.User("@me")
	if err != nil {
		fmt.Println(err)
	}
	conf.BotID = u.ID

	discord.AddHandler(messageHandler)
	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening discord: %v", err.Error())
	}

	fmt.Println("wow i am working :D ")
	<-make(chan struct{})
}
