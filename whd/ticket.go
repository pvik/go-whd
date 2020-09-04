package whd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type ProblemType struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"detailDisplayName,omitempty"`
}

type Location struct {
	Id         int    `json:"id,omitempty"`
	Type       string `json:"type,omitempty"`
	Address    string `json:"address,omitempty"`
	City       string `json:"city,omitempty"`
	Name       string `json:"locationName,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	State      string `json:"state,omitempty"`
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
	Id    int    `json:"definitionId"`
	Value string `json:"restValue"`
}

type OrionAlert struct {
	Id   string            `json:"id"`
	Data map[string]string `json:"data"`
}

type Note struct {
	Id                  int          `json:"id,omitempty"`
	Date                time.Time    `json:"date,omitempty"`
	MobileNoteText      string       `json:"mobileNoteText,omitempty"` // Used for reading notes FROM whd
	PrettyUpdatedString string       `json:"prettyUpdatedString,omitempty"`
	NoteText            string       `json:"noteText,omitempty"` // Used to Create note TO whd
	Attachments         []Attachment `json:"attachments,omitempty"`
	IsHidden            bool         `json:"isHidden"`
	IsTechNote          bool         `json:"isTechNote"`
	JobTicket           struct {
		Id   int    `json:"id,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"jobticket,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type Attachment struct {
	Id            int       `json:"id,omitempty"`
	FileName      string    `json:"fileName,omitempty"`
	SizeString    string    `json:"sizeString,omitempty"`
	UploadDateUtc time.Time `json:"uploadDateUtc,omitempty"`
}

type ClientTech struct {
	Id          int    `json:"id,omitempty"`
	Type        string `json:"type,omitempty"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

type TechGroupLevel struct {
	Id             int    `json:"id,omitempty"`
	Type           string `json:"type,omitempty"`
	Level          int    `json:"level,omitempty"`
	LevelName      string `json:"email,omitempty"`
	ShortLevelName string `json:"displayName,omitempty"`
}

type Ticket struct {
	Id             int            `json:"id,omitempty"`
	Detail         string         `json:"detail,omitempty"`
	Subject        string         `json:"subject,omitempty"`
	LastUpdated    time.Time      `json:"lastUpdated,omitempty"`
	ReportDateUtc  string         `json:"reportDateUtc,omitempty"`
	LocationId     int            `json:"locationId,omitempty"`
	Location       Location       `json:"location,omitempty"`
	StatusTypeId   int            `json:"statusTypeId,omitempty"`
	StatusType     StatusType     `json:"statustype,omitempty"`
	PriorityTypeId int            `json:"priorityTypeId,omitempty"`
	PriorityType   PriorityType   `json:"prioritytype,omitempty"`
	ProblemType    ProblemType    `json:"problemtype,omitempty"`
	CustomFields   []CustomField  `json:"ticketCustomFields,omitempty"`
	Notes          []Note         `json:"notes,omitempty"`
	Attachments    []Attachment   `json:"attachments,omitempty"`
	ClientTech     ClientTech     `json:"clientTech,omitempty"`
	TechGroupLevel TechGroupLevel `json:"techGroupLevel,omitempty"`
	OrionAlert     OrionAlert     `json:"orionAlert,omitempty"`
	EmailTech      bool           `json:"emailTech,omitempty"`
	EmailClient    bool           `json:"emailClient"`
}

func CreateNote(uri string, user User, whdTicketId int, noteTxt string) (int, error) {
	var note Note
	note.JobTicket.Id = whdTicketId
	note.JobTicket.Type = "JobTicket"
	note.NoteText = noteTxt
	note.IsHidden = false
	return createNote(uri, user, whdTicketId, note)
}

func CreateHiddenNote(uri string, user User, whdTicketId int, noteTxt string) (int, error) {
	var note Note
	note.JobTicket.Id = whdTicketId
	note.JobTicket.Type = "JobTicket"
	note.NoteText = noteTxt
	note.IsHidden = true
	return createNote(uri, user, whdTicketId, note)
}

func createNote(uri string, user User, whdTicketId int, note Note) (int, error) {
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
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return 0, err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	//log.Println("Data:", string(data))
	if err = json.Unmarshal(data, &note); err != nil {
		log.Printf("Error unmarshalling response for create note: %s\n%s", string(data), err)
		return 0, fmt.Errorf("Error unmarshalling response for create note: %v\n", string(data))
	}

	if note.Reason == "" {
		return note.Id, nil
	} else {
		return 0, fmt.Errorf("Unable to create note: %s", note.Reason)
	}
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
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	if err = json.Unmarshal(data, &ticket); err != nil {
		log.Println("error unmarshalling: ", err)
		return err
	}

	return nil
}

// GetTickets allows you to query WHD for a list of tickets which matches
// a qualifier
// sample qualifier:
//  - all tickets including deleted: ((deleted %3D null) or (deleted %3D 0) or (deleted %3D 1))
//  - tickets in location ATL (location.locationName %3D 'ATL')
//  - tickets in stauts New (statustype.statusTypeName %3D 'Open')
// limit - limits the number of tickets returned, default is 25, max value is 100
// page  - Page of results to retrieve. Returns `limit` number of items, starting
//   with item `(page*limit)` of the search results
func GetTickets(uri string, user User, qualifier string, limit uint, page uint, ticket *[]Ticket) error {
	req, err := http.NewRequest("GET", uri+urn+"Tickets", nil)
	if err != nil {
		return err
	}

	WrapAuth(req, user)

	if limit == 0 {
		limit = 25
	} else if limit > 100 {
		limit = 100
	}

	if page == 0 {
		page = 1
	}

	q := req.URL.Query()
	q.Add("qualifier", qualifier)
	q.Add("limit", strconv.FormatUint(uint64(limit), 10))
	q.Add("page", strconv.FormatUint(uint64(page), 10))
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	if err = json.Unmarshal(data, &ticket); err != nil {
		log.Println("error unmarshalling: ", err)
		return err
	}

	return nil
}

func CreateUpdateTicket(uri string, user User, whdTicket Ticket) (int, error) {
	whdTicketMap := make(map[string]interface{})

	// reportDateUTC cannot be set when sending create/update transaction to WHD
	whdTicket.ReportDateUtc = ""

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
	//defer resp.Body.Close()
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return 0, err
	}
	defer resp.Body.Close()

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
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	var ticket Ticket
	if err = json.Unmarshal(data, &ticket); err != nil {
		log.Printf("error unmarshalling: %s\n%s", string(data), err)
		return 0, fmt.Errorf("Error: %v\n", string(data))
	}

	return ticket.Id, nil
}

func GetAttachment(uri string, user User, attachmentId int) ([]byte, error) {
	req, err := http.NewRequest("GET", uri+urn+"TicketAttachments/"+strconv.Itoa(attachmentId), nil)
	if err != nil {
		return nil, err
	}

	WrapAuth(req, user)
	req.Header.Set("accept", "application/octet")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	return data, nil
}

func GetAttachmentAsBase64(uri string, user User, attachmentId int) (string, error) {
	data, err := GetAttachment(uri, user, attachmentId)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func UploadAttachment(uri string, user User, ticketId int, filename string, filedata []byte) (int, error) {
	return UploadAttachmentToEntity(uri, user, "jobTicket", ticketId, filename, filedata)
}

func UploadAttachmentToNote(uri string, user User, noteId int, filename string, filedata []byte) (int, error) {
	return UploadAttachmentToEntity(uri, user, "techNote", noteId, filename, filedata)
}

func UploadAttachmentToNoteFromFile(uri string, user User, noteId int, filename string, fullFilePath string, deleteFileAfter bool) (int, error) {
	// read in file
	filedata, err := ioutil.ReadFile(fullFilePath)

	if err != nil {
		return 0, fmt.Errorf("unable to read PDF file: %+v", err)
	}

	attId, err := UploadAttachmentToNote(uri, user, noteId, filename, filedata)

	if err != nil {
		return 0, err
	}

	if deleteFileAfter {
		os.Remove(fullFilePath)
	}

	return attId, nil
}

func UploadAttachmentToTicketFromFile(uri string, user User, ticketId int, filename string, fullFilePath string, deleteFileAfter bool) (int, error) {
	// read in file
	filedata, err := ioutil.ReadFile(fullFilePath)

	if err != nil {
		return 0, fmt.Errorf("unable to read PDF file: %+v", err)
	}

	attId, err := UploadAttachment(uri, user, ticketId, filename, filedata)

	if err != nil {
		return 0, err
	}

	if deleteFileAfter {
		os.Remove(fullFilePath)
	}

	return attId, nil
}

func UploadAttachmentToEntity(uri string, user User, entity string, entityId int, filename string, filedata []byte) (int, error) {
	cookieJar, _ := cookiejar.New(nil)

	// get session key to get JSESSIONID and wosid
	req, err := http.NewRequest("GET", uri+urn+"Session", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("accept", "application/json")
	WrapAuth(req, user)

	client := &http.Client{
		Jar: cookieJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("error getting session key: bad status: %s", resp.Status)
	}

	data, _ := ioutil.ReadAll(resp.Body)
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		log.Println("error unmarshalling: ", err)
		return 0, err
	}

	sessionKey, ok := dataMap["sessionKey"].(string)
	if !ok {
		log.Println("invalid sessionKey in map")
		return 0, err
	}
	log.Printf("sessionKey retrieved: %s", sessionKey)
	defer TerminateSession(uri, sessionKey)

	// Upload attachment
	cookies := cookieJar.Cookies(req.URL)

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			log.Printf("Add JSESSIONID cookie to request")
			cookies = append(cookies, &http.Cookie{
				Name:  "JSESSIONID",
				Value: cookie.Value,
				//Path:   "/helpdesk",
				//Domain: host,
			})
		}
	}

	log.Printf("Cookies: %+v", cookies)
	cookies = append(cookies, &http.Cookie{
		Name:  "wosid",
		Value: sessionKey,
		Path:  "/helpdesk",
		//Domain: host,
	})
	log.Printf("Cookies with wosid: %+v", cookies)

	// check if tmp directory exists and create it
	if _, err := os.Stat("tmp"); os.IsNotExist(err) {
		err := os.Mkdir("tmp", os.ModeDir|os.ModeSetuid|os.ModeSetgid|0777)
		if err != nil {
			return 0, fmt.Errorf("Unable to create tmp directory to store attachments: %s", err)
		}
	}

	// save file
	filename = filepath.FromSlash("tmp/" + filename) // save in tmp directory
	err = ioutil.WriteFile(filename, filedata, 0644)
	if err != nil {
		log.Println("unable to save file")
		return 0, err
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Println("unable to read file")
		return 0, err
	}
	defer file.Close()

	// Prepare a form that you will submit to that URL.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return 0, err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return 0, err
	}

	//log.Printf("Body: %+v", body)
	postUrl := fmt.Sprintf("%s%s/helpdesk/attachment/upload?type=%s&entityId=%d&returnFields=id&sessionKey=%s", uri, urn, entity, entityId, sessionKey)
	log.Printf("Sending Attachment POST to: %s", postUrl)
	req2, err := http.NewRequest("POST", postUrl, body)
	if err != nil {
		return 0, err
	}
	cookieJar.SetCookies(req2.URL, cookies)
	client2 := &http.Client{
		Jar: cookieJar,
	}

	req2.Header.Set("accept", "text/html,image/gif,image/jpeg,*;q=.2,*/*;q=.2")
	req2.Header.Set("Pragma", "no-cache")
	req2.Header.Set("Connection", "keep-alive")
	// Don't forget to set the content type, this will contain the boundary.
	req2.Header.Set("Content-Type", writer.FormDataContentType())

	resp2, err := client2.Do(req2)
	if err != nil {
		log.Printf("The HTTP request failed when uploading attachment: %s\n", err)
		return 0, err
	}
	defer resp.Body.Close()

	// if resp2.StatusCode != http.StatusOK {
	// 	err = fmt.Errorf("error uploading attachment: bad status: %s", resp2.Status)
	// 	return 0, err
	// }

	data2, _ := ioutil.ReadAll(resp2.Body)
	log.Printf("attachment upload response (%d): %s", resp2.StatusCode, string(data2))
	var dataMap2 map[string]interface{}
	if err := json.Unmarshal(data2, &dataMap2); err != nil {
		log.Println("error unmarshalling att upload response: ", err)
		return 0, err
	}

	attIdFloat, ok := dataMap2["id"].(float64)
	if !ok {
		reasonStr, ok := dataMap2["reason"].(string)
		if !ok {
			log.Println("invalid attachment id in map")
			return 0, fmt.Errorf("Invalid attachment id in response")
		}

		log.Printf("Unable to Upload Attachment: %s", reasonStr)
		return 0, fmt.Errorf("Unable to upload attachment: %s", reasonStr)
	}

	return int(attIdFloat), nil
}
