package usecases

// LegacyChannelRequest defines the request payload for the legacy system.
type LegacyChannelRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	LevelName   string         `json:"levelName"`
	Config      LegacyConfig   `json:"config"`
	SendList    []SendListItem `json:"sendList"`
}

// LegacyConfig defines the config for the legacy system.
type LegacyConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Secure       bool   `json:"secure"`
	Method       string `json:"method"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	SenderEmail  string `json:"senderEmail"`
	EmailSubject string `json:"emailSubject"`
	Template     string `json:"template"`
}

// SendListItem defines a recipient for the legacy system.
type SendListItem struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	RecipientType string `json:"recipientType"`
	Target        string `json:"target"`
}