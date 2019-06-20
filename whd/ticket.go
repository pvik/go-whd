package whd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ProblemType struct {
	Id   int
	Name string `json:"detailDisplayName"`
}

type CustomField struct {
	Id    int         `json:"definitionId"`
	Value interface{} `json:"restValue"`
}

type Note struct {
	Id             int `json:"id"`
	Date           time.Time
	MobileNoteText string
}

type Ticket struct {
	Id             int
	Detail         string
	Subject        string
	LastUpdated    time.Time
	LocationId     int
	StatusTypeId   int
	PriorityTypeId int
	ProblemType    ProblemType
	CustomFields   []CustomField `json:"ticketCustomFields"`
	Notes          []Note
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
