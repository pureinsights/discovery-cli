//go:build ignore
// +build ignore

package discovery

import (
	"fmt"
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
	} else {
		fmt.Println(string(mass))
	}
}

func secretCRUD() {
	core := newClient("http://localhost:8080/v2", "")

	secret, err := execute(core, "POST", "/secret", WithBody(`{
  "name": "test-secret",
  "active": true,
  "content": {
    "mechanism": "SCRAM-SHA-1", 
    "username": "user",
    "password": "password"
  }
}`), WithHeader("Content-Type", "application/json"))

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(secret.String())
	}

	secretId := secret.Get("id").String()

	getSecret, err := core.execute("GET", "/secret/"+secretId, []RequestOption{}...)

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(string(getSecret))
	}

	putSecret, err := core.execute("PUT", "/secret/"+secretId, WithBody(`{
	"name": "test-secret-2",
	"active": true,
	"content": {
		"mechanism": "SCRAM-SHA-1", 
		"username": "user",
		"password": "password"
	}
	}`), WithHeader("Content-Type", "application/json"))

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(string(putSecret))
	}

	getSecret, err = core.execute("GET", "/secret/"+secretId, []RequestOption{}...)

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(string(getSecret))
	}

	deleteSecret, err := core.execute("DELETE", "/secret/"+secretId, []RequestOption{}...)

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(string(deleteSecret))
	}

	_, err = core.execute("GET", "/secret/"+secretId, []RequestOption{}...)

	if err != nil {
		fmt.Println(err)
		return
	}
}

func fileCRUD() {
	core := newClient("http://localhost:8080/v2", "")

	file1, err := core.execute("PUT", "/file/test.txt", WithFile("discovery/files/testFile.txt"))

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("PUT Test File 1: " + string(file1))
	}

	file1, err = core.execute("GET", "/file/test.txt")

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("GET Test File 1: " + string(file1))
	}

	file2, err := core.execute("PUT", "/file/test.txt", WithFile("discovery/files/testFile2.txt"))

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("PUT Test file 2: " + string(file2))
	}

	file2, err = core.execute("GET", "/file/test.txt")

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("GET Test file 2: " + string(file2))
	}

	if string(file1) != string(file2) {
		fmt.Println("The files are different.")
	}

	file2, err = core.execute("DELETE", "/file/test.txt")

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("DELETE Test file: " + string(file2))
	}

	_, err = core.execute("GET", "/file/test.txt")

	if err != nil {
		fmt.Println("Error in GET: " + err.Error())
		return
	}
}

func codesTest() {
	// The Ingestion API must have no pipelines and run on the port 8081
	ingestion := newClient("http://localhost:8081/v2/pipeline", "")

	noContent, err := ingestion.execute("GET", "")

	if err != nil {
		fmt.Println(err)
		return
	} else {
		stringNoContent := string(noContent)
		fmt.Println(stringNoContent)
		if stringNoContent != "" {
			fmt.Printf("No Content Test failed: Expected an empty body, got %s \n", stringNoContent)
			return
		} else {
			fmt.Println("No Content Received")
		}
	}

	core := newClient("http://localhost:8080/v2", "")

	_, err = core.execute("GET", "/secret/5f125024-1e5e-4591-9fee-365dc20eeeed")

	if err != nil {
		fmt.Println("Error in GET: " + err.Error())
		errorStruct, ok := err.(Error)
		if ok {
			if errorStruct.Status != 404 {
				fmt.Printf("Incorrect error: expected 404, got %d \n", errorStruct.Status)
				return
			} else {
				fmt.Println("Error 404 was correctly received")
			}
		}
	}

	putSecret, err := core.execute("POST", "/secret", WithBody(`{
  "name": "mongo-secret",
  "active": true,
  "content": {
    "mechanism": "SCRAM-SHA-1", 
    "username": "user",
    "password": "password"
  }
}`), WithHeader("Content-Type", "application/json"))

	if err != nil {
		fmt.Println("Error in POST: " + err.Error())
		errorStruct, ok := err.(Error)
		if ok {
			if errorStruct.Status != 409 {
				fmt.Printf("Incorrect error: expected 409, got %d \n", errorStruct.Status)
				return
			} else {
				fmt.Println("Error 409 was correctly received")
			}
		}
	} else {
		fmt.Println(string(putSecret))
		fmt.Println("Duplicated name failed: Should have received an error.")
		return
	}

	putSecret, err = core.execute("PUT", "/secret", WithBody(`{
  "name": "mongo-secret",
  "active": true,
  "content": {
    "mechanism": "SCRAM-SHA-1", 
    "username": "user",
    "password": "password"
  }
}`), WithHeader("Content-Type", "application/json"))

	if err != nil {
		fmt.Println("Error in PUT: " + err.Error())
		errorStruct, ok := err.(Error)
		if ok {
			if errorStruct.Status != 405 {
				fmt.Printf("Incorrect error: expected 405, got %d \n", errorStruct.Status)
				return
			} else {
				fmt.Println("Error 405 was correctly received")
			}
		}
	} else {
		fmt.Println(string(putSecret))
		fmt.Println("Method not allowed failed: Should have received an error.")
		return
	}

	twitterAPI := newClient("https://api.twitter.com/2/users/me", "")

	_, err = twitterAPI.execute("GET", "")

	if err != nil {
		fmt.Println("Error in GET: " + err.Error())
		errorStruct, ok := err.(Error)
		if ok {
			if errorStruct.Status != 401 {
				fmt.Printf("Incorrect error: expected 401, got %d \n", errorStruct.Status)
				return
			} else {
				fmt.Println("Error 401 was correctly received")
			}
		}
	}
}
