package main

func main() {
	app := NewApp()
	_, err := app.AddRoom(30000)
	if err != nil {
		panic(err)
	}
	_, err = app.AddRoom(30002)
	if err != nil {
		panic(err)
	}
	_, err = app.AddRoom(30004)
	if err != nil {
		panic(err)
	}
	_, err = app.AddRoom(30006)
	if err != nil {
		panic(err)
	}

	NewHTTPServer(app)
}
