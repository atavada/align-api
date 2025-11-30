package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/atavada/project-management-saas/internal/models"
	"github.com/atavada/project-management-saas/internal/repository"
	"github.com/gofiber/fiber/v3"
	svix "github.com/svix/svix-webhooks/go"
)

type WebhookHandler struct {
	userRepo 		*repository.UserRepository
	orgRepo  		*repository.OrganizationRepository
	memberRepo 	*repository.OrganizationMemberRepository
	webhookSecret string
}

func NewWebhookHandler(
	userRepo *repository.UserRepository,
	orgRepo *repository.OrganizationRepository,
	memberRepo *repository.OrganizationMemberRepository,
	webhookSecret string,
) *WebhookHandler {
	return &WebhookHandler{
		userRepo:      userRepo,
		orgRepo:       orgRepo,
		memberRepo:    memberRepo,
		webhookSecret: webhookSecret,
	}
}

// HandlerClerkWebhook processes incoming Clerk webhooks
func (h *WebhookHandler) HandlerClerkWebhook(c fiber.Ctx) error {
	ctx := context.Background()

	// Get Svix headers for verif
	svixID := c.Get("svix-Id")
	svixTimeStamp := c.Get("svix-timestamp")
	svixSignature := c.Get("svix-signature")

	if svixID == "" || svixTimeStamp == "" || svixSignature == "" {
		log.Println("Missing svix headers")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing webhook headers",
		})
	}

	// Verify webhook signature
	webhooks, err := svix.NewWebhook(h.webhookSecret)
	if err != nil {
		log.Printf("Error creating webhook verifier: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Webhook verifier error",
		})
	}

	payload := c.Body()
	headers := http.Header{}
	headers.Set("svix-id", svixID)
	headers.Set("svix-timestamp", svixTimeStamp)
	headers.Set("svix-signature", svixSignature)

	var event map[string]interface{}
	err = webhooks.Verify(payload, headers)
	if err != nil {
		log.Printf("Webhook verification failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid webhook signature",
		})
	}

	// Parse the event
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Printf("Error parsing webhook payload: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	eventType, ok := event["type"].(string)
	if !ok {
		log.Println("Missing event type")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing event type",
		})
	}

	data := event["data"]

	// Route to appropriate handler
	log.Printf("Processing webhook event: %s", eventType)

	switch eventType {
	case "user.created", "user.updated":
			return h.handleUserEvent(ctx, c, data)
	case "organization.created", "organization.updated":
			return h.handleOrganizationEvent(ctx, c, data)
	case "organizationMembership.created":
			return h.handleMembershipCreated(ctx, c, data)
	case "organizationMembership.deleted":
			return h.handleMembershipDeleted(ctx, c, data)
	default:
			log.Printf("Unhandled webhook event type: %s", eventType)
			return c.SendStatus(fiber.StatusOK)
	}
}

func (h *WebhookHandler) handleUserEvent(ctx context.Context, c fiber.Ctx, data interface{}) error {
	userData, ok := data.(map[string]interface{})
	if !ok {
		log.Println("Invalid user data format")
		return c.SendStatus(fiber.StatusOK)
	}

	// Extract user info
	clerkUserID, _ := userData["id"].(string)

	// Get primary email
	var email string
	if emailAddresses, ok := userData["email_addresses"].([]interface{}); ok && len(emailAddresses) > 0 {
		if emailData, ok := emailAddresses[0].(map[string]interface{}); ok {
			email, _ = emailData["email_address"].(string)
		}
	}

	firstName, _ := userData["first_name"].(string)
	lastName, _ := userData["last_name"].(string)
	imageURL, _ := userData["image_url"].(string)

	if clerkUserID == "" || email == "" {
		log.Println("Missing required user fields")
		return c.SendStatus(fiber.StatusOK)
	}

	// Upsert user to DB
	user := &models.CreateUserRequest{
		ClerkUserID: clerkUserID,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		AvatarURL:   imageURL,
	}

	_, err := h.userRepo.Upsert(ctx, user)
	if err != nil {
		log.Printf("Error upserting user: %v", err)
		return c.SendStatus(fiber.StatusOK) // return 200 to avoid retries
	}

	log.Printf("User synced: %s (%s)", email, clerkUserID)
	return c.SendStatus(fiber.StatusOK)
}

