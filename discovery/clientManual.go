package discovery

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func Manual() {
	codesTest()
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

	file1, err := core.execute("PUT", "/file/test.txt", WithFile("discovery/files/testFile.txt"))

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("PUT Test File 1: " + string(file1.([]byte)))
	}

	file1, err = core.execute("GET", "/file/test.txt")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("GET Test File 1: " + string(file1.([]byte)))
	}

	file2, err := core.execute("PUT", "/file/test.txt", WithFile("discovery/files/testFile2.txt"))

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("PUT Test file 2: " + string(file2.([]byte)))
	}

	file2, err = core.execute("GET", "/file/test.txt")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("GET Test file 2: " + string(file2.([]byte)))
	}

	if string(file1.([]byte)) != string(file2.([]byte)) {
		fmt.Println("The files are different.")
	}

	file2, err = core.execute("DELETE", "/file/test.txt")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("DELETE Test file: " + string(file2.([]byte)))
	}

	_, err = core.execute("GET", "/file/test.txt")

	if err != nil {
		fmt.Println("Error in GET: " + err.Error())
	}
}

func codesTest() {
	queryflow := newClient("http://localhost:8088/v2/api", "")

	noContent := newSubClient(queryflow, "/blogs-search")

	mass, err := noContent.execute("GET", "", WithQueryParameters(map[string][]string{
		"q": {"google"},
	}))

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(mass.([]byte)))
	}
}
