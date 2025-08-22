package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"notification/internal/application/template/dtos"
	"notification/internal/application/template/usecases"
	"notification/internal/domain/channel"
	"notification/internal/domain/shared"
	"notification/internal/domain/template"
	"notification/internal/infrastructure/repository"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// TemplateHandlerTestSuite is the test suite for the TemplateNATSHandler
type TemplateHandlerTestSuite struct {
	ChannelHandlerTestSuite // Embed the channel suite to reuse its setup
	templateHandler         *TemplateNATSHandler
	templateRepo            template.TemplateRepository
}

// TestTemplateHandlerTestSuite runs the entire test suite
func TestTemplateHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(TemplateHandlerTestSuite))
}

// SetupSuite runs once before the entire test suite
func (suite *TemplateHandlerTestSuite) SetupSuite() {
	suite.ChannelHandlerTestSuite.SetupSuite() // Run the parent suite's setup

	suite.templateRepo = repository.NewTemplateRepositoryImpl(suite.db)

	// Instantiate the TemplateNATSHandler
	createUseCase := usecases.NewCreateTemplateUseCase(suite.templateRepo)
	getUseCase := usecases.NewGetTemplateUseCase(suite.templateRepo)
	listUseCase := usecases.NewListTemplatesUseCase(suite.templateRepo)
	updateUseCase := usecases.NewUpdateTemplateUseCase(suite.templateRepo, suite.channelRepo, suite.appConfig)
	deleteUseCase := usecases.NewDeleteTemplateUseCase(suite.templateRepo, suite.channelRepo, suite.appConfig)

	handler := NewTemplateNATSHandler(
		createUseCase,
		getUseCase,
		listUseCase,
		updateUseCase,
		deleteUseCase,
		suite.natsConn,
	)
	err := handler.RegisterHandlers()
	suite.Require().NoError(err)
	suite.templateHandler = handler
}

func (suite *TemplateHandlerTestSuite) TestCreateTemplate() {
	createReq := dtos.CreateTemplateRequest{
		Name:        "Test_Template_Create",
		ChannelType: shared.ChannelTypeEmail,
		Subject:     "Test Subject",
		Content:     "Hello {Name}",
		Tags:        []string{"test", "template"},
	}

	reqData, err := json.Marshal(NATSRequest{ReqSeqId: uuid.NewString(), Data: createReq})
	suite.Require().NoError(err)

	msg, err := suite.natsConn.Request("eco1j.infra.eventcenter.template.create", reqData, 5*time.Second)
	suite.Require().NoError(err)

	var natsResp NATSResponse
	err = json.Unmarshal(msg.Data, &natsResp)
	suite.Require().NoError(err)
	suite.True(natsResp.Success, "NATS response should be successful. Error: %v", natsResp.Error)

	var respData dtos.TemplateResponse
	respBytes, _ := json.Marshal(natsResp.Data)
	json.Unmarshal(respBytes, &respData)
	suite.Equal(createReq.Name, respData.Name)

	templateID, err := template.NewTemplateIDFromString(respData.ID)
	suite.Require().NoError(err)
	_, err = suite.templateRepo.FindByID(context.Background(), templateID)
	suite.Require().NoError(err)
}

func (suite *TemplateHandlerTestSuite) TestUpdateTemplateSyncsLegacyChannel() {
	// 1. Create a template
	tmplName, err := template.NewTemplateName("Original_Template")
	suite.Require().NoError(err)
	tmplSub, err := template.NewSubject("Original Subject")
	suite.Require().NoError(err)
	tmplCont, err := template.NewTemplateContent("Original Content")
	suite.Require().NoError(err)
	tmpl, err := template.NewTemplate(tmplName, nil, shared.ChannelTypeEmail, tmplSub, tmplCont, nil)
	suite.Require().NoError(err)
	err = suite.templateRepo.Save(context.Background(), tmpl)
	suite.Require().NoError(err)

	// 2. Create a channel linked to the template
	chanName, err := channel.NewChannelName("Channel_To_Sync")
	suite.Require().NoError(err)
	chn, err := channel.NewChannel(chanName, nil, true, shared.ChannelTypeEmail, tmpl.ID(), &shared.CommonSettings{Timeout: 10}, channel.NewChannelConfig(map[string]interface{}{"host": "a", "port": 1, "username": "a", "password": "a", "senderEmail": "a", "secure": true, "method": "ssl"}), nil, nil)
	suite.Require().NoError(err)
	err = suite.channelRepo.Save(context.Background(), chn)
	suite.Require().NoError(err)

	// 3. Update the template
	updatedName := "Updated_Template_Name"
	updateMap := map[string]interface{}{
		"templateId": tmpl.ID().String(),
		"name":       updatedName,
	}
	reqData, err := json.Marshal(NATSRequest{ReqSeqId: uuid.NewString(), Data: updateMap})
	suite.Require().NoError(err)

	msg, err := suite.natsConn.Request("eco1j.infra.eventcenter.template.update", reqData, 5*time.Second)
	suite.Require().NoError(err)

	var natsResp NATSResponse
	err = json.Unmarshal(msg.Data, &natsResp)
	suite.Require().NoError(err)
	suite.True(natsResp.Success, "NATS response should be successful. Error: %v", natsResp.Error)

	// 4. Verify legacy API was called to sync the channel
	select {
	case received := <-suite.receivedLegacy:
		suite.Equal(http.MethodPut, received.Method)
		suite.Equal(fmt.Sprintf("/Groups/%s", chn.ID().String()), received.Path)
	case <-time.After(2 * time.Second):
		suite.Fail("Did not receive request on legacy API mock to sync channel")
	}
}
