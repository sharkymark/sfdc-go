package main

import (
	"fmt"
	"os"
	"net/http"
	"net/url"
	"strings"
	"io/ioutil"
	"encoding/json"
	
)

const salesforceAPIBaseURL = "/services/data/v54.0"

type Salesforce struct {
	Url string
	ConsumerKey string
	ConsumerSecret string
	AccessToken string
}

// Define a generic interface to handle different Salesforce objects
type SalesforceObject interface{}

// Define an Account struct
type Account struct {
    Id          string `json:"Id"`
    Name        string `json:"Name"`
    Type        string `json:"Type"`
    Description string `json:"Description"`
    Website     string `json:"Website"`
    Industry    string `json:"Industry"`
}

// Define a Contact struct
type Contact struct {
    Id        	string `json:"Id"`
    FirstName 	string `json:"FirstName"`
    LastName  	string `json:"LastName"`
	Account 	Account `json:"Account"`
    Email     	string `json:"Email"`
    Phone     	string `json:"Phone"`
	Description string `json:"Description"`
}

//
// printEnvVars prints Salesforce credentials
//
// It prints the contents of the Salesforce struct
// 


func printSalesforceCreds(s *Salesforce) {

	fmt.Println("Salesforce URL:", s.Url)
	fmt.Println("Salesforce Consumer Key:", s.ConsumerKey)
	fmt.Println("Salesforce Consumer Secret:", s.ConsumerSecret)
	fmt.Println("Generated Salesforce Access Token:", s.AccessToken)

}

//
// GetEnvVars retrieves Salesforce credentials from environment variables
//
// Required environment variables:
// - SALESFORCE_URL_1
// - SALESFORCE_CONSUMER_KEY_1
// - SALESFORCE_CONSUMER_SECRET_1

func getEnvVars() Salesforce{

	s := Salesforce{}

	requiredVars := []string{"SALESFORCE_URL_1", "SALESFORCE_CONSUMER_KEY_1", "SALESFORCE_CONSUMER_SECRET_1"}
	missingVars := []string{}

	for _, varName := range requiredVars {
			value := os.Getenv(varName)
			if value == "" {
					missingVars = append(missingVars, varName)
			} else {
					switch varName {
					case "SALESFORCE_URL_1":
							s.Url = value
					case "SALESFORCE_CONSUMER_KEY_1":
							s.ConsumerKey = value
					case "SALESFORCE_CONSUMER_SECRET_1":
							s.ConsumerSecret = value
					}
			}
	}

    if len(missingVars) > 0 {
        fmt.Println("\nError: Missing required environment variables:\n")
        for _, varName := range missingVars {
                fmt.Printf("  - %s\n", varName)
        }
		fmt.Printf("\n\nExiting %s ...\n\n", os.Args[0])
        os.Exit(1)
    }

	return s


}

// getAccessToken retrieves an access token from Salesforce
//
//	Requires:
//		- Salesforce struct with
//			Consumer Key, Consumer Secret, Url
//
func getAccessToken(s *Salesforce) (string, error) {
    form := url.Values{}
    form.Add("grant_type", "client_credentials")
    form.Add("client_id", s.ConsumerKey)
    form.Add("client_secret", s.ConsumerSecret)

    // 1. Print request details for debugging
    // fmt.Printf("Sending POST request to: %s\n", s.Url)
    // fmt.Printf("Form data: %v\n", form)

    req, err := http.NewRequest("POST", s.Url+"/services/oauth2/token", strings.NewReader(form.Encode()))
    if err != nil {
        return "", fmt.Errorf("error creating request: %w", err)
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("error making request: %w", err)
    }
    defer resp.Body.Close()


    // 2. Check for successful response status code
    if resp.StatusCode != http.StatusOK {
        defer resp.Body.Close() // Close body even on errors
        body, _ := ioutil.ReadAll(resp.Body)
        return "", fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
    }

    // 3. Print response details for debugging
    // fmt.Printf("Received response with status code: %d\n", resp.StatusCode)



    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("error reading response body: %w", err)
    }
    // fmt.Println("Received response body:")
    // fmt.Println(string(body))

    var result map[string]interface{}
	err = json.Unmarshal(body, &result)
    if err != nil {
        return "", fmt.Errorf("error parsing JSON response: %w, response body: %s", err, string(body))
    }

    accessToken, ok := result["access_token"].(string)
    if !ok {
        return "", fmt.Errorf("couldn't parse access token, response body: %s", string(body))
    }

	s.AccessToken = accessToken

    return accessToken, nil
}

