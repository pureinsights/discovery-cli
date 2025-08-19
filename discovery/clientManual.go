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
	}

	fmt.Println(string(mass.([]byte)))
}

type secret struct {
	Name                 string `json:"name"`
	Labels               string `json:"labels"`
	Active               bool   `json:"active"`
	Id                   string `json:"id"`
	CreationTimestamp    string `json:"creationTimestamp"`
	LastUpdatedTimestamp string `json:"lastUpdatedTimestamp"`
}

func (s secret) String() string {
	return fmt.Sprintf(
		"{Name: %q, Labels: %q, Active: %t, Id: %q, CreationTimestamp: %q, LastUpdatedTimestamp: %q}",
		s.Name,
		s.Labels,
		s.Active,
		s.Id,
		s.CreationTimestamp,
		s.LastUpdatedTimestamp,
	)
}

func secretCRUD() {
	core := newClient("http://localhost:8080/v2", "")

	secret, err := execute[secret](core, "POST", "/secret", WithBody(`{
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

	fmt.Println(secret.String())

	secretId := secret.Id

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
	} else {
		fmt.Println(string(getSecret.([]byte)))
	}

	deleteSecret, err := core.execute("DELETE", "/secret/"+secretId, []RequestOption{}...)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(deleteSecret.([]byte)))
	}

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
	// The Ingestion API must have no pipelines and run on the port 8081
	ingestion := newClient("http://localhost:8081/v2/pipeline", "")

	noContent, err := ingestion.execute("GET", "")

	if err != nil {
		fmt.Println(err)
	} else {
		stringNoContent := string(noContent.([]byte))
		fmt.Println(stringNoContent)
		if stringNoContent != "" {
			fmt.Printf("No Content Test failed: Expected an empty body, got %s \n", stringNoContent)
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
				fmt.Printf("Incorrect error: expected 404, got %d", errorStruct.Status)
			} else {
				fmt.Println("Error 404 was correctly received")
			}
		}
	}

	putSecret, err := core.execute("PUT", "/secret", WithBody(`{
  "name": "mongo-secret",
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

}
