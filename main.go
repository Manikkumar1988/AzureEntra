package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/joho/godotenv"
	auth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	tenantID := os.Getenv("AZURE_TENANT_ID")

	if clientID == "" || clientSecret == "" || tenantID == "" {
		log.Fatal("AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, and AZURE_TENANT_ID must be set in .env")
	}

	// Create a new client credentials credential
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		log.Fatalf("Error creating client secret credential: %v", err)
	}

	log.Printf("cred: %v", cred)
	// Create a new Azure authentication provider
	authProvider, err := auth.NewAzureIdentityAuthenticationProvider(cred)
	if err != nil {
		log.Fatalf("Error creating authentication provider: %v", err)
	}
	log.Printf("authProvider: %v", authProvider)
	// Create a new Graph client
	requestAdapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
	if err != nil {
		log.Fatalf("Error creating request adapter: %v", err)
	}

	// Create a new Graph client from the request adapter
	graphClient := msgraphsdk.NewGraphServiceClient(requestAdapter)
	// Define the new user details
	newUser := graphmodels.NewUser()
	displayName := "Go App User"
	mailNickname := "goappuser"                                            // Must be unique within your tenant
	userPrincipalName := mailNickname + "@" + os.Getenv("AZURE_AD_DOMAIN") // Replace with your verified domain
	password := "SecurePa$$word123!"                                       // Consider using a more secure way to generate/handle passwords

	newUser.SetAccountEnabled(new(bool))
	*newUser.GetAccountEnabled() = true
	newUser.SetDisplayName(&displayName)
	newUser.SetMailNickname(&mailNickname)
	newUser.SetUserPrincipalName(&userPrincipalName)

	passwordProfile := graphmodels.NewPasswordProfile()
	passwordProfile.SetForceChangePasswordNextSignIn(new(bool))
	*passwordProfile.GetForceChangePasswordNextSignIn() = true // Force user to change password on first login
	passwordProfile.SetPassword(&password)
	newUser.SetPasswordProfile(passwordProfile)

	log.Printf("Attempting to create user: %s (UPN: %s)", *newUser.GetDisplayName(), *newUser.GetUserPrincipalName())

	// Create the user in Azure AD
	createdUser, err := graphClient.Users().Post(context.Background(), newUser, nil)
	if err != nil {
		log.Fatalf("Error creating user: %v", err)
	}

	fmt.Printf("Successfully created user: %s (ID: %s, UPN: %s)\n",
		*createdUser.GetDisplayName(), *createdUser.GetId(), *createdUser.GetUserPrincipalName())
}
