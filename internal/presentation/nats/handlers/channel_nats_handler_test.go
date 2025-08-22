package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"notification/internal/application/channel/usecases"
	"notification/internal/domain/channel"
	"notification/internal/domain/services"
	"notification/internal/domain/shared"
	"notification/internal/domain/shared/channel_types"
	"notification/internal/infrastructure/models"
	"notification/internal/infrastructure/repository"
	"notification/pkg/config"
	"notification/pkg/database"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"notification/internal/application/channel/dtos"
)

// MockLegacyAPIRequest stores information about a request received by the mock legacy API
type MockLegacyAPIRequest struct {
	Method string
	Path   string
	Body   []byte
}

// ChannelHandlerTestSuite is the test suite for the ChannelNATSHandler
type ChannelHandlerTestSuite struct {
	suite.Suite
	db             *gorm.DB
	channelRepo    channel.ChannelRepository
	natsServer     *server.Server
	natsConn       *nats.Conn
	handler        *ChannelNATSHandler
	legacyAPI      *httptest.Server
	legacyAPIURL   string
	receivedLegacy chan MockLegacyAPIRequest
	appConfig      *config.Config
}

// newTestPostgresDB creates a new PostgreSQL database connection for testing.
func newTestPostgresDB(t *suite.Suite) *gorm.DB {
	testConfig := &config.DatabaseConfig{
		Type:           "postgres",
		Host:           "localhost",
		Port:           5432,
		User:           "admin",
		Password:       "admin",
		DBName:         "unitest",
		SSLMode:        "disable",
		MigrationsPath: "../../../../migrations",
	}

	db, err := database.NewGormDB(testConfig)
	t.Require().NoError(err, "Failed to create test GORM DB")

	// Run migrations
	err = db.RunMigrations()
	t.Require().NoError(err, "Failed to run migrations")

	return db.DB
}

// SetupSuite runs once before the entire test suite
func (suite *ChannelHandlerTestSuite) SetupSuite() {
	// 0. Register channel types
	channel_types.RegisterDefaultChannelTypes()

	// 1. Setup in-memory NATS server
	opts := &server.Options{Host: "127.0.0.1", Port: -1}
	ns, err := server.NewServer(opts)
	suite.Require().NoError(err)
	go ns.Start()
	if !ns.ReadyForConnections(5 * time.Second) {
		suite.FailNow("NATS server not ready")
	}
	suite.natsServer = ns

	nc, err := nats.Connect(ns.ClientURL())
	suite.Require().NoError(err)
	suite.natsConn = nc

	// 2. Setup PostgreSQL test database
	suite.db = newTestPostgresDB(&suite.Suite)
	suite.channelRepo = repository.NewChannelRepositoryImpl(suite.db)

	// 3. Setup Mock Legacy API Server
	suite.receivedLegacy = make(chan MockLegacyAPIRequest, 10)
	mockLegacyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		bodyCopy := make([]byte, len(body))
		copy(bodyCopy, body)

		suite.receivedLegacy <- MockLegacyAPIRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Body:   bodyCopy,
		}

		w.Header().Set("Content-Type", "application/json")
		var response interface{}
		var statusCode = http.StatusOK

		switch r.Method {
		case http.MethodPost:
			statusCode = http.StatusCreated
			// Return array format for legacy message response
			response = []map[string]interface{}{
				{
					"groupId": uuid.New().String(),
					"result": []map[string]interface{}{
						{"statusCode": 200, "message": "Message sent successfully"},
					},
				},
			}
		case http.MethodPut:
			response = map[string]interface{}{"id": uuid.New().String(), "status": "success"}
		case http.MethodDelete:
			response = map[string]interface{}{"message": "Group deleted"}
		default: // GET
			response = map[string]interface{}{"groups": []string{}, "total": 0}
		}
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}))
	suite.legacyAPI = mockLegacyServer
	suite.legacyAPIURL = mockLegacyServer.URL

	// 4. Dependency Injection
	suite.appConfig = &config.Config{
		LegacySystem: config.LegacySystemConfig{
			URL: suite.legacyAPIURL,
		},
	}

	templateRepo := repository.NewTemplateRepositoryImpl(suite.db)
	validator := services.NewChannelValidator(suite.channelRepo, templateRepo)

	createUseCase := usecases.NewCreateChannelUseCase(suite.channelRepo, templateRepo, validator, suite.appConfig)
	getUseCase := usecases.NewGetChannelUseCase(suite.channelRepo)
	listUseCase := usecases.NewListChannelsUseCase(suite.channelRepo)
	updateUseCase := usecases.NewUpdateChannelUseCase(suite.channelRepo, templateRepo, validator, suite.appConfig)
	deleteUseCase := usecases.NewDeleteChannelUseCase(suite.channelRepo, validator, suite.appConfig)

	// 5. Instantiate Handler
	suite.handler = NewChannelNATSHandler(
		createUseCase,
		getUseCase,
		listUseCase,
		updateUseCase,
		deleteUseCase,
		suite.natsConn,
	)
	err = suite.handler.RegisterHandlers()
	suite.Require().NoError(err)
}

