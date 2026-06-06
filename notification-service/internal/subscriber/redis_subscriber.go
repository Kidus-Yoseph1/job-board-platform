package subscriber

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	db "github.com/kidus-yoseph1/job-board-platform/notification-service/db/generated"
	jobGrpc "github.com/kidus-yoseph1/job-board-platform/notification-service/internal/grpc"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/repository"
	ws "github.com/kidus-yoseph1/job-board-platform/notification-service/internal/websocket"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/logger"
)

// EventPayload contains the textual information and metadata of the event.
type EventPayload struct {
	TargetUserID string `json:"target_user_id"`
	Title        string `json:"title"`
	Message      string `json:"message"`
}

// RedisEvent represents the JSON wrapper envelope for messages published over Redis Pub/Sub.
type RedisEvent struct {
	EventType string       `json:"event_type"`
	Payload   EventPayload `json:"payload"`
}

// RedisSubscriber manages the persistent Redis Pub/Sub subscription channel.
type RedisSubscriber struct {
	redisClient      *redis.Client
	hub              *ws.Hub
	notificationRepo *repository.NotificationRepo
	emailJobRepo     *repository.EmailJobRepo
	grpcClient       *jobGrpc.JobBoardClient
	log              *logger.Logger
}

// NewRedisSubscriber creates a new RedisSubscriber instance.
func NewRedisSubscriber(
	redisClient *redis.Client,
	hub *ws.Hub,
	repo *repository.NotificationRepo,
	emailRepo *repository.EmailJobRepo,
	grpcClient *jobGrpc.JobBoardClient,
) *RedisSubscriber {
	return &RedisSubscriber{
		redisClient:      redisClient,
		hub:              hub,
		notificationRepo: repo,
		emailJobRepo:     emailRepo,
		grpcClient:       grpcClient,
		log:              logger.Get(),
	}
}

// Start opens the subscription to Redis and processes messages in a blocking loop.
func (s *RedisSubscriber) Start(ctx context.Context) {
	pubsub := s.redisClient.Subscribe(ctx, "job_events")
	defer pubsub.Close()

	s.log.Infow("redis subscriber successfully listening on channel", "channel", "job_events")

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			s.log.Infow("stopping redis subscriber...")
			return
		case msg, ok := <-ch:
			if !ok {
				s.log.Warnw("redis pubsub channel was closed unexpectedly")
				return
			}
			s.log.Infow("received event from Redis", "channel", msg.Channel)
			s.handleEvent(ctx, msg.Payload)
		}
	}
}

// handleEvent processes the raw JSON string received from Redis, logs it to PostgreSQL,
// and routes it to any active WebSocket connection.
func (s *RedisSubscriber) handleEvent(ctx context.Context, rawPayload string) {
	var event RedisEvent
	if err := json.Unmarshal([]byte(rawPayload), &event); err != nil {
		s.log.Errorw("failed to parse incoming Redis JSON event", "error", err, "rawPayload", rawPayload)
		return
	}

	targetUUID, err := uuid.Parse(event.Payload.TargetUserID)
	if err != nil {
		s.log.Errorw("target_user_id in event payload is not a valid UUID", "error", err, "targetUserID", event.Payload.TargetUserID)
		return
	}

	dbParams := db.CreateNotificationParams{
		UserID: event.Payload.TargetUserID,
		Type:   event.EventType,
		Title:  event.Payload.Title,
		Body:   event.Payload.Message,
	}

	dbNotif, err := s.notificationRepo.CreateNotification(ctx, dbParams)
	if err != nil {
		s.log.Errorw("failed to save notification record to database", "error", err, "targetUserID", event.Payload.TargetUserID)
	} else {
		s.log.Infow("notification saved to database history", "notificationID", dbNotif.ID)
	}

	s.hub.SendNotification(targetUUID, ws.NotificationPayload{
		Type:    event.EventType,
		Message: event.Payload.Message,
		Data:    dbNotif,
	})

	recipientUser, err := s.grpcClient.GetUserDetails(ctx, event.Payload.TargetUserID)
	if err != nil {
		s.log.Errorw("failed to fetch user details from job-service via gRPC", "error", err, "targetUserID", event.Payload.TargetUserID)
		return
	}

	emailParams := db.CreateEmailJobParams{
		ToEmail: recipientUser.Email,
		Subject: event.Payload.Title,
		Body:    event.Payload.Message,
	}

	emailJob, err := s.emailJobRepo.CreateEmailJob(ctx, emailParams)
	if err != nil {
		s.log.Errorw("failed to enqueue email job in database", "error", err, "toEmail", recipientUser.Email)
	} else {
		s.log.Infow("email job enqueued successfully", "emailJobID", emailJob.ID, "toEmail", recipientUser.Email)
	}
}
