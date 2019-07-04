package whd

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type Location struct {
	Id   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

type PriorityType struct {
	Id   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

type StatusType struct {
	Id   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

type CustomField struct {
	Id    int         `json:"definitionId"`
	Value interface{} `json:"restValue"`
}

type Note struct {
	Id             int       `json:"id,omitempty"`
	Date           time.Time `json:"date,omitempty"`
	MobileNoteText string    `json:"mobileNoteText,omitempty` // Used for reading notes FROM whd
	NoteText       string    `json:notetext,omitempty`        // Used to Create note TO whd
	JobTicket      struct {
		Id   int    `json:id,omitempty`
		Type string `json:type,omitempty`
	} `json:jobticket,omitempty`
}

type Ticket struct {
	Id             int           `json:"id,omitempty"`
	Detail         string        `json:"detail,omitempty"`
	Subject        string        `json:"subject,omitempty"`
	LastUpdated    time.Time     `json:"lastUpdated,omitempty"`
	LocationId     int           `json:"locationId,omitempty"`
	Location       Location      `json:"location,omitempty"`
	StatusTypeId   int           `json:"statusTypeId,omitempty"`
	StatusType     StatusType    `json:"statustype,omitempty"`
	PriorityTypeId int           `json:"priorityTypeId,omitempty"`
	PriorityType   PriorityType  `json:"prioritytype,omitempty"`
	ProblemType    ProblemType   `json:"problemtype,omitempty"`
	CustomFields   []CustomField `json:"ticketCustomFields,omitempty"`
	Notes          []Note        `json:"notes,omitempty"`
}

func CreateNote(uri string, user User, whdTicketId int, noteTxt string) (int, error) {
	var note Note
	note.JobTicket.Id = whdTicketId
	note.JobTicket.Type = "JobTicket"
	note.NoteText = noteTxt

	noteJsonStr, _ := json.Marshal(note)
	log.Printf("JSON Sent to WHD: %s", noteJsonStr)
	req, err := http.NewRequest("POST", uri+urn+"TechNotes", bytes.NewBuffer(noteJsonStr))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	WrapAuth(req, user)

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return 0, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	log.Println("Data:", string(data))
	if err = json.Unmarshal(data, &note); err != nil {
		log.Printf("Error unmarshalling response for create note: %s\n%s", string(data), err)
		return 0, fmt.Errorf("Error unmarshalling response for create note: %v\n", string(data))
	}

	return note.Id, nil
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

	if whdTicket.LocationId != 0 {
		whdTicket.Location = Location{
			Id:   whdTicket.LocationId,
			Type: "Location",
		}
	}

	if whdTicket.PriorityTypeId != 0 {
		whdTicket.PriorityType = PriorityType{
			Id:   whdTicket.PriorityTypeId,
			Type: "PriorityType",
		}
	}

	if whdTicket.StatusTypeId != 0 {
		whdTicket.StatusType = StatusType{
			Id:   whdTicket.StatusTypeId,
			Type: "StatusType",
		}
	}

	interim, _ := json.Marshal(whdTicket)
	json.Unmarshal(interim, &whdTicketMap)

	delete(whdTicketMap, "lastUpdated")
	whdTicketMap["customFields"] = whdTicketMap["ticketCustomFields"]
	delete(whdTicketMap, "ticketCustomFields")

	if whdTicket.ProblemType.Id == 0 {
		delete(whdTicketMap, "problemtype")
	}
	if whdTicket.Location.Id == 0 {
		delete(whdTicketMap, "location")
	}
	if whdTicket.PriorityTypeId == 0 {
		delete(whdTicketMap, "prioritytype")
	}
	if whdTicket.StatusTypeId == 0 {
		delete(whdTicketMap, "statustype")
	}

	ticketJsonStr, _ := json.Marshal(whdTicketMap)
	log.Printf("JSON Sent to WHD: %s", ticketJsonStr)
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
	defer resp.Body.Close()
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return 0, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	log.Println("Data:", string(data))
	var ticket Ticket
	if err = json.Unmarshal(data, &ticket); err != nil {
		log.Printf("error unmarshalling: %s\n%s", string(data), err)
		return 0, fmt.Errorf("Error: %v\n", string(data))
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
		log.Printf("error unmarshalling: %s\n%s", string(data), err)
		return 0, fmt.Errorf("Error: %v\n", string(data))
	}

	return ticket.Id, nil
}