// TearDownSuite runs once after the entire test suite
func (suite *ChannelHandlerTestSuite) TearDownSuite() {
	suite.natsConn.Close()
	suite.natsServer.Shutdown()
	// Clean up the database
	suite.db.Exec("DROP TABLE IF EXISTS channels, templates, messages, message_results, schema_migrations CASCADE")
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
	suite.legacyAPI.Close()
}

// SetupTest runs before each test
func (suite *ChannelHandlerTestSuite) SetupTest() {
	// Clean data from tables before each test
	suite.db.Exec("DELETE FROM channels")
	suite.db.Exec("DELETE FROM templates")
	suite.db.Exec("DELETE FROM messages")
	suite.db.Exec("DELETE FROM message_results")
	// Clear the received requests channel
	for len(suite.receivedLegacy) > 0 {
		<-suite.receivedLegacy
	}
}

// TestChannelHandlerTestSuite runs the entire test suite
func TestChannelHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ChannelHandlerTestSuite))
}

func (suite *ChannelHandlerTestSuite) TestCreateChannel() {
	createReq := dtos.CreateChannelRequest{
		ChannelName: "Test_Email_Channel",
		Description: "A channel for testing",
		Enabled:     true,
		ChannelType: "email",
		CommonSettings: dtos.CommonSettingsDTO{
			Timeout:       10,
			RetryAttempts: 3,
			RetryDelay:    5,
		},
		Config: map[string]interface{}{
			"host":        "smtp.test.com",
			"port":        float64(587),
			"secure":      true,
			"method":      "tls",
			"username":    "test@test.com",
			"password":    "password",
			"senderEmail": "test@test.com",
		},
		Tags:       []string{"test", "email"},
		Recipients: []dtos.RecipientDTO{{Name: "test", Target: "test@test.com", Type: "to"}},
	}

	reqData, err := json.Marshal(NATSRequest{
		ReqSeqId:  uuid.NewString(),
		Data:      createReq,
		Timestamp: time.Now().UnixMilli(),
	})
	suite.Require().NoError(err)

	// Send NATS request
	msg, err := suite.natsConn.Request("eco1j.infra.eventcenter.channel.create", reqData, 5*time.Second)
	suite.Require().NoError(err)

	// Assert NATS response
	var natsResp NATSResponse
	err = json.Unmarshal(msg.Data, &natsResp)
	suite.Require().NoError(err)
	suite.True(natsResp.Success, "NATS response should be successful. Error: %v", natsResp.Error)
	suite.Require().NotNil(natsResp.Data)

	var respData dtos.ChannelResponse
	respBytes, _ := json.Marshal(natsResp.Data)
	json.Unmarshal(respBytes, &respData)
	suite.Equal(createReq.ChannelName, respData.ChannelName)

	// Assert database record
	channelID, err := channel.NewChannelIDFromString(respData.ChannelID)
	suite.Require().NoError(err)
	_, err = suite.channelRepo.FindByID(context.Background(), channelID)
	suite.Require().NoError(err)

	// Assert legacy API call
	select {
	case received := <-suite.receivedLegacy:
		suite.Equal(http.MethodPost, received.Method)
		suite.Equal("/Groups", received.Path)
	case <-time.After(2 * time.Second):
		suite.Fail("Did not receive request on legacy API mock")
	}
}

