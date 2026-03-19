package main

import (
	"fmt"
	"log"
	"time"

	faker "github.com/go-faker/faker/v4"
)

func controlerService() {
	// Insert 10 users with random data
	for i := 0; i < 10; i++ {
		user := &User{
			Email: faker.Email(),
			Name:  faker.Name(),
			Role:  UserRoleUser, // Default role for seeded users
		}
		fmt.Print(user)
		err := CreateUser(user)
		if err != nil {
			panic(err)
		}
	}

	// Insert 1 Admin user with random data
	adminUser := &User{
		Email: faker.Email(),
		Name:  faker.Name(),
		Role:  UserRoleAdmin,
	}
	err := CreateUser(adminUser)
	if err != nil {
		panic(err)
	}

	log.Println("Database seeded with users")

	// Create a template
	template := &Template{
		UserID:      adminUser.ID,
		Name:        "Welcome Template",
		Description: "A template for welcoming new users",
		Content:     "Hello {{name}}, welcome to our service!",
	}
	err = CreateTemplate(template)
	if err != nil {
		panic(err)
	}

	log.Println("Database seeded with template")

	// List all templates for the admin user
	templates, err := GetTemplatesByUser(adminUser.ID)
	if err != nil {
		panic(err)
	}

	log.Printf("Templates for user %d: %+v\n", adminUser.ID, templates)

	users, _ := GetAllRegularUsers()

	// Create a campaign targeting all users
	campaign := &Campaign{
		UserID:        adminUser.ID,
		TemplateID:    template.ID,
		Name:          "Welcome Campaign",
		Description:   "A campaign to welcome all users",
		Status:        CampaignStatusDraft,
		RecipientType: CampaignRecipientTypeAllUsers,
		Priority:      CampaignPriorityP1,
	}
	userIds := make([]int64, len(users))
	for i, user := range users {
		userIds[i] = user.ID
	}
	err = CreateCampaign(campaign, userIds)
	if err != nil {
		panic(err)
	}
}

func controlerServiceLoop() {
	// Keep it running in while loop with 0.5 second sleep
	for {
		controlerService()
		time.Sleep(10 * time.Millisecond)
	}
}
