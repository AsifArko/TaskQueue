package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gitlab.com/pub-sub/utility"
	"log"
	"net/http"
)

const (
	PUBLISH     = "publish"
	SUBSCRIBE   = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
)

type PubSubController struct {
	Clients       []Client
	Subscriptions []Subscription
}

type Client struct {
	Id         string
	Connection *websocket.Conn
}

type Message struct {
	Action  string          `json:"action"`
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

type Subscription struct {
	Topic  string
	Client *Client
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func GetPubSubController() *PubSubController {
	return &PubSubController{}
}

func (ps *PubSubController) WebSocketHandler(w http.ResponseWriter, r *http.Request) {

	// Check origin if false then unauthorized
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	// Get websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// If a connection is established then a Client is created , unique ID is given to the client and the connection is assigned to the client
	client := Client{
		Id:         utility.GenerateUUID(),
		Connection: conn,
	}

	// When a client is created then adds the client to our client array
	ps.AddClient(client)

	// Notifies if a new client is added
	msg := fmt.Sprintf("New client %s is connected to the server . \tTotal client : %d", client.Id, len(ps.Clients))
	fmt.Println(msg)

	// When a client connection is established server listens for client messages
	for {
		// Try to read message from the client connection
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			log.Println("Could not read message from the client . Removing the client", err)

			// Remove the client
			ps.RemoveClient(client)

			msg := fmt.Sprintf("Total Client : %d\nTotal subscriptions : %d", len(ps.Clients), len(ps.Subscriptions))
			fmt.Println(msg)
			return
		}

		// If the message is parsed successfully then handle the message with MessageHandler
		ps.MessageHandler(client, messageType, payload)
	}
}

func (ps *PubSubController) MessageHandler(client Client, messageType int, payload []byte) *PubSubController {

	// Client message will unmarshal into our message struct
	var message Message
	err := json.Unmarshal(payload, &message)
	if err != nil {
		errorHandler(err, "Unmarshal Error")
		return ps
	}

	// Handle different operation based on different action of the user
	switch message.Action {
	// If the action is publish then publish the message on the channel
	case PUBLISH:
		ps.Publish(message.Topic, message.Message, nil)
		break
	// If action is subscribe client subscribe a particular topic
	case SUBSCRIBE:
		ps.Subscribe(&client, message.Topic)
		break
	// If action is unsubscribe client unsubscribes a particular topic
	case UNSUBSCRIBE:
		msg := fmt.Sprintf("Client %s want to unsubscribe the topic %s ", client.Id, message.Topic)
		fmt.Println(msg)
		ps.Unsubscribe(&client, message.Topic)
		break

	default:
		break
	}

	// After applying the appropriate action the pub-sub controller is returned to the WebSocket handler
	return ps
}

func (ps *PubSubController) AddClient(client Client) *PubSubController {

	// Appends a client to a topic/channel after client subscribes for a topic
	ps.Clients = append(ps.Clients, client)

	//  After appending a client to an existing client for a channel/topic a welcome payload message is sent to the client
	msg := fmt.Sprintf("Welcome Client %s" , client.Id)
	payload := []byte(msg)

	// Send the message to the newly added client
	client.Connection.WriteMessage(1, payload)

	return ps
}

func (ps *PubSubController) RemoveClient(client Client) *PubSubController {

	// first remove all subscriptions by this client
	for index, subscriber := range ps.Subscriptions {
		if client.Id == subscriber.Client.Id {
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	// remove client from the list
	for index, cli := range ps.Clients {
		if cli.Id == client.Id {
			ps.Clients = append(ps.Clients[:index], ps.Clients[index+1:]...)
		}
	}

	return ps
}

func (ps *PubSubController) GetSubscriptions(topic string, client *Client) []Subscription {

	var subscriptionList []Subscription

	for _, subscriber := range ps.Subscriptions {

		if client != nil {
			if subscriber.Client.Id == client.Id && subscriber.Topic == topic {
				subscriptionList = append(subscriptionList, subscriber)
			}
		} else {
			if subscriber.Topic == topic {
				subscriptionList = append(subscriptionList, subscriber)
			}
		}
	}
	return subscriptionList
}

func (ps *PubSubController) Subscribe(client *Client, topic string) *PubSubController {

	clientSubs := ps.GetSubscriptions(topic, client)

	if len(clientSubs) > 0 {
		msg := fmt.Sprintf("Client has subscribed this topic before")
		fmt.Println(msg)
		return ps
	}

	// create a new subscription
	newSubscription := Subscription{
		Topic:  topic,
		Client: client,
	}

	// Append the newly created subscription
	ps.Subscriptions = append(ps.Subscriptions, newSubscription)

	msg := fmt.Sprintf("New subscriber %s on topic %s , Total Clients : %d",client.Id, topic, len(ps.Subscriptions))
	fmt.Println(msg)

	return ps
}

func (ps *PubSubController) Publish(topic string, message []byte, excludeClient *Client) {

	subscriptions := ps.GetSubscriptions(topic, nil)

	for _, sub := range subscriptions {
		fmt.Printf("Client :: %s :: Message :: %s \n", sub.Client.Id, message)
		sub.Client.Send(message)
	}

}

func (client *Client) Send(message []byte) error {
	// Sends a message to a channel
	return client.Connection.WriteMessage(1, message)
}

func (ps *PubSubController) Unsubscribe(client *Client, topic string) *PubSubController {

	// If the client is in the subscription list remove it from the subscription array
	for index, sub := range ps.Subscriptions {
		if sub.Client.Id == client.Id && sub.Topic == topic {
			// found this subscription from client and we do need remove it
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}
	return ps

}

func errorHandler(err error, context string) error {
	msg := fmt.Sprintf("PubSub Controller :: %s::%s", context, err.Error())
	fmt.Println(msg)
	return err
}
