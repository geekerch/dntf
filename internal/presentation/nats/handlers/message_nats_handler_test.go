package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"notification/internal/application/message/dtos"
	"notification/internal/application/message/usecases"
	"notification/internal/domain/channel"
	"notification/internal/domain/services"
	"notification/internal/domain/shared"
	"notification/internal/domain/template"
	"notification/internal/infrastructure/external"
	"notification/internal/infrastructure/repository"
	"notification/pkg/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// MockSMTP captures sent emails
type MockSMTP struct {
	mu         sync.Mutex
	listener   *net.TCPListener
	addr       string
	sentEmails []CapturedEmail
}

// CapturedEmail represents a sent email
type CapturedEmail struct {
	From    string
	To      []string
	Subject string
	Body    string
}

// NewMockSMTP creates a new MockSMTP server
func NewMockSMTP() *MockSMTP {
	return &MockSMTP{
		sentEmails: make([]CapturedEmail, 0),
	}
}

// Start starts the mock SMTP server
func (m *MockSMTP) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0") // Listen on random port
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	m.listener = listener
	m.addr = listener.Addr().String()

	go func() {
		for {
			conn, err := m.listener.Accept()
			if err != nil {
				// Listener closed
				return
			}
			go m.handleConn(conn)
		}
	}()
	return nil
}

// Addr returns the address the mock SMTP server is listening on
func (m *MockSMTP) Addr() string {
	return m.addr
}

// Stop stops the mock SMTP server
func (m *MockSMTP) Stop() error {
	if m.listener != nil {
		return m.listener.Close()
	}
	return nil
}

// Clear clears captured emails
func (m *MockSMTP) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentEmails = make([]CapturedEmail, 0)
}

// Emails returns captured emails
func (m *MockSMTP) Emails() []CapturedEmail {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sentEmails
}

func (m *MockSMTP) handleConn(conn net.Conn) {
	defer conn.Close()

	tc := textproto.NewConn(conn)
	defer tc.Close()

	// SMTP handshake
	tc.PrintfLine("220 %s ESMTP Mock", m.addr)
	tc.ReadLine()
	tc.PrintfLine("250 %s", m.addr)
	tc.ReadLine()

	// MAIL FROM
	line, _ := tc.ReadLine()
	from := strings.TrimPrefix(line, "MAIL FROM:<")
	from = strings.TrimSuffix(from, ">")
	from = strings.TrimSpace(from)
	tc.PrintfLine("250 Ok")

	// RCPT TO
	to := []string{}
	for {
		line, _ = tc.ReadLine()
		if strings.HasPrefix(line, "RCPT TO:<") {
			recipient := strings.TrimPrefix(line, "RCPT TO:<")
			recipient = strings.TrimSuffix(recipient, ">")
			recipient = strings.TrimSpace(recipient)
			to = append(to, recipient)
			tc.PrintfLine("250 Ok")
		} else if line == "DATA" {
			break
		} else {
			// Handle other commands or errors
			break
		}
	}

	// DATA
	tc.PrintfLine("354 End data with <CR><LF>.<CR><LF>")
	data := tc.DotReader()
	bodyBytes, _ := io.ReadAll(data)
	body := string(bodyBytes)
	tc.PrintfLine("250 Ok")
	tc.ReadLine()

	// Extract subject (more robust parsing)
	subject := ""
	lines := strings.Split(body, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Subject: ") {
			subject = strings.TrimPrefix(line, "Subject: ")
			break
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentEmails = append(m.sentEmails, CapturedEmail{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
	})

	// QUIT
	tc.PrintfLine("221 Bye")
}

// MockNotificationService implements external.NotificationService for testing
type MockNotificationService struct {
	mu           sync.Mutex
	sentRequests []*external.SendRequest
}

// NewMockNotificationService creates a new MockNotificationService
func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		sentRequests: make([]*external.SendRequest, 0),
	}
}

// SendNotification captures the request and returns a successful result
func (m *MockNotificationService) SendNotification(ctx context.Context, requests []*external.SendRequest) ([]*external.SendResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentRequests = append(m.sentRequests, requests...)
	results := make([]*external.SendResult, len(requests))
	for i := range requests {
		results[i] = &external.SendResult{Success: true, Message: "Sent successfully", SentAt: time.Now().UnixMilli()}
	}
	return results, nil
}

// SendSingleNotification captures the request and returns a successful result
func (m *MockNotificationService) SendSingleNotification(ctx context.Context, request *services.SendRequest) *services.SendResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Convert to external.SendRequest for storage
	extRequest := &external.SendRequest{
		Channel:   request.Channel,
		Content:   request.Content,
		Variables: request.Variables,
	}
	m.sentRequests = append(m.sentRequests, extRequest)
	return &services.SendResult{Success: true, Message: "Sent successfully", SentAt: time.Now().UnixMilli()}
}

// ValidateChannel always returns nil for testing
func (m *MockNotificationService) ValidateChannel(ch *channel.Channel) error {
	return nil
}

// Clear clears captured requests
func (m *MockNotificationService) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentRequests = make([]*external.SendRequest, 0)
}

// SentRequests returns captured requests
func (m *MockNotificationService) SentRequests() []*external.SendRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sentRequests
}

// MessageHandlerTestSuite is the test suite for the MessageNATSHandler
type MessageHandlerTestSuite struct {
	ChannelHandlerTestSuite
	messageHandler          *MessageNATSHandler
	mockSMTPServer          *MockSMTP
	mockNotificationService *MockNotificationService
	templateRepo            template.TemplateRepository
}