func (suite *ChannelHandlerTestSuite) TestUpdateChannel() {
	// 1. Create a channel first
	name, _ := channel.NewChannelName("Initial_Name")
	desc, _ := channel.NewDescription("desc")
	initialChannel, err := channel.NewChannel(
		name,
		desc,
		true,
		shared.ChannelTypeEmail,
		nil, // templateID
		&shared.CommonSettings{Timeout: 10},
		channel.NewChannelConfig(map[string]interface{}{"host": "a", "port": 1, "username": "a", "password": "a", "senderEmail": "a", "secure": true, "method": "ssl"}),
		channel.NewRecipients([]*channel.Recipient{}),
		channel.NewTags([]string{}),
	)
	suite.Require().NoError(err)
	err = suite.channelRepo.Save(context.Background(), initialChannel)
	suite.Require().NoError(err)

	updateReq := dtos.UpdateChannelRequest{
		ChannelID:   initialChannel.ID().String(),
		ChannelName: "Updated_Channel_Name",
		Description: "Updated description",
		Enabled:     false,
		ChannelType: "email", // Must provide channel type on update for validation
		CommonSettings: dtos.CommonSettingsDTO{
			Timeout:       20,
			RetryAttempts: 5,
			RetryDelay:    10,
		},
		Config: map[string]interface{}{"host": "b", "port": 2, "username": "b", "password": "b", "senderEmail": "b", "secure": false, "method": "tls"},
	}

	reqData, err := json.Marshal(NATSRequest{
		ReqSeqId:  uuid.NewString(),
		Data:      updateReq,
		Timestamp: time.Now().UnixMilli(),
	})
	suite.Require().NoError(err)

	// 2. Send NATS update request
	msg, err := suite.natsConn.Request("eco1j.infra.eventcenter.channel.update", reqData, 5*time.Second)
	suite.Require().NoError(err)

	// 3. Assert NATS response
	var natsResp NATSResponse
	err = json.Unmarshal(msg.Data, &natsResp)
	suite.Require().NoError(err)
	suite.True(natsResp.Success, "NATS response should be successful. Error: %v", natsResp.Error)

	// 4. Assert database record
	updatedChannel, err := suite.channelRepo.FindByID(context.Background(), initialChannel.ID())
	suite.Require().NoError(err)
	suite.Equal("Updated_Channel_Name", updatedChannel.Name().String())
	suite.False(updatedChannel.IsEnabled())

	// 5. Assert legacy API call
	select {
	case received := <-suite.receivedLegacy:
		suite.Equal(http.MethodPut, received.Method)
		suite.Equal(fmt.Sprintf("/Groups/%s", initialChannel.ID().String()), received.Path)
	case <-time.After(2 * time.Second):
		suite.Fail("Did not receive request on legacy API mock")
	}
}

func (suite *ChannelHandlerTestSuite) TestDeleteChannel() {
	// 1. Create a channel first
	name, _ := channel.NewChannelName("To_Be_Deleted")
	desc, _ := channel.NewDescription("desc")
	initialChannel, err := channel.NewChannel(
		name,
		desc,
		true,
		shared.ChannelTypeSMS,
		nil, // templateID
		&shared.CommonSettings{Timeout: 10},
		channel.NewChannelConfig(map[string]interface{}{"provider": "a", "apiKey": "a", "apiSecret": "a"}),
		channel.NewRecipients([]*channel.Recipient{}),
		channel.NewTags([]string{}),
	)
	suite.Require().NoError(err)
	err = suite.channelRepo.Save(context.Background(), initialChannel)
	suite.Require().NoError(err)

	deleteReq := map[string]interface{}{"channelId": initialChannel.ID().String()}
	reqData, err := json.Marshal(NATSRequest{
		ReqSeqId:  uuid.NewString(),
		Data:      deleteReq,
		Timestamp: time.Now().UnixMilli(),
	})
	suite.Require().NoError(err)

	// 2. Send NATS delete request
	msg, err := suite.natsConn.Request("eco1j.infra.eventcenter.channel.delete", reqData, 5*time.Second)
	suite.Require().NoError(err)

	// 3. Assert NATS response
	var natsResp NATSResponse
	err = json.Unmarshal(msg.Data, &natsResp)
	suite.Require().NoError(err)
	suite.True(natsResp.Success, "NATS response should be successful. Error: %v", natsResp.Error)

	// 4. Assert database record is soft-deleted
	var channelModel models.ChannelModel
	err = suite.db.Unscoped().First(&channelModel, "id = ?", initialChannel.ID().String()).Error
	suite.Require().NoError(err)
	suite.NotNil(channelModel.DeletedAt)

	// 5. Assert legacy API call
	select {
	case received := <-suite.receivedLegacy:
		suite.Equal(http.MethodDelete, received.Method)
		suite.Equal("/Groups", received.Path)
		var body []string
		err := json.Unmarshal(received.Body, &body)
		suite.Require().NoError(err)
		suite.Require().Len(body, 1)
		suite.Equal(initialChannel.ID().String(), body[0])
	case <-time.After(2 * time.Second):
		suite.Fail("Did not receive request on legacy API mock")
	}
}
