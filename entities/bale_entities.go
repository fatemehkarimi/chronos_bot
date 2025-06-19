package entities

type User struct {
	Id           int     `json:"id"`
	IsBot        bool    `json:"is_bot"`
	FirstName    string  `json:"first_name"`
	LastName     *string `json:"last_name,omitempty"`
	UserName     *string `json:"username,omitempty"`
	LanguageCode *string `json:"language_code,omitempty"`
}

type ChatPhoto struct {
	SmallFileId       string `json:"small_file_id"`
	SmallFileUniqueId string `json:"small_file_unique_id"`
	BigFileId         string `json:"big_file_id"`
	BigFileUniqueId   string `json:"big_file_unique_id"`
}

type Chat struct {
	Id        int64      `json:"id"`
	Type      string     `json:"type"`
	Title     *string    `json:"title,omitempty"`
	Username  *string    `json:"username,omitempty"`
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	Photo     *ChatPhoto `json:"photo,omitempty"`
}

type PhotoSize struct {
	FileId       string `json:"file_id"`
	FileUniqueId string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     *int64 `json:"file_size,omitempty"`
}

type Animation struct {
	FileId       string     `json:"file_id"`
	FileUniqueId string     `json:"file_unique_id"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	Duration     int        `json:"duration"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
	FileName     *string    `json:"file_name,omitempty"`
	MimeType     *string    `json:"mime_type,omitempty"`
	FileSize     *int64     `json:"file_size,omitempty"`
}

type Audio struct {
	FileID       string  `json:"file_id"`
	FileUniqueID string  `json:"file_unique_id"`
	Duration     int     `json:"duration"`
	Title        *string `json:"title,omitempty"`
	FileName     *string `json:"file_name,omitempty"`
	MimeType     *string `json:"mime_type,omitempty"`
	FileSize     *int64  `json:"file_size,omitempty"`
}

type Document struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
	FileName     *string    `json:"file_name,omitempty"`
	MimeType     *string    `json:"mime_type,omitempty"`
	FileSize     *int64     `json:"file_size,omitempty"`
}

type Sticker struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Type         string `json:"type"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     *int64 `json:"file_size,omitempty"`
}

type Video struct {
	FileID       string  `json:"file_id"`
	FileUniqueID string  `json:"file_unique_id"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	Duration     int     `json:"duration"`
	FileName     *string `json:"file_name,omitempty"`
	MimeType     *string `json:"mime_type,omitempty"`
	FileSize     *int64  `json:"file_size,omitempty"`
}

type Voice struct {
	FileId       string `json:"file_id"`
	FileUniqueId string `json:"file_unique_id"`
}

type Contact struct {
	PhoneNumber string  `json:"phone_number"`
	FirstName   string  `json:"first_name"`
	LastName    *string `json:"last_name,omitempty"`
	UserId      *int32  `json:"user_id,omitempty"`
}

type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type Invoice struct{}

type SuccessfulPayment struct {
	Currency                string `json:"currency"`
	TotalAmount             int    `json:"total_amount"`
	InvoicePayload          string `json:"invoice_payload"`
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
	ProviderPaymentChargeID string `json:"provider_payment_charge_id"`
}

type WebAppData struct {
	Data string `json:"data"`
}

type WebAppInfo struct {
	URL string `json:"url"`
}

type CopyTextButton struct {
	Text string `json:"text"`
}

type ReplyMarkup interface {
	isReplyMarkup()
}

type KeyboardButton struct {
	Text            string      `json:"text"`
	RequestContact  bool        `json:"request_contact,omitempty"`
	RequestLocation bool        `json:"request_location,omitempty"`
	WebApp          *WebAppInfo `json:"web_app,omitempty"`
}

type ReplyKeyboardMarkup struct {
	Keyboard []KeyboardButton `json:"keyboard"`
}

func (i ReplyKeyboardMarkup) isReplyMarkup() {}

type ReplyKeyboardRemove struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
}

func (r ReplyKeyboardRemove) isReplyMarkup() {}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

func (i InlineKeyboardMarkup) isReplyMarkup() {}

type InlineKeyboardButton struct {
	Text         string          `json:"text"`
	URL          *string         `json:"url,omitempty"`
	CallbackData *string         `json:"callback_data,omitempty"`
	WebApp       *WebAppInfo     `json:"web_app,omitempty"`
	CopyText     *CopyTextButton `json:"copy_text,omitempty"`
}

func (i InlineKeyboardButton) isReplyMarkup() {}

type CallbackQuery struct {
	Id      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    *string  `json:"data,omitempty"`
}

type PreCheckoutQuery struct {
	Id             string `json:"id"`
	From           User   `json:"from"`
	Currency       string `json:"currency"`
	TotalAmount    int    `json:"total_amount"`
	InvoicePayload string `json:"invoice_payload"`
}

type Message struct {
	MessageId            int                   `json:"message_id"`
	From                 User                  `json:"from"`
	Date                 int                   `json:"date"`
	Chat                 Chat                  `json:"chat"`
	ForwardFrom          *User                 `json:"forward_from,omitempty"`
	ForwardFromChat      *Chat                 `json:"forward_from_chat,omitempty"`
	ForwardFromMessageId *int                  `json:"forward_from_message_id,omitempty"`
	ForwardDate          *int                  `json:"forward_date,omitempty"`
	ReplyToMessage       *Message              `json:"reply_to_message,omitempty"`
	EditDate             *int                  `json:"edit_date,omitempty"`
	Text                 *string               `json:"text,omitempty"`
	Animation            *Animation            `json:"animation,omitempty"`
	Audio                *Audio                `json:"audio,omitempty"`
	Document             *Document             `json:"document,omitempty"`
	Photo                []PhotoSize           `json:"photo,omitempty"`
	Sticker              *Sticker              `json:"sticker,omitempty"`
	Video                *Video                `json:"video,omitempty"`
	Voice                *Voice                `json:"voice,omitempty"`
	Caption              *string               `json:"caption,omitempty"`
	Contact              *Contact              `json:"contact,omitempty"`
	Location             *Location             `json:"location,omitempty"`
	NewChatMembers       []User                `json:"new_chat_members,omitempty"`
	LeftChatMember       *User                 `json:"left_chat_member,omitempty"`
	Invoice              *Invoice              `json:"invoice,omitempty"`
	SuccessfulPayment    *SuccessfulPayment    `json:"successful_payment,omitempty"`
	WebAppData           *WebAppData           `json:"web_app_data,omitempty"`
	ReplyMarkup          *InlineKeyboardButton `json:"reply_markup,omitempty"`
}

type Update struct {
	UpdateId         int               `json:"update_id"`
	Message          *Message          `json:"message,omitempty"`
	EditedMessage    *Message          `json:"edited_message,omitempty"`
	CallbackQuery    *CallbackQuery    `json:"callback_query,omitempty"`
	PreCheckoutQuery *PreCheckoutQuery `json:"pre_checkout_query,omitempty"`
}
