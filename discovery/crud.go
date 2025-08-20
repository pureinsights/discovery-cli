package discovery

import (
	"net/http"

	"github.com/tidwall/gjson"

	"github.com/google/uuid"
)

type getter struct {
	client
}

func (getter getter) Get(id uuid.UUID) (gjson.Result, error) {
	response, err := execute(getter.client, http.MethodGet, "/"+id.String())
	if err != nil {
		return gjson.Result{}, err
	} else {
		return response, nil
	}
}

func (getter getter) GetAll() ([]gjson.Result, error) {
	response, err := execute(getter.client, http.MethodGet, "")
	if err != nil {
		return []gjson.Result{}, err
	} else {
		return response.Array(), nil
	}
}

type crud struct {
	getter
}

func (crud crud) Create(config gjson.Result) (gjson.Result, error) {
	response, err := execute(crud.client, http.MethodPost, "", WithJSONBody(config.Raw))
	if err != nil {
		return gjson.Result{}, err
	} else {
		return response, nil
	}
}

func (crud crud) Update(id uuid.UUID, config gjson.Result) (gjson.Result, error) {
	response, err := execute(crud.client, http.MethodPut, "/"+id.String(), WithJSONBody(config.Raw))
	if err != nil {
		return gjson.Result{}, err
	} else {
		return response, nil
	}
}

func (crud crud) Delete(id uuid.UUID) (gjson.Result, error) {
	response, err := execute(crud.client, http.MethodDelete, "/"+id.String())
	if err != nil {
		return gjson.Result{}, err
	} else {
		return response, nil
	}
}
