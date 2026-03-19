package main

import (
	"context"
	"log"
	"strings"
)

// Create Template
func CreateTemplate(template *Template) error {
	return DB.QueryRow(
		context.Background(),
		`INSERT INTO templates (user_id, name, description, content)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		template.UserID,
		template.Name,
		template.Description,
		template.Content,
	).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)
}

// Get Template
func GetTemplate(templateID int64) (*Template, error) {
	template := &Template{}

	err := DB.QueryRow(
		context.Background(),
		`SELECT id, user_id, name, description, content, created_at, updated_at
		 FROM templates
		 WHERE id = $1`,
		templateID,
	).Scan(
		&template.ID,
		&template.UserID,
		&template.Name,
		&template.Description,
		&template.Content,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return template, nil
}

// Get All Templates for User
func GetTemplatesByUser(userID int64) ([]*Template, error) {
	rows, err := DB.Query(
		context.Background(),
		`SELECT id, user_id, name, description, content, created_at, updated_at
		 FROM templates
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]*Template, 0)
	for rows.Next() {
		template := &Template{}
		if err := rows.Scan(
			&template.ID,
			&template.UserID,
			&template.Name,
			&template.Description,
			&template.Content,
			&template.CreatedAt,
			&template.UpdatedAt,
		); err != nil {
			return nil, err
		}

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return templates, nil
}

// Update Template
func UpdateTemplate(template *Template) error {
	return DB.QueryRow(
		context.Background(),
		`UPDATE templates
		 SET user_id = $1,
		     name = $2,
		     description = $3,
		     content = $4,
		     updated_at = NOW()
		 WHERE id = $5
		 RETURNING updated_at`,
		template.UserID,
		template.Name,
		template.Description,
		template.Content,
		template.ID,
	).Scan(&template.UpdatedAt)
}

// Delete Template
func DeleteTemplate(templateID int64) error {
	_, err := DB.Exec(
		context.Background(),
		`DELETE FROM templates WHERE id = $1`,
		templateID,
	)
	return err
}

