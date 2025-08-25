package discovery

type labelsClient struct {
    crud
}

type secretsClient struct {
    crud
}

type credentialsClient struct {
    crud
    searcher
    cloner
}

type serversClient struct {
    crud
    cloner
    searcher
}
Ping(id uuid.UUID) serverClient

type filesClient struct {
    client
}
Upload() ({JsonObject}, error)
Retrieve(key string) ([]byte, error)
List() ({JsonArray}, error)
Delete() ({JsonObject}, error)

type maintenanceClient struct {
     client
}
Log(componentName, level, loggerName string) ({JsonArray}, error)