func (h *WebhookHandler) handleOrganizationEvent(ctx context.Context, c fiber.Ctx, data interface{}) error {
	orgData, ok := data.(map[string]interface{})
	if !ok {
		log.Println("Invalid organization data format")
		return c.SendStatus(fiber.StatusOK)
	}

	clerkOrgID, _ := orgData["id"].(string)
	name, _ := orgData["name"].(string)
	slug, _ := orgData["slug"].(string)
	imageURL, _ := orgData["image_url"].(string)
	createdBy, _ := orgData["created_by"].(string)

	if clerkOrgID == "" || name == "" || slug == "" {
		log.Println("Missing required organization fields")
		return c.SendStatus(fiber.StatusOK)
	}

	// Upsert organization
	org := &models.CreateOrganizationRequest{
		ClerkOrgID: clerkOrgID,
		Name:       name,
		Slug:       slug,
		Description: "",
		LogoURL: imageURL,
	}

	createdOrg, err := h.orgRepo.Upsert(ctx, org)
	if err != nil {
		log.Printf("Error upserting organization: %v", err)
		return c.SendStatus(fiber.StatusOK)
	}

	// For new organizations, add creator as owner
	if createdBy != "" {
		creator, err := h.userRepo.GetByClerkID(ctx, createdBy)
		if err == nil && creator != nil {
			member := &models.OrganizationMember{
				OrganizationID: createdOrg.ID,
				UserID: creator.ID,
				Role: models.RoleOwner,
			}

			if err := h.memberRepo.Create(ctx, member); err != nil {
				log.Printf("Error creating owner membership: %v", err)
			} else {
				log.Printf("Owner memberhsip created for org: %s", name)
			}
		}
	}

	log.Printf("Organization synced: %s (%s)", name, clerkOrgID)
	return c.SendStatus(fiber.StatusOK)
}

func (h *WebhookHandler) handleMembershipCreated(ctx context.Context, c fiber.Ctx, data interface{}) error {
	membershipData, ok := data.(map[string]interface{})
	if !ok {
		log.Println("Invalid membership data format")
		return c.SendStatus(fiber.StatusOK)
	}

	clerkMembershipID, _ := membershipData["id"].(string)

	// Get organization and user info
	var clerkOrgID, clerkUserID string
	if org, ok := membershipData["organization"].(map[string]interface{}); ok {
		clerkOrgID, _ = org["id"].(string)
	}
	if publicUserData, ok := membershipData["public_user_data"].(map[string]interface{}); ok {
		clerkUserID, _ = publicUserData["user_id"].(string)
	}

	role, _ := membershipData["role"].(string)

	if clerkOrgID == "" || clerkUserID == "" {
		log.Println("Missing required membership fields")
		return c.SendStatus(fiber.StatusOK)
	}

	// Get local org and user IDs
	org, err := h.orgRepo.GetByClerkID(ctx, clerkOrgID)
	if err != nil || org == nil {
		log.Printf("Organization not found: %s", clerkOrgID)
		return c.SendStatus(fiber.StatusOK)
	}

	user, err := h.userRepo.GetByClerkID(ctx, clerkUserID)
	if err != nil || user == nil {
		log.Printf("User not found: %s", clerkUserID)
		return c.SendStatus(fiber.StatusOK)
	}

	// Map clerk role to our role
	var memberRole models.OrganizationRole
	switch role {
	case "admin":
		memberRole = models.RoleAdmin
	case "basic_member":
		memberRole = models.RoleMember
	default:
		memberRole = models.RoleMember
	}

	// Create membership
	member := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           memberRole,
		ClerkMembershipID: clerkMembershipID,
	}

	if err := h.memberRepo.Create(ctx, member); err != nil {
		log.Printf("Error creating membership: %v", err)
		return c.SendStatus(fiber.StatusOK)
	}

	log.Printf("Membership created: %s in %s", user.Email, org.Name)
	return c.SendStatus(fiber.StatusOK)
}

func (h *WebhookHandler) handleMembershipDeleted(ctx context.Context, c fiber.Ctx, data interface{}) error {
	membershipData, ok := data.(map[string]interface{})
	if !ok {
		log.Println("Invalid membership data format")
		return c.SendStatus(fiber.StatusOK)
	}

	var clerkOrgID, clerkUserID string
	if org, ok := membershipData["organization"].(map[string]interface{}); ok {
		clerkOrgID, _ = org["id"].(string)
	}
	if publicUserData, ok := membershipData["public_user_data"].(map[string]interface{}); ok {
		clerkUserID, _ = publicUserData["user_id"].(string)
	}
	
	if clerkOrgID == "" || clerkUserID == "" {
		log.Println("Missing required membership fields for deletion")
		return c.SendStatus(fiber.StatusOK)
	}

	// Get local org and user IDs
	org, err := h.orgRepo.GetByClerkID(ctx, clerkOrgID)
	if err != nil || org == nil {
		log.Printf("Organization not found: %s", clerkOrgID)
		return c.SendStatus(fiber.StatusOK)
	}

	user, err := h.userRepo.GetByClerkID(ctx, clerkUserID)
	if err != nil || user == nil {
		log.Printf("User not found: %s", clerkUserID)
		return c.SendStatus(fiber.StatusOK)
	}

	// Delete membership
	if err := h.memberRepo.Delete(ctx, org.ID, user.ID); err != nil {
		log.Printf("Error deleting membership: %v", err)
		return c.SendStatus(fiber.StatusOK)
	}
	
	log.Printf("Membership deleted: %s from %s", user.Email, org.Name)
	return c.SendStatus(fiber.StatusOK)
}

