package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// var host string = "http://34.222.135.134:8080"
var host string = "http://localhost:8080"
var testAccount string = "secondtest"
var testEmail string = "secondtest@test.com"
var testPassword string = "test"
var testGrindID string = "44c1c728-54ad-4809-ba30-a8cc23aea4f1"

func testRunning() {
	req, err := http.NewRequest("GET", host+"/api/v1/ping", nil)
	if err != nil {
		fmt.Println("could not create request")
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("could not send request")
		return
	}
	defer resp.Body.Close()

	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("could not read body")
		return
	}

	fmt.Println("Response body:", string(bodyResponse))
}

func testRegisterAPI() {
	body := map[string]string{
		"username": testAccount,
		"email":    testEmail,
		"password": testPassword,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Println("could not marshal request body")
		return
	}
	req, err := http.NewRequest("POST", host+"/api/v1/register", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Println("could not create request")
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("could not send request")
		return
	}
	defer resp.Body.Close()

	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("could not read body")
		return
	}

	fmt.Println("Response body:", string(bodyResponse))
}

func testLoginAPI() string {
	body := map[string]any{
		"email":    testEmail,
		"password": testPassword,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Println("could not marshal request body", err)
		return ""
	}

	req, err := http.NewRequest("POST", host+"/api/v1/login", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Println("could not create request", err)
		return ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("could not send request", err)
		return ""
	}
	defer resp.Body.Close()

	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("could not read body", err)
		return ""
	}

	var response map[string]any
	err = json.Unmarshal(bodyResponse, &response)
	if err != nil {
		fmt.Println("could not unmarshal response body", err)
		return ""
	}

	fmt.Println("Response token:", response["token"])
	return response["token"].(string)
}

func testCreateGrindAPI() {
	token := testLoginAPI()
	if token == "" {
		fmt.Println("could not get token")
		return
	}
	body := map[string]any{
		"startDate":    time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
		"duration":     30,
		"budget":       100,
		"participants": []string{testEmail},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Println("could not marshal request body", err)
		return
	}
	req, err := http.NewRequest("POST", host+"/api/v1/grinds", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if err != nil {
		fmt.Println("could not create request", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("could not send request", err)
		return
	}
	defer resp.Body.Close()
	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("could not read body", err)
		return
	}
	fmt.Println("Response body:", string(bodyResponse))
}

func testGetGrindAPI() {
	token := testLoginAPI()
	if token == "" {
		fmt.Println("could not get token")
		return
	}
	req, err := http.NewRequest("GET", host+"/api/v1/grinds/current", nil)
	if err != nil {
		fmt.Println("could not create request", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("could not send request", err)
		return
	}
	defer resp.Body.Close()
	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("could not read body", err)
		return
	}
	fmt.Println("Response body:", string(bodyResponse))
}

func testDeleteUserAPI() {
	req, err := http.NewRequest("DELETE", host+"/api/v1/users/delete", nil)
	if err != nil {
		fmt.Println("could not create request", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("could not send request", err)
		return
	}
	defer resp.Body.Close()
	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("could not read body", err)
		return
	}
	fmt.Println("Response body:", string(bodyResponse))
}

func testDeleteAllGrindsAPI() {
	req, err := http.NewRequest("DELETE", host+"/api/v1/grinds/delete-all", nil)
	if err != nil {
		fmt.Println("could not create request", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("could not send request", err)
		return
	}
	defer resp.Body.Close()
	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("could not read body", err)
		return
	}
	fmt.Println("Response body:", string(bodyResponse))
}

func testCreateInvitationAPI() {
	token := testLoginAPI()
	if token == "" {
		fmt.Println("could not get token")
		return
	}

	body := map[string]any{
		"grindID":          testGrindID,
		"participantEmail": "test@test.com",
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Println("could not marshal request body", err)
		return
	}
	req, err := http.NewRequest("POST", host+"/api/v1/messages/invitation/create", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Println("could not create request", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("could not send request", err)
		return
	}
	defer resp.Body.Close()
	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("could not read body", err)
		return
	}
	fmt.Println("Response body:", string(bodyResponse))
}

func main() {
	// testRunning()
	// testRegisterAPI()
	// testLoginAPI()
	// testCreateGrindAPI()
	// testGetGrindAPI()
	// testDeleteUserAPI()
	// testDeleteAllGrindsAPI()
	// testCreateInvitationAPI()

}
