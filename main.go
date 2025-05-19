package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"main.go/dto"
)

// Global connection pool
// var dbPool *pgxpool.Pool

// func initDB() {

// 	var err error
// 	err = godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}
// 	connString := os.Getenv("DB_CREDS")
// 	if connString == "" {
// 		log.Fatal("DB_CREDS not set in .env file")
// 	}
// 	dbPool, err = pgxpool.New(context.Background(), connString)
// 	if err != nil {
// 		log.Fatalf("Unable to create connection pool: %v", err)
// 	}

// 	fmt.Println("Connected to PostgreSQL with connection pool!")
// }

// func userExists(chatID int64, userID int) (bool, error) {
// 	query := `SELECT COUNT(*) FROM fpt_user.user WHERE chat_id = $1 AND user_id = $2`
// 	var count int

// 	err := dbPool.QueryRow(context.Background(), query, chatID, userID).Scan(&count)
// 	if err != nil {
// 		return false, err
// 	}
// 	return count > 0, nil
// }

// func insertUserIfNotExists(chatID int64, userID int, firstName, lastName, languageCode string) error {
// 	exists, err := userExists(chatID, userID)
// 	if err != nil {
// 		return err
// 	}

// 	if exists {
// 		fmt.Println("User already exists, skipping insert.")
// 		return nil
// 	}

// 	query := `INSERT INTO fpt_user.user (chat_id, user_id, first_name, last_name, lang_code)
// 	          VALUES ($1, $2, $3, $4, $5)`

// 	_, err = dbPool.Exec(context.Background(), query, chatID, userID, firstName, lastName, languageCode)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("New user inserted successfully!")
// 	return nil
// }

func main() {

	//Load the .env file
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	go func() {
		//HTTP server so that port doesn't remain open on render
		http.HandleFunc("/", aboutHandler)
		http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Pong")
		})

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		fmt.Println("Web server listening on port:", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}()

	// Get the bot token from the environment
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not set in .env file")
	}

	// Initialize the bot
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Set debug mode (optional)
	bot.Debug = false
	fmt.Println("Bot is running...")

	// Set up updates using long polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	var userState = make(map[int]string)
	var userSelections = make(map[int]map[string]string)

	// initDB()             // Initialize connection pool
	// defer dbPool.Close() // Close pool when the program exits

	// Listen for updates
	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		userID := update.Message.From.ID

		// firstName := update.Message.From.FirstName
		// lastName := update.Message.From.LastName
		// languageCode := update.Message.From.LanguageCode

		// err := insertUserIfNotExists(chatID, userID, firstName, lastName, languageCode)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		if _, exists := userSelections[userID]; !exists {
			userSelections[userID] = make(map[string]string)
		}

		switch userState[userID] {
		case "waiting_for_from":
			userSelections[userID]["from"] = update.Message.Text
			msg := tgbotapi.NewMessage(chatID, "Please enter the 'To' location:")
			bot.Send(msg)
			userState[userID] = "waiting_for_to"
		case "waiting_for_to":
			userSelections[userID]["to"] = update.Message.Text
			msg := tgbotapi.NewMessage(chatID, "Please enter the departure date (YYYY-MM-DD):")
			bot.Send(msg)
			userState[userID] = "waiting_for_date"
		case "waiting_for_date":
			userSelections[userID]["date"] = update.Message.Text
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Your trip details:\nFrom: %s\nTo: %s\nDate: %s\nAre you sure? (Yes/No)", userSelections[userID]["from"], userSelections[userID]["to"], userSelections[userID]["date"]))
			bot.Send(msg)
			userState[userID] = "waiting_for_confirmation"
		case "waiting_for_confirmation":
			if strings.EqualFold(update.Message.Text, "Yes") {
				msg := tgbotapi.NewMessage(chatID, "Fetching flights....")
				bot.Send(msg)
				fmt.Printf("User confirmed:\nFrom: %s\nTo: %s\nDate: %s\n", userSelections[userID]["from"], userSelections[userID]["to"], userSelections[userID]["date"])
				fromLocation := geminiApiToFetchIATA(userSelections[userID]["from"])
				toLocation := geminiApiToFetchIATA(userSelections[userID]["to"])
				message, err := getLowestPriceItinerary(fromLocation, toLocation, userSelections[userID]["date"])
				if err != nil {
					log.Fatal(err)
				}
				msg = tgbotapi.NewMessage(chatID, message)

				fromLocationLower := strings.ToLower(fromLocation)
				toLocationLower := strings.ToLower(toLocation)
				dateForUrl := dateFormatChangeForURL(userSelections[userID]["date"])
				// CTA to skyscanner site
				skyscannerURL := fmt.Sprintf("https://www.skyscanner.co.in/transport/flights/%s/%s/%s/", fromLocationLower, toLocationLower, dateForUrl)
				button := tgbotapi.NewInlineKeyboardButtonURL("🔗 View on Skyscanner", skyscannerURL)
				row := tgbotapi.NewInlineKeyboardRow(button)
				keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
				// 🔁 Prompt user to /start again after 30 seconds
				go func() {
					time.Sleep(30 * time.Second)
					reminderMsg := tgbotapi.NewMessage(chatID, "Want to search again? Type /start ✈️")
					bot.Send(reminderMsg)
				}()

				// Clean up user data
				delete(userState, userID)
				delete(userSelections, userID)
			} else {
				msg := tgbotapi.NewMessage(chatID, "Trip details discarded. Start again with /start")
				bot.Send(msg)
				delete(userState, userID)
				delete(userSelections, userID)
			}
			userState[userID] = ""
		default:
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(chatID, "Welcome! Please enter the 'From' location:")
				bot.Send(msg)
				userState[userID] = "waiting_for_from"
			} else {
				msg := tgbotapi.NewMessage(chatID, "Please use the provided options to select locations and date.")
				bot.Send(msg)
			}
		}
	}
}