// Create Campaign
func CreateCampaign(campaign *Campaign, userIDs []int64) error {
	// Start Transaction
	tx, err := DB.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Insert Campaign
	err = tx.QueryRow(
		context.Background(),
		`INSERT INTO campaigns (user_id, template_id, name, description, status, recipient_type, priority, total_recipients, scheduled_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		campaign.UserID,
		campaign.TemplateID,
		campaign.Name,
		campaign.Description,
		campaign.Status,
		campaign.RecipientType,
		campaign.Priority,
		len(userIDs),
		campaign.ScheduledAt,
	).Scan(&campaign.ID, &campaign.CreatedAt, &campaign.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert Campaign Recipients
	for _, userID := range userIDs {
		_, err := tx.Exec(
			context.Background(),
			`INSERT INTO campaign_recipients (campaign_id, user_id, status)
			 VALUES ($1, $2, $3)`,
			campaign.ID,
			userID,
			RecipientStatusPending,
		)
		if err != nil {
			return err
		}
	}

	// Commit Transaction
	if err := tx.Commit(context.Background()); err != nil {
		return err
	}

	// Add to the queue

	queueName := ResolveRabbitMQQueue("bulk-filter")

	rabbitMQ, err := GetRabbitMQClient(queueName)
	if err != nil {
		log.Fatal(err)
	}

	bulkFilterJob := BulkFilterJob{
		CampaignID:    campaign.ID,
		TemplateID:    campaign.TemplateID,
		UserId:        campaign.UserID,
		Name:          campaign.Name,
		RecipientType: campaign.RecipientType,
		Priority:      campaign.Priority,
	}

	if err := rabbitMQ.EnqueueMessage(bulkFilterJob); err != nil {
		log.Fatal(err)
	}

	log.Printf("Enqueued campaign %d to %s", campaign.ID, queueName)

	return nil

}

// Get Campaign
func GetCampaign(campaignID int64, userID int64) (*Campaign, error) {
	campaign := &Campaign{}

	err := DB.QueryRow(
		context.Background(),
		`SELECT id, user_id, template_id, name, description, status, recipient_type, priority, total_recipients, created_at, updated_at, started_at, scheduled_at, completed_at, sent_count, failed_count, skipped_count
		 FROM campaigns
		 WHERE id = $1 AND user_id = $2`,
		campaignID,
		userID,
	).Scan(
		&campaign.ID,
		&campaign.UserID,
		&campaign.TemplateID,
		&campaign.Name,
		&campaign.Description,
		&campaign.Status,
		&campaign.RecipientType,
		&campaign.Priority,
		&campaign.TotalRecipients,
		&campaign.CreatedAt,
		&campaign.UpdatedAt,
		&campaign.StartedAt,
		&campaign.ScheduledAt,
		&campaign.CompletedAt,
		&campaign.SentCount,
		&campaign.FailedCount,
		&campaign.SkippedCount,
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

// Get All Campaigns for User
func GetCampaignsByUser(userID int64) ([]*Campaign, error) {
	rows, err := DB.Query(
		context.Background(),
		`SELECT id, user_id, template_id, name, description, status, recipient_type, priority, total_recipients, created_at, updated_at, started_at, scheduled_at, completed_at, sent_count, failed_count, skipped_count
		 FROM campaigns
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	campaigns := make([]*Campaign, 0)
	for rows.Next() {
		campaign := &Campaign{}
		if err := rows.Scan(
			&campaign.ID,
			&campaign.UserID,
			&campaign.TemplateID,
			&campaign.Name,
			&campaign.Description,
			&campaign.Status,
			&campaign.RecipientType,
			&campaign.Priority,
			&campaign.TotalRecipients,
			&campaign.CreatedAt,
			&campaign.UpdatedAt,
			&campaign.StartedAt,
			&campaign.ScheduledAt,
			&campaign.CompletedAt,
			&campaign.SentCount,
			&campaign.FailedCount,
			&campaign.SkippedCount,
		); err != nil {
			return nil, err
		}

		campaigns = append(campaigns, campaign)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return campaigns, nil
}

// Update Campaign
func UpdateCampaign(campaign *Campaign, userIDs []int64) error {
	tx, err := DB.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	err = tx.QueryRow(
		context.Background(),
		`UPDATE campaigns
		 SET template_id = $1,
		     name = $2,
		     description = $3,
		     status = $4,
		     recipient_type = $5,
		     priority = $6,
		     total_recipients = $7,
		     scheduled_at = $8,
		     updated_at = NOW()
		 WHERE id = $9 AND user_id = $10
		 RETURNING updated_at`,
		campaign.TemplateID,
		campaign.Name,
		campaign.Description,
		campaign.Status,
		campaign.RecipientType,
		campaign.Priority,
		len(userIDs),
		campaign.ScheduledAt,
		campaign.ID,
		campaign.UserID,
	).Scan(&campaign.UpdatedAt)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(
		context.Background(),
		`DELETE FROM campaign_recipients WHERE campaign_id = $1`,
		campaign.ID,
	); err != nil {
		return err
	}

	for _, userID := range userIDs {
		if _, err := tx.Exec(
			context.Background(),
			`INSERT INTO campaign_recipients (campaign_id, user_id, status)
			 VALUES ($1, $2, $3)`,
			campaign.ID,
			userID,
			RecipientStatusPending,
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return err
	}

	return nil
}

// Delete Campaign
func DeleteCampaign(campaignID int64, userID int64) error {
	_, err := DB.Exec(
		context.Background(),
		`DELETE FROM campaigns WHERE id = $1 AND user_id = $2`,
		campaignID,
		userID,
	)
	return err
}

// Create User
func CreateUser(user *User) error {
	return DB.QueryRow(
		context.Background(),
		`INSERT INTO users (email, name, role)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		user.Email,
		user.Name,
		user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// Get User
func GetUser(userID int64) (*User, error) {
	user := &User{}

	err := DB.QueryRow(
		context.Background(),
		`SELECT id, email, name, role, created_at, updated_at
		 FROM users
		 WHERE id = $1`,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Get All Users
func GetAllUsers() ([]*User, error) {
	rows, err := DB.Query(
		context.Background(),
		`SELECT id, email, name, role, created_at, updated_at
		 FROM users
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func GetAllAdmins() ([]*User, error) {
	rows, err := DB.Query(
		context.Background(),
		`SELECT id, email, name, role, created_at, updated_at
		 FROM users
		 WHERE role = $1
		 ORDER BY created_at DESC`,
		UserRoleAdmin,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	admins := make([]*User, 0)
	for rows.Next() {
		admin := &User{}
		if err := rows.Scan(
			&admin.ID,
			&admin.Email,
			&admin.Name,
			&admin.Role,
			&admin.CreatedAt,
			&admin.UpdatedAt,
		); err != nil {
			return nil, err
		}

		admins = append(admins, admin)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return admins, nil
}

func GetAllRegularUsers() ([]*User, error) {
	rows, err := DB.Query(
		context.Background(),
		`SELECT id, email, name, role, created_at, updated_at
		 FROM users
		 WHERE role = $1
		 ORDER BY created_at DESC`,
		UserRoleUser,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Update User
func UpdateUser(user *User) error {
	return DB.QueryRow(
		context.Background(),
		`UPDATE users
		 SET email = $1,
		     name = $2,
		     role = $3,
		     updated_at = NOW()
		 WHERE id = $4
		 RETURNING updated_at`,
		user.Email,
		user.Name,
		user.Role,
		user.ID,
	).Scan(&user.UpdatedAt)
}

// Delete User
func DeleteUser(userID int64) error {
	_, err := DB.Exec(
		context.Background(),
		`DELETE FROM users WHERE id = $1`,
		userID,
	)
	return err
}

// Only Admins can manage templates and campaigns, so we can add a simple check here
func IsAdmin(userID int64) (bool, error) {
	var role UserRole
	err := DB.QueryRow(
		context.Background(),
		`SELECT role FROM users WHERE id = $1`,
		userID,
	).Scan(&role)
	if err != nil {
		return false, err
	}

	return role == UserRoleAdmin, nil
}

func Login(email string) (*User, error) {
	user := &User{}

	err := DB.QueryRow(
		context.Background(),
		`SELECT id, email, name, role, created_at, updated_at
		 FROM users
		 WHERE email = $1`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func Logout(userID int64) error {
	// For simplicity, we are not implementing actual session management here
	return nil
}

func GetCampaignRecipients(campaignID int64) ([]*CampaignRecipient, error) {
	rows, err := DB.Query(
		context.Background(),
		`SELECT id, campaign_id, user_id, status, provider_message_id, error_message, retry_count, scheduled_at, sent_at, created_at, updated_at
		 FROM campaign_recipients
		 WHERE campaign_id = $1
		 ORDER BY id ASC`,
		campaignID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipients := make([]*CampaignRecipient, 0)
	for rows.Next() {
		recipient := &CampaignRecipient{}
		if err := rows.Scan(
			&recipient.ID,
			&recipient.CampaignID,
			&recipient.UserID,
			&recipient.Status,
			&recipient.ProviderMessageID,
			&recipient.ErrorMessage,
			&recipient.RetryCount,
			&recipient.ScheduledAt,
			&recipient.SentAt,
			&recipient.CreatedAt,
			&recipient.UpdatedAt,
		); err != nil {
			return nil, err
		}

		recipients = append(recipients, recipient)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return recipients, nil
}

func UpdateCampaignStatus(campaignID int64, status CampaignStatus) error {
	_, err := DB.Exec(
		context.Background(),
		`UPDATE campaigns
		 SET status = $1,
		     updated_at = NOW(),
		     started_at = CASE
		       WHEN $1 = 'running' AND started_at IS NULL THEN NOW()
		       ELSE started_at
		     END,
		     completed_at = CASE
		       WHEN $1 IN ('completed', 'failed', 'cancelled') THEN NOW()
		       ELSE completed_at
		     END
		 WHERE id = $2`,
		status,
		campaignID,
	)
	return err
}

func UpdateCampaignRecipientStatus(recipientID int64, status RecipientStatus) error {
	_, err := DB.Exec(
		context.Background(),
		`UPDATE campaign_recipients
		 SET status = $1,
		     updated_at = NOW()
		 WHERE id = $2`,
		status,
		recipientID,
	)
	return err
}

func MarkCampaignRecipientSent(recipientID int64, providerMessageID string) error {
	_, err := DB.Exec(
		context.Background(),
		`UPDATE campaign_recipients
		 SET status = $1,
		     provider_message_id = $2,
		     error_message = NULL,
		     sent_at = NOW(),
		     updated_at = NOW()
		 WHERE id = $3`,
		RecipientStatusSent,
		providerMessageID,
		recipientID,
	)
	return err
}

func MarkCampaignRecipientFailed(recipientID int64, errorMessage string) error {
	_, err := DB.Exec(
		context.Background(),
		`UPDATE campaign_recipients
		 SET status = $1,
		     error_message = $2,
		     retry_count = retry_count + 1,
		     updated_at = NOW()
		 WHERE id = $3`,
		RecipientStatusFailed,
		errorMessage,
		recipientID,
	)
	return err
}

func SyncCampaignDeliveryStats(campaignID int64) error {
	_, err := DB.Exec(
		context.Background(),
		`WITH stats AS (
			SELECT
				campaign_id,
				COUNT(*) FILTER (WHERE status = 'sent') AS sent_count,
				COUNT(*) FILTER (WHERE status = 'failed') AS failed_count,
				COUNT(*) FILTER (WHERE status = 'skipped') AS skipped_count,
				COUNT(*) FILTER (WHERE status IN ('sent', 'failed', 'skipped')) AS finalized_count
			FROM campaign_recipients
			WHERE campaign_id = $1
			GROUP BY campaign_id
		)
		UPDATE campaigns AS c
		SET sent_count = COALESCE(stats.sent_count, 0),
		    failed_count = COALESCE(stats.failed_count, 0),
		    skipped_count = COALESCE(stats.skipped_count, 0),
		    status = CASE
		    	WHEN COALESCE(stats.finalized_count, 0) < c.total_recipients THEN 'running'
		    	WHEN COALESCE(stats.sent_count, 0) > 0 THEN 'completed'
		    	WHEN COALESCE(stats.failed_count, 0) > 0 THEN 'failed'
		    	ELSE c.status
		    END,
		    updated_at = NOW(),
		    started_at = CASE
		    	WHEN COALESCE(stats.finalized_count, 0) < c.total_recipients AND c.started_at IS NULL THEN NOW()
		    	ELSE c.started_at
		    END,
		    completed_at = CASE
		    	WHEN COALESCE(stats.finalized_count, 0) >= c.total_recipients THEN NOW()
		    	ELSE c.completed_at
		    END
		FROM stats
		WHERE c.id = stats.campaign_id`,
		campaignID,
	)
	return err
}

func deriveUserName(user *User) string {
	if user == nil {
		return ""
	}

	localPart, _, found := strings.Cut(user.Email, "@")
	if !found {
		return user.Email
	}

	return localPart
}
