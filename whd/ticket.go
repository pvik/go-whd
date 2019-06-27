package whd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ProblemType struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"detailDisplayName,omitempty"`
}

type CustomField struct {
	Id    int         `json:"definitionId"`
	Value interface{} `json:"restValue"`
}

type Note struct {
	Id             int       `json:"id"`
	Date           time.Time `json:"date"`
	MobileNoteText string    `json:"mobileNoteText`
}

type Ticket struct {
	Id             int           `json:"id,omitempty"`
	Detail         string        `json:"detail"`
	Subject        string        `json:"subject"`
	LastUpdated    time.Time     `json:"lastUpdated"`
	LocationId     int           `json:"locationId"`
	StatusTypeId   int           `json:"statusTypeId"`
	PriorityTypeId int           `json:"priorityTypeId"`
	ProblemType    ProblemType   `json:"problemtype"`
	CustomFields   []CustomField `json:"ticketCustomFields,omitempty"`
	Notes          []Note        `json:"notes,omitempty"`
}

func GetTicket(uri string, user User, id int, ticket *Ticket) error {
	req, err := http.NewRequest("GET", uri+urn+"Ticket/"+strconv.Itoa(id), nil)
	if err != nil {
		return err
	}

	WrapAuth(req, user)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return err
	}

	data, _ := ioutil.ReadAll(resp.Body)

	if err = json.Unmarshal(data, &ticket); err != nil {
		log.Println("error unmarshalling: ", err)
		return err
	}

	return nil
}

func CreateUpdateTicket(uri string, user User, whdTicket Ticket) (int, error) {
	whdTicketMap := make(map[string]interface{})
	interim, _ := json.Marshal(whdTicket)
	json.Unmarshal(interim, &whdTicketMap)

	delete(whdTicketMap, "lastUpdated")
	whdTicketMap["customFields"] = whdTicketMap["ticketCustomFields"]
	delete(whdTicketMap, "ticketCustomFields")

	ticketJsonStr, _ := json.Marshal(whdTicketMap)
	if whdTicket.Id == 0 {
		return createTicket(uri, user, []byte(ticketJsonStr))
	} else {
		return updateTicket(uri, user, whdTicket.Id, []byte(ticketJsonStr))
	}
}

func createTicket(uri string, user User, ticketJsonStr []byte) (int, error) {
	req, err := http.NewRequest("POST", uri+urn+"Ticket", bytes.NewBuffer(ticketJsonStr))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	WrapAuth(req, user)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return 0, err
	}

	data, _ := ioutil.ReadAll(resp.Body)

	var ticket Ticket
	if err = json.Unmarshal(data, &ticket); err != nil {
		log.Println("error unmarshalling: ", err)
		return 0, err
	}

	return ticket.Id, nil
}

func updateTicket(uri string, user User, id int, ticketJsonStr []byte) (int, error) {
	req, err := http.NewRequest("PUT", uri+urn+"Ticket/"+strconv.Itoa(id), bytes.NewBuffer(ticketJsonStr))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	WrapAuth(req, user)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return 0, err
	}

	data, _ := ioutil.ReadAll(resp.Body)

	var ticket Ticket
	if err = json.Unmarshal(data, &ticket); err != nil {
		log.Println("error unmarshalling: ", err)
		return 0, err
	}

	return ticket.Id, nil
}
