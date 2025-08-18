package discovery

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func Manual() {
	fileCRUD()
}

func tutorialTest() {
	queryflow := newClient("http://localhost:8088/v2/api", "")

	wiki := newSubClient(queryflow, "/wikis-search")

	mass, err := wiki.execute("GET", "", WithQueryParameters(map[string][]string{
		"q": {"Where is most mass located in the solar system?"},
	}))

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(mass.([]byte)))
}

func secretCRUD() {
	core := newClient("http://localhost:8080/v2", "")

	secret, err := core.execute("POST", "/secret", WithBody(`{
  "name": "test-secret",
  "active": true,
  "content": {
    "mechanism": "SCRAM-SHA-1", 
    "username": "user",
    "password": "password"
  }
}`), WithContentType("application/json"))

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(secret.([]byte)))

	secretId := gjson.Get(string(secret.([]byte)), "id").String()

	getSecret, err := core.execute("GET", "/secret/"+secretId, []RequestOption{}...)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(getSecret.([]byte)))

	putSecret, err := core.execute("PUT", "/secret/"+secretId, WithBody(`{
  "name": "test-secret-2",
  "active": true,
  "content": {
    "mechanism": "SCRAM-SHA-1", 
    "username": "user",
    "password": "password"
  }
}`), WithContentType("application/json"))

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(putSecret.([]byte)))

	getSecret, err = core.execute("GET", "/secret/"+secretId, []RequestOption{}...)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(getSecret.([]byte)))

	deleteSecret, err := core.execute("DELETE", "/secret/"+secretId, []RequestOption{}...)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(deleteSecret.([]byte)))

	_, err = core.execute("GET", "/secret/"+secretId, []RequestOption{}...)

	if err != nil {
		fmt.Println(err)
	}
}

func fileCRUD() {
	core := newClient("http://localhost:8080/v2", "")

	file, err := core.execute("PUT", "/file/test.txt", WithFile("test.txt", "discovery/testFile.txt"))

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(file.([]byte)))
	}
}