// func createKeyboard(options []string) tgbotapi.ReplyKeyboardMarkup {
// 	var keyboardRows [][]tgbotapi.KeyboardButton
// 	for _, option := range options {
// 		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(option)))
// 	}
// 	return tgbotapi.ReplyKeyboardMarkup{
// 		Keyboard:        keyboardRows,
// 		ResizeKeyboard:  true,
// 		OneTimeKeyboard: true,
// 	}
// }

func fetchFlightDataSSApi(from string, to string, departDate string) (*dto.SSFlightResponse, error) {
	baseURL := "https://sky-scanner3.p.rapidapi.com/flights/search-one-way"
	params := url.Values{}
	params.Add("fromEntityId", from)
	params.Add("toEntityId", to)
	params.Add("departDate", departDate)
	params.Add("market", "IN")
	params.Add("currency", "INR")
	params.Add("stops", "direct")
	// params.Add("includeOriginNearbyAirports", false)
	params.Add("sort", "cheapest_first")

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// err = godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	// Get the bot token from the environment
	skyScannerKey := os.Getenv("SKY_SCANNER_KEY")

	req.Header.Add("x-rapidapi-host", "sky-scanner3.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", skyScannerKey) // Replace with your actual API key

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var flightData dto.SSFlightResponse
	err = json.Unmarshal(body, &flightData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &flightData, nil
}

func getLowestPriceItinerary(from string, to string, departDate string) (string, error) {
	var message string
	flightData, err := fetchFlightDataSSApi(from, to, departDate)
	if err != nil {
		return "", err
	}

	if flightData != nil && flightData.Status && len(flightData.Data.Itineraries) > 0 {
		lowestPriceItinerary := &flightData.Data.Itineraries[0]

		for i := 1; i < len(flightData.Data.Itineraries); i++ {
			if flightData.Data.Itineraries[i].Price.Raw < lowestPriceItinerary.Price.Raw {
				lowestPriceItinerary = &flightData.Data.Itineraries[i]
			}
		}
		if lowestPriceItinerary != nil {
			leg := lowestPriceItinerary.Legs[0] // Assuming single leg
			carrier := leg.Carriers.Marketing[0]
			segment := leg.Segments[0]

			// departureTime, _ := time.Parse(time.RFC3339, segment.Departure)
			// arrivalTime, _ := time.Parse(time.RFC3339, segment.Arrival)

			message = fmt.Sprintf(
				"Cheapest Flight found on skyscanner: %s, Operated by %s, Flight Number: %s, Departs: %s, Arrives: %s, Duration: %d minutes.",
				lowestPriceItinerary.Price.Formatted,
				carrier.Name,
				segment.FlightNumber,
				segment.Departure,
				segment.Arrival,
				leg.DurationInMinutes,
			)
			fmt.Println(message)
		}
		return message, nil
	} else if flightData != nil {
		return "", fmt.Errorf("API request unsuccessful. Message: %s", flightData.Message)
	} else {
		return "", fmt.Errorf("Flight data is nil.")
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	aboutHTML := `
        <!DOCTYPE html>
        <html>
        <head>
                <title>About FlightPriceTracker</title>
                <style>
                        body {
                                background-color: #121212; /* Dark background */
                                color: #e0e0e0; /* Light text */
                                font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"; /* Apple-style font */
                                display: flex;
                                justify-content: center;
                                align-items: center;
                                height: 100vh;
                                margin: 0;
                        }
                        .container {
                                text-align: center;
                                max-width: 600px;
                                padding: 20px;
                                border-radius: 8px;
                                background-color: #1e1e1e; /* Darker container background */
                                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.5);
                        }
                        h1 {
                                color:rgb(6, 56, 97); /* Highlight color */
                                margin-bottom: 20px;
                        }
                        ul {
                                list-style-type: none;
                                padding: 0;
                                margin-bottom: 20px;
                        }
                        li {
                                margin: 5px 0;
                        }
                        a {
                                color: #64b5f6;
                                text-decoration: none;
                        }
                        a:hover {
                                text-decoration: underline;
                        }
                </style>
        </head>
        <body>
                <div class="container">
                        <h1>About FlightPriceTracker Bot</h1>
                        <p>This bot helps you know the cheapest flight details and send you alerts when price goes down.</p>
                        <ul>
                                <li>Allows users to enter flight details.</li>
                                <li>Provides confirmation of entered details.</li>
                        </ul>
                        <p>Contact: <a href="mailto:veerujadhav879@gmail.com">veerujadhav879@gmail.com</a></p>
                </div>
        </body>
        </html>
        `

	fmt.Fprint(w, aboutHTML)
}

const apiKey = "AIzaSyBhvNjYFwC0vVBauRDOdjVgeNroe_mahZ8"

type Response struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func geminiApiToFetchIATA(userInput string) string {

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + apiKey

	// JSON payload
	requestBody := fmt.Sprintf(`{
			"contents": [{
				"parts":[{"text": "what is iata code for %s airport. just give the iata code"}]
			}]
		}`, userInput)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("GEMINI_API_KEY", apiKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return ""
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return ""
	}

	// Parse JSON
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return ""
	}

	var iataCode string
	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		iataCode = strings.TrimSpace(response.Candidates[0].Content.Parts[0].Text)
		fmt.Println("IATA Code:", iataCode)
	} else {
		fmt.Println("No IATA code found")
	}

	return iataCode
}

func dateFormatChangeForURL(dateStr string) string {

	// Parse the input string into time.Time
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic(err)
	}

	// Format as yyMMdd
	formatted := t.Format("060102")
	return formatted
}
