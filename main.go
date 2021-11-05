package main

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/olekukonko/tablewriter"
)

type Leaderboard struct {
	Top15s []string
}

type Sort struct {
	Username string
	Wpm      float64
}

type Points struct {
	Username string
	Points   int
	Medals   map[string]int
}

func NewPoint() Points {
	medals := make(map[string]int)
	return Points{Medals: medals}
}

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
		case conf.Prefix + "lb":
			LeaderboardHandler(s, m)
		case conf.Prefix + "help":
			HelpHandler(s, m)
		default:
			_, _ = s.ChannelMessageSend(m.ChannelID, "codice sconosciuto, usa !help per sapere i codici che puoi usare")
		}
	}
}

//return a table of the personal bests of a specific user
func generatePBmessage(personalBest PB) string {
	buf := new(bytes.Buffer)
	data := [][]string{}

	//time section
	if len(personalBest.Time.T15) != 0 {
		for _, t := range personalBest.Time.T15 {
			data = append(data, []string{"15s", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Time.T30) != 0 {
		for _, t := range personalBest.Time.T30 {
			data = append(data, []string{"30s", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Time.T60) != 0 {
		for _, t := range personalBest.Time.T60 {
			data = append(data, []string{"60s", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Time.T120) != 0 {
		for _, t := range personalBest.Time.T120 {
			data = append(data, []string{"120s", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	//words section
	if len(personalBest.Words.W10) != 0 {
		for _, t := range personalBest.Words.W10 {
			data = append(data, []string{"10w", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Words.W25) != 0 {
		for _, t := range personalBest.Words.W25 {
			data = append(data, []string{"25w", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Words.W50) != 0 {
		for _, t := range personalBest.Words.W50 {
			data = append(data, []string{"50w", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	if len(personalBest.Words.W100) != 0 {
		for _, t := range personalBest.Words.W100 {
			data = append(data, []string{"100w", t.Language, fmt.Sprint(t.Wpm), fmt.Sprint(t.Accuracy)})
		}
	}

	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"TIPO", "LINGUA", "WPM", "PRECISIONE"})
	table.SetAutoMergeCellsByColumnIndex([]int{0, 0})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()

	return buf.String()
}

//handler for the /pb command
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

	table := generatePBmessage(u.PersonalBest)

	message := fmt.Sprintf("Punteggi migliori di %s\n\n```%s```", u.Mention(s), table)

	_, _ = s.ChannelMessageSend(m.ChannelID, message)
}

func SortUsersByLangAndType(users []User, lang string, t string) []Sort {
	var toSort []Sort
	var stats []Stats
	for _, u := range users {
		switch t {
		case "15 seconds":
			stats = u.PersonalBest.Time.T15
		case "30 seconds":
			stats = u.PersonalBest.Time.T30
		case "60 seconds":
			stats = u.PersonalBest.Time.T60
		case "120 seconds":
			stats = u.PersonalBest.Time.T120
		case "10 words":
			stats = u.PersonalBest.Words.W10
		case "25 words":
			stats = u.PersonalBest.Words.W25
		case "50 words":
			stats = u.PersonalBest.Words.W50
		case "100 words":
			stats = u.PersonalBest.Words.W100
		}

		for i, t := range stats {
			if t.Language == lang {
				var toAppend Sort
				toAppend.Username = u.Username
				toAppend.Wpm = stats[i].Wpm
				toSort = append(toSort, toAppend)
				break
			}
		}
	}

	sort.Slice(toSort, func(i, j int) bool {
		return toSort[i].Wpm > toSort[j].Wpm
	})
	return toSort
}

func GenerateLeaderboard(s *discordgo.Session, users []User) string {
	leaderboard := "**MonkeyType WPM  |  Score Board:**\n"
	lang := []string{"english", "italian"}
	types := []string{"15 seconds", "30 seconds", "60 seconds", "120 seconds", "10 words", "25 words", "50 words", "100 words"}

	var points []Points
	pointsValues := []int{4, 2, 1}

	for _, u := range users {
		toAppend := NewPoint()
		toAppend.Username = u.Username
		// toAppend.Medals["Gold"] = 0
		// toAppend.Medals["Silver"] = 0
		// toAppend.Medals["Bronze"] = 0
		points = append(points, toAppend)
	}

	for _, l := range lang {
		for _, t := range types {
			sorted := SortUsersByLangAndType(users, l, t)
			if len(sorted) > 0 {
				if strings.HasSuffix(t, "words") {
					leaderboard += "\n" + t + " mode - " + l + ": :pencil: "
				} else {
					leaderboard += "\n" + t + " mode - " + l + ": :alarm_clock: "
				}

				if l == "italian" {
					leaderboard += ":flag_it: \n"
				} else {
					leaderboard += ":flag_gb: \n"
				}

				for i, u := range sorted {
					switch i {
					case 0:
						leaderboard += ":first_place:"
						for j := 0; j < len(points); j++ {
							if points[j].Username == u.Username {
								points[j].Points += pointsValues[i]
								points[j].Medals["Gold"] += 1
							}
						}
					case 1:
						leaderboard += ":second_place:"
						for j := 0; j < len(points); j++ {
							if points[j].Username == u.Username {
								points[j].Points += pointsValues[i]
								points[j].Medals["Silver"] += 1
							}
						}
					case 2:
						leaderboard += ":third_place:"
						for j := 0; j < len(points); j++ {
							if points[j].Username == u.Username {
								points[j].Points += pointsValues[i]
								points[j].Medals["Bronze"] += 1
							}
						}
					}
					leaderboard += fmt.Sprintf(" %s %.2fwpm\n", u.Username, u.Wpm) //.Mention(s)
				}
			}
		}
	}

	leaderboard += "\n\n**MonkeyType WPM  |  Medal Table:**\n\n"

	sort.Slice(points, func(i, j int) bool {
		return points[i].Points > points[j].Points
	})

	for i, p := range points {
		switch i {
		case 0:
			leaderboard += ":first_place:"
		case 1:
			leaderboard += ":second_place:"
		case 2:
			leaderboard += ":third_place:"
		}
		leaderboard += fmt.Sprintf(" %s %dpt. (%d Gold, %d Silver, %d Bronze)\n", p.Username, p.Points, p.Medals["Gold"], p.Medals["Silver"], p.Medals["Bronze"]) //.Mention(s)
	}

	return leaderboard
}

func LeaderboardHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	users, err := GetAllTypers()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hey c'é stato un problema\n\n errore: %v", err.Error()))
		return
	}

	for _, u := range users {
		u.GetPersonaBest()
	}

	// fmt.Println(GenerateLeaderboard(s, users))
	s.ChannelMessageSend(m.ChannelID, GenerateLeaderboard(s, users))
}

//register a new user in the database
func PartecipateHandler(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	var u User
	messageID := m.ID
	err := s.ChannelMessageDelete(m.ChannelID, messageID)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("problema nel cancellare il messaggio: %v", err.Error()))
	}
	u.Userid = m.Author.ID
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
	u.Userid = m.Author.ID
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
	u.Userid = m.Author.ID
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
	_, _ = s.ChannelMessageSend(m.ChannelID, "```Ciao, sono Monkey-Bot v1.0 e sono un bot scritto da vano-.- :D.\nTerró traccia dei punti e delle run fatti dai giocatori UwU.\nI codici disponibili sono:\n\t⸭ /partecipate || <email monkey-type> <password monkey-type> || => Aggiunge un giocatore alla competizione e viene assegnato il tag @typer (tranquillo, il messaggio verrá cancellato automaticamente in poco tempo, quindi le tue credenziali sono al sicuro :D )\n\t⸭ /quit => rimuove il giocatore dalla competizione (puoi rientrare in qualsiasi momento)\n\t⸭ /update || <email monkey-type> <password monkey-type> || => modifica le vecchie credenziali con quelle nuove\n\t⸭ /pb => mostra una tabella con tutti le migliori run di un giocatore\n\t⸭ /lb => mostra la leaderboard completa tra tutti gli utenti e in fondo al messaggio i punti totalizzati dagli utenti\n\nSe trovi bug o problemi usando questo bot scrivi un dm a vano-.-\n\nEd ora... che la battagl... hum la competizione abbia iniziooo :D```")
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
