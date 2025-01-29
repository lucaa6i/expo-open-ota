package main

import (
	"expo-open-ota/internal/services"
	"fmt"
	"github.com/manifoldco/promptui"
	"log"
)

func main() {
	auth := retrieveExpoAuth()
	apps, err := services.GetExpoApplications(*auth)
	if err != nil {
		log.Fatalf("Error fetching expo applications: %v", err)
		return
	}

	if len(apps) == 0 {
		log.Println("No Expo applications found.")
		return
	}

	items := make([]string, len(apps))
	for i, app := range apps {
		items[i] = fmt.Sprintf("%s (ID: %s)", app.Name, app.Id)
	}

	prompt := promptui.Select{
		Label: "Select an Expo application",
		Items: items,
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
		return
	}

	var selectedAppId string
	for _, app := range apps {
		if result == fmt.Sprintf("%s (ID: %s)", app.Name, app.Id) {
			selectedAppId = app.Id
			break
		}
	}
	fmt.Println("Please copy the following line to your .env or add it to your environment variables:")
	fmt.Printf("EXPO_APP_ID=%s\n", selectedAppId)
}
