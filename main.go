package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/olekukonko/tablewriter"
)

//handle all the messages coming and if it's a valid command run the command handler
func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, conf.Prefix) {
		if m.Author.ID == conf.BotID {
			return
		}

		elements := strings.Split(m.Content, " ")

		switch elements[0] {
		case conf.Prefix + "partecipate":
			PartecipateHandler(s, m, elements[1:])
		case conf.Prefix + "update":
			UpdateHandler(s, m, elements[1:])
		case conf.Prefix + "quit":
			QuitHandler(s, m)
		case conf.Prefix + "pb":
			PBHandler(s, m)
		default:
			_, _ = s.ChannelMessageSend(m.ChannelID, "codice sconosciuto, usa !help per sapere i codici che puoi usare")
		}
	}
}

//return time table and word table
func generatePBmessage(personalBest PB) (string, string) {
	bufTime := new(bytes.Buffer)
	bufWords := new(bytes.Buffer)

	dataTime := [][]string{}

	//time section
	if len(personalBest.Time.T15) != 0 {
		for _, t := range personalBest.Time.T15 {
			dataTime = append(dataTime, []string{"15 sec", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Time.T30) != 0 {
		for _, t := range personalBest.Time.T30 {
			dataTime = append(dataTime, []string{"30 sec", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Time.T60) != 0 {
		for _, t := range personalBest.Time.T60 {
			dataTime = append(dataTime, []string{"60 sec", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Time.T120) != 0 {
		for _, t := range personalBest.Time.T120 {
			dataTime = append(dataTime, []string{"120 sec", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	table := tablewriter.NewWriter(bufTime)
	table.SetHeader([]string{"TIPO", "LINGUA", "WPM", "PRECISIONE"})
	table.SetAutoMergeCells(true)

	for _, v := range dataTime {
		table.Append(v)
	}
	table.Render()

	//-------------------------------------------
	dataWords := [][]string{}

	//words section
	if len(personalBest.Words.W10) != 0 {
		for _, t := range personalBest.Words.W10 {
			dataWords = append(dataWords, []string{"10 parole", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Words.W25) != 0 {
		for _, t := range personalBest.Words.W25 {
			dataWords = append(dataWords, []string{"25 parole", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Words.W50) != 0 {
		for _, t := range personalBest.Words.W50 {
			dataWords = append(dataWords, []string{"50 parole", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Words.W100) != 0 {
		for _, t := range personalBest.Words.W100 {
			dataWords = append(dataWords, []string{"100 parole", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	table = tablewriter.NewWriter(bufWords)
	table.SetHeader([]string{"TIPO", "LINGUA", "WPM", "PRECISIONE"})
	table.SetAutoMergeCells(true)

	for _, v := range dataWords {
		table.Append(v)
	}
	table.Render()

	return bufTime.String(), bufWords.String()
}

func PBHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	u, err := GetUser(m.Author.ID)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hey c'é stato un problema\n\n errore: %v", err.Error()))
		return
	}
	err = u.GetPersonaBest()
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hey c'é stato un problema\n\n errore: %v", err.Error()))
		return
	}

	time, words := generatePBmessage(u.PersonalBest)

	message := fmt.Sprintf("Punteggi migliori di %s\n\n**tempo:**\n```%s```\n**parole:**\n```%s```", u.Mention(s), time, words)

	_, _ = s.ChannelMessageSend(m.ChannelID, message)
}

//register a new user in the database
func PartecipateHandler(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	var u User
	messageID := m.ID
	err := s.ChannelMessageDelete(m.ChannelID, messageID)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("problema nel cancellare il messaggio: %v", err.Error()))
	}
	u.UserID = m.Author.ID
	u.Username = m.Author.Username
	u.Email = params[1]
	u.Password = params[2]

	err = u.AddToDb()
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hey c'é stato un problema nella registrazione\n\n errore: %v", err.Error()))
		return
	}

	err = u.AddTyperRole(s)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hey c'é stato un problema nell'aggiungere il ruolo\n\n errore: %v", err.Error()))
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, u.Mention(s)+" é stato registrato correttamante")
}

//update will update user's info (discord username, password and email)
func UpdateHandler(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	var u User
	messageID := m.ID
	err := s.ChannelMessageDelete(m.ChannelID, messageID)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("problema nel cancellare il messaggio: %v", err.Error()))
	}
	u.UserID = m.Author.ID
	u.Username = m.Author.Username
	u.Email = params[1]
	u.Password = params[2]

	err = u.UpdateUser()
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hey c'é stato un problema nella modifica\n\n errore: %v", err.Error()))
		return
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, u.Mention(s)+", il tuo account é stato modificato correttamente")
}

//remove the typer from the database
func QuitHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	var u User
	u.UserID = m.Author.ID
	err := u.RemoveFromDB()
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hey c'é stato un problema nella rimozione\n\n errore: %v", err.Error()))
		return
	}
	err = u.RemoveTyperRole(s)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hey c'é stato un problema nella rimozione del ruolo\n\n errore: %v", err.Error()))
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, u.Mention(s)+", sei stato rimosso dalla classifica correttamente")
}

//return all codes knows by the bot
func HelpHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
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

	discord.AddHandler(MessageHandler)
	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening discord: %v", err.Error())
	}

	fmt.Println("wow i am working :D ")
	<-make(chan struct{})
}
