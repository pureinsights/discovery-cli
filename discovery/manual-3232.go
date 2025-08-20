//go:build ignore
// +build ignore

package discovery

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

func Manual() {
	fmt.Println("\nTEST: Secret Create")
	secretCRUD := crud{
		getter{
			client: newClient("http://localhost:8080/v2/secret", ""),
		},
	}

	secret, err := secretCRUD.Create(gjson.Parse(`{
	"name": "test-secret",
	"active": true,
	"content": {
		"mechanism": "SCRAM-SHA-1", 
		"username": "user",
		"password": "password"
	}
	}`))

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(secret.String())
	}

	fmt.Println("\nTEST: GET created secret")
	secretId, err := uuid.Parse(secret.Get("id").String())

	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}

	getSecret, err := secretCRUD.Get(secretId)

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(getSecret.Raw)
	}

	fmt.Println("\nTEST: PUT to update secret")

	putSecret, err := secretCRUD.Update(secretId, gjson.Parse(`{
	"name": "test-secret-2",
	"active": true,
	"content": {
		"mechanism": "SCRAM-SHA-1", 
		"username": "user",
		"password": "key"
	}
	}`))

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(putSecret.Raw)
	}

	fmt.Println("\nTEST: GET updated secret")
	getSecret, err = secretCRUD.Get(secretId)

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(getSecret.Raw)
	}

	fmt.Println("\nTEST: DELETE secret")
	deleteSecret, err := secretCRUD.Delete(secretId)

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(deleteSecret.Raw)
	}

	fmt.Println("\nTEST: GET deleted secret")
	getSecret, err = secretCRUD.Get(secretId)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(getSecret.Raw)
		return
	}

	fmt.Println("\nTEST: GET All Secrets")
	secrets, err := secretCRUD.GetAll()

	if err != nil {
		fmt.Println(err)
		return
	} else {
		for i := range secrets {
			fmt.Println(secrets[i].Raw)
		}
	}
}