// querySalesforce executes SOQL queries
// 
// Requires:
//		- Salesforce struct with access token
//		- A string with the SOQL query
//		- A destination interface to store the query results
//
func querySalesforce(salesforce *Salesforce, soql string, dest interface{}) error {
    // Create the HTTP request
    req, err := http.NewRequest("GET", salesforce.Url+salesforceAPIBaseURL+"/query?q="+url.QueryEscape(soql), nil)
    if err != nil {
        return fmt.Errorf("error creating request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+salesforce.AccessToken)

    // Make the API call
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("error making request: %w", err)
    }
    defer resp.Body.Close()

    // Check for successful response
    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        return fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
    }

    // Parse the JSON response into the provided destination
    err = json.NewDecoder(resp.Body).Decode(dest)
    if err != nil {
        return fmt.Errorf("error parsing JSON response: %w", err)
    }

    return nil
}



// getAccountsByName retrieves a list of accounts from Salesforce
//
// Requires:
//   - 	Salesforce struct with access token
//	 - 	A string with the account name filter
//

func getAccountsByName(salesforce *Salesforce, nameFilter string) ([]Account, error) {
    soql := fmt.Sprintf("SELECT Id, Name, Type, Description, Website, Industry FROM Account WHERE Name LIKE '%%%s%%' ORDER BY Name", nameFilter)
    var accountsResponse struct {
        Records []Account `json:"records"`
    }
    err := querySalesforce(salesforce, soql, &accountsResponse)
    return accountsResponse.Records, err
}

// getContactsByName retrieves a list of accounts from Salesforce
//
// Requires:
//   - 	Salesforce struct with access token
//	 - 	A string with the account name filter
//
func getContacts(salesforce *Salesforce, contactFilter string) ([]Contact, error) {
    soql := fmt.Sprintf("SELECT Id, FirstName, LastName, Email, Account.Name, Phone, Description FROM Contact "+
	"WHERE LastName LIKE '%%%s%%' OR FirstName LIKE '%%%s%%'"+
	"OR Account.Name LIKE '%%%s%%' OR Email LIKE '%%%s%%' ORDER BY LastName",
	contactFilter,contactFilter,contactFilter,contactFilter)
    var contactsResponse struct {
        Records []Contact `json:"records"`
    }
    err := querySalesforce(salesforce, soql, &contactsResponse)
    return contactsResponse.Records, err
}

// printAccounts prints a list of accounts
//
// Requires:
//   - 	A slice of Account structs
//

func printAccounts(accounts []Account) {
    for _, account := range accounts {
        fmt.Printf("\nName: %s\nIndustry: %s\nType: %s\nWebsite: %s\nDescription:\n\n%s\n\n", account.Name, account.Industry, account.Type, account.Website, account.Description)
    }
}

// printContacts prints a list of contacts
//
// Requires:
//   - 	A slice of Contact structs
//

func printContacts(contacts []Contact) {
    for _, contact := range contacts {
        fmt.Printf("\nContact Name: %s, %s\nAccount: %s\nEmail: %s\nDescription:\n\n%s\n\n", contact.LastName, contact.FirstName, contact.Account.Name, contact.Email, contact.Description)
    }
}

// main is the entry point of the Salesforce CLI tool.
//
// Requires:
//   - 	A Salesforce Connected App with OAuth 2.0 to generate
//		the Consumer Key and Consumer Secret
//

func main() {

	fmt.Println("\n\nSalesforce CLI")
	fmt.Println("--------------\n\n")

	// Get and print environment variables with Salesforce credentials
	salesforceCreds := getEnvVars()

	// Get access token
	_, err := getAccessToken(&salesforceCreds)
	if err != nil {
		fmt.Println("Error getting access token:", err)
		return
	}

	printSalesforceCreds(&salesforceCreds)

	// Main menu
	for {
		fmt.Println("\nMain Menu:\n")
		fmt.Println("1. Search accounts")
		fmt.Println("2. Search contacts")
		fmt.Println("3. Exit\n")

		var option string
		fmt.Print("Enter your option: ")
		fmt.Scanln(&option)

		switch option {
		case "1":
			
			var nameFilter string
			fmt.Print("\nEnter account name filter: ")
			fmt.Scanln(&nameFilter)
		
			accounts, err := getAccountsByName(&salesforceCreds, nameFilter)
			if err != nil {
				fmt.Println("Error retrieving accounts:", err)
				return
			}

			printAccounts(accounts)

		case "2":
			
			var contactFilter string
			fmt.Print("\nEnter contact first, last name, email or account name filter: ")
			fmt.Scanln(&contactFilter)
		
			contacts, err := getContacts(&salesforceCreds, contactFilter)
			if err != nil {
				fmt.Println("Error retrieving contacts:", err)
				return
			}

			printContacts(contacts)

		case "3":
			fmt.Println("\nExiting...")
			return
		default:
			fmt.Println("\nInvalid option")
		}
	}

}
