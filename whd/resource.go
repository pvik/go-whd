package whd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type RequestType struct {
	Id       int
	ParentId int
	Name     string `json:"problemTypeName"`
}

func (rt RequestType) String() string {
	return rt.Name
}

func GetRequestTypeList(uri string, user User, result *map[int]RequestType) error {
	limit := 75

	resMap := make(map[int][]byte)
	if err := getResourceList(uri, user, "RequestTypes", limit, map[string]string{"list": "all"}, &resMap); err != nil {
		log.Printf("error retrieving resource list: %s\n", err)
		return err
	}

	for _, data := range resMap {
		l := make([]RequestType, 0, limit)

		if err := json.Unmarshal(data, &l); err != nil {
			log.Println("error unmarshalling: ", err)
			return err
		}

		// log.Printf("pg: %d | str: %s\n", pg, string(data))
		// log.Printf("unmarshalled: %+v\n", l)
		for _, rt := range l {
			(*result)[rt.Id] = rt
		}
	}

	return nil
}

func GetStatusTypeList(uri string, user User, list *map[int]string) error {
	limit := 50

	resMap := make([]interface{}, 0, limit)
	if err := getResourceListMap(uri, user, "StatusTypes", limit, nil, &resMap); err != nil {
		log.Printf("error retrieving resource list: %s\n", err)
		return err
	}

	parseResourceListMap("id", "statusTypeName", &resMap, list)
	return nil
}

func GetCustomFieldList(uri string, user User, list *map[int]string) error {
	limit := 50

	resMap := make([]interface{}, 0, limit)
	if err := getResourceListMap(uri, user, "CustomFieldDefinitions", limit, nil, &resMap); err != nil {
		log.Printf("error retrieving resource list: %s\n", err)
		return err
	}

	parseResourceListMap("id", "label", &resMap, list)
	return nil
}

func GetLocationList(uri string, user User, list *map[int]string) error {
	limit := 250

	resMap := make([]interface{}, 0, limit)
	if err := getResourceListMap(uri, user, "Locations", limit, map[string]string{"qualifier": "((deleted=null)or(deleted=0))"}, &resMap); err != nil {
		log.Printf("error retrieving resource list: %s\n", err)
		return err
	}

	parseResourceListMap("id", "locationName", &resMap, list)
	return nil
}

func GetPriorityTypeList(uri string, user User, list *map[int]string) error {
	limit := 10

	resMap := make([]interface{}, 0, limit)
	if err := getResourceListMap(uri, user, "PriorityTypes", limit, nil, &resMap); err != nil {
		log.Printf("error retrieving resource list: %s\n", err)
		return err
	}

	parseResourceListMap("id", "priorityTypeName", &resMap, list)
	return nil
}

func parseResourceListMap(idLabel string, valueLabel string, resMap *[]interface{}, list *map[int]string) {
	for _, data := range *resMap {
		v := data.(map[string]interface{})
		(*list)[int(v[idLabel].(float64))] = v[valueLabel].(string)
	}
}

func getResourceList(uri string, user User, resource string, limit int, params map[string]string, result *map[int][]byte) error {
	tmp := make([]interface{}, limit, limit)

	for pg := 1; len(tmp) == limit; pg++ {

		data, err := getResourceListPage(uri, user, resource, limit, pg, params)
		if err != nil {
			log.Printf("error retrieving: %s\n")
			return err
		}

		if err = json.Unmarshal(data, &tmp); err != nil {
			log.Println("error unmarshalling: ", err)
			return err
		}
		(*result)[pg] = data
	}

	return nil
}

func getResourceListMap(uri string, user User, resource string, limit int, params map[string]string, result *[]interface{}) error {
	tmp := make([]interface{}, limit, limit)

	for pg := 1; len(tmp) == limit; pg++ {

		data, err := getResourceListPage(uri, user, resource, limit, pg, params)
		if err != nil {
			log.Printf("error retrieving: %s\n")
			return err
		}

		if err = json.Unmarshal(data, &tmp); err != nil {
			log.Printf("error unmarshalling: %s\nError: %s", string(data), err)
			return err
		}
		*result = append(*result, tmp...)
	}

	return nil
}

func getResourceListPage(uri string, user User, resource string, limit int, page int, params map[string]string) ([]byte, error) {
	log.Printf("Get %s | limit: %d | page: %d", resource, limit, page)

	req, err := http.NewRequest("GET", uri+urn+resource, nil)
	if err != nil {
		return nil, err
	}

	WrapAuth(req, user)

	q := req.URL.Query()
	q.Add("limit", strconv.Itoa(limit))
	q.Add("page", strconv.Itoa(page))

	for p, v := range params {
		q.Add(p, v)
	}

	req.URL.RawQuery = q.Encode()
	//log.Printf("URL: %s\n", req.URL.String())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return nil, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	return data, nil

}
