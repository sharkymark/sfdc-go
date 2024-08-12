package main

import (
	"fmt"
	"os"
	"net/http"
	"net/url"
	"strings"
	"io/ioutil"
	"encoding/json"
    "time"
    //"io"
	
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

// Define a Task struct
type Task struct {
    Id        	string `json:"Id"`
    Subject 	string `json:"Subject"`
    Description string `json:"Description"`
    Who 		Contact `json:"Who"`
    CreatedBy 	Contact `json:"CreatedBy"`
    Account 	Account `json:"Account"`
    CreatedAt 	string `json:"CreatedDate"`
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

func setVars(deployment *Salesforce, requiredVars []string, optional bool) {
    missingVars := []string{}

    for _, varName := range requiredVars {
        value := os.Getenv(varName)
        if value == "" {
            missingVars = append(missingVars, varName)
        } else {
            switch varName {
            case requiredVars[0]:
                deployment.Url = value
            case requiredVars[1]:
                deployment.ConsumerKey = value
            case requiredVars[2]:
                deployment.ConsumerSecret = value
            }
        }
    }

    // If it's not optional and variables are missing, print an error and exit
    if !optional && len(missingVars) > 0 {
        fmt.Println("\nError: Missing required environment variables for deployment:\n")
        for _, varName := range missingVars {
            fmt.Printf("  - %s\n", varName)
        }
        os.Exit(1)
    }
}

// isValidDeployment checks if the given Salesforce deployment has valid credentials
func isValidDeployment(s *Salesforce) bool {
    return s.Url != "" && s.ConsumerKey != "" && s.ConsumerSecret != ""
}

//
// GetEnvVars retrieves Salesforce credentials from environment variables for up to 2 deployments
//
// Required environment variables:
// - SALESFORCE_URL_1
// - SALESFORCE_CONSUMER_KEY_1
// - SALESFORCE_CONSUMER_SECRET_1
// 
// Optional environment variables:
// - SALESFORCE_URL_2
// - SALESFORCE_CONSUMER_KEY_2
// - SALESFORCE_CONSUMER_SECRET_2

func getEnvVars(d1, d2 *Salesforce) Salesforce {

	requiredVars1 := []string{"SALESFORCE_URL_1", "SALESFORCE_CONSUMER_KEY_1", "SALESFORCE_CONSUMER_SECRET_1"}
    requiredVars2 := []string{"SALESFORCE_URL_2", "SALESFORCE_CONSUMER_KEY_2", "SALESFORCE_CONSUMER_SECRET_2"}

    // Set the first deployment (required)
    setVars(d1, requiredVars1, false)

    // Set the second deployment (optional)
    setVars(d2, requiredVars2, true)

    // Check if credentials are valid and take the first valid deployment
    if isValidDeployment(d1) {
        return *d1
    } else if isValidDeployment(d2) {
        return *d2
    }

     // If neither deployment is valid, return an error
     fmt.Println("Error: Missing required environment variables for both deployments.")
     os.Exit(1)

     return Salesforce{}

}

func changeDeployment(d1, d2 *Salesforce) Salesforce {
    fmt.Println("\nAvailable Deployments:\n")
    if d1.Url != "" {
        fmt.Println("1. ", d1.Url)
    } else {
        fmt.Println("1.  Not configured")
    }
    if d2.Url != "" {
        fmt.Println("2. ", d2.Url)
    } else {
        fmt.Println("2.  Not configured")
    }

    fmt.Print("\nSelect the deployment you want to switch to (1 or 2): ")
    var choice string
    fmt.Scanln(&choice)

    switch choice {
    case "1":
        if d1.Url == "" {
            fmt.Println("\nDeployment 1 is not configured.")
            return *d2
        }
        return *d1
    case "2":
        if d2.Url == "" {
            fmt.Println("\nDeployment 2 is not configured.")
            return *d1
        }
        return *d2
    default:
        fmt.Println("Invalid choice. No changes made.")
        return *d1 // Return the current deployment if invalid choice
    }
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
    //fmt.Printf("Sending POST request to: %s\n", s.Url)
    //fmt.Printf("Form data: %v\n", form)

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
func querySalesforce(s *Salesforce, soql string, dest interface{}) error {
    // Create the HTTP request
    req, err := http.NewRequest("GET", s.Url+salesforceAPIBaseURL+"/query?q="+url.QueryEscape(soql), nil)
    if err != nil {
        return fmt.Errorf("error creating request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+s.AccessToken)

    // Make the API call
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("error making request: %w", err)
    }
    defer resp.Body.Close()

    // Check for successful response
    if resp.StatusCode == http.StatusUnauthorized {
        // 401 Unauthorized indicates session expired, try to refresh the token
        fmt.Println("Session expired, refreshing token...\n")

        // Get a new access token
        _, err := getAccessToken(s)
        if err != nil {
            return fmt.Errorf("error refreshing access token: %w", err)
        }

        // Retry the request with the new token
        req.Header.Set("Authorization", "Bearer "+s.AccessToken)
        resp, err = client.Do(req)
        if err != nil {
            return fmt.Errorf("error making request after token refresh: %w", err)
        }
        defer resp.Body.Close()

        // Check the response again after retrying
        if resp.StatusCode != http.StatusOK {
            body, _ := ioutil.ReadAll(resp.Body)
            return fmt.Errorf("unexpected status code after token refresh: %d, response body: %s", resp.StatusCode, string(body))
        }
    } else if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        return fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
    }

    // Parse the JSON response into the provided destination

    /* for debugging json payload 
    // Step 1: Read the entire response body
    responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("error reading response body: %w", err)
    }

    // Step 2: Print the raw JSON response
    fmt.Println("Raw JSON response:", string(responseBody))
    */

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

// getContacts retrieves a list of contacts from Salesforce
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

// getContactsByName retrieves a list of accounts from Salesforce
//
// Requires:
//   - 	Salesforce struct with access token
//	 - 	A string with the account name filter
//
func getTasks(salesforce *Salesforce, taskFilter string) ([]Task, error) {
    soql := fmt.Sprintf("SELECT Id, Subject, Description, Who.FirstName, Who.LastName, CreatedBy.FirstName, CreatedBy.LastName, Account.Name, CreatedDate FROM Task "+
	"WHERE Subject LIKE '%%%s%%' OR "+
    "Who.LastName LIKE '%%%s%%' OR Who.FirstName LIKE '%%%s%%' "+
	"OR Account.Name LIKE '%%%s%%' ORDER BY CreatedDate ASC",
	taskFilter,taskFilter,taskFilter,taskFilter)
    var tasksResponse struct {
        Records []Task `json:"records"`
    }
    err := querySalesforce(salesforce, soql, &tasksResponse)
    return tasksResponse.Records, err
}

// FormatCreatedAt takes a date string and returns it formatted as "YYYY-MM-DD HH:MM AM/PM".
func FormatCreatedAt(dateStr string) (string, error) {
    // Parse the input date string (adjust the layout according to your input format)
    createdAt, err := time.Parse("2006-01-02T15:04:05.000-0700", dateStr) // Adjust as necessary
    if err != nil {
        return "", fmt.Errorf("error parsing date: %w", err)
    }

    // Format the date to the desired output
    formattedDate := createdAt.Format("2006-01-02 03:04 PM")
    return formattedDate, nil
}

// printAccounts prints a list of accounts
//
// Requires:
//   - 	A slice of Account structs
//

func printAccounts(accounts []Account) {

    if len(accounts) == 0 {
        fmt.Println("\nNo accounts found.")
        return
    }

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

    if len(contacts) == 0 {
        fmt.Println("\nNo contacts found.")
        return
    }

    for _, contact := range contacts {
        fmt.Printf("\nContact Name: %s, %s\nAccount: %s\nEmail: %s\nDescription:\n\n%s\n\n", contact.LastName, contact.FirstName, contact.Account.Name, contact.Email, contact.Description)
    }
}

// printTasks prints a list of tasks
//
// Requires:
//   - 	A slice of Task structs
//

func printTasks(tasks []Task) {

    if len(tasks) == 0 {
        fmt.Println("\nNo tasks found.")
        return
    }

    for _, task := range tasks {

        // Call the reusable date formatting function
        formattedDate, err := FormatCreatedAt(task.CreatedAt)
        if err != nil {
            fmt.Println(err) // Handle the error as needed
            continue // Skip this task if there's an error
        }

        // Initialize the contact information variables
        contactInfo := ""
        if task.Who.FirstName != "" || task.Who.LastName != "" {
            contactInfo = fmt.Sprintf("Contact: %s, %s\n", task.Who.FirstName, task.Who.LastName)
        }

        fmt.Printf("\n%s\nCreatedBy: %s %s\nAccount: %s\n%sSubject: %s\nDescription:\n\n%s\n\n", formattedDate, task.CreatedBy.FirstName, task.CreatedBy.LastName, task.Account.Name, contactInfo, task.Subject, task.Description)
    }
}

func printObjectCounts(salesforce *Salesforce) {
    // Define a list of SOQL queries for counting different objects
    queries := map[string]string{
        "accounts":      "SELECT COUNT() FROM Account",
        "contacts":      "SELECT COUNT() FROM Contact",
        "opportunities": "SELECT COUNT() FROM Opportunity",
        "tasks":         "SELECT COUNT() FROM Task",
    }

    // Iterate through the queries and print counts
    fmt.Println("\nDeployment counts:\n")
    for object, query := range queries {
        var countResponse struct {
            TotalSize int `json:"totalSize"`
        }
        
        // Execute the query and handle errors
        err := querySalesforce(salesforce, query, &countResponse)
        if err != nil {
            fmt.Printf("Error retrieving count for %s: %s\n", object, err)
            continue // Skip to the next object on error
        }

        // Print the count for the object
        fmt.Printf("  %s: %d\n", object, countResponse.TotalSize)
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

    // Define pointers to Salesforce structs for each deployment and current deployment
    var deployment1, deployment2, currentDeployment Salesforce

	// Get and print environment variables with Salesforce credentials
	currentDeployment = getEnvVars(&deployment1, &deployment2)

	// Get access token
	_, err := getAccessToken(&currentDeployment)
	if err != nil {
		fmt.Println("Error getting access token:", err)
		return
	}

    printSalesforceCreds(&currentDeployment)
    printObjectCounts(&currentDeployment)

    // Check if both deployments are valid
    bothDeploymentsValid := isValidDeployment(&deployment1) && isValidDeployment(&deployment2)
    if bothDeploymentsValid {
        print("\nYou have two valid Salesforce deployments:\n")
        print("\n  ", deployment1.Url)
        print("\n  ", deployment2.Url)
        print("\n")
    }

	// Main menu
	for {
		fmt.Println("\nMain Menu:\n")
		fmt.Println("1. Search accounts")
		fmt.Println("2. Search contacts")
        fmt.Println("3. Search tasks")
        fmt.Println("4. Change Salesforce deployment")
		fmt.Println("5. Exit\n")

		var option string
		fmt.Print("Enter your option: ")
		fmt.Scanln(&option)

		switch option {
		case "1":
			
			var nameFilter string
			fmt.Print("\nEnter account name filter: ")
			fmt.Scanln(&nameFilter)
		
			accounts, err := getAccountsByName(&currentDeployment, nameFilter)
			if err != nil {
				fmt.Println("Error retrieving accounts:", err)
				return
			}

			printAccounts(accounts)

		case "2":
			
			var contactFilter string
			fmt.Print("\nEnter contact first, last name, email or account name filter: ")
			fmt.Scanln(&contactFilter)
		
			contacts, err := getContacts(&currentDeployment, contactFilter)
			if err != nil {
				fmt.Println("Error retrieving contacts:", err)
				return
			}

			printContacts(contacts)

        case "3":

            var taskFilter string
            fmt.Print("\nEnter task subject, contact first, last name, email or account name filter: ")
            fmt.Scanln(&taskFilter)

            tasks, err := getTasks(&currentDeployment, taskFilter)
            if err != nil {
                fmt.Println("Error retrieving tasks:", err)
                return
            }

            printTasks(tasks)

        case "4":

            currentDeployment = changeDeployment(&deployment1, &deployment2)
            fmt.Println("\nSwitched to deployment:", currentDeployment.Url)
            printObjectCounts(&currentDeployment)
            // Get access token for the new deployment
            _, err := getAccessToken(&currentDeployment)
            if err != nil {
                fmt.Println("Error getting access token:", err)
                return
            }

		case "5":
			fmt.Println("\nExiting...")
			return
		default:
			fmt.Println("\nInvalid option")
		}
	}

}
