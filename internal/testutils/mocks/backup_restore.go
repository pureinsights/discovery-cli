package mocks

import (
	"net/http"

	"github.com/tidwall/gjson"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
)

// WorkingBackupRestore mocks a working backup restore.
type WorkingBackupRestore struct{}

// Export returns zip bytes as if the request worked successfully.
func (g *WorkingBackupRestore) Export() ([]byte, string, error) {
	return []byte("Exportfiles"), "export-20251110T1455.zip", nil
}

// Import implements the interface.
func (g *WorkingBackupRestore) Import(discoveryPackage.OnConflict, string) (gjson.Result, error) {
	return gjson.Parse("{\n  \"Credential\": [\n    {\n      \"id\": \"3b32e410-2f33-412d-9fb8-17970131921c\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"458d245a-6ed2-4c2b-a73f-5540d550a479\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"46cb4fff-28be-4901-b059-1dd618e74ee4\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"4957145b-6192-4862-a5da-e97853974e9f\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"5c09589e-b643-41aa-a766-3b7fc3660473\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"6dd2177f-0196-42d8-9468-0053a5c1127a\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"822b2d33-20a2-4df4-aebf-a1cee5acdce7\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"837196a6-1ac5-4b0c-a24a-4b9d092e6260\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"84f66cd4-a28b-4e66-94e1-a3dc9f083bbd\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"8c243a1d-9384-421d-8f99-4ef28d4e0ab0\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"9be0e625-a510-46c5-8130-438823f849c2\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"9d438628-5981-49c5-9426-0d328fd16370\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"b4d9ee85-9775-49fa-8dfb-b3e5ce2f619e\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"f643fe55-18db-48e4-9d3f-335d0f5f5348\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"f64a5451-3716-45c4-8158-350f30e1cbdb\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"f6c4585b-4e65-4359-9aee-e995ba09f69e\",\n      \"status\": 200\n    }\n  ],\n  \"Server\": [\n    {\n      \"id\": \"21029da3-041c-43b5-a67e-870251f2f6a6\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"226e8a0b-5016-4ebe-9963-1461edd39d0a\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"2b839453-ddad-4ced-8e13-2c7860af60a7\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"3ab2e3c0-5459-4f19-9e89-f8282d111eba\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"3edc9c72-a875-49d7-8929-af09f3e9c01c\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"6f2ddfd5-154a-4534-8f29-b1569ac23b8a\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"6ffc7784-481e-4da8-8ee3-6817d15a757c\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"74160a12-bcf6-4778-8944-4a4b2a7c4be1\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"741df47e-208f-47c1-812f-53cc62c726af\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"7cd191c0-d8ab-44f7-923f-2e32d044ced2\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"8f14c11c-bb66-49d3-aa2a-dedff4608c17\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"a798cd5b-aa7a-4fc5-9292-1de6fe8e8b7f\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"f6950327-3175-4a98-a570-658df852424a\",\n      \"status\": 200\n    }\n  ]\n}\n"), nil
}

// FailingBackupRestore mocks a failing backup restore.
type FailingBackupRestore struct{}

// Export returns an error as if the request failed.
func (g *FailingBackupRestore) Export() ([]byte, string, error) {
	return []byte(nil), "discovery.zip", discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// Import implements the interface.
func (g *FailingBackupRestore) Import(discoveryPackage.OnConflict, string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}
