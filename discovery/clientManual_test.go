package discovery

func main() {
	url := "http://localhost:8088"
	path := "/wiki"

	parent := newClient(url, "key")
	sub := newSubClient(parent, path)
}