// TestMessageHandlerTestSuite runs the entire test suite
func TestMessageHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MessageHandlerTestSuite))
}

// SetupSuite runs once before the entire test suite
func (suite *MessageHandlerTestSuite) SetupSuite() {
	suite.ChannelHandlerTestSuite.SetupSuite()

	// Start mock SMTP server
	server := NewMockSMTP()
	suite.Require().NoError(server.Start())
	suite.mockSMTPServer = server

	// Instantiate the MessageNATSHandler
	messagingRepo := repository.NewMessageRepositoryImpl(suite.db)
	channelRepo := repository.NewChannelRepositoryImpl(suite.db)
	templateRepo := repository.NewTemplateRepositoryImpl(suite.db)
	suite.templateRepo = templateRepo
	mockNotificationService := NewMockNotificationService()
	suite.mockNotificationService = mockNotificationService

	// Create enhanced message sender
	renderer := services.NewDefaultTemplateRenderer()
	appLogger := logger.GetGlobalLogger()
	enhancedMessageSender := services.NewEnhancedMessageSender(channelRepo, templateRepo, messagingRepo, renderer, mockNotificationService, appLogger)
	
	sendUseCase := usecases.NewSendMessageUseCase(messagingRepo, channelRepo, templateRepo, enhancedMessageSender, suite.appConfig)
	getUseCase := usecases.NewGetMessageUseCase(messagingRepo)
	listUseCase := usecases.NewListMessagesUseCase(messagingRepo)

	handler := NewMessageNATSHandler(
		sendUseCase,
		getUseCase,
		listUseCase,
		suite.natsConn,
	)
	err := handler.RegisterHandlers()
	suite.Require().NoError(err)
	suite.messageHandler = handler
}

// TearDownSuite runs once after the suite
func (suite *MessageHandlerTestSuite) TearDownSuite() {
	suite.mockSMTPServer.Stop()
	suite.ChannelHandlerTestSuite.TearDownSuite()
}

// SetupTest runs before each test
func (suite *MessageHandlerTestSuite) SetupTest() {
	suite.ChannelHandlerTestSuite.SetupTest()
	suite.mockSMTPServer.Clear()
	suite.mockNotificationService.Clear()
}

func (suite *MessageHandlerTestSuite) TestSendMessage_WithTemplate() {
	// 1. Create a template
	tmplName, _ := template.NewTemplateName("Message_Template")
	tmplSub, _ := template.NewSubject("Hello {{.name}} from template")
	tmplCont, _ := template.NewTemplateContent("Your code is {{.code}}")
	tmpl, _ := template.NewTemplate(tmplName, nil, shared.ChannelTypeEmail, tmplSub, tmplCont, nil)
	suite.templateRepo.Save(context.Background(), tmpl)

	// 2. Create a channel pointing to the mock SMTP server
	host, portStr, _ := strings.Cut(suite.mockSMTPServer.Addr(), ":")
	port, _ := strconv.Atoi(portStr)
	chanName, _ := channel.NewChannelName("SMTP_Channel")
	chn, _ := channel.NewChannel(chanName, nil, true, shared.ChannelTypeEmail, tmpl.ID(), &shared.CommonSettings{Timeout: 10}, channel.NewChannelConfig(map[string]interface{}{"host": host, "port": float64(port), "username": "testuser", "password": "testpass", "senderEmail": "from@test.com"}), nil, nil)
	suite.channelRepo.Save(context.Background(), chn)

	// 3. Send message
	sendReq := dtos.SendMessageRequest{
		ChannelIDs: []string{chn.ID().String()},
		Recipients: []map[string]interface{}{{"name": "John Doe", "target": "to@test.com", "type": "to"}},
		Variables:  map[string]interface{}{"name": "John", "code": "12345"},
		TemplateID: tmpl.ID().String(),
	}
	reqData, _ := json.Marshal(NATSRequest{ReqSeqId: uuid.NewString(), Data: sendReq})

	msg, err := suite.natsConn.Request("eco1j.infra.eventcenter.message.send", reqData, 5*time.Second)
	suite.Require().NoError(err)

	var natsResp NATSResponse
	err = json.Unmarshal(msg.Data, &natsResp)
	suite.Require().NoError(err)
	suite.True(natsResp.Success, "NATS response should be successful. Error: %v", natsResp.Error)

	// 4. Verify email was sent via mock NotificationService
	suite.Require().Len(suite.mockNotificationService.SentRequests(), 1)
	sentNotificationRequest := suite.mockNotificationService.SentRequests()[0]

	suite.Equal(chn.ID().String(), sentNotificationRequest.Channel.ID().String())
	suite.Equal("Hello John from template", sentNotificationRequest.Content.Subject)
	suite.Contains(sentNotificationRequest.Content.Content, "Your code is 12345")
	suite.Equal("John", sentNotificationRequest.Variables["name"])
	suite.Equal("12345", sentNotificationRequest.Variables["code"])

	// 5. Verify email was sent via mock SMTP server (actual SMTP interaction)
	suite.Require().Len(suite.mockSMTPServer.Emails(), 1)
	sentEmail := suite.mockSMTPServer.Emails()[0]

	suite.Contains(sentEmail.To, "to@test.com")
	suite.Equal("Hello {{.name}} from template", sentEmail.Subject) // Subject should be raw template subject
	suite.Contains(sentEmail.Body, "Your code is {{.code}}")        // Body should be raw template content
}
