package whd

import (
	"bytes"
	"crypto/tls"
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

	"github.com/hashicorp/go-retryablehttp"
)

type ProblemType struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"detailDisplayName,omitempty"`
}

type Location struct {
	Id           int           `json:"id,omitempty"`
	Type         string        `json:"type,omitempty"`
	Address      string        `json:"address,omitempty"`
	City         string        `json:"city,omitempty"`
	Name         string        `json:"locationName,omitempty"`
	PostalCode   string        `json:"postalCode,omitempty"`
	State        string        `json:"state,omitempty"`
	Country      string        `json:"country,omitempty"`
	CustomFields []CustomField `json:"locationCustomFields,omitempty"`
}

type Asset struct {
	Id             int           `json:"id,omitempty"`
	Type           string        `json:"type,omitempty"`
	AssetNumber    string        `json:"assetNumber,omitempty"`
	SerialNumber   string        `json:"serialNumber,omitempty"`
	NetworkAddress string        `json:"networkAddress,omitempty"`
	NetworkName    string        `json:"networkName,omitempty"`
	Location       Location      `json:"location,omitempty"`
	CustomFields   []CustomField `json:"assetCustomFields,omitempty"`
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
	Id    int    `json:"definitionId,omitempty"`
	Value string `json:"restValue,omitempty"`
}

type OrionAlert struct {
	Id   string            `json:"id,omitempty"`
	Data map[string]string `json:"data,omitempty"`
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
	Assets         []Asset        `json:"assets,omitempty"`
	Notes          []Note         `json:"notes,omitempty"`
	Attachments    []Attachment   `json:"attachments,omitempty"`
	ClientTech     ClientTech     `json:"clientTech,omitempty"`
	TechGroupLevel TechGroupLevel `json:"techGroupLevel,omitempty"`
	OrionAlert     OrionAlert     `json:"orionAlert,omitempty"`
	EmailTech      bool           `json:"emailTech,omitempty"`
	EmailClient    bool           `json:"emailClient"`
}

func CreateNote(uri string, user User, whdTicketId int, noteTxt string, sslVerify bool) (int, error) {
	var note Note
	note.JobTicket.Id = whdTicketId
	note.JobTicket.Type = "JobTicket"
	note.NoteText = noteTxt
	note.IsHidden = false
	return createNote(uri, user, whdTicketId, note, sslVerify)
}

func CreateHiddenNote(uri string, user User, whdTicketId int, noteTxt string, sslVerify bool) (int, error) {
	var note Note
	note.JobTicket.Id = whdTicketId
	note.JobTicket.Type = "JobTicket"
	note.NoteText = noteTxt
	note.IsHidden = true
	return createNote(uri, user, whdTicketId, note, sslVerify)
}

func createNote(uri string, user User, whdTicketId int, note Note, sslVerify bool) (int, error) {
	noteJsonStr, _ := json.Marshal(note)
	log.Printf("JSON Sent to WHD: %s", noteJsonStr)
	req, err := retryablehttp.NewRequest("POST", uri+urn+"TechNotes", bytes.NewBuffer(noteJsonStr))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	WrapAuth(req, user)

	var client *http.Client
	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 60,
		}
	} else {
		client = &http.Client{
			Timeout: time.Second * 60,
		}
	}

	retryclient := retryablehttp.NewClient()
	retryclient.RetryMax = RETRY_MAX
	retryclient.HTTPClient = client

	resp, err := retryclient.Do(req)
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

func GetNotes(uri string, user User, ticketID int, notes *[]Note, sslVerify bool) error {
	req, err := retryablehttp.NewRequest("GET", uri+urn+"TicketNotes", nil)
	if err != nil {
		return err
	}

	WrapAuth(req, user)

	q := req.URL.Query()

	q.Add("jobTicketId", fmt.Sprintf("%d", ticketID))
	q.Add("limit", "1000")

	req.URL.RawQuery = q.Encode()

	var client *http.Client
	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 90,
		}
	} else {
		client = &http.Client{
			Timeout: time.Second * 90,
		}
	}

	retryclient := retryablehttp.NewClient()
	retryclient.RetryMax = RETRY_MAX
	retryclient.HTTPClient = client

	resp, err := retryclient.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	if err = json.Unmarshal(data, &notes); err != nil {
		log.Println("Invalid JSON from WHD: ", data)
		return fmt.Errorf("Invalid JSON from WHD: %s", data)
	}

	return nil
}

func GetTicket(uri string, user User, id int, ticket *Ticket, sslVerify bool) error {
	req, err := retryablehttp.NewRequest("GET", uri+urn+"Ticket/"+strconv.Itoa(id), nil)
	if err != nil {
		return err
	}

	WrapAuth(req, user)

	var client *http.Client
	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 30,
		}
	} else {
		client = &http.Client{
			Timeout: time.Second * 60,
		}
	}

	retryclient := retryablehttp.NewClient()
	retryclient.RetryMax = RETRY_MAX
	retryclient.HTTPClient = client

	resp, err := retryclient.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	if err = json.Unmarshal(data, &ticket); err != nil {
		log.Println("Invalid JSON from WHD: ", data)
		return fmt.Errorf("Invalid JSON from WHD: %s", data)
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
func GetTickets(uri string, user User, qualifier string, limit uint, page uint, ticket *[]Ticket, sslVerify bool) error {
	req, err := retryablehttp.NewRequest("GET", uri+urn+"Tickets", nil)
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

	var client *http.Client
	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 90,
		}
	} else {
		client = &http.Client{}
	}

	retryclient := retryablehttp.NewClient()
	retryclient.RetryMax = RETRY_MAX
	retryclient.HTTPClient = client

	resp, err := retryclient.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	if err = json.Unmarshal(data, &ticket); err != nil {
		log.Println("Invalid JSON from WHD: ", data)
		return fmt.Errorf("Invalid JSON from WHD: %s", data)
	}

	return nil
}

func CreateUpdateTicket(uri string, user User, whdTicket Ticket, sslVerify bool) (int, error) {
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

	// remove custom fields with empty value
	tempCfs := make([]CustomField, 0, 10)
	for _, cf := range whdTicket.CustomFields {
		if cf.Value != "" {
			tempCfs = append(tempCfs, cf)
		}
	}
	whdTicket.CustomFields = tempCfs

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
		return createTicket(uri, user, []byte(ticketJsonStr), sslVerify)
	} else {
		return updateTicket(uri, user, whdTicket.Id, []byte(ticketJsonStr), sslVerify)
	}
}

func createTicket(uri string, user User, ticketJsonStr []byte, sslVerify bool) (int, error) {
	req, err := retryablehttp.NewRequest("POST", uri+urn+"Ticket", bytes.NewBuffer(ticketJsonStr))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	WrapAuth(req, user)

	var client *http.Client
	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	retryclient := retryablehttp.NewClient()
	retryclient.RetryMax = RETRY_MAX
	retryclient.HTTPClient = client

	resp, err := retryclient.Do(req)
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

func updateTicket(uri string, user User, id int, ticketJsonStr []byte, sslVerify bool) (int, error) {
	req, err := retryablehttp.NewRequest("PUT", uri+urn+"Ticket/"+strconv.Itoa(id), bytes.NewBuffer(ticketJsonStr))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	WrapAuth(req, user)

	var client *http.Client
	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 90,
		}
	} else {
		client = &http.Client{
			Timeout: time.Second * 90,
		}
	}

	retryclient := retryablehttp.NewClient()
	retryclient.RetryMax = RETRY_MAX
	retryclient.HTTPClient = client

	resp, err := retryclient.Do(req)
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

func GetAttachment(uri string, user User, attachmentId int, sslVerify bool) ([]byte, error) {
	req, err := retryablehttp.NewRequest("GET", uri+urn+"TicketAttachments/"+strconv.Itoa(attachmentId), nil)
	if err != nil {
		return nil, err
	}

	WrapAuth(req, user)
	req.Header.Set("accept", "application/octet")

	var client *http.Client
	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 120,
		}
	} else {
		client = &http.Client{
			Timeout: time.Second * 120,
		}
	}

	retryclient := retryablehttp.NewClient()
	retryclient.RetryMax = RETRY_MAX
	retryclient.HTTPClient = client

	resp, err := retryclient.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	return data, nil
}

func GetAttachmentAsBase64(uri string, user User, attachmentId int, sslVerify bool) (string, error) {
	data, err := GetAttachment(uri, user, attachmentId, sslVerify)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func UploadAttachment(uri string, user User, ticketId int, filename string, filedata []byte, sslVerify bool) (int, error) {
	return UploadAttachmentToEntity(uri, user, "jobTicket", ticketId, filename, filedata, sslVerify)
}

func UploadAttachmentToNote(uri string, user User, noteId int, filename string, filedata []byte, sslVerify bool) (int, error) {
	return UploadAttachmentToEntity(uri, user, "techNote", noteId, filename, filedata, sslVerify)
}

func UploadAttachmentToNoteFromFile(uri string, user User, noteId int, filename string, fullFilePath string, deleteFileAfter bool, sslVerify bool) (int, error) {
	// read in file
	filedata, err := ioutil.ReadFile(fullFilePath)

	if err != nil {
		return 0, fmt.Errorf("unable to read PDF file: %+v", err)
	}

	attId, err := UploadAttachmentToNote(uri, user, noteId, filename, filedata, sslVerify)

	if err != nil {
		return 0, err
	}

	if deleteFileAfter {
		os.Remove(fullFilePath)
	}

	return attId, nil
}

func UploadAttachmentToTicketFromFile(uri string, user User, ticketId int, filename string, fullFilePath string, deleteFileAfter bool, sslVerify bool) (int, error) {
	// read in file
	filedata, err := ioutil.ReadFile(fullFilePath)

	if err != nil {
		return 0, fmt.Errorf("unable to read PDF file: %+v", err)
	}

	attId, err := UploadAttachment(uri, user, ticketId, filename, filedata, sslVerify)

	if err != nil {
		return 0, err
	}

	if deleteFileAfter {
		os.Remove(fullFilePath)
	}

	return attId, nil
}

func UploadAttachmentToEntity(uri string, user User, entity string, entityId int, filename string, filedata []byte, sslVerify bool) (int, error) {
	cookieJar, _ := cookiejar.New(nil)

	// get session key to get JSESSIONID and wosid
	req, err := retryablehttp.NewRequest("GET", uri+urn+"Session", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("accept", "application/json")
	WrapAuth(req, user)

	client := &http.Client{
		Timeout: time.Second * 30,
		Jar:     cookieJar,
	}

	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	retryclient := retryablehttp.NewClient()
	retryclient.RetryMax = RETRY_MAX
	retryclient.HTTPClient = client

	resp, err := retryclient.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("error getting session key: bad status: %s", resp.Status)
	}

	data, _ := ioutil.ReadAll(resp.Body)
	log.Printf("session key resp: %s\n", data)

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
		Path:  "/",
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

	writer.WriteField("wosid", sessionKey)

	//log.Printf("Body: %+v", body)
	postUrl := fmt.Sprintf("%s/helpdesk/attachment/upload", uri)
	log.Printf("Sending Attachment POST to: %s", postUrl)
	req2, err := retryablehttp.NewRequest("POST", postUrl, body)
	if err != nil {
		return 0, err
	}
	cookieJar.SetCookies(req2.URL, cookies)
	client2 := &http.Client{
		Jar:     cookieJar,
		Timeout: time.Second * 120,
	}

	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client2.Transport = tr
	}

	req2.Header.Set("User-Agent", "Java/1.7.0_55")
	req2.Header.Set("accept", "application/json")
	req2.Header.Set("Pragma", "no-cache")
	req2.Header.Set("Cache-Control", "no-cache")
	req2.Header.Set("Connection", "keep-alive")
	req2.Header.Set("Accept", "text/html,image/gif,image/jpeg,*;q=.2,*/*;q=.2")
	// Don't forget to set the content type, this will contain the boundary.
	req2.Header.Set("Content-Type", writer.FormDataContentType())

	q := req2.URL.Query()

	q.Add("type", entity)
	q.Add("entityId", fmt.Sprintf("%d", entityId))
	q.Add("returnFields", "id")
	q.Add("sessionKey", sessionKey)

	// q.Add("subscriberId", "2")

	req2.URL.RawQuery = q.Encode()

	// WrapAuth(req2, User{
	// 	Name: user.Name,
	// 	Pass: sessionKey,
	// 	Type: SessionKeyAuth,
	// })

	log.Printf("Sending Attachment POST to: %+v", req2.URL)

	retryclient2 := retryablehttp.NewClient()
	retryclient2.RetryMax = 10
	retryclient2.HTTPClient = client2

	resp2, err := retryclient2.Do(req2)
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
	log.Printf("attachment upload response(%d): %s", resp2.StatusCode, string(data2))
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
