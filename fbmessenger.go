package fbmessenger

import (
	//"reflect"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"errors"
)

var accessToken = os.Getenv("ACCESS_TOKEN")
var verifyToken = os.Getenv("VERIFY_TOKEN")

// const ...
const (
	EndPoint = "https://graph.facebook.com/v2.6/me/messages"
)

// receivedMessage ...
type receivedMessage struct {
	Object string  `json:"object"`
	Entry  []entry `json:"entry"`
}

// entry ...
type entry struct {
	ID        int64        `json:"id"`
	Time      int64        `json:"time"`
	Messagings []messaging `json:"messaging"`
}

// messaging ...
type messaging struct {
	Sender    Sender    `json:"sender"`
	Recepient Recepient `json:"recipient"`
	Timestamp int64     `json:"timestamp"`
	Message   message   `json:"message"`
}

// Sender ...
type Sender struct {
	ID int64 `json:"id"`
}

// Recepient ...
type Recepient struct {
	ID int64 `json:"id"`
}

// message ...
type message struct {
	MID  string `json:"mid"`
	Seq  int64  `json:"seq"`
	Text string `json:"text"`
}

// sendMessage ...
type sendMessage struct {
	Recepient Recepient `json:"recipient"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
}

// Event express one messaging
type Event struct {
	Sender Sender
	Recepient Recepient
	Content interface{}
}

// TextContent express content of one message
type TextContent struct {
	Text string
}

func (_messaging *messaging) Event(_recepient Recepient) Event {
	event := Event{}
	event.Sender = _messaging.Sender
	event.Recepient = _recepient
	event.Content = TextContent{_messaging.Message.Text}
	return event
}

// Listen call callback function given when requested from Facebook service.
func Listen(callback func(Event)) {
	handleReceiveMessage = callback
	http.HandleFunc("/", webhookHandler)
	http.HandleFunc("/webhook", webhookHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

var handleReceiveMessage func(Event)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if r.URL.Query().Get("hub.verify_token") == verifyToken {
			fmt.Fprintf(w, r.URL.Query().Get("hub.challenge"))
		} else {
			fmt.Fprintf(w, "Error, wrong validation token")
		}
	}
	if r.Method == "POST" {
		var receivedMessage receivedMessage
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Print(err)
		}
		if err = json.Unmarshal(b, &receivedMessage); err != nil {
			log.Print(err)
		}
		messagingEvents := receivedMessage.Entry[0].Messagings
		for _, event := range messagingEvents {
			if &event.Message != nil && event.Message.Text != "" {
				handleReceiveMessage(event.Event(Recepient{0}))
			}
		}
		fmt.Fprintf(w, "Success")
	}
}

// Send method send Event
func Send(event Event) error {
	switch content := event.Content.(type) {
	case TextContent:
		sendTextMessage(event.Recepient, content.Text)
	default:
		return errors.New("Event.Content type is invalid")
	}
	return nil
}

func sendTextMessage(recepient Recepient, sendText string) {
	m := new(sendMessage)
	m.Recepient = recepient

	log.Print("------------------------------------------------------------")
	log.Print(m.Message.Text)
	log.Print("------------------------------------------------------------")

	m.Message.Text = sendText

	log.Print(m.Message.Text)

	b, err := json.Marshal(m)
	if err != nil {
		log.Print(err)
	}
	req, err := http.NewRequest("POST", EndPoint, bytes.NewBuffer(b))
	if err != nil {
		log.Print(err)
	}
	values := url.Values{}
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{Timeout: time.Duration(30 * time.Second)}
	res, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	defer res.Body.Close()
	var result map[string]interface{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Print(err)
	}
	log.Print(result)
}
