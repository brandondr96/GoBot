/*
This bot was inspired by the GroupMe-bot created by adammohammed.
It additionally contains an altered version of TextGenerate to allow
it to respond to replies in interesting ways. This basic bot can be 
extended to respond to many prompts in a manner similar to an AI.
*/

package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"strings"
	"os"
	"net"
    "math/rand"
    "time"
)

const FILE_NAME1 string = "default.txt"
const RESP_LENGTH int = 3
const BOT_ID = ""
const GROUP_ID = ""

// Structure for bot
type GroupMeBot struct {
	ID       string
	GroupID  string
}

// Structure of inbound message
type InboundMessage struct {
	Id           string                   `json:"id"`
	Avatar_url   string                   `json:"avatar_url"`
	Name         string                   `json:"name"`
	Sender_id    string                   `json:"sender_id"`
	Sender_type  string                   `json:"sender_type"`
	System       bool                     `json:"system"`
	Text         string                   `json:"text"`
	Source_guid  string                   `json:"source_guid"`
	Created_at   int                      `json:"created_at"`
	User_id      string                   `json:"user_id"`
	Group_id     string                   `json:"group_id"`
	Favorited_by []string                 `json:"favorited_by"`
	Attachments  []map[string]interface{} `json:"attachments"`
}

//Structure to hold the words
type data struct {
	word string
	nextWords []string
}

//Get next word from structure
func (dataS data) next(w []data) string {
	l := len(dataS.nextWords)
	if l == 0 {
		toPick := rand.Intn(len(w))
		return w[toPick].word
	}
	toPick := rand.Intn(l)
	return dataS.nextWords[toPick]
}

//Add word to repository
func (dataS data) addWord(toAdd string) data{
	dataS.nextWords = append(dataS.nextWords, toAdd)
	return dataS
}

//Default error checker
func check(e error) {
    if e != nil {
        panic(e)
    }
}

//Initialize conversational data from file
func initData(w []data, filename string) []data{
	dat, err := ioutil.ReadFile(filename)
    check(err)
    toParse := string(dat)
    toParse = strings.Replace(toParse,"---","",-1)
    ret := learnData(toParse,w)
    return ret
}

//Add message sender's name to response
func initName(w []data, name string) []data{
	toParse := name+", you"
	ret := learnData(toParse,w)
	return ret
} 

//Parse string to add to data
func learnData(toParse string, w []data) []data{
	allWords := strings.Fields(toParse)
	for i, aW := range allWords {
		//Check to ensure that data exists for word
		var t []string
		current := data{word: "", nextWords: t}
		bmark := false
		for _, d := range w {
			if d.word == aW {
				current = d //Since it takes latest appended entry, stale entries ignored
				bmark = true
			}
		}
		//Add word to wordlist if it doesn't exist
		if !bmark {
			current = data{word: aW, nextWords: t}
			w = append(w, current)
		}
		//Set the next word for data
		if i<len(allWords)-1 {
			current = current.addWord(allWords[i+1])
			w = append(w, current)
		}
	}
	return w
}

//Structure the response based on data
func respond(w []data) string{
	response := ""
	var t []string
	current := data{word: "", nextWords: t}
	l := len(w)
	if l==0 {
		return ""
	}
	j := 0
	for j<1 {
		toPick := rand.Intn(l)
		current = w[toPick]
		if current.word[0]<90 {
			break
		}
	}
	for i:=0;i<RESP_LENGTH;i++ {
		response = response + current.word
		if current.word != "" {
			response = response + " "
		}
		nextWord := current.next(w)
		for _, d := range w {
			if d.word == nextWord {
				current = d
			}
		}
	}
	response = response + current.word
	if current.word != "" {
		response = response + " "
	}
	if !strings.Contains(current.word,".") {
		for j<1 {
			nextWord := current.next(w)
			for _, d := range w {
				if d.word == nextWord {
					current = d
				}
			}
			response = response + current.word
			if strings.Contains(current.word,".") {
				break
			}
			if current.word != "" {
				response = response + " "
			}
		}
	}
	return response
}

// Structure of message to send
type OutboundMessage struct {
	ID   string `json:"bot_id"`
	Text string `json:"text"`
}

// Formats message and sends to group
func (b *GroupMeBot) SendMessage(outMessage string) (*http.Response, error) {
	msg := OutboundMessage{b.ID, outMessage}
	toSend, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	j_tosend := string(toSend)
	return http.Post("https://api.groupme.com/v3/bots/post", "application/json", strings.NewReader(j_tosend))
}

// Determines how to respond based on message 
func (b *GroupMeBot) HandleMessage(msg InboundMessage) {
	resp := ""
	////////////////////////////////////////////////
	//Insert criteria for response in this region
	// 
	// Example of simple response:
	if(strings.Contains(msg.Text,"Hello")){
		resp = fmt.Sprintf("Hi %v", msg.Name)
	}
	// Example of randomized response:
	if(strings.Contains(msg.Text,"Talk to me")){
		wordList := make([]data, 1)
		var t []string
		temp := data{word: "", nextWords: t}
		wordList[0] = temp
		wordList = initData(wordList,FILE_NAME1)
		wordList = initName(wordList,msg.Name)
		resp = fmt.Sprintf("%v", respond(wordList))
	}
	//
	////////////////////////////////////////////////
	if len(resp) > 0 {
		fmt.Println("Sending message: "+resp)
		_, err := b.SendMessage(resp)
		if err != nil {
			fmt.Println("Error sending message.")
		}
	}
}

// Determines if request should be responded to
func (b *GroupMeBot) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			defer req.Body.Close()
			var msg InboundMessage
			err := json.NewDecoder(req.Body).Decode(&msg)
			if err != nil {
				fmt.Println("Error parsing request.")
				msg.Sender_type = "bot"
			}
			if msg.Sender_type != "bot" {
				b.HandleMessage(msg)
			}
		} else {
			fmt.Println("Message does not require response.")
		}
	}
}

func main() {
	// Create random seed for responses
	t := time.Now().Hour()*time.Now().Second()+time.Now().Minute()
	rand.Seed(int64(t))

	// Create the GroupMeBot to use
	bot := GroupMeBot{ID: BOT_ID, GroupID: GROUP_ID}

	// Determine usable IP/port automatically
    conn, err := net.Dial("udp", "8.8.8.8:80")
    check(err)
    defer conn.Close()
    localAddr := conn.LocalAddr().String()
    idx := strings.LastIndex(localAddr, ":")
    pip := localAddr[0:idx]
    port := os.Getenv("PORT")

	// Create server to listen for posts in the group
	fmt.Println("Listening on: "+pip+":"+port)
	http.HandleFunc("/", bot.Handler())
	http.ListenAndServe(pip+":"+port, nil)
}
