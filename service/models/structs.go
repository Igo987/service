package models

// Структура для парсинга названия и кода страны by alpha2 (https://github.com/atnmorrison/country-locale-map/blob/master/countries.json)
type Country struct {
	Code    string `json:"name"`
	Country string `json:"alpha2"`
}

type SMSData struct { // СТРУКТУРА SMS она же и MMS
	Country      string `json:"country"`
	Provider     string `json:"provider"`
	Bandwidth    string `json:"bandwidth"`
	ResponseTime string `json:"response_time"`
}

type MMSData struct { // СТРУКТУРА SMS она же и MMS
	Country      string `json:"country"`
	Provider     string `json:"provider"`
	Bandwidth    string `json:"bandwidth"`
	ResponseTime string `json:"response_time"`
}

type VoiceCall struct { // СТРУКТУРА VoiceCall
	Country                 string  `json:"country"`
	Load                    int     `json:"load"`
	ResponseTime            int     `json:"response_time"`
	Provider                string  `json:"provider"`
	Connection_Stability    float32 `json:"coonection_stability"`
	TTFB                    int     `json:"ttfb"`
	Purity_Of_Communication int     `json:"purity_of_communication"`
	Median                  int     `json:"median"`
}

type EmailData struct {
	Country      string `json:"country"`
	Provider     string `json:"provider"`
	DeliveryTime int    `json:"delivery_time"`
}

type SupportData struct {
	Topic         string `json:"topic"`
	ActiveTickets int    `json:"active_tickets"`
}

type BillingData struct {
	CreateCustomer bool `json:"create_customer"`
	Purchase       bool `json:"purchase"`
	Payout         bool `json:"payout"`
	Recurring      bool `json:"recurring"`
	FraudControl   bool `json:"fraud_control"`
	CheckoutPage   bool `json:"checkout_page"`
}

type IncidentData struct {
	Topic  string `json:"topic"`
	Status string `json:"status"`
}

type SortSMSData [][]SMSData
type SortMMSData [][]MMSData
type VoiceCallData []VoiceCall

type ResultSetT struct {
	SMS       [][]SMSData              `json:"sms"`
	MMS       [][]MMSData              `json:"mms"`
	VoiceCall VoiceCallData            `json:"voice_call"`
	Email     map[string][][]EmailData `json:"email"`
	Billing   BillingData              `json:"billing"`
	Support   []int                    `json:"support"`
	Incidents []IncidentData           `json:"incident"`
}

type ResultT struct {
	Status bool        `json:"status"` // True, если все этапы сбора данных прошли успешно, False во всех остальных случаях
	Data   *ResultSetT `json:"data"`   // Заполнен, если все этапы сбора  данных прошли успешно, nil во всех остальных случаях
	Error  string      `json:"error"`  // Пустая строка, если все этапы сбора данных прошли успешно, в случае ошибки заполнено текстом ошибки
}

// type ResultSetT struct {
// 	SMS       [][]SMSData              `json:"sms"`
// 	MMS       [][]MMSData              `json:"mms"`
// 	VoiceCall VoiceCallData            `json:"voice_call"`
// 	Email     map[string][][]EmailData `json:"email"`
// 	Billing   BillingData              `json:"billing"`
// 	Support   []int                    `json:"support"`
// 	Incidents []IncidentData           `json:"incident"`
// }

type ResultEmail map[string][][]EmailData
