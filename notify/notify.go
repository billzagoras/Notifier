package notify

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Notify library default constructor.
func New() API {
	api := API{}

	return api
}

func (*API) Notify(request *Request, responseChannel chan interface{}, url string) {

	// Turn on logging and define/initialize variables.
	log.SetOutput(io.MultiWriter(os.Stdout))
	var (
		response *Response
		method   = "POST"
	)

	// Marshal request object to json format.
	payload, err := json.Marshal(&request)
	if err != nil {
		// Feed response channel.
		responseChannel <- response
		return
	}

	// Set up a basic http client and the request.
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		responseChannel <- response
		return
	}
	log.Println("url")
	log.Println(url)

	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Perform the request.
	res, err := client.Do(req)
	if err != nil {
		responseChannel <- response
		return
	}
	defer res.Body.Close()

	// Read the response object.
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		responseChannel <- response
		return
	}

	// Unmarshal the response body to a json response object.
	err = json.Unmarshal(body, &response)
	if err != nil {
		responseChannel <- response
		return
	}

	responseChannel <- response
	return
}
