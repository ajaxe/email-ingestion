package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
	"github.com/jhillyerd/enmime"
)

func NewEmailProcessor(queries *public.Queries, storageService *storage.S3StorageService) *EmailProcessor {
	return &EmailProcessor{
		queries:        queries,
		storageService: storageService,
	}
}

type EmailProcessor struct {
	queries        *public.Queries
	storageService *storage.S3StorageService
}

func (e *EmailProcessor) Process(ctx context.Context, payload *model.IngestEmailPayload) error {
	spoolID, err := uuid.Parse(payload.SpoolID)
	if err != nil {
		return fmt.Errorf("invalid spool ID: %w", err)
	}

	slog.InfoContext(ctx, "begin processing email", "spool_id", spoolID, "upload_key", payload.UploadKey)

	stream, err := e.storageService.DownloadObject(ctx, payload.UploadKey)
	if err != nil {
		_ = e.queries.UpdateSpooledEmailStatus(ctx, public.UpdateSpooledEmailStatusParams{
			ID:               spoolID,
			Status:           public.SpoolStatusPENDING,
			LastErrorMessage: err.Error(),
		})
		return model.NewRetryableError(fmt.Errorf("failed to download spool object: %w", err))
	}
	defer stream.Close()

	slog.InfoContext(ctx, "parsing email envelope", "spool_id", spoolID)
	env, err := enmime.ReadEnvelope(stream)
	if err != nil {
		_ = e.queries.UpdateSpooledEmailStatus(ctx, public.UpdateSpooledEmailStatusParams{
			ID:               spoolID,
			Status:           public.SpoolStatusFAILED,
			LastErrorMessage: fmt.Sprintf("failed to parse mime: %v", err),
		})
		return nil // terminal error
	}

	toAddress := env.GetHeader("Delivered-To")
	if toAddress == "" {
		toAddress = env.GetHeader("To")
	}

	addr, err := mail.ParseAddress(toAddress)
	if err != nil {
		_ = e.queries.UpdateSpooledEmailStatus(ctx, public.UpdateSpooledEmailStatusParams{
			ID:               spoolID,
			Status:           public.SpoolStatusFAILED,
			LastErrorMessage: "invalid to address",
		})
		return nil // terminal error
	}

	localPart := strings.Split(addr.Address, "@")[0]
	referenceToken := ""
	if idx := strings.Index(localPart, "+"); idx != -1 {
		localPart = localPart[:idx]
		referenceToken = localPart[idx+1:]
	}

	assignedEmail, err := e.queries.GetAssignedEmailByLocalPart(ctx, localPart)
	if err != nil {
		_ = e.queries.UpdateSpooledEmailStatus(ctx, public.UpdateSpooledEmailStatusParams{
			ID:               spoolID,
			Status:           public.SpoolStatusFAILED,
			LastErrorMessage: err.Error(),
		})
		slog.ErrorContext(ctx, "permanent failure: failed to get assigned email by local part", "spool_id", spoolID, "localPart", localPart, "error", err)
		return nil // terminal error
	}

	if assignedEmail.ApplicationID == uuid.Nil {
		_ = e.queries.UpdateSpooledEmailStatus(ctx, public.UpdateSpooledEmailStatusParams{
			ID:               spoolID,
			Status:           public.SpoolStatusFAILED,
			LastErrorMessage: "unassigned email address",
		})
		slog.ErrorContext(ctx, "permanent failure: unassigned email address", "spool_id", spoolID, "localPart", localPart, "error", err)
		return nil // terminal error
	}

	slog.InfoContext(ctx, "found assigned email address", "spool_id", spoolID, "localPart", localPart, "application_id", assignedEmail.ApplicationID)

	msgID := env.GetHeader("Message-Id")
	msgID = strings.Trim(msgID, "<>")
	if msgID == "" {
		slog.InfoContext(ctx, "falling back to spoolID as Message-ID", "spool_id", spoolID)
		msgID = spoolID.String()
	}

	basePath := util.ProcessedEmailS3KeyPrefix(assignedEmail.ApplicationID.String(), msgID)
	slog.InfoContext(ctx, "upload key path for email content json", "spool_id", spoolID, "base_path", basePath, "application_id", assignedEmail.ApplicationID)

	headerMap := make(map[string]string)
	for _, key := range env.GetHeaderKeys() {
		headerMap[key] = env.GetHeader(key)
	}

	contentBody := map[string]interface{}{
		"text":    env.Text,
		"html":    env.HTML,
		"headers": headerMap,
	}

	contentJSON, _ := json.Marshal(contentBody)

	contentKey := fmt.Sprintf("%s/contents.json", basePath)
	_, err = e.storageService.UploadObject(ctx, contentKey, bytes.NewReader(contentJSON), "application/json")
	if err != nil {
		return model.NewRetryableError(fmt.Errorf("failed to upload contents: %w", err))
	}
	atchPrefix := util.ProcessedAttachmentS3KeyPrefix(basePath)
	for _, att := range env.Attachments {
		attKey := fmt.Sprintf("%s/%s", atchPrefix, att.FileName)
		_, err = e.storageService.UploadObject(ctx, attKey, bytes.NewReader(att.Content), att.ContentType)
		if err != nil {
			return model.NewRetryableError(fmt.Errorf("failed to upload attachment: %w", err))
		}
	}
	for _, in := range env.Inlines {
		attKey := fmt.Sprintf("%s/%s", atchPrefix, in.FileName)
		_, err = e.storageService.UploadObject(ctx, attKey, bytes.NewReader(in.Content), in.ContentType)
		if err != nil {
			return model.NewRetryableError(fmt.Errorf("failed to upload inline attachment: %w", err))
		}
	}

	ingestedEmail, err := e.queries.CreateIngestedEmail(ctx, public.CreateIngestedEmailParams{
		ApplicationID:   assignedEmail.ApplicationID,
		AssignedEmailID: assignedEmail.ID,
		ReferenceToken:  referenceToken,
		FromAddress:     env.GetHeader("From"),
		Subject:         env.GetHeader("Subject"),
		MessageID:       msgID,
		S3KeyPrefix:     basePath,
	})
	if err != nil {
		return model.NewRetryableError(fmt.Errorf("failed to create ingested email record: %w", err))
	}

	/*
		// webhook processing to come soon.
		_, err = e.queries.EnqueueWebhookJob(ctx, public.EnqueueWebhookJobParams{
			ApplicationID:   assignedEmail.ApplicationID,
			IngestedEmailID: ingestedEmail.ID,
		})
		if err != nil {
			return model.NewRetryableError(fmt.Errorf("failed to enqueue webhook job: %w", err))
		}*/

	// TODO: publish a job on webhok-stream, another job type to be processed by the worker.

	_ = e.queries.UpdateSpooledEmailStatus(ctx, public.UpdateSpooledEmailStatusParams{
		ID:               spoolID,
		Status:           public.SpoolStatusSUCCESS,
		LastErrorMessage: "",
	})
	_ = e.storageService.DeleteObject(ctx, payload.UploadKey)

	slog.InfoContext(ctx, "email processed successfully, raw email deleted", "spoolID", spoolID, "uploadKey", payload.UploadKey, "applicationID", assignedEmail.ApplicationID, "ingestedEmailID", ingestedEmail.ID)

	return nil
}